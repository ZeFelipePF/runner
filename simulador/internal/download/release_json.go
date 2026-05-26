package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// URLReleaseJsonPadrao e o endpoint default para o release.json conforme a
// estrategia sugerida em US-03 da especificacao.
// Sobrescritivel via env HUBSAUDE_RELEASE_JSON_URL.
const URLReleaseJsonPadrao = "https://raw.githubusercontent.com/hubsaude/simulador/main/release.json"

// ReleaseJson e o formato esperado do release.json conforme a especificacao:
//
//	{
//	  "jar": { "url": "...", "version": "1.2.0", "sha256": "..." },
//	  "jre": { "windows_x64": "...", "linux_x64": "...", "mac_x64": "..." }
//	}
//
// Os campos "sha256" e "jre" sao opcionais.
type ReleaseJson struct {
	Jar struct {
		URL     string `json:"url"`
		Version string `json:"version"`
		SHA256  string `json:"sha256,omitempty"`
	} `json:"jar"`
	JRE map[string]string `json:"jre,omitempty"`
}

// BuscarReleaseJson tenta carregar e parsear o release.json em urlOrPath.
// Aceita http(s):// e file://.
func BuscarReleaseJson(ctx context.Context, client *http.Client, urlOrPath string) (*ReleaseJson, error) {
	if urlOrPath == "" {
		urlOrPath = URLReleaseJsonOuPadrao()
	}
	var data []byte
	if strings.HasPrefix(urlOrPath, "file://") {
		path := strings.TrimPrefix(urlOrPath, "file://")
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("ao ler release.json local: %w", err)
		}
		data = b
	} else {
		if _, err := url.Parse(urlOrPath); err != nil {
			return nil, fmt.Errorf("URL release.json invalida: %w", err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlOrPath, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ao buscar release.json: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("release.json retornou HTTP %d", resp.StatusCode)
		}
		b, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if err != nil {
			return nil, err
		}
		data = b
	}
	var rj ReleaseJson
	if err := json.Unmarshal(data, &rj); err != nil {
		return nil, fmt.Errorf("ao decodificar release.json: %w", err)
	}
	if rj.Jar.URL == "" {
		return nil, fmt.Errorf("release.json sem jar.url")
	}
	if rj.Jar.Version == "" {
		rj.Jar.Version = "release-json"
	}
	return &rj, nil
}

// URLReleaseJsonOuPadrao retorna a URL configurada via env, ou a default.
func URLReleaseJsonOuPadrao() string {
	if u := strings.TrimSpace(os.Getenv("HUBSAUDE_RELEASE_JSON_URL")); u != "" {
		return u
	}
	return URLReleaseJsonPadrao
}
