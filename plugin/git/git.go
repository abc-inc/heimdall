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

//go:build !no_git

package git

import "github.com/spf13/cobra"

const envHelp = `
GIT_USERNAME  <USERNAME>
GIT_PASSWORD  <PASSWORD>
`

func NewGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "git <subcommand>",
		Short:       "Query information from Git repositories",
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"help:environment": envHelp},
	}

	cmd.AddCommand(
		NewCommitsCmd(),
	)

	return cmd
}
