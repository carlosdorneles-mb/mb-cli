package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("update_repo", func(fl validator.FieldLevel) bool {
		return validateUpdateRepo(fl.Field().String()) == nil
	})
}

// ValidateUpdateRepo checks owner/repo style (at least two non-empty path segments).
// Exported for tests.
func ValidateUpdateRepo(s string) error {
	return validateUpdateRepo(s)
}

func validateUpdateRepo(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("cannot be empty")
	}
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return errors.New("expected owner/repo (e.g. org/repo)")
	}
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			return errors.New("invalid owner/repo")
		}
	}
	return nil
}

// validateFileConfig runs struct validation and returns an error suitable for config.yaml context.
func validateFileConfig(cfg *fileConfig) error {
	if err := validate.Struct(cfg); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			var msgs []string
			for _, e := range errs {
				// e.Field() is "DocsURL" or "UpdateRepo"; e.Tag() is "http_url" or "update_repo"
				msgs = append(
					msgs,
					fmt.Sprintf(
						"%s: %s",
						yamlKeyForField(e.Field()),
						tagToMessage(e.Tag(), e.Param()),
					),
				)
			}
			return fmt.Errorf("%s", strings.Join(msgs, "; "))
		}
		return err
	}
	return nil
}

func yamlKeyForField(field string) string {
	switch field {
	case "DocsURL":
		return "docs_url"
	case "UpdateRepo":
		return "update_repo"
	default:
		return strings.ToLower(field)
	}
}

func tagToMessage(tag, _ string) string {
	switch tag {
	case "http_url":
		return "must be an absolute http or https URL"
	case "update_repo":
		return "expected owner/repo (e.g. org/repo)"
	default:
		return "invalid"
	}
}
