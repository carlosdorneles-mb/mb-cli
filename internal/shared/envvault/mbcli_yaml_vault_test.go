package envvault

import "testing"

func TestValidateMbcliYAMLProjectNestedSuffix(t *testing.T) {
	t.Parallel()
	if err := ValidateMbcliYAMLProjectNestedSuffix("staging"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMbcliYAMLProjectNestedSuffix(""); err == nil {
		t.Fatal("expected error for empty suffix")
	}
	if err := ValidateMbcliYAMLProjectNestedSuffix("   "); err == nil {
		t.Fatal("expected error for whitespace-only suffix")
	}
	if err := ValidateMbcliYAMLProjectNestedSuffix("../x"); err == nil {
		t.Fatal("expected error for invalid name")
	}
}
