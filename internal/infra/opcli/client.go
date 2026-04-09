package opcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"mb/internal/ports"
)

// ErrNotInstalled is returned when the `op` binary is not on PATH.
var ErrNotInstalled = errors.New("comando op (1Password CLI) não encontrado no PATH")

// Client runs the 1Password CLI. It is safe for concurrent use only if the underlying
// `op` session is; the MB CLI invokes one command at a time per process.
type Client struct{}

// New returns a Client using the system `op` binary.
func New() *Client {
	return &Client{}
}

var _ ports.OnePasswordEnv = (*Client)(nil)

// EnsureAvailable returns ErrNotInstalled if `op` is missing.
func (c *Client) EnsureAvailable() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return fmt.Errorf(
			"%w: instale com «mb tools 1password-cli» ou veja https://developer.1password.com/docs/cli/: %w",
			ErrNotInstalled,
			err,
		)
	}
	return nil
}

func itemTitle(keyringGroup string) string {
	return "mb-cli env / " + keyringGroup
}

// PutSecret implements [ports.OnePasswordEnv].
func (c *Client) PutSecret(keyringGroup, key, value string) (string, error) {
	if err := c.EnsureAvailable(); err != nil {
		return "", err
	}
	ctx := context.Background()
	title := itemTitle(keyringGroup)

	raw, getErr := opItemGetJSON(ctx, title)
	if getErr != nil {
		if !isItemNotFound(getErr) {
			return "", getErr
		}
		if err := opItemCreatePassword(ctx, title); err != nil {
			return "", fmt.Errorf("criar item 1Password %q: %w", title, err)
		}
		raw, getErr = opItemGetJSON(ctx, title)
		if getErr != nil {
			return "", getErr
		}
	}

	updated, err := upsertConcealedFieldInItemJSON(raw, key, value)
	if err != nil {
		return "", err
	}
	if err := opItemEditStdin(ctx, title, updated); err != nil {
		return "", fmt.Errorf("atualizar segredo no 1Password: %w", err)
	}

	raw2, err := opItemGetJSON(ctx, title)
	if err != nil {
		return "", err
	}
	return fieldReferenceFromItemJSON(raw2, key)
}

// RemoveSecretField implements [ports.OnePasswordEnv].
func (c *Client) RemoveSecretField(keyringGroup, key string) error {
	if err := c.EnsureAvailable(); err != nil {
		return err
	}
	ctx := context.Background()
	title := itemTitle(keyringGroup)
	raw, err := opItemGetJSON(ctx, title)
	if err != nil {
		if isItemNotFound(err) {
			return nil
		}
		return err
	}
	updated, err := removeConcealedFieldFromItemJSON(raw, key)
	if err != nil {
		return err
	}
	return opItemEditStdin(ctx, title, updated)
}

// ReadOPReference implements [ports.OnePasswordEnv].
func (c *Client) ReadOPReference(ref string) (string, error) {
	if err := c.EnsureAvailable(); err != nil {
		return "", err
	}
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "op", "read", ref)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("op read: %w: %s", err, msg)
		}
		return "", fmt.Errorf("op read: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func opItemGetJSON(ctx context.Context, title string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "op", "item", "get", title, "--format", "json")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		combined := stderr.String()
		if len(out) > 0 {
			combined += string(out)
		}
		return out, fmt.Errorf("op item get %q: %w\n%s", title, err, strings.TrimSpace(combined))
	}
	return out, nil
}

func isItemNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "could not find") ||
		strings.Contains(s, "is not an item") ||
		strings.Contains(s, "isn't an item") ||
		strings.Contains(s, "doesn't exist") ||
		strings.Contains(s, "not found") ||
		strings.Contains(s, "não foi possível encontrar") // future localized
}

func opItemCreatePassword(ctx context.Context, title string) error {
	cmd := exec.CommandContext(ctx, "op", "item", "create",
		"--category=Password",
		"--title", title,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(out) > 0 {
			msg += "\n" + strings.TrimSpace(string(out))
		}
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(msg))
	}
	return nil
}

func opItemEditStdin(ctx context.Context, title string, jsonData []byte) error {
	cmd := exec.CommandContext(ctx, "op", "item", "edit", title)
	cmd.Stdin = bytes.NewReader(jsonData)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(out) > 0 {
			msg += "\n" + strings.TrimSpace(string(out))
		}
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(msg))
	}
	return nil
}
