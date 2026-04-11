package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/fang/v2"

	"mb/internal/bootstrap"
	"mb/internal/shared/ui"
	"mb/internal/shared/version"
)

func main() {
	ctx := context.Background()

	// Check for verbose flag before bootstrap to enable FX lifecycle logging.
	verbose := hasVerboseFlag(os.Args[1:])
	fxApp, rootCmd, err := bootstrap.Bootstrap(verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.RenderError(fmt.Sprintf("bootstrap failure: %v", err)))
		os.Exit(1)
	}

	if err := fxApp.Start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, ui.RenderError(fmt.Sprintf("startup failure: %v", err)))
		os.Exit(1)
	}

	defer func() {
		_ = fxApp.Stop(ctx)
	}()

	v := version.Version
	if v == "" {
		v = "dev"
	}
	opts := []fang.Option{
		fang.WithoutManpage(),
		fang.WithErrorHandler(ui.ErrorHandlerPT),
		fang.WithVersion(v),
		fang.WithColorSchemeFunc(ui.MBHelpTheme()),
	}
	if err := fang.Execute(ctx, rootCmd, opts...); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if code := exitErr.ExitCode(); code >= 0 {
				os.Exit(code)
			}
		}
		os.Exit(1)
	}
}

// hasVerboseFlag checks if -v or --verbose appears in the initial args.
// This is a best-effort check to enable FX logging before Cobra parses flags.
func hasVerboseFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-v" || arg == "--verbose" {
			return true
		}
		// Stop at first subcommand (non-flag token)
		if !strings.HasPrefix(arg, "-") {
			break
		}
	}
	return false
}
