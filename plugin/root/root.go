// Copyright 2025 The Heimdall authors
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

package root

import (
	"os"
	"path/filepath"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/docs"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/artifactory"
	"github.com/abc-inc/heimdall/plugin/confluence"
	"github.com/abc-inc/heimdall/plugin/cyclonedx"
	"github.com/abc-inc/heimdall/plugin/docker"
	"github.com/abc-inc/heimdall/plugin/echo"
	"github.com/abc-inc/heimdall/plugin/eval"
	"github.com/abc-inc/heimdall/plugin/example"
	"github.com/abc-inc/heimdall/plugin/git"
	"github.com/abc-inc/heimdall/plugin/golang"
	"github.com/abc-inc/heimdall/plugin/html"
	"github.com/abc-inc/heimdall/plugin/java"
	"github.com/abc-inc/heimdall/plugin/jira"
	"github.com/abc-inc/heimdall/plugin/keyring"
	"github.com/abc-inc/heimdall/plugin/parse"
	"github.com/abc-inc/heimdall/plugin/ssh"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func GetRootCmd() *cobra.Command {
	if rootCmd != nil {
		return rootCmd
	}

	rootCmd = &cobra.Command{
		Use:   "heimdall <command>",
		Short: "Perform compliance checks and report evidences",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cli.SetFormat(map[string]any{
				"jq":     cmd.Flag("jq"),
				"output": cmd.Flag("output"),
				"pretty": cmd.Flag("pretty"),
				"query":  cmd.Flag("query")},
			)
		},
	}

	rootCmd.AddGroup(
		&cobra.Group{ID: cli.FileGroup, Title: cli.FileGroup + ":"},
		&cobra.Group{ID: cli.HeimdallGroup, Title: cli.HeimdallGroup + ":"},
		&cobra.Group{ID: cli.MiscGroup, Title: cli.MiscGroup + ":"},
		&cobra.Group{ID: cli.ServiceGroup, Title: cli.ServiceGroup + ":"},
		&cobra.Group{ID: cli.SoftwareGroup, Title: cli.SoftwareGroup + ":"},
	)

	rootCmd.AddCommand(
		artifactory.NewArtifactoryCmd(),
		confluence.NewConfluenceCmd(),
		cyclonedx.NewCycloneDXCmd(),
		docker.NewDockerCmd(),
		echo.NewEchoCmd(),
		example.NewExampleCmd(),
		eval.NewEvalCmd(),
		git.NewGitCmd(),
		golang.NewGoCmd(),
		html.NewHTMLCmd(),
		java.NewJavaCmd(),
		jira.NewJiraCmd(),
		keyring.NewKeyringCmd(),
		parse.NewParseCmd(),
		ssh.NewSSHCmd(),
	)

	for _, t := range docs.Topics {
		rootCmd.AddCommand(docs.NewHelpTopicCmd(t))
	}

	rootCmd.SetHelpFunc(cli.RootHelpFunc)

	cfgDir := filepath.Join(internal.Must(os.UserConfigDir()), "heimdall")
	rootCmd.PersistentFlags().String("config", cfgDir, "Location of the Heimdall config directory")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error, fatal)")
	return rootCmd
}
