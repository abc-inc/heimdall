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

package git

import (
	"reflect"

	"github.com/abc-inc/heimdall/console"
	"github.com/spf13/cobra"
)

func NewGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git <subcommand>",
		Short: "Query information from Git repositories",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewCommitsCmd(),
	)

	return cmd
}

func init() {
	console.EnableMarshalling(reflect.TypeOf(Commit{}))
}
