package runtimeflags

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"mb/internal/deps"
)

// FlagSpec defines a single runtime flag. This is the single source of truth
// for all MB global flags. When adding a flag here, update RegisterRuntimePersistentFlags
// and ParseLeadingRuntimeFlags accordingly.
type FlagSpec struct {
	Long         string
	Short        string
	Help         string
	SetFromPflag func(fs *pflag.FlagSet, rt *deps.RuntimeConfig)
	Parse        func(rt *deps.RuntimeConfig, args []string, i int) (newI int, err error)
}

// AllFlagSpecs is the authoritative list of MB runtime flags.
var AllFlagSpecs = []FlagSpec{
	{
		Long:  "verbose",
		Short: "v",
		Help:  "Ativa logs verbosos",
		SetFromPflag: func(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
			fs.BoolVarP(&rt.Verbose, "verbose", "v", false, "Ativa logs verbosos")
		},
		Parse: func(rt *deps.RuntimeConfig, args []string, i int) (int, error) {
			rt.Verbose = true
			return i + 1, nil
		},
	},
	{
		Long:  "quiet",
		Short: "q",
		Help:  "Não exibir nenhuma mensagem",
		SetFromPflag: func(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
			fs.BoolVarP(&rt.Quiet, "quiet", "q", false, "Não exibir nenhuma mensagem")
		},
		Parse: func(rt *deps.RuntimeConfig, args []string, i int) (int, error) {
			rt.Quiet = true
			return i + 1, nil
		},
	},
	{
		Long: "env-file",
		Help: "Carrega as váriaveis de um arquivo específico. Ex.: .env.local",
		SetFromPflag: func(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
			fs.StringVar(
				&rt.EnvFilePath,
				"env-file",
				"",
				"Carrega as váriaveis de um arquivo específico. Ex.: .env.local",
			)
		},
		Parse: func(rt *deps.RuntimeConfig, args []string, i int) (int, error) {
			val, consumed := parseValueFlag(args, i, "--env-file")
			if consumed == 0 {
				return 0, fmt.Errorf("flag needs an argument: --env-file")
			}
			rt.EnvFilePath = val
			return consumed, nil
		},
	},
	{
		Long: "env-vault",
		Help: "Carrega as variáveis de um vault específico. Ex.: staging",
		SetFromPflag: func(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
			fs.StringVar(
				&rt.EnvVault,
				"env-vault",
				"",
				"Carrega as variáveis de um vault específico. Ex.: staging",
			)
		},
		Parse: func(rt *deps.RuntimeConfig, args []string, i int) (int, error) {
			val, consumed := parseValueFlag(args, i, "--env-vault")
			if consumed == 0 {
				return 0, fmt.Errorf("flag needs an argument: --env-vault")
			}
			rt.EnvVault = val
			return consumed, nil
		},
	},
	{
		Long:  "env",
		Short: "e",
		Help:  "Define variável na execução do processo atual. Ex.: KEY=VALUE",
		SetFromPflag: func(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
			fs.StringArrayVarP(
				&rt.InlineEnvValues,
				"env",
				"e",
				nil,
				"Define variável na execução do processo atual. Ex.: KEY=VALUE",
			)
		},
		Parse: func(rt *deps.RuntimeConfig, args []string, i int) (int, error) {
			val, consumed := parseValueFlag(args, i, "-e", "--env")
			if consumed == 0 {
				return 0, fmt.Errorf("flag needs an argument: %q", args[i])
			}
			rt.InlineEnvValues = append(rt.InlineEnvValues, val)
			return consumed, nil
		},
	},
}

// parseValueFlag handles both --flag value and --flag=value forms.
// Returns (value, nextIndex). If not consumed, returns ("", 0).
func parseValueFlag(args []string, i int, longNames ...string) (string, int) {
	arg := args[i]

	// Check for --name=value form
	for _, name := range longNames {
		if prefix := name + "="; strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):], i + 1
		}
	}

	// Short form -e=value
	if strings.HasPrefix(arg, "-e=") {
		return arg[3:], i + 1
	}

	// --name value form
	if i+1 < len(args) {
		return args[i+1], i + 2
	}
	return "", 0
}

// findSpecByArg finds the FlagSpec that matches a flag argument.
func findSpecByArg(arg string) *FlagSpec {
	if strings.HasPrefix(arg, "--") {
		name := strings.SplitN(arg[2:], "=", 2)[0]
		for i := range AllFlagSpecs {
			if AllFlagSpecs[i].Long == name {
				return &AllFlagSpecs[i]
			}
		}
	} else if strings.HasPrefix(arg, "-") {
		name := arg[1:]
		// Could be short or long form
		for i := range AllFlagSpecs {
			if AllFlagSpecs[i].Short == name || AllFlagSpecs[i].Long == name {
				return &AllFlagSpecs[i]
			}
		}
	}
	return nil
}

// RegisterRuntimePersistentFlags binds the MB global runtime/env flags used by the root command.
// All flag definitions come from AllFlagSpecs.
func RegisterRuntimePersistentFlags(fs *pflag.FlagSet, rt *deps.RuntimeConfig) {
	for _, spec := range AllFlagSpecs {
		spec.SetFromPflag(fs, rt)
	}
}

// ParseLeadingRuntimeFlags parses MB global flags from the start of args (before the subprocess name),
// merges them into rt, and returns the remaining tokens for exec. Stops at the first non-flag token
// (or "--", or a lone "-") so flags meant for the child stay after the executable name.
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

		spec := findSpecByArg(arg)
		if spec == nil {
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}

		newI, err := spec.Parse(&peel, args, i)
		if err != nil {
			return nil, err
		}
		i = newI
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
