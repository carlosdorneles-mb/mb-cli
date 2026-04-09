package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCanonicalSemver(t *testing.T) {
	t.Parallel()
	if s, ok := CanonicalSemver("1.2.3"); !ok || s != "v1.2.3" {
		t.Fatalf("got %q %v", s, ok)
	}
	if s, ok := CanonicalSemver("v4.0.0"); !ok || s != "v4.0.0" {
		t.Fatalf("got %q %v", s, ok)
	}
	if _, ok := CanonicalSemver("dev"); ok {
		t.Fatal("expected false")
	}
}

func TestShouldFetchNewRelease(t *testing.T) {
	t.Parallel()
	tests := []struct {
		local, latest string
		want          bool
	}{
		{"dev", "v1.0.0", true},
		{"", "v1.0.0", true},
		{"v0.9.0", "v1.0.0", true},
		{"0.9.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.1", "v1.0.0", false},
		{"v2.0.0", "v1.9.9", false},
	}
	for _, tc := range tests {
		if got := ShouldFetchNewRelease(tc.local, tc.latest); got != tc.want {
			t.Errorf("ShouldFetchNewRelease(%q,%q)=%v want %v", tc.local, tc.latest, got, tc.want)
		}
	}
}

func TestFetchLatestTag(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.5.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	tag, err := FetchLatestTag(context.Background(), cfg)
	if err != nil || tag != "v0.5.0" {
		t.Fatalf("tag=%q err=%v", tag, err)
	}
}

func TestFetchLatestTag_notFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	_, err := FetchLatestTag(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExpectedSHA256FromChecksums(t *testing.T) {
	t.Parallel()
	artifact := "mb_1.0.0_linux_amd64.tar.gz"
	hash := bytes.Repeat([]byte("ab"), sha256.Size/2)
	line := hex.EncodeToString(hash) + "  " + artifact + "\n"
	got, err := expectedSHA256FromChecksums([]byte(line), artifact)
	if err != nil || !bytes.Equal(got, hash) {
		t.Fatalf("err=%v got=%x", err, got)
	}
}

func tarGzMB(payload []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name: "mb",
		Mode: 0o755,
		Size: int64(len(payload)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write(payload); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestExtractMBBinary(t *testing.T) {
	t.Parallel()
	tgz, err := tarGzMB([]byte("#!/bin/sh\necho hi"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := ExtractMBBinary(tgz)
	if err != nil || string(out) != "#!/bin/sh\necho hi" {
		t.Fatalf("err=%v out=%q", err, out)
	}
}

func TestRun_mockedDownloadAndInstall(t *testing.T) {
	goos, goarch, err := Platform()
	if err != nil {
		t.Skip(err)
	}
	const tag = "v9.9.9"
	ver := "9.9.9"
	art := fmt.Sprintf("mb_%s_%s_%s.tar.gz", ver, goos, goarch)
	payload := []byte("#!/bin/sh\necho mock-mb-999")
	tgz, err := tarGzMB(payload)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(tgz)
	checksums := hex.EncodeToString(sum[:]) + "  " + art + "\n"

	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"` + tag + `"}`))
	})
	mux.HandleFunc("/dl/"+tag+"/"+art, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(tgz)
	})
	mux.HandleFunc("/dl/"+tag+"/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(checksums))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	dest := filepath.Join(t.TempDir(), "mb")
	cfg := &Config{
		LatestReleaseURL:    srv.URL + "/latest",
		ReleaseDownloadBase: srv.URL + "/dl",
		DestPath:            dest,
	}

	out, err := Run(context.Background(), cfg, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected stdout message")
	}
	b, err := os.ReadFile(dest)
	if err != nil || !bytes.Equal(b, payload) {
		t.Fatalf("installed binary: err=%v len=%d", err, len(b))
	}
}

func TestRun_alreadyLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	out, err := Run(context.Background(), cfg, "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected message")
	}
}

func TestArtifactName(t *testing.T) {
	t.Parallel()
	if got := artifactName("v2.3.4", "linux", "arm64"); got != "mb_2.3.4_linux_arm64.tar.gz" {
		t.Fatal(got)
	}
}

func TestRun_newerLocalThanRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	out, err := Run(context.Background(), cfg, "v2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected message")
	}
}

func TestPlatform_unsupportedOS(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		_, _, err := Platform()
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunCheckOnly_updateAvailable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	out, code, err := RunCheckOnly(context.Background(), cfg, "v1.0.0")
	if err != nil || code != ExitCodeUpdateAvailable {
		t.Fatalf("err=%v code=%d", err, code)
	}
	if !strings.Contains(out, "Atualização disponível") || !strings.Contains(out, "v2.0.0") {
		t.Fatalf("unexpected out: %q", out)
	}
}

func TestRunCheckOnly_alreadyLatest(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	out, code, err := RunCheckOnly(context.Background(), cfg, "v1.0.0")
	if err != nil || code != 0 {
		t.Fatalf("err=%v code=%d", err, code)
	}
	if !strings.Contains(out, "versão mais recente") {
		t.Fatalf("out=%q", out)
	}
}

func TestRunCheckOnly_newerLocalThanRelease(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	out, code, err := RunCheckOnly(context.Background(), cfg, "v2.0.0")
	if err != nil || code != 0 {
		t.Fatalf("err=%v code=%d", err, code)
	}
	if !strings.Contains(out, "mais recente que a última release") {
		t.Fatalf("out=%q", out)
	}
}

func TestCheckOnlyDetails_matchesRunCheckOnly(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v2.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	rep, msg, code, err := CheckOnlyDetails(context.Background(), cfg, "v1.0.0")
	if err != nil || code != ExitCodeUpdateAvailable {
		t.Fatalf("err=%v code=%d", err, code)
	}
	if rep.LocalVersion != "v1.0.0" || rep.RemoteVersion != "v2.0.0" || !rep.UpdateAvailable {
		t.Fatalf("report: %+v", rep)
	}
	msg2, code2, err2 := RunCheckOnly(context.Background(), cfg, "v1.0.0")
	if err2 != nil || code2 != code || msg != msg2 {
		t.Fatalf("RunCheckOnly diverged: err2=%v code2=%d msg2=%q", err2, code2, msg2)
	}
}

func TestRunCheckOnly_apiError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	cfg := &Config{LatestReleaseURL: srv.URL}
	_, _, err := RunCheckOnly(context.Background(), cfg, "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
}
