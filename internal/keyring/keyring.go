// Package keyring provides a thin wrapper around the system keyring for MB CLI secrets.
// Service is fixed as keyringService; item id is "group:key" (e.g. "default:API_TOKEN").
package keyring

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const keyringService = "mb-cli"

func itemID(group, envKey string) string {
	if group == "" {
		group = "default"
	}
	return group + ":" + envKey
}

// Set stores a secret for the given group and key.
func Set(group, envKey, value string) error {
	if strings.Contains(envKey, ":") {
		return fmt.Errorf("key não pode conter ':'")
	}
	return keyring.Set(keyringService, itemID(group, envKey), value)
}

// Get retrieves a secret for the given group and key.
func Get(group, envKey string) (string, error) {
	return keyring.Get(keyringService, itemID(group, envKey))
}

// Delete removes a secret for the given group and key.
func Delete(group, envKey string) error {
	return keyring.Delete(keyringService, itemID(group, envKey))
}
