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

package keyring

import (
	"os"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

func NewGetCmd() *cobra.Command {
	cfg := keyringCfg{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve secrets from the system keyring",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			_ = internal.Must(console.Msg(get(cfg)))
			if isatty.IsTerminal(os.Stdout.Fd()) {
				_ = internal.Must(console.Msg("\n"))
			}
		},
	}

	cmd.Flags().StringVarP(&cfg.service, "service", "s", cfg.service, "Service / application name")
	cmd.Flags().StringVarP(&cfg.user, "username", "u", cfg.user, "Username")

	internal.MustNoErr(cmd.MarkFlagRequired("service"))
	internal.MustNoErr(cmd.MarkFlagRequired("username"))
	return cmd
}

func get(cfg keyringCfg) string {
	return internal.Must(keyring.Get(cfg.service, cfg.user))
}
