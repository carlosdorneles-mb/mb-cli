package aliases

import "strings"

// storeKeySep separates env_vault from display name in map keys. Neither vault nor
// ValidateName names may contain this rune.
const storeKeySep = "\x1e"

// StoreKey returns the canonical map key for (env_vault, displayName).
// env_vault is the empty string for aliases without a vault.
func StoreKey(envVault, displayName string) string {
	return envVault + storeKeySep + displayName
}

// ParseStoreKey splits a StoreKey into env_vault and displayName.
func ParseStoreKey(key string) (envVault, displayName string, ok bool) {
	i := strings.Index(key, storeKeySep)
	if i < 0 {
		return "", "", false
	}
	return key[:i], key[i+len(storeKeySep):], true
}

// EntryStoreKey returns StoreKey(e.EnvVault, displayName).
func EntryStoreKey(displayName string, e Entry) string {
	return StoreKey(e.EnvVault, displayName)
}
