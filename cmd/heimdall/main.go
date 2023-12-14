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
	"strings"
	"time"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/docs"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/docker"
	"github.com/abc-inc/heimdall/plugin/echo"
	"github.com/abc-inc/heimdall/plugin/eval"
	"github.com/abc-inc/heimdall/plugin/example"
	"github.com/abc-inc/heimdall/plugin/format"
	"github.com/abc-inc/heimdall/plugin/git"
	"github.com/abc-inc/heimdall/plugin/github"
	"github.com/abc-inc/heimdall/plugin/golang"
	"github.com/abc-inc/heimdall/plugin/java"
	"github.com/abc-inc/heimdall/plugin/jira"
	"github.com/abc-inc/heimdall/plugin/json"
	"github.com/abc-inc/heimdall/plugin/keyring"
	"github.com/abc-inc/heimdall/plugin/properties"
	"github.com/abc-inc/heimdall/plugin/run"
	"github.com/abc-inc/heimdall/plugin/ssh"
	"github.com/abc-inc/heimdall/plugin/toml"
	"github.com/abc-inc/heimdall/plugin/xml"
	"github.com/abc-inc/heimdall/plugin/yaml"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const version = "0.0"

func main() {
	d := internal.Must(time.Parse(time.DateOnly, "2024-01-01"))
	if time.Now().After(d) {
		log.Warn().Msgf("This is an experimental build and stopped working after %s.", d.Format(time.DateOnly))
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "heimdall <command>",
		Short: "Perform compliance checks and report evidences",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			console.SetFormat(map[string]any{
				"jq":     cmd.Flag("jq"),
				"output": cmd.Flag("output"),
				"query":  cmd.Flag("query")},
			)
		},
	}

	rootCmd.AddCommand(
		docker.NewDockerCmd(),
		echo.NewEchoCmd(),
		example.NewExampleCmd(),
		eval.NewEvalCmd(),
		format.NewFormatCmd(),
		git.NewGitCmd(),
		github.NewGitHubCmd(),
		golang.NewGoCmd(),
		java.NewJavaCmd(),
		jira.NewJiraCmd(),
		json.NewJSONCmd(),
		keyring.NewKeyringCmd(),
		properties.NewPropertiesCmd(),
		run.NewRunCmd(),
		ssh.NewSSHCmd(),
		toml.NewTOMLCmd(),
		xml.NewXMLCmd(),
		yaml.NewYAMLCmd(),
		versionCmd(),
	)

	for _, t := range docs.Topics {
		rootCmd.AddCommand(docs.NewHelpTopicCmd(t))
	}

	rootCmd.SetHelpFunc(console.RootHelpFunc)

	cfgDir := filepath.Join(internal.Must(os.UserConfigDir()), "heimdall")
	rootCmd.PersistentFlags().String("config", cfgDir, "Location of the Heimdall config directory")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error, fatal)")

	cobra.OnInitialize(func() { initConfig(rootCmd) })

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
	var envRepl = make([]string, 2*len(console.StyleProps))
	for i, p := range console.StyleProps {
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
		if f.Name == "silent" {
			fmt.Println("FLAG: ", f.Name, f.Value, viper.GetString(f.Name))
		}
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			internal.MustNoErr(cmd.Flags().Set(f.Name, viper.GetString(f.Name)))
			internal.MustOkMsgf[any](nil, cmd.Flags().Changed(f.Name), "flag %s does not exist", f.Name)
		}
	})
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show Heimdall version",
		Args:  cobra.ExactArgs(0),
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
				fmt.Printf("Heimdall version %s (Git commit: %s, Date: %s)\n", version, rev[0:10], t)
			} else {
				fmt.Printf("Heimdall version %s\n", version)
			}
		},
	}
}
