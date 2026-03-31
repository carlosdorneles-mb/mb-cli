package ports

// SecretStore persists sensitive values outside plain env files (OS keyring).
type SecretStore interface {
	Set(group, envKey, value string) error
	Get(group, envKey string) (string, error)
	Delete(group, envKey string) error
}
