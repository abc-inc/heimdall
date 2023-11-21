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
	"encoding/csv"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type CovRec struct {
	Group     string `json:"group"`
	Pkg       string `json:"pkg"`
	Class     string `json:"class"`
	InstrMis  int    `json:"instruction_missed"`
	InstrCov  int    `json:"instruction_covered"`
	BranchMis int    `json:"branch_missed"`
	BranchCov int    `json:"branch_covered"`
	LineMis   int    `json:"line_missed"`
	LineCov   int    `json:"line_covered"`
	ComplMis  int    `json:"complexity_missed"`
	ComplCov  int    `json:"complexity_covered"`
	MethMis   int    `json:"method_missed"`
	MethCov   int    `json:"method_covered"`
}

type jaCoCoCfg struct {
	console.OutCfg
	file    string
	exclude string
	summary bool
}

func NewJaCoCoCmd() *cobra.Command {
	cfg := jaCoCoCfg{}
	cmd := &cobra.Command{
		Use:   "jacoco",
		Short: "Parse and aggregate Java code coverage reports",
		Example: heredoc.Doc(`
			heimdall java jacoco -f jacoco.csv --summary --output text --jq ".line_covered / (.line_covered + .line_missed)"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			printJaCoCo(cfg)
		},
	}

	cmd.Flags().BoolVarP(&cfg.summary, "summary", "s", false, "Aggregate the report")
	cmd.Flags().StringVarP(&cfg.exclude, "exclude", "x", "generated", "Packages to exclude")

	console.AddFileFlag(cmd, &cfg.file, "Path to the JaCoCo CSV file or URL")
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagFilename("file", "csv"))
	return cmd
}

func printJaCoCo(cfg jaCoCoCfg) {
	crs := processJaCoCo(cfg)
	if cfg.summary && len(crs) == 1 {
		console.Fmtln(crs[0])
	} else {
		console.Fmtln(crs)
	}
}

func processJaCoCo(cfg jaCoCoCfg) (crs []CovRec) {
	var test func(rec []string) bool
	if len(cfg.exclude) > 0 {
		test = func(rec []string) bool {
			return len(rec) > 1 && !strings.Contains(rec[1], cfg.exclude)
		}
	}

	for _, p := range strings.Split(cfg.file, ":") {
		for _, f := range internal.Must(filepath.Glob(p)) {
			c := jaCoCoCfg{file: f, exclude: cfg.exclude, summary: cfg.summary}
			crs = append(crs, loadJaCoCoCSV(c.file, test)...)
		}
	}
	if len(crs) == 0 {
		log.Fatal().Str("file", cfg.file).Msg("Cannot load JaCoCo report or file does not contain any matching lines")
	}

	if cfg.summary {
		crs = []CovRec{aggregateJaCoCo(crs...)}
	}
	return
}

func loadJaCoCoCSV(uri string, test func([]string) bool) (crs []CovRec) {
	f := internal.Must(res.Open(uri))
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	_ = internal.Must(r.Read())
	for {
		rec, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		internal.MustNoErr(err)
		if test != nil && !test(rec) {
			continue
		}

		crs = append(crs, CovRec{
			Group:     rec[0],
			Pkg:       rec[1],
			Class:     rec[2],
			InstrMis:  internal.Must(strconv.Atoi(rec[3])),
			InstrCov:  internal.Must(strconv.Atoi(rec[4])),
			BranchMis: internal.Must(strconv.Atoi(rec[5])),
			BranchCov: internal.Must(strconv.Atoi(rec[6])),
			LineMis:   internal.Must(strconv.Atoi(rec[7])),
			LineCov:   internal.Must(strconv.Atoi(rec[8])),
			ComplMis:  internal.Must(strconv.Atoi(rec[9])),
			ComplCov:  internal.Must(strconv.Atoi(rec[10])),
			MethMis:   internal.Must(strconv.Atoi(rec[11])),
			MethCov:   internal.Must(strconv.Atoi(rec[12])),
		})
	}
	return
}

func aggregateJaCoCo(crs ...CovRec) (tot CovRec) {
	for _, cr := range crs {
		tot.InstrMis += cr.InstrMis
		tot.InstrCov += cr.InstrCov
		tot.BranchMis += cr.BranchMis
		tot.BranchCov += cr.BranchCov
		tot.LineMis += cr.LineMis
		tot.LineCov += cr.LineCov
		tot.ComplMis += cr.ComplMis
		tot.ComplCov += cr.ComplCov
		tot.MethMis += cr.MethMis
		tot.MethCov += cr.MethCov
	}
	return
}
