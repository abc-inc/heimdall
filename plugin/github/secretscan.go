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

//go:build !no_github

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/secret_scanning.go ../../docs

package github

import (
	"reflect"
	"strings"

	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewSecretScanCmd() *cobra.Command {
	cfg := newGHCfg()
	cmd := &cobra.Command{
		Use:   "secret-scanning",
		Short: "Handles communication with the secret scanning related methods.",
		Args:  cobra.ExactArgs(0),
	}

	ops := []string{
		"alert",
		"alerts-for-repo",
		"locations-for-alert",
	}

	svcTyp := reflect.TypeOf(&github.SecretScanningService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execSecretScan, ops)...)
	for _, sub := range cmd.Commands() {
		addRepoFlags(cfg, sub)
		if sub.Name() == "alerts-for-repo" {
			sub.Flags().StringVar(cfg.state, "state", *cfg.state, "State of the secret scanning alerts to list.")
			sub.Flags().StringSliceVar(&cfg.secretType, "secret-type", cfg.secretType, "A comma-separated list of secret types to return.")
			sub.Flags().StringSliceVar(&cfg.resolution, "resolution", cfg.resolution, "false_positive, wont_fix, revoked, pattern_edited, pattern_deleted or used_in_tests.")
		}
		if !strings.Contains(sub.Name(), "-") {
			addItemFlags(cfg, sub)
		}
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execSecretScan(cfg *ghCfg, cmd *cobra.Command) (a any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.SecretScanning

	switch cmd.Name() {
	case "alert":
		a, _, err = svc.GetAlert(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "alerts-for-repo":
		a, _, err = svc.ListAlertsForRepo(getCtx(cfg), cfg.owner, cfg.repo, &github.SecretScanningAlertListOptions{
			State:             *cfg.state,
			SecretType:        strings.Join(cfg.secretType, ","),
			Resolution:        strings.Join(cfg.resolution, ","),
			ListCursorOptions: github.ListCursorOptions{},
			ListOptions:       github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "locations-for-alert":
		a, _, err = svc.ListLocationsForAlert(getCtx(cfg), cfg.owner, cfg.repo, cfg.id,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	default:
		panic("Unsupported operation: " + cmd.Name())
	}
	return
}
