package runtimeflags

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"mb/internal/deps"
)

// RegisterRuntimePersistentFlags binds the MB global runtime/env flags used by the root command
// and by mb run (leading prefix parse). Keep definitions in sync when adding flags.
func RegisterRuntimePersistentFlags(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
	fs.BoolVarP(&rt.Verbose, "verbose", "v", false, "Ativa logs verbosos")
	fs.BoolVarP(&rt.Quiet, "quiet", "q", false, "Não exibir nenhuma mensagem")
	fs.StringVar(
		&rt.EnvFilePath,
		"env-file",
		"",
		"Carrega as váriaveis de um arquivo específico. Ex.: .env.local",
	)
	fs.StringVar(
		&rt.EnvVault,
		"env-vault",
		"",
		"Carrega as variáveis de um vault específico. Ex.: staging",
	)
	fs.StringArrayVarP(
		&rt.InlineEnvValues,
		"env",
		"e",
		nil,
		"Define variável na execução do processo atual. Ex.: KEY=VALUE",
	)
}

// ParseLeadingRuntimeFlags parses MB global flags from the start of args (before the subprocess name),
// merges them into rt, and returns the remaining tokens for exec. Stops at the first non-flag token
// (or "--", or a lone "-") so flags meant for the child stay after the executable name (e.g. mb run grep -r).
func ParseLeadingRuntimeFlags(rt *deps.RuntimeConfig, args []string) (rest []string, err error) {
	var peel deps.RuntimeConfig
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			i++
			break
		}
		if arg == "-" {
			break
		}
		if !strings.HasPrefix(arg, "-") {
			break
		}
		switch arg {
		case "-v", "--verbose":
			peel.Verbose = true
			i++
		case "-q", "--quiet":
			peel.Quiet = true
			i++
		case "-e", "--env":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %q", arg)
			}
			peel.InlineEnvValues = append(peel.InlineEnvValues, args[i+1])
			i += 2
		default:
			if v, ok := strings.CutPrefix(arg, "-e="); ok {
				peel.InlineEnvValues = append(peel.InlineEnvValues, v)
				i++
				break
			}
			if v, ok := strings.CutPrefix(arg, "--env="); ok {
				peel.InlineEnvValues = append(peel.InlineEnvValues, v)
				i++
				break
			}
			if arg == "--env-file" {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag needs an argument: %q", arg)
				}
				peel.EnvFilePath = args[i+1]
				i += 2
				break
			}
			if v, ok := strings.CutPrefix(arg, "--env-file="); ok {
				peel.EnvFilePath = v
				i++
				break
			}
			if arg == "--env-vault" {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag needs an argument: %q", arg)
				}
				peel.EnvVault = args[i+1]
				i += 2
				break
			}
			if v, ok := strings.CutPrefix(arg, "--env-vault="); ok {
				peel.EnvVault = v
				i++
				break
			}
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}

	rt.Verbose = rt.Verbose || peel.Verbose
	rt.Quiet = rt.Quiet || peel.Quiet
	if peel.EnvFilePath != "" {
		rt.EnvFilePath = peel.EnvFilePath
	}
	if peel.EnvVault != "" {
		rt.EnvVault = peel.EnvVault
	}
	rt.InlineEnvValues = append(rt.InlineEnvValues, peel.InlineEnvValues...)
	return args[i:], nil
}
