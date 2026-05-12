// Package download obtem o simulador.jar via GitHub Releases (ou URL custom),
// com cache local em ~/.hubsaude/simulador/ e verificacao SHA-256.
package download

import (
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
	"strings"
	"time"

	"github.com/hubsaude/runner/simulador/internal/state"
)

// URLPadraoGitHub aponta para o repo da disciplina (ajustar conforme necessario).
// Sobrescritivel via env HUBSAUDE_SIMULADOR_REPO.
const URLPadraoGitHub = "https://api.github.com/repos/hubsaude/simulador/releases"

// NomeArquivo e o nome canonico do jar dentro do diretorio de cache.
const NomeArquivo = "simulador.jar"

// MetadadosArquivo contem informacoes sobre o jar cacheado.
const MetadadosArquivo = "simulador-meta.json"

// Metadados sobre a versao cacheada.
type Metadados struct {
	Versao    string    `json:"versao"`
	URL       string    `json:"url"`
	SHA256    string    `json:"sha256"`
	BaixadoEm time.Time `json:"baixadoEm"`
	TamanhoB  int64     `json:"tamanhoBytes"`
}

// Opcoes configuram um download/resolucao do simulador.jar.
type Opcoes struct {
	// HubsaudeDir e o diretorio base (~/.hubsaude). Resolvido via state.Dir() se vazio.
	HubsaudeDir string
	// Source sobrescreve a URL de download. Aceita http(s):// ou file://.
	// Quando definido, verificacao de versao via GitHub Releases e ignorada.
	Source string
	// Versao a baixar. Vazio ou "latest" = ultima release.
	Versao string
	// SHA256Esperado, se nao-vazio, e comparado apos o download.
	SHA256Esperado string
	// LogProgresso recebe mensagens. Pode ser nil.
	LogProgresso io.Writer
	// HTTPClient sobrescreve o client padrao (para testes).
	HTTPClient *http.Client
	// ForcarRedownload ignora o cache.
	ForcarRedownload bool
}

// Resultado descreve o jar disponivel localmente.
type Resultado struct {
	JarPath   string
	Metadados Metadados
	UsouCache bool
}

// Garantir retorna um Resultado apontando para o jar pronto para uso.
// Estrategia:
//  1. Se Source != "": baixa direto (ignora versao).
//  2. Senao: consulta GitHub Releases pela versao indicada.
//  3. Cache local: reusa se versao + SHA256 conferem (a menos que ForcarRedownload).
func Garantir(ctx context.Context, opc Opcoes) (*Resultado, error) {
	if opc.HTTPClient == nil {
		opc.HTTPClient = http.DefaultClient
	}
	dirBase := opc.HubsaudeDir
	if dirBase == "" {
		d, err := state.EnsureDir()
		if err != nil {
			return nil, err
		}
		dirBase = d
	}
	cacheDir := filepath.Join(dirBase, "simulador")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("ao criar diretorio de cache: %w", err)
	}
	jarPath := filepath.Join(cacheDir, NomeArquivo)
	metaPath := filepath.Join(cacheDir, MetadadosArquivo)

	if opc.Source != "" {
		log(opc, "obtendo simulador.jar de fonte custom: %s", opc.Source)
		return baixarFonteDireta(ctx, opc, jarPath, metaPath)
	}

	versao := opc.Versao
	if versao == "" {
		versao = "latest"
	}

	if !opc.ForcarRedownload {
		if r, ok := tentarCache(metaPath, jarPath, versao, opc.SHA256Esperado); ok {
			log(opc, "reusando simulador.jar em cache (versao=%s)", r.Metadados.Versao)
			return r, nil
		}
	}

	log(opc, "consultando GitHub Releases (%s)...", versao)
	rel, err := buscarRelease(ctx, opc, versao)
	if err != nil {
		return nil, err
	}
	asset, err := selecionarAsset(rel)
	if err != nil {
		return nil, err
	}
	log(opc, "baixando %s (%s)", asset.Nome, asset.URL)
	soma, n, err := baixarPara(ctx, opc.HTTPClient, asset.URL, jarPath)
	if err != nil {
		return nil, err
	}
	if opc.SHA256Esperado != "" && !strings.EqualFold(soma, opc.SHA256Esperado) {
		os.Remove(jarPath)
		return nil, fmt.Errorf("SHA-256 nao confere: esperado=%s, obtido=%s",
			opc.SHA256Esperado, soma)
	}
	meta := Metadados{
		Versao:    rel.TagName,
		URL:       asset.URL,
		SHA256:    soma,
		BaixadoEm: time.Now().UTC(),
		TamanhoB:  n,
	}
	if err := salvarMeta(metaPath, meta); err != nil {
		return nil, err
	}
	return &Resultado{JarPath: jarPath, Metadados: meta}, nil
}

