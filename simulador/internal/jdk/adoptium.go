package jdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hubsaude/runner/simulador/internal/state"
)

// BaseURLAdoptium e o endpoint padrao da API Adoptium v3. Sobrescritivel via
// env HUBSAUDE_ADOPTIUM_URL para testes.
const BaseURLAdoptium = "https://api.adoptium.net/v3"

type adoptiumAsset struct {
	ReleaseName string `json:"release_name"`
	Binary      struct {
		Package struct {
			Link     string `json:"link"`
			Checksum string `json:"checksum"`
			Name     string `json:"name"`
		} `json:"package"`
	} `json:"binary"`
	Version struct {
		Semver string `json:"semver"`
		Major  int    `json:"major"`
	} `json:"version"`
}

// baixarAdoptium consulta a API Adoptium, baixa o JDK, valida SHA-256, extrai
// para ~/.hubsaude/jdk/<release> e retorna o JDK.
func baixarAdoptium(ctx context.Context, opc Opcoes) (*JDK, error) {
	min := coalesce(opc.VersaoMinima, VersaoMinima)
	base := strings.TrimRight(os.Getenv("HUBSAUDE_ADOPTIUM_URL"), "/")
	if base == "" {
		base = BaseURLAdoptium
	}

	osNome, arch, ext, err := plataformaAdoptium()
	if err != nil {
		return nil, err
	}

	// US-03 da especificacao exige JRE (nao JDK completo) — Adoptium image_type=jre.
	url := fmt.Sprintf("%s/assets/latest/%d/hotspot?architecture=%s&image_type=jre&os=%s&vendor=eclipse",
		base, min, arch, osNome)
	log(opc, "consultando Adoptium: %s", url)
	assets, err := buscarAssets(ctx, url)
	if err != nil {
		return nil, err
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("Adoptium nao retornou nenhum asset para %s/%s", osNome, arch)
	}
	asset := assets[0]
	if asset.Binary.Package.Link == "" {
		return nil, errors.New("Adoptium: asset sem link de download")
	}

	dirBase := opc.HubsaudeDir
	if dirBase == "" {
		dirBase, err = state.EnsureDir()
		if err != nil {
			return nil, err
		}
	}
	raiz := filepath.Join(dirBase, "jdk")
	if err := os.MkdirAll(raiz, 0o755); err != nil {
		return nil, fmt.Errorf("ao criar %s: %w", raiz, err)
	}

	tmp, err := os.CreateTemp(raiz, "download-*."+ext)
	if err != nil {
		return nil, fmt.Errorf("ao criar arquivo temporario: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	log(opc, "baixando %s", asset.Binary.Package.Link)
	if err := baixarComChecksum(ctx, asset.Binary.Package.Link, asset.Binary.Package.Checksum, tmp); err != nil {
		_ = tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}

	releaseDir := filepath.Join(raiz, sanitizar(asset.ReleaseName))
	log(opc, "extraindo em %s", releaseDir)
	if err := os.RemoveAll(releaseDir); err != nil {
		return nil, err
	}
	if err := extrair(tmpPath, releaseDir, ext); err != nil {
		return nil, err
	}

	javaBin := localizarJavaBin(releaseDir)
	if javaBin == "" {
		return nil, fmt.Errorf("java nao encontrado apos extracao em %s", releaseDir)
	}

	// Persistir metadados.
	meta := map[string]any{
		"release":   asset.ReleaseName,
		"versao":    asset.Version.Semver,
		"major":     asset.Version.Major,
		"javaBin":   javaBin,
		"baixadoEm": time.Now().UTC().Format(time.RFC3339),
		"checksum":  asset.Binary.Package.Checksum,
		"origem":    "adoptium",
		"imageType": "jre",
	}
	if data, err := json.MarshalIndent(meta, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(releaseDir, "hubsaude-meta.json"), data, 0o644)
	}

	full, major, err := versaoDoBinario(javaBin)
	if err != nil {
		return nil, err
	}
	return &JDK{JavaBin: javaBin, VersaoFull: full, VersaoMajor: major, Origem: "adoptium"}, nil
}

func buscarAssets(ctx context.Context, urlStr string) ([]adoptiumAsset, error) {
	if _, err := url.Parse(urlStr); err != nil {
		return nil, fmt.Errorf("URL Adoptium invalida: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ao consultar Adoptium: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("Adoptium retornou HTTP %d: %s", resp.StatusCode, body)
	}
	var assets []adoptiumAsset
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, fmt.Errorf("ao decodificar resposta Adoptium: %w", err)
	}
	return assets, nil
}

func baixarComChecksum(ctx context.Context, urlStr, esperado string, w io.Writer) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ao baixar JDK: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("download retornou HTTP %d", resp.StatusCode)
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(w, h), resp.Body); err != nil {
		return fmt.Errorf("ao gravar download: %w", err)
	}
	if esperado != "" {
		obtido := hex.EncodeToString(h.Sum(nil))
		if !strings.EqualFold(obtido, esperado) {
			return fmt.Errorf("checksum SHA-256 nao confere (esperado=%s, obtido=%s)", esperado, obtido)
		}
	}
	return nil
}

