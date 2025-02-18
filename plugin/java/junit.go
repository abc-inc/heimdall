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

//go:build !no_java

package java

import (
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/joshdk/go-junit"
	"github.com/spf13/cobra"
)

type junitCfg struct {
	cli.OutCfg
	files      []string
	summary    bool
	withOutput bool
}

func NewJUnitCmd() *cobra.Command {
	cfg := junitCfg{}
	cmd := &cobra.Command{
		Use:   "junit",
		Short: "Parse and aggregate JUnit test reports",
		Example: heredoc.Doc(`
			heimdall java junit --summary --output text --jq ".[] | .totals.passed / .totals.tests" reports/TEST-*.xml
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.files = args
			cli.Fmtln(readJUnit(cfg))
		},
	}

	cmd.Flags().BoolVarP(&cfg.summary, "summary", "s", false, "Aggregate the report")
	cmd.Flags().BoolVar(&cfg.withOutput, "include-output", false, "Include standard output and standard error")

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func readJUnit(cfg junitCfg) []junit.Suite {
	ss := loadSuites(cfg.files)
	if cfg.summary {
		ss = []junit.Suite{{Suites: ss}}
		ss[0].Aggregate()
	}

	if !cfg.withOutput {
		for i := range ss {
			removeOutput(&ss[i])
		}
	}
	return ss
}

// loadSuites loads all files, and files in the given directories, recursively.
func loadSuites(names []string) (ss []junit.Suite) {
	for _, n := range names {
		if stat := internal.Must(os.Stat(n)); stat.IsDir() {
			ss = append(ss, internal.Must(junit.IngestDir(n))...)
		} else {
			ss = append(ss, internal.Must(junit.IngestFile(n))...)
		}
	}
	return
}

func removeOutput(s *junit.Suite) {
	s.SystemOut, s.SystemErr = "", ""
	for _, t := range s.Tests {
		t.SystemOut, t.SystemErr = "", ""
	}
	for _, c := range s.Suites {
		removeOutput(&c)
	}
}
