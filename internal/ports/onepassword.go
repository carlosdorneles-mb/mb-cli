package ports

// OnePasswordEnv integrates mb env secrets with the 1Password CLI (`op`).
// Implementations return op:// references for storage in the OS keyring.
type OnePasswordEnv interface {
	// EnsureAvailable returns an error if the `op` binary is missing or unusable.
	EnsureAvailable() error
	// PutSecret stores value in a Password item (one per keyring group); returns an op:// reference.
	PutSecret(keyringGroup, key, value string) (opReference string, err error)
	// RemoveSecretField removes the MB CLI custom field for key from the group's item.
	RemoveSecretField(keyringGroup, key string) error
	// ReadOPReference resolves an op:// reference to the secret value (via `op read`).
	ReadOPReference(ref string) (string, error)
}
