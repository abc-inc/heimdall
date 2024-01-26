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

//go:build !no_run

package run

import "github.com/spf13/cobra"

func NewRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "run",
		Short:              "Run a Heimdall task file",
		Run:                func(cmd *cobra.Command, args []string) { runTask() },
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}

	return cmd
}
