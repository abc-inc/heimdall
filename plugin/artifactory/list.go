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

package artifactory

import (
	"slices"
	"strings"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

type listCfg struct {
	console.OutCfg
	pattern string
}

func NewListCmd() *cobra.Command {
	cfg := listCfg{}
	cmd := &cobra.Command{
		Use:   "list <subcommand>",
		Short: "List artifacts in Artifactory",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(find(cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.pattern, "pattern", "p", cfg.pattern, "List only assets that match a pattern")

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("pattern"))
	return cmd
}

func find(cfg listCfg) (its []*utils.ResultItem) {
	rtManager := newRtManager()

	params := services.NewSearchParams()
	params.Pattern = cfg.pattern
	params.Recursive = true
	params.SortBy = []string{"path"}
	params.SortOrder = "asc"

	fs := internal.Must(rtManager.SearchFiles(params))
	defer func() { _ = fs.Close() }()

	for item := new(utils.ResultItem); fs.NextRecord(item) == nil; item = new(utils.ResultItem) {
		its = append(its, item)
	}
	return sortVersion(its)
}

func sortVersion(its []*utils.ResultItem) []*utils.ResultItem {
	slices.SortFunc(its, func(a, b *utils.ResultItem) int {
		v1 := strings.ReplaceAll(a.Name, "b", "+")
		v2 := strings.ReplaceAll(b.Name, "b", "+")
		return semver.Compare("v"+v2, "v"+v1)
	})
	return its
}
