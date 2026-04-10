package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"mb/internal/infra/selfupdate"
	"mb/internal/shared/system"
	"mb/internal/shared/version"
)

func TestRun_CheckOnlyJSON_AlreadyLatest(t *testing.T) {
	old := version.Version
	version.Version = "v1.0.0"
	t.Cleanup(func() { version.Version = old })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	d := testDepsForUpdateCLI(t)
	log := system.NewLogger(false, false, &strings.Builder{})

	err := Run(ctx, cmd, d, log, Options{
		OnlyCLI:   true,
		CheckOnly: true,
		JSON:      true,
		SelfUpdate: &selfupdate.Config{
			LatestReleaseURL: srv.URL,
		},
		RunAllGitPlugins: func(context.Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	var rep selfupdate.CheckOnlyReport
	if err := json.Unmarshal([]byte(strings.TrimSpace(out.String())), &rep); err != nil {
		t.Fatalf("stdout JSON: %v; raw=%q", err, out.String())
	}
	if rep.LocalVersion != "v1.0.0" || rep.RemoteVersion != "v1.0.0" || rep.UpdateAvailable {
		t.Fatalf("unexpected report: %+v", rep)
	}
}

func TestRun_JSONRequiresOnlyCLIAndCheckOnly(t *testing.T) {
	ctx := context.Background()
	cmd := &cobra.Command{}
	d := testDepsForUpdateCLI(t)
	log := system.NewLogger(false, false, &strings.Builder{})

	err := Run(ctx, cmd, d, log, Options{
		JSON:             true,
		RunAllGitPlugins: func(context.Context) error { return nil },
	})
	if err == nil || !strings.Contains(err.Error(), "--json") {
		t.Fatalf("expected --json validation error, got: %v", err)
	}
}
