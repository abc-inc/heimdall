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

//go:build !no_docker

package docker

import (
	"github.com/abc-inc/heimdall/cli"
	"github.com/spf13/cobra"
)

type dockerCfg struct {
	cli.OutCfg
	file        string
	invertMatch bool
}

func NewDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docker <subcommand>",
		Short:   "Process Docker-related files",
		GroupID: cli.SoftwareGroup,
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewContextCmd(),
		NewFileCmd(),
		NewIgnoreCmd(),
	)

	return cmd
}
