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
	"io"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/spf13/cobra"
)

func NewFileCmd() *cobra.Command {
	cfg := dockerCfg{file: "Dockerfile"}
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Parse a Dockerfile and print its instructions",
		Example: heredoc.Doc(`
			heimdall docker file -f Dockerfile
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(readDockerfile(cfg))
		},
	}

	console.AddFileFlag(cmd, &cfg.file, "Path to the Dockerfile", console.Optional)
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

type line struct {
	Cmd       string   `json:"cmd,omitempty" yaml:"cmd,omitempty"`               // lowercase command name, e.g., "from"
	SubCmd    string   `json:"sub_cmd,omitempty" yaml:"sub_cmd,omitempty"`       // ONBUILD only, holds the sub-command
	JSON      bool     `json:"json,omitempty" yaml:"json,omitempty"`             // whether the value is written in json form
	Original  string   `json:"line,omitempty" yaml:"line,omitempty"`             // original source line
	StartLine int      `json:"start_line,omitempty" yaml:"start_line,omitempty"` // original source line number which starts this command
	EndLine   int      `json:"end_line,omitempty" yaml:"end_line,omitempty"`     // original source line number which ends this command
	Flags     []string `json:"flags,omitempty" yaml:"flags,omitempty"`           // any flags such as "--from=..." for "COPY"
	Value     []string `json:"value,omitempty" yaml:"value,omitempty"`           // contents of the command, e.g., "scratch"
}

func (l line) String() string {
	return l.Original
}

func readDockerfile(cfg dockerCfg) []line {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()
	return parseDockerfile(r)
}

func parseDockerfile(r io.Reader) (cmds []line) {
	f := internal.Must(parser.Parse(r))

	for _, child := range f.AST.Children {
		c := line{
			Cmd:       child.Value,
			Original:  child.Original,
			StartLine: child.StartLine,
			EndLine:   child.EndLine,
			Flags:     child.Flags,
		}

		// Only happens for ONBUILD
		if child.Next != nil && len(child.Next.Children) > 0 {
			child = child.Next.Children[0]
			c.SubCmd = child.Value
		}

		c.JSON = child.Attributes["json"]
		for n := child.Next; n != nil; n = n.Next {
			c.Value = append(c.Value, n.Value)
		}

		cmds = append(cmds, c)
	}
	return
}
