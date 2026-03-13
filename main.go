package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/fang"

	"mb/internal/app"
	"mb/internal/ui"
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

	opts := []fang.Option{
		fang.WithoutManpage(),
	}
	if err := fang.Execute(ctx, rootCmd, opts...); err != nil {
		os.Exit(1)
	}
}
