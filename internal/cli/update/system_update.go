package update

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"mb/internal/shared/system"
)

// RunSystemUpdate runs OS-level package updates: Homebrew and mas on darwin;
// apt-get (via sudo), flatpak, and snap refresh (via sudo) on Linux when each tool is available.
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

func runCmd(ctx context.Context, bin string, args ...string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func runDarwinSystemUpdate(ctx context.Context, log *system.Logger) error {
	brewPath, err := exec.LookPath("brew")
	if err != nil {
		_ = log.Warn(ctx, "Homebrew não encontrado; ignorando atualização do sistema.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando Homebrew...")

	if err := runCmd(ctx, brewPath, "update"); err != nil {
		return fmt.Errorf("brew update: %w", err)
	}
	if err := runCmd(ctx, brewPath, "upgrade"); err != nil {
		return fmt.Errorf("brew upgrade: %w", err)
	}

	masPath, err := exec.LookPath("mas")
	if err != nil {
		_ = log.Debug(ctx, "mas não encontrado; ignorando atualização do sistema.")
		return nil
	}

	_ = log.Print(ctx, "=> Atualizando App Store...")

	if err := runCmd(ctx, masPath, "upgrade"); err != nil {
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

	env := append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	runApt := func(args ...string) error {
		sudoArgs := append([]string{aptPath}, args...)
		cmd := exec.CommandContext(ctx, sudoPath, sudoArgs...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Env = env
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

	if err := runCmd(ctx, fp, "update", "-y"); err != nil {
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

	cmd := exec.CommandContext(ctx, sudoPath, snapPath, "refresh")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("snap refresh: %w", err)
	}
	return nil
}
