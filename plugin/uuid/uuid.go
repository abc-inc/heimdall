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

//go:build !no_uuid

package uuid

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewUUIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uuid",
		Short: "Generate a Universally Unique Identifier",
		Example: heredoc.Doc(`
			heimdall uuid"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = cli.Msg(uuidgen() + "\n")
		},
	}

	return cmd
}

func uuidgen() string {
	return uuid.NewString()
}
