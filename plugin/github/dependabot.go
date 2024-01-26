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

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/dependabot\*.go ../../docs

package github

import (
	"reflect"

	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewDependabotCmd() *cobra.Command {
	cfg := newGHCfg()
	cmd := &cobra.Command{
		Use:   "dependabot",
		Short: "Handles communication with the Dependabot related methods.",
		Args:  cobra.ExactArgs(0),
	}

	ops := []string{
		"repo-alert",
		"repo-alerts",
		"repo-secret",
		"repo-secrets",
	}

	svcTyp := reflect.TypeOf(&github.DependabotService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execDependabot, ops)...)
	for _, sub := range cmd.Commands() {
		addRepoFlags(cfg, sub)

		switch sub.Name() {
		case "repo-alert":
			addItemFlags(cfg, sub)
		case "repo-alerts":
			sub.Flags().StringVar(cfg.state, "state", *cfg.state, "Comma-separated list of states (auto_dismissed, dismissed, fixed, open)")
			sub.Flags().StringVar(cfg.severity, "severity", *cfg.severity, "Comma-separated list of severities (low, medium, high, critical)")
			sub.Flags().StringVar(cfg.ecosystem, "ecosystem", *cfg.ecosystem, "Comma-separated list of ecosystems. If specified, only alerts for these ecosystems will be returned (composer, go, maven, npm, nuget, pip, pub, rubygems, rust)")
			sub.Flags().StringVar(cfg.pkg, "package", *cfg.pkg, "Comma-separated list of package names")
			sub.Flags().StringVar(cfg.scope, "scope", *cfg.scope, "Scope of the vulnerable dependency (development, runtime)")
			sub.Flags().StringVar(cfg.sort, "sort", *cfg.sort, "The property by which to sort the results (created, updated)")
			sub.Flags().StringVar(cfg.direction, "direction", *cfg.direction, "The direction to sort the results by (asc, desc)")
		case "repo-secret":
			sub.Flags().StringVar(&cfg.name, "secret-name", cfg.name, "Name of the secret")
		}
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execDependabot(cfg *ghCfg, cmd *cobra.Command) (a any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.Dependabot

	switch cmd.Name() {
	case "repo-alert":
		a, _, err = svc.GetRepoAlert(getCtx(cfg), cfg.owner, cfg.repo, int(cfg.id))
	case "repo-alerts":
		a, _, err = svc.ListRepoAlerts(getCtx(cfg), cfg.owner, cfg.repo, &github.ListAlertsOptions{
			State:             cfg.state,
			Severity:          cfg.severity,
			Ecosystem:         cfg.ecosystem,
			Package:           cfg.pkg,
			Scope:             cfg.scope,
			Sort:              cfg.sort,
			Direction:         cfg.direction,
			ListOptions:       github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
			ListCursorOptions: github.ListCursorOptions{},
		})
	case "repo-secret":
		a, _, err = svc.GetRepoSecret(getCtx(cfg), cfg.owner, cfg.repo, cfg.name)
	case "repo-secrets":
		a, _, err = svc.ListRepoSecrets(getCtx(cfg), cfg.owner, cfg.repo, &github.ListOptions{
			Page: cfg.page, PerPage: cfg.perPage})
	default:
		panic("Unsupported operation: " + cmd.Name())
	}
	return
}