func plataformaAdoptium() (osNome, arch, ext string, err error) {
	switch runtime.GOARCH {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", "", "", fmt.Errorf("arquitetura nao suportada: %s", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "linux":
		return "linux", arch, "tar.gz", nil
	case "darwin":
		return "mac", arch, "tar.gz", nil
	case "windows":
		return "windows", arch, "zip", nil
	}
	return "", "", "", fmt.Errorf("sistema operacional nao suportado: %s", runtime.GOOS)
}

func extrair(arquivo, destino, ext string) error {
	if err := os.MkdirAll(destino, 0o755); err != nil {
		return err
	}
	switch ext {
	case "tar.gz":
		return extrairTarGz(arquivo, destino)
	case "zip":
		return extrairZip(arquivo, destino)
	}
	return fmt.Errorf("extensao nao suportada: %s", ext)
}

func extrairTarGz(arquivo, destino string) error {
	f, err := os.Open(arquivo)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// Remove o primeiro componente do caminho (ex: jdk-21.0.3+9/).
		nome := descascarPrefixo(h.Name)
		if nome == "" {
			continue
		}
		alvo := filepath.Join(destino, nome)
		if !strings.HasPrefix(filepath.Clean(alvo), filepath.Clean(destino)) {
			return fmt.Errorf("tar com caminho suspeito: %s", h.Name)
		}
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(alvo, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(alvo), 0o755); err != nil {
				return err
			}
			modo := os.FileMode(h.Mode) & 0o777
			if modo == 0 {
				modo = 0o644
			}
			out, err := os.OpenFile(alvo, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, modo)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			_ = os.MkdirAll(filepath.Dir(alvo), 0o755)
			_ = os.Symlink(h.Linkname, alvo)
		}
	}
}

func extrairZip(arquivo, destino string) error {
	r, err := zip.OpenReader(arquivo)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		nome := descascarPrefixo(f.Name)
		if nome == "" {
			continue
		}
		alvo := filepath.Join(destino, nome)
		if !strings.HasPrefix(filepath.Clean(alvo), filepath.Clean(destino)) {
			return fmt.Errorf("zip com caminho suspeito: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(alvo, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(alvo), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(alvo, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return err
		}
		rc.Close()
		out.Close()
	}
	return nil
}

func descascarPrefixo(nome string) string {
	nome = strings.TrimPrefix(nome, "./")
	idx := strings.IndexAny(nome, "/\\")
	if idx < 0 {
		return ""
	}
	return nome[idx+1:]
}

func localizarJavaBin(raiz string) string {
	candidatos := []string{
		filepath.Join(raiz, "bin", binarioJava()),
		filepath.Join(raiz, "Contents", "Home", "bin", binarioJava()), // macOS .pkg layout
	}
	for _, c := range candidatos {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func sanitizar(nome string) string {
	r := strings.NewReplacer(string(filepath.Separator), "_", " ", "_", ":", "_")
	out := r.Replace(nome)
	if out == "" {
		return "temurin-21"
	}
	return out
}
