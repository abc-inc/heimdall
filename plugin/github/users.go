// Copyright 2024 The Heimdall authors
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

//go:build !no_github

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/users\*.go ../../docs

package github

import (
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewUsersCmd() *cobra.Command {
	cfg := newGHCfg()
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Handles communication with the user related methods.",
		Example: heredoc.Doc(`
			heimdall github users list
		`),
		Args: cobra.ExactArgs(0),
	}

	ops := []string{
		"all",
		"get",
	}

	svcTyp := reflect.TypeOf(&github.UsersService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execUsers, ops)...)

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execUsers(cfg *ghCfg, cmd *cobra.Command) (x any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.Users

	switch cmd.Name() {
	case "all":
		x, _, err = svc.ListAll(getCtx(cfg), nil)
	case "get":
		x, _, err = svc.Get(getCtx(cfg), cfg.user)
	}
	return
}
