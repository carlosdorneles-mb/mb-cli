package env

import (
	"fmt"
	"strings"
)

func ParseInlinePairs(pairs []string) (map[string]string, error) {
	parsed := map[string]string{}
	for _, pair := range pairs {
		idx := strings.Index(pair, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --env value %q, expected KEY=VALUE", pair)
		}
		key := pair[:idx]
		value := pair[idx+1:]
		parsed[key] = value
	}
	return parsed, nil
}

func AsMap(values []string) map[string]string {
	out := map[string]string{}
	for _, value := range values {
		idx := strings.Index(value, "=")
		if idx <= 0 {
			continue
		}
		out[value[:idx]] = value[idx+1:]
	}
	return out
}

func AsList(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for key, value := range values {
		out = append(out, key+"="+value)
	}
	return out
}

func Merge(system []string, envFile map[string]string, cli map[string]string) []string {
	merged := AsMap(system)
	for key, value := range envFile {
		merged[key] = value
	}
	for key, value := range cli {
		merged[key] = value
	}
	return AsList(merged)
}
