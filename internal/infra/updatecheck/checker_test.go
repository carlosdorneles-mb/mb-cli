package updatecheck

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestShouldCheck_FirstTime(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")
	if !c.ShouldCheck() {
		t.Error("ShouldCheck should return true when no previous check exists")
	}
}

func TestShouldCheck_Recent(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	// Grava check agora
	now := time.Now().Unix()
	_ = os.WriteFile(filepath.Join(tmp, lastCheckFile), []byte(strconv.FormatInt(now, 10)), 0o644)

	if c.ShouldCheck() {
		t.Error("ShouldCheck should return false when check was recent")
	}
}

func TestShouldCheck_Old(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	// Grava check há 2 horas
	old := time.Now().Add(-2 * time.Hour).Unix()
	_ = os.WriteFile(filepath.Join(tmp, lastCheckFile), []byte(strconv.FormatInt(old, 10)), 0o644)

	if !c.ShouldCheck() {
		t.Error("ShouldCheck should return true when check is older than interval")
	}
}

func TestShouldCheck_CorruptFile(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	_ = os.WriteFile(filepath.Join(tmp, lastCheckFile), []byte("not-a-number"), 0o644)

	if !c.ShouldCheck() {
		t.Error("ShouldCheck should return true when file is corrupt")
	}
}

func TestIsUpdateAvailable_NoFile(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	if tag, ok := c.IsUpdateAvailable(); ok || tag != "" {
		t.Errorf(
			"IsUpdateAvailable should return false when no file exists, got tag=%q ok=%v",
			tag,
			ok,
		)
	}
}

func TestIsUpdateAvailable_WithTag(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	_ = c.RecordAvailable("v1.2.0")

	tag, ok := c.IsUpdateAvailable()
	if !ok || tag != "v1.2.0" {
		t.Errorf("IsUpdateAvailable should return v1.2.0, got tag=%q ok=%v", tag, ok)
	}
}

func TestRecordCheck_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	if err := c.RecordCheck(); err != nil {
		t.Fatalf("RecordCheck failed: %v", err)
	}

	// Verifica que ShouldCheck agora retorna false
	if c.ShouldCheck() {
		t.Error("ShouldCheck should be false after RecordCheck")
	}
}

func TestRecordAvailable_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	if err := c.RecordAvailable("v2.0.0"); err != nil {
		t.Fatalf("RecordAvailable failed: %v", err)
	}

	tag, ok := c.IsUpdateAvailable()
	if !ok || tag != "v2.0.0" {
		t.Errorf("Expected v2.0.0, got tag=%q ok=%v", tag, ok)
	}
}

func TestClearAvailable(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	_ = c.RecordAvailable("v1.5.0")
	c.ClearAvailable()

	if tag, ok := c.IsUpdateAvailable(); ok {
		t.Errorf(
			"IsUpdateAvailable should be false after ClearAvailable, got tag=%q ok=%v",
			tag,
			ok,
		)
	}
}

func TestIsDisabled_EnvNotSet(t *testing.T) {
	// Garante que env não está set
	_ = os.Unsetenv(DisableEnvVar)

	if IsDisabled() {
		t.Error("IsDisabled should return false when env is not set")
	}
}

func TestIsDisabled_True(t *testing.T) {
	_ = os.Setenv(DisableEnvVar, "1")
	defer os.Unsetenv(DisableEnvVar)

	if !IsDisabled() {
		t.Error("IsDisabled should return true when env is '1'")
	}
}

func TestIsDisabled_TrueUpperCase(t *testing.T) {
	_ = os.Setenv(DisableEnvVar, "TRUE")
	defer os.Unsetenv(DisableEnvVar)

	if !IsDisabled() {
		t.Error("IsDisabled should return true when env is 'TRUE'")
	}
}

func TestIsDisabled_False(t *testing.T) {
	_ = os.Setenv(DisableEnvVar, "0")
	defer os.Unsetenv(DisableEnvVar)

	if IsDisabled() {
		t.Error("IsDisabled should return false when env is '0'")
	}
}

func TestRun_Disabled(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv(DisableEnvVar, "1")
	defer os.Unsetenv(DisableEnvVar)

	c := NewChecker(tmp, "v1.0.0")

	err := c.Run(context.Background())
	if err != nil {
		t.Fatalf("Run should not error when disabled, got: %v", err)
	}

	// Quando disabled, não deve tentar API nem gravar check
	// Mas ShouldCheck ainda retorna true porque não gravamos
	// O importante é que não houve erro
}

func TestRun_NoPreviousCheck(t *testing.T) {
	tmp := t.TempDir()
	c := NewChecker(tmp, "v1.0.0")

	// Não deve error mesmo sem API
	// O run vai tentar API mas vai falhar silenciosamente
	err := c.Run(context.Background())
	if err != nil {
		t.Fatalf("Run should not error on API failure, got: %v", err)
	}
}
