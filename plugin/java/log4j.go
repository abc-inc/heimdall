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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/gobwas/glob"
	"github.com/google/log4jscanner/jar"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type log4jCfg struct {
	console.OutCfg
	skip []string
}

type jarFileRep struct {
	Path        string `json:"Path" yaml:"path"`
	*jar.Report `json:"" yaml:",inline"`
}

func NewLog4jCmd() *cobra.Command {
	cfg := log4jCfg{}
	cmd := &cobra.Command{
		Use:   "log4j",
		Short: "Scan the filesystem for log4j vulnerabilities",
		Example: heredoc.Doc(`
			heimdall java log4j --skip '**/wrapper/dists/gradle-*' ~/.gradle ~/.m2/repository
		`),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				args = append(args, ".")
			}
			reps := scan(cfg, args...)
			if len(reps) > 0 {
				console.Fmtln(reps)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringSliceVarP(&cfg.skip, "skip", "s", cfg.skip, "Glob pattern to skip when scanning (e.g. '/var/run/*'). May be provided multiple times.")
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func scan(cfg log4jCfg, dirs ...string) []jarFileRep {
	skipDirs := map[string]bool{
		".git":         true,
		".hg":          true,
		".idea":        true,
		".svn":         true,
		"node_modules": true,
	}

	var gs []glob.Glob
	for _, p := range cfg.skip {
		gs = append(gs, glob.MustCompile(p, filepath.Separator))
	}

	rs := make([]jarFileRep, 0)
	walker := jar.Walker{
		Rewrite: false,
		SkipDir: func(path string, de fs.DirEntry) bool {
			if !de.IsDir() {
				return false
			}
			for _, g := range gs {
				if ok := g.Match(path); ok {
					return true
				}
			}
			return skipDirs[filepath.Base(path)]
		},
		HandleError: func(path string, err error) {
			log.Warn().Str("file", path).Err(err).Send()
		},
		HandleReport: func(path string, r *jar.Report) {
			rs = append(rs, jarFileRep{path, r})
		},
	}

	for _, d := range dirs {
		log.Debug().Str("directory", d).Msg("Scanning for log4j vulnerabilities")
		internal.MustNoErr(walker.Walk(d))
	}
	return rs
}
