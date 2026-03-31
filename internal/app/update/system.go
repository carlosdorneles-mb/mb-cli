package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"mb/internal/shared/system"
)

// RunSystemUpdate runs OS-level package updates: Homebrew and mas on darwin; apt-get (via sudo),
// flatpak, and snap refresh (via sudo) on Linux when each tool is on PATH.
func RunSystemUpdate(ctx context.Context, log *system.Logger) error {
	_ = log.Info(ctx, "Atualizando pacotes do sistema...")

	switch runtime.GOOS {
	case "darwin":
		return runDarwinSystemUpdate(ctx, log)
	case "linux":
		return runLinuxSystemUpdate(ctx, log)
	default:
		_ = log.Warn(ctx, "SO não suportado para atualização automática (%s).", runtime.GOOS)
		return nil
	}
}

// linuxPackageEnv returns os.Environ with CI=1 and optional extra KEY=value pairs for non-interactive tools.
func linuxPackageEnv(extra ...string) []string {
	return append(append(os.Environ(), "CI=1"), extra...)
}

// runCmd runs an external command with stdin from /dev/null and stdout/stderr to os.Stderr.
func runCmd(ctx context.Context, env []string, bin string, args ...string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	if env == nil {
		env = os.Environ()
	}
	cmd.Env = env
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer stdin.Close()
	cmd.Stdin = stdin
	return cmd.Run()
}

func runDarwinSystemUpdate(ctx context.Context, log *system.Logger) error {
	brewPath, err := exec.LookPath("brew")
	if err != nil {
		_ = log.Warn(ctx, "Homebrew não encontrado; ignorando atualização do sistema.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando Homebrew...")

	if err := runCmd(ctx, nil, brewPath, "update"); err != nil {
		return fmt.Errorf("brew update: %w", err)
	}
	if err := runCmd(ctx, nil, brewPath, "upgrade"); err != nil {
		return fmt.Errorf("brew upgrade: %w", err)
	}

	masPath, err := exec.LookPath("mas")
	if err != nil {
		_ = log.Debug(ctx, "mas não encontrado; ignorando atualização do sistema.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando App Store...")

	if err := runCmd(ctx, nil, masPath, "upgrade"); err != nil {
		return fmt.Errorf("mas upgrade: %w", err)
	}
	return nil
}

func runLinuxSystemUpdate(ctx context.Context, log *system.Logger) error {
	sudoPath, errSudo := exec.LookPath("sudo")
	if errSudo != nil {
		_ = log.Warn(ctx, "sudo não encontrado no PATH; ignorando snap refresh.")
		return nil
	}

	if err := runLinuxAPT(ctx, log, sudoPath); err != nil {
		return err
	}

	if err := runLinuxFlatpak(ctx, log); err != nil {
		return err
	}

	return runLinuxSnap(ctx, log, sudoPath)
}

func runLinuxAPT(ctx context.Context, log *system.Logger, sudoPath string) error {
	aptPath, errApt := exec.LookPath("apt-get")
	if errApt != nil {
		_ = log.Debug(ctx, "apt-get não encontrado no PATH; ignorando APT.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando APT...")

	env := linuxPackageEnv("DEBIAN_FRONTEND=noninteractive")
	runApt := func(args ...string) error {
		sudoArgs := append([]string{aptPath}, args...)
		cmd := exec.CommandContext(ctx, sudoPath, sudoArgs...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Env = env
		stdin, err := os.Open(os.DevNull)
		if err != nil {
			return err
		}
		defer stdin.Close()
		cmd.Stdin = stdin
		return cmd.Run()
	}

	if err := runApt("update"); err != nil {
		return fmt.Errorf("apt-get update: %w", err)
	}

	if err := runApt("upgrade", "-y"); err != nil {
		return fmt.Errorf("apt-get upgrade: %w", err)
	}
	return nil
}

func runLinuxFlatpak(ctx context.Context, log *system.Logger) error {
	fp, err := exec.LookPath("flatpak")
	if err != nil {
		_ = log.Debug(ctx, "flatpak não encontrado no PATH; ignorando flatpak.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando Flatpak...")

	if err := runCmd(ctx, linuxPackageEnv(), fp, "update", "-y", "--noninteractive"); err != nil {
		return fmt.Errorf("flatpak update: %w", err)
	}
	return nil
}

func runLinuxSnap(ctx context.Context, log *system.Logger, sudoPath string) error {
	snapPath, errSnap := exec.LookPath("snap")
	if errSnap != nil {
		_ = log.Debug(ctx, "snap não encontrado no PATH; ignorando snap.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando Snap...")

	cmd := exec.CommandContext(
		ctx,
		sudoPath,
		snapPath,
		"refresh",
		"--color=never",
		"--unicode=never",
	)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = linuxPackageEnv()
	stdin, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer stdin.Close()
	cmd.Stdin = stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("snap refresh: %w", err)
	}
	return nil
}
