package main

import (
	"context"
	"fmt"
	"os"

	"charm.land/fang/v2"

	"mb/internal/app"
	"mb/internal/shared/ui"
	"mb/internal/version"
)

func main() {
	ctx := context.Background()

	fxApp, rootCmd, err := app.Bootstrap()
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
		os.Exit(1)
	}
}
