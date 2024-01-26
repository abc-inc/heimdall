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

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/code-scanning.go ../../docs

package github

import (
	"reflect"

	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewCodeScanCmd() *cobra.Command {
	var branch string
	cfg := newGHCfg()
	cfg.branch = &branch

	cmd := &cobra.Command{
		Use:   "code-scanning",
		Short: "Handles communication with the code scanning related methods.",
		Args:  cobra.ExactArgs(0),
	}

	ops := []string{
		"alert",
		"alert-instances",
		"alerts-for-repo",
		"analysis",
		"analyses-for-repo",
		"default-setup-configuration",
		"sarif",
	}

	svcTyp := reflect.TypeOf(&github.CodeScanningService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execCodeScan, ops)...)
	for _, sub := range cmd.Commands() {
		addRepoFlags(cfg, sub)

		switch sub.Name() {
		case "alert":
			addItemFlags(cfg, sub)
		case "alert-instances":
			addItemFlags(cfg, sub)
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "alerts-for-repo":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
			sub.Flags().StringVar(cfg.state, "state", *cfg.state, "Alert state")
			sub.Flags().StringVar(cfg.severity, "severity", *cfg.severity, "Severity")
		case "analyses-for-repo":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
			sub.Flags().StringVar(&cfg.sarifID, "sarif", cfg.sarifID, "SARIF ID obtained after uploading")
		case "analysis:":
			addItemFlags(cfg, sub)
		case "sarif":
			sub.Flags().StringVar(&cfg.sarifID, "sarif", cfg.sarifID, "SARIF ID obtained after uploading")
		}
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execCodeScan(cfg *ghCfg, cmd *cobra.Command) (a any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.CodeScanning

	if cmd.Flag("branch") != nil {
		branchOrDefault(cfg, cfg.client.Repositories)
	}

	switch cmd.Name() {
	case "alert":
		a, _, err = svc.GetAlert(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "alert-instances":
		a, _, err = svc.ListAlertInstances(getCtx(cfg), cfg.owner, cfg.repo, cfg.id, &github.AlertInstancesListOptions{
			Ref:         *cfg.branch,
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "alerts-for-repo":
		a, _, err = svc.ListAlertsForRepo(getCtx(cfg), cfg.owner, cfg.repo, &github.AlertListOptions{
			State:             *cfg.state,
			Ref:               *cfg.branch,
			Severity:          *cfg.severity,
			ToolName:          *cfg.toolName,
			ListCursorOptions: github.ListCursorOptions{},
			ListOptions:       github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "analyses-for-repo":
		a, _, err = svc.ListAnalysesForRepo(getCtx(cfg), cfg.owner, cfg.repo, &github.AnalysesListOptions{
			SarifID:     &cfg.sarifID,
			Ref:         cfg.branch,
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "analysis":
		a, _, err = svc.GetAnalysis(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "default-setup-configuration":
		a, _, err = svc.GetDefaultSetupConfiguration(getCtx(cfg), cfg.owner, cfg.repo)
	case "sarif":
		a, _, err = svc.GetSARIF(getCtx(cfg), cfg.owner, cfg.repo, cfg.sarifID)
	default:
		panic("Unsupported operation: " + cmd.Name())
	}
	return a, err
}
