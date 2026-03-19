// Package selfupdate implements checking and installing newer MB CLI releases from GitHub.
package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"mb/internal/config"
	"mb/internal/version"
)

// Config controls where to fetch release metadata and artifacts (for tests, use custom URLs).
type Config struct {
	Repo string // owner/repo; if empty, version.UpdateRepo or config.DefaultUpdateRepo

	// LatestReleaseURL if non-empty overrides the GitHub API latest release URL.
	LatestReleaseURL string
	// ReleaseDownloadBase is the prefix before "/{tag}/{artifact}".
	// Example: https://github.com/owner/repo/releases/download
	ReleaseDownloadBase string

	HTTPClient *http.Client

	// DestPath if set replaces the install target (for tests only).
	DestPath string
}

func (c *Config) repo() string {
	r := strings.TrimSpace(c.Repo)
	if r != "" {
		return r
	}
	r = strings.TrimSpace(version.UpdateRepo)
	if r != "" {
		return r
	}
	return config.DefaultUpdateRepo
}

func (c *Config) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func latestAPIURL(cfg *Config) string {
	if cfg.LatestReleaseURL != "" {
		return cfg.LatestReleaseURL
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", cfg.repo())
}

func downloadBase(cfg *Config) string {
	if cfg.ReleaseDownloadBase != "" {
		return strings.TrimSuffix(cfg.ReleaseDownloadBase, "/")
	}
	return fmt.Sprintf("https://github.com/%s/releases/download", cfg.repo())
}

// CanonicalSemver returns "vX.Y.Z" if s is a valid semver, else ("", false).
func CanonicalSemver(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	candidate := "v" + s
	if semver.IsValid(candidate) {
		return candidate, true
	}
	return "", false
}

// ShouldFetchNewRelease reports whether we should download the latest release tarball.
func ShouldFetchNewRelease(localVersion, latestTag string) bool {
	latest := strings.TrimSpace(latestTag)
	if latest == "" {
		return false
	}
	if !semver.IsValid(latest) {
		lt := "v" + strings.TrimPrefix(latest, "v")
		if semver.IsValid(lt) {
			latest = lt
		} else {
			return true
		}
	}
	local := strings.TrimSpace(localVersion)
	if local == "" || strings.EqualFold(local, "dev") {
		return true
	}
	if strings.Contains(local, "(devel)") {
		return true
	}
	lv, ok := CanonicalSemver(local)
	if !ok {
		return true
	}
	return semver.Compare(latest, lv) > 0
}

// Platform maps runtime to release artifact OS/ARCH (linux/darwin, amd64/arm64).
func Platform() (goos, goarch string, err error) {
	switch runtime.GOOS {
	case "linux", "darwin":
		goos = runtime.GOOS
	default:
		return "", "", fmt.Errorf(
			"sistema operational não suportado para atualização automática: %s (suportado: linux, darwin)",
			runtime.GOOS,
		)
	}
	switch runtime.GOARCH {
	case "amd64", "arm64":
		goarch = runtime.GOARCH
	default:
		return "", "", fmt.Errorf(
			"arquitetura não suportada para atualização automática: %s (suportado: amd64, arm64)",
			runtime.GOARCH,
		)
	}
	return goos, goarch, nil
}

type releaseLatestJSON struct {
	TagName string `json:"tag_name"`
}

// FetchLatestTag returns the tag_name of the latest stable GitHub release.
func FetchLatestTag(ctx context.Context, cfg *Config) (string, error) {
	url := latestAPIURL(cfg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mb-cli-self-update")
	if tok := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := cfg.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("falha ao contactar o GitHub: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	switch resp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusForbidden:
		if bytes.Contains(bytes.ToLower(body), []byte("rate limit")) {
			return "", errors.New(
				"limite de pedidos à API do GitHub excedido; tente mais tarde ou defina GITHUB_TOKEN",
			)
		}
		return "", fmt.Errorf("acesso negado pela API do GitHub (HTTP %d)", resp.StatusCode)
	case http.StatusNotFound:
		return "", errors.New("repositório ou release não encontrado no GitHub")
	default:
		return "", fmt.Errorf("API do GitHub retornou HTTP %d", resp.StatusCode)
	}

	var out releaseLatestJSON
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("resposta inválida da API do GitHub: %w", err)
	}
	tag := strings.TrimSpace(out.TagName)
	if tag == "" {
		return "", errors.New("release sem tag_name na resposta da API")
	}
	return tag, nil
}

func artifactName(tag string, goos, goarch string) string {
	ver := strings.TrimPrefix(strings.TrimSpace(tag), "v")
	return fmt.Sprintf("mb_%s_%s_%s.tar.gz", ver, goos, goarch)
}

func expectedSHA256FromChecksums(checksums []byte, artifact string) ([]byte, error) {
	base := filepath.Base(artifact)
	for _, line := range bytes.Split(checksums, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		parts := strings.Fields(string(line))
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		nameField := parts[len(parts)-1]
		nameField = strings.TrimPrefix(nameField, "*")
		if filepath.Base(nameField) != base {
			continue
		}
		b, err := hex.DecodeString(hash)
		if err != nil || len(b) != sha256.Size {
			continue
		}
		return b, nil
	}
	return nil, fmt.Errorf("checksum não encontrado para %s em checksums.txt", artifact)
}

func verifyTarballSHA256(tarball []byte, expected []byte) error {
	sum := sha256.Sum256(tarball)
	if !bytes.Equal(sum[:], expected) {
		return errors.New("checksum SHA256 do arquivo baixado não confere")
	}
	return nil
}

// ExtractMBBinary reads a tar.gz release archive and returns the mb executable bytes.
func ExtractMBBinary(tarGz []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(tarGz))
	if err != nil {
		return nil, fmt.Errorf("arquivo compactado inválido: %w", err)
	}
	defer func() { _ = zr.Close() }()
	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != 0 { // 0 = legacy TypeRegA
			continue
		}
		if filepath.Base(hdr.Name) != "mb" {
			continue
		}
		const maxBinary = 128 << 20
		data, err := io.ReadAll(io.LimitReader(tr, maxBinary))
		if err != nil {
			return nil, err
		}
		if len(data) < 4 {
			continue
		}
		return data, nil
	}
	return nil, errors.New("binário mb não encontrado no arquivo de release")
}

