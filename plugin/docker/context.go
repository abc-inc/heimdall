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

//go:build !no_docker

package docker

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/jfrog/gofrog/io"
	"github.com/moby/patternmatcher"
	"github.com/spf13/cobra"
)

func NewContextCmd() *cobra.Command {
	cfg := dockerCfg{file: ".dockerignore"}
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Parse a .dockerignore file and list matching files",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(listMatches(cfg))
		},
	}

	cmd.Flags().BoolVarP(&cfg.invertMatch, "invert-match", "v", cfg.invertMatch, "Selected files are those matching any of the specified patterns")

	console.AddFileFlag(cmd, &cfg.file, "Path to the .dockerignore file", console.Optional)
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func listMatches(cfg dockerCfg) (fs []string) {
	fi, err := os.Stat(cfg.file)
	if err == nil && !fi.Mode().IsRegular() {
		return
	}

	pm := newParser(cfg.file)
	dir := filepath.Dir(cfg.file) + string(filepath.Separator)
	internal.MustNoErr(io.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if errors.Is(err, io.ErrSkipDir) {
			return err
		}

		path = strings.TrimPrefix(path, dir)
		var ok bool
		if ok, err = pm.MatchesOrParentMatches(path); err == nil && ok == cfg.invertMatch {
			fs = append(fs, path)
		}
		return err
	}, false))

	if len(fs) > 0 && fs[0] == "" {
		fs = fs[1:]
	}
	return
}

func newParser(name string) *patternmatcher.PatternMatcher {
	r, err := res.Open(name)
	if err != nil && !res.IsURL(name) {
		return internal.Must(patternmatcher.New(nil))
	}

	internal.MustNoErr(err)
	defer func() { _ = r.Close() }()

	return internal.Must(patternmatcher.New(parseIgnore(r)))
}
