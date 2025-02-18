// Copyright 2023 The Heimdall authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	_ "github.com/abc-inc/heimdall/plugin/github"
	"github.com/abc-inc/heimdall/plugin/root"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const version = "0.0"

func main() {
	cli.Version = version

	rootCmd := root.GetRootCmd()
	rootCmd.AddCommand(versionCmd())
	cobra.OnInitialize(func() { initConfig(rootCmd) })

	if strings.Contains(os.Args[0], "___go_build_") {
		os.Args = os.Args[1:]
	}
	exe := filepath.Base(os.Args[0])
	if strings.HasPrefix(exe, "hd-") || strings.HasPrefix(exe, "heimdall-") {
		parts := strings.Split(exe, "-")[1:]
		os.Args = slices.Insert(os.Args, 1, parts...)
	} else if !strings.HasPrefix(exe, "hd") && !strings.HasPrefix(exe, "heimdall") {
		parts := strings.Split(exe, "-")
		os.Args = slices.Insert(os.Args, 1, parts...)
	}

	if err := rootCmd.Execute(); err != nil &&
		!strings.HasPrefix(err.Error(), "unknown command ") &&
		!strings.HasPrefix(err.Error(), "unknown shorthand flag: ") &&
		!strings.HasPrefix(err.Error(), "no flags provided") &&
		!strings.HasPrefix(err.Error(), "unknown flag: ") &&
		!strings.HasPrefix(err.Error(), "required flag(s) ") &&
		!strings.HasPrefix(err.Error(), "accepts 1 arg(s),") {

		internal.MustNoErr(err)
	}
}

// initConfig reads the config file and environment variables, if set.
func initConfig(rootCmd *cobra.Command) {
	// Attempt to load environment variables from files.
	cfgDir := internal.Must(rootCmd.PersistentFlags().GetString("config"))
	_ = godotenv.Overload("/etc/heimdall/heimdall.env")
	_ = godotenv.Load(".heimdall.env")
	_ = godotenv.Load(internal.Must(filepath.Glob(filepath.Join(cfgDir, "*.env")))...)
	_ = godotenv.Load("/etc/heimdall/default.env")

	// Import existing environment variables.
	envPrefix := "HEIMDALL"
	var envRepl = make([]string, 2*len(cli.StyleProps))
	for i, p := range cli.StyleProps {
		if envRepl[2*i] == "" {
			envRepl[2*i], envRepl[2*i+1] = envPrefix+"_"+strings.ReplaceAll(p, "_", "-"), p
		}
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(envRepl...))
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	// Search config with name "heimdall" (without extension).
	viper.AddConfigPath(cfgDir)
	viper.SetConfigName("heimdall")

	// If a config file is found, read it.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Loaded config")
	}

	postInitCommands(rootCmd)

	l := internal.Must(rootCmd.PersistentFlags().GetString("log-level"))
	zerolog.SetGlobalLevel(internal.Must(zerolog.ParseLevel(l)))
}

func postInitCommands(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		presetFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands()...)
		}
	}
}

// presetFlags sets the values of flags, where the value is provided by a
// corresponding environment variable.
func presetFlags(cmd *cobra.Command) {
	internal.MustNoErr(viper.BindPFlags(cmd.Flags()))
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			internal.MustNoErr(cmd.Flags().Set(f.Name, viper.GetString(f.Name)))
			internal.MustOkMsgf[any](nil, cmd.Flags().Changed(f.Name), "flag %s does not exist", f.Name)
		}
	})
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Show Heimdall version",
		GroupID: cli.HeimdallGroup,
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			t, rev := "", ""
			if info, ok := debug.ReadBuildInfo(); ok {
				for _, s := range info.Settings {
					if s.Key == "vcs.revision" {
						rev = s.Value
					} else if s.Key == "vcs.time" {
						t = s.Value
					}
				}
			}

			if len(rev) > 10 {
				_, _ = cli.Msg(fmt.Sprintf("Heimdall version %s (Git commit: %s, Date: %s)\n", version, rev[0:10], t))
			} else {
				_, _ = cli.Msg(fmt.Sprintf("Heimdall version %s\n", version))
			}
		},
	}
}
