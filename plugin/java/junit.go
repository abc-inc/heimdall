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

package java

import (
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/joshdk/go-junit"
	"github.com/spf13/cobra"
)

type junitCfg struct {
	console.OutCfg
	file       string
	summary    bool
	withOutput bool
}

func NewJUnitCmd() *cobra.Command {
	cfg := junitCfg{}
	cmd := &cobra.Command{
		Use:   "junit",
		Short: "Parse and aggregate JUnit test reports",
		Example: heredoc.Doc(`
			heimdall java junit -f "reports/TEST-*.xml" --summary --output text --jq ".[] | .totals.passed / .totals.tests"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(readJUnit(cfg))
		},
	}

	cmd.Flags().BoolVarP(&cfg.summary, "summary", "s", false, "Aggregate the report")
	cmd.Flags().BoolVar(&cfg.withOutput, "include-output", false, "Include standard output and standard error")

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	console.AddFileFlag(cmd, &cfg.file, "Path to the JUnit report directory or file")
	return cmd
}

func readJUnit(cfg junitCfg) []junit.Suite {
	ss := loadSuites(cfg.file)
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

// loadSuites recursively loads all files in the given directory, recursively.
// If a globing pattern is used, then it ingests only the matching files.
func loadSuites(dir string) []junit.Suite {
	if !strings.ContainsRune(dir, '*') {
		return internal.Must(junit.IngestDir(dir))
	}
	fs := internal.Must(filepath.Glob(dir))
	return internal.Must(junit.IngestFiles(fs))
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
