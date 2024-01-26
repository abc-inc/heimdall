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

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/dependency_graph.go ../../docs

package github

import (
	"reflect"

	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewDepGraphCmd() *cobra.Command {
	cfg := newGHCfg()
	cmd := &cobra.Command{
		Use:   "dependency-graph",
		Short: "Handles communication with the dependency graph related methods.",
		Args:  cobra.ExactArgs(0),
	}

	ops := []string{
		"sbom",
	}

	svcTyp := reflect.TypeOf(&github.DependencyGraphService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execDepGraph, ops)...)
	for _, sub := range cmd.Commands() {
		addRepoFlags(cfg, sub)
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execDepGraph(cfg *ghCfg, cmd *cobra.Command) (a any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.DependencyGraph

	switch cmd.Name() {
	case "sbom":
		a, _, err = svc.GetSBOM(getCtx(cfg), cfg.owner, cfg.repo)
	default:
		panic("Unsupported operation: " + cmd.Name())
	}
	return
}
