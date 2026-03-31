package keyring

import "mb/internal/ports"

// SystemKeyring implements ports.SecretStore via the OS keyring.
type SystemKeyring struct{}

func (SystemKeyring) Set(group, envKey, value string) error {
	return Set(group, envKey, value)
}

func (SystemKeyring) Get(group, envKey string) (string, error) {
	return Get(group, envKey)
}

func (SystemKeyring) Delete(group, envKey string) error {
	return Delete(group, envKey)
}

var _ ports.SecretStore = SystemKeyring{}