// DownloadReleaseArtifact fetches tarball and checksums for the given tag.
func DownloadReleaseArtifact(
	ctx context.Context,
	cfg *Config,
	tag, goos, goarch string,
) (tarball []byte, err error) {
	art := artifactName(tag, goos, goarch)
	base := downloadBase(cfg)
	tarURL := fmt.Sprintf("%s/%s/%s", base, tag, art)
	sumURL := fmt.Sprintf("%s/%s/checksums.txt", base, tag)

	tarball, err = httpGetBody(ctx, cfg.client(), tarURL)
	if err != nil {
		return nil, fmt.Errorf("download do release: %w", err)
	}
	sums, err := httpGetBody(ctx, cfg.client(), sumURL)
	if err != nil {
		return nil, fmt.Errorf("download de checksums.txt: %w", err)
	}
	exp, err := expectedSHA256FromChecksums(sums, art)
	if err != nil {
		return nil, err
	}
	if err := verifyTarballSHA256(tarball, exp); err != nil {
		return nil, err
	}
	return tarball, nil
}

func httpGetBody(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mb-cli-self-update")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d ao obter %s", resp.StatusCode, url)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 256<<20))
}

// InstallBinary writes newExe into destPath (same directory as running binary, same basename).
func InstallBinary(destPath string, newExe []byte) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("criar diretório de destino: %w", err)
	}
	f, err := os.CreateTemp(dir, ".mb-update-*")
	if err != nil {
		return fmt.Errorf("criar arquivo temporário: %w", err)
	}
	tmpPath := f.Name()
	_, werr := f.Write(newExe)
	cerr := f.Chmod(0o755)
	_ = f.Close()
	if werr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("gravar binário: %w", werr)
	}
	if cerr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("definir permissões: %w", cerr)
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf(
			"substituir binário (verifique permissões de escrita em %s): %w",
			dir,
			err,
		)
	}
	return nil
}

// DestExecutablePath returns the path to replace (resolved symlink, full path).
func DestExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// ExitCodeUpdateAvailable is returned by RunCheckOnly (and mb self update --check-only) when a newer release exists.
const ExitCodeUpdateAvailable = 2

// MessageNoUpdateNeeded returns the user-facing message when ShouldFetchNewRelease is false.
func MessageNoUpdateNeeded(localVersion, tag string) string {
	lt := tag
	if !strings.HasPrefix(lt, "v") {
		lt = "v" + strings.TrimPrefix(lt, "v")
	}
	if lv, ok := CanonicalSemver(
		localVersion,
	); ok && semver.IsValid(lt) &&
		semver.Compare(lt, lv) < 0 {
		return fmt.Sprintf(
			"Sua versão (%s) é mais recente que a última release estável (%s). Nenhuma ação necessária.\n",
			strings.TrimSpace(localVersion),
			tag,
		)
	}
	return fmt.Sprintf("MB CLI já está na versão mais recente (%s).\n", tag)
}

// RunCheckOnly queries GitHub and compares versions without downloading. exitCode is ExitCodeUpdateAvailable when an update exists, else 0.
func RunCheckOnly(
	ctx context.Context,
	cfg *Config,
	localVersion string,
) (stdout string, exitCode int, err error) {
	tag, err := FetchLatestTag(ctx, cfg)
	if err != nil {
		return "", 0, err
	}
	if ShouldFetchNewRelease(localVersion, tag) {
		return fmt.Sprintf(
			"Atualização disponível para o MB CLI.\nVersão instalada: %s\nÚltima release estável: %s\nExecute: mb self update\n",
			strings.TrimSpace(localVersion),
			tag,
		), ExitCodeUpdateAvailable, nil
	}
	return MessageNoUpdateNeeded(localVersion, tag), 0, nil
}

// Run checks GitHub, downloads if needed, and installs. Returns messages for stdout (already updated / up to date).
func Run(ctx context.Context, cfg *Config, localVersion string) (stdout string, err error) {
	goos, goarch, err := Platform()
	if err != nil {
		return "", err
	}
	tag, err := FetchLatestTag(ctx, cfg)
	if err != nil {
		return "", err
	}
	if !ShouldFetchNewRelease(localVersion, tag) {
		return MessageNoUpdateNeeded(localVersion, tag), nil
	}
	tarball, err := DownloadReleaseArtifact(ctx, cfg, tag, goos, goarch)
	if err != nil {
		return "", err
	}
	mbData, err := ExtractMBBinary(tarball)
	if err != nil {
		return "", err
	}
	dest := strings.TrimSpace(cfg.DestPath)
	if dest == "" {
		var err error
		dest, err = DestExecutablePath()
		if err != nil {
			return "", fmt.Errorf("localizar executável: %w", err)
		}
	}
	if err := InstallBinary(dest, mbData); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"MB CLI atualizado para %s.\nReexecute o commando mb para usar a nova versão.\n",
		tag,
	), nil
}