func tentarCache(metaPath, jarPath, versao, esperado string) (*Resultado, bool) {
	if _, err := os.Stat(jarPath); err != nil {
		return nil, false
	}
	meta, err := lerMeta(metaPath)
	if err != nil {
		return nil, false
	}
	if versao != "latest" && meta.Versao != versao {
		return nil, false
	}
	if esperado != "" && !strings.EqualFold(meta.SHA256, esperado) {
		return nil, false
	}
	// Re-valida hash do arquivo on-disk para detectar corrupcao.
	soma, err := sha256DeArquivo(jarPath)
	if err != nil || !strings.EqualFold(soma, meta.SHA256) {
		return nil, false
	}
	return &Resultado{JarPath: jarPath, Metadados: meta, UsouCache: true}, true
}

func baixarFonteDireta(ctx context.Context, opc Opcoes, jarPath, metaPath string) (*Resultado, error) {
	soma, n, err := baixarPara(ctx, opc.HTTPClient, opc.Source, jarPath)
	if err != nil {
		return nil, err
	}
	if opc.SHA256Esperado != "" && !strings.EqualFold(soma, opc.SHA256Esperado) {
		os.Remove(jarPath)
		return nil, fmt.Errorf("SHA-256 nao confere: esperado=%s, obtido=%s",
			opc.SHA256Esperado, soma)
	}
	versao := opc.Versao
	if versao == "" {
		versao = "custom"
	}
	meta := Metadados{
		Versao:    versao,
		URL:       opc.Source,
		SHA256:    soma,
		BaixadoEm: time.Now().UTC(),
		TamanhoB:  n,
	}
	if err := salvarMeta(metaPath, meta); err != nil {
		return nil, err
	}
	return &Resultado{JarPath: jarPath, Metadados: meta}, nil
}

type assetRelease struct {
	Nome string
	URL  string
}

type releaseGitHub struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func buscarRelease(ctx context.Context, opc Opcoes, versao string) (*releaseGitHub, error) {
	base := strings.TrimRight(os.Getenv("HUBSAUDE_SIMULADOR_REPO"), "/")
	if base == "" {
		base = URLPadraoGitHub
	}
	var endpoint string
	if versao == "latest" {
		endpoint = base + "/latest"
	} else {
		endpoint = base + "/tags/" + versao
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := opc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ao consultar GitHub Releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %q nao encontrada", versao)
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub Releases retornou HTTP %d: %s", resp.StatusCode, body)
	}
	var rel releaseGitHub
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("ao decodificar release: %w", err)
	}
	return &rel, nil
}

func selecionarAsset(rel *releaseGitHub) (*assetRelease, error) {
	for _, a := range rel.Assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".jar") {
			return &assetRelease{Nome: a.Name, URL: a.URL}, nil
		}
	}
	return nil, fmt.Errorf("nenhum asset .jar em %s", rel.TagName)
}

func baixarPara(ctx context.Context, client *http.Client, urlStr, dest string) (string, int64, error) {
	if strings.HasPrefix(urlStr, "file://") {
		return copiarLocal(strings.TrimPrefix(urlStr, "file://"), dest)
	}
	if _, err := url.Parse(urlStr); err != nil {
		return "", 0, fmt.Errorf("URL invalida: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("ao baixar %s: %w", urlStr, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", 0, fmt.Errorf("download retornou HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("ao gravar download: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func copiarLocal(src, dest string) (string, int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", 0, err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, h), in)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func sha256DeArquivo(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func salvarMeta(path string, m Metadados) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func lerMeta(path string) (Metadados, error) {
	var m Metadados
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	if m.SHA256 == "" {
		return m, errors.New("meta sem SHA-256")
	}
	return m, nil
}

func log(opc Opcoes, formato string, args ...any) {
	if opc.LogProgresso == nil {
		return
	}
	fmt.Fprintf(opc.LogProgresso, "[i] "+formato+"\n", args...)
}
