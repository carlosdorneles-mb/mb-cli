package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"charm.land/fang/v2"

	"mb/internal/bootstrap"
	"mb/internal/shared/ui"
	"mb/internal/shared/version"
)

func main() {
	ctx := context.Background()

	fxApp, rootCmd, err := bootstrap.Bootstrap()
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
