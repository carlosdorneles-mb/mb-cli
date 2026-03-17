package version

import "testing"

func TestIsReleaseBuild(t *testing.T) {
	orig := Version
	t.Cleanup(func() { Version = orig })

	Version = ""
	if IsReleaseBuild() {
		t.Fatal("empty Version should not be release")
	}
	Version = "   "
	if IsReleaseBuild() {
		t.Fatal("whitespace Version should not be release")
	}
	Version = "v0.1.0"
	if !IsReleaseBuild() {
		t.Fatal("tag Version should be release")
	}
}
