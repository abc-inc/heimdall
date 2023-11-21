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

package java

import "github.com/spf13/cobra"

func NewJavaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "java <subcommand>",
		Short: "Process Java-related files",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewJaCoCoCmd(),
		NewJUnitCmd(),
		NewLog4jCmd(),
		NewWebXMLCmd(),
	)

	return cmd
}
