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

package docs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/abc-inc/heimdall"
	"github.com/abc-inc/heimdall/internal"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/gum/pager"
	"github.com/charmbracelet/gum/style"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type Topic struct {
	File string
	Desc string
}

var Topics = []Topic{
	{File: "man/expr-lang.md", Desc: "Overview for expr (built-in expression language)"},
	{File: "contributing.md", Desc: "Information for improving Heimdall"},
	{File: "formatting.md", Desc: "Description of output formats and filters"},
	{File: "source.md", Desc: "Instructions for building Heimdall from source"},
	{File: "why.md", Desc: fmt.Sprintf("Why %s?", color.New(color.Italic).Sprint("Heimdall"))},
}

func init() {
	for _, f := range internal.Must(heimdall.StaticFS.ReadDir("docs/examples")) {
		d := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		Topics = append(Topics, Topic{File: path.Join("examples", f.Name()), Desc: "Example for " + d})
	}
}

type helpTopicCfg struct {
	style   string
	noPager bool
}

func NewHelpTopicCmd(t Topic) *cobra.Command {
	cfg := helpTopicCfg{style: "github"}
	t.File = filepath.Clean(t.File) // prevent path traversal
	n := strings.TrimSuffix(t.File, path.Ext(t.File))

	cmd := &cobra.Command{
		Use:    "help:" + n,
		Short:  t.Desc,
		Hidden: true,
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		printTopic(cfg, t, cmd)
	})

	cmd.Flags().BoolVar(&cfg.noPager, "no-pager", cfg.noPager, "Do not pipe output into a pager")
	cmd.Flags().StringVar(&cfg.style, "style", cfg.style,
		fmt.Sprintf("Style to use for formatting (%s)", strings.Join(styles.Names(), ", ")))

	return cmd
}

func printTopic(cfg helpTopicCfg, t Topic, cmd *cobra.Command) {
	str := string(internal.Must(heimdall.StaticFS.ReadFile(path.Join("docs", t.File))))
	ext := strings.TrimPrefix(filepath.Ext(t.File), ".")
	if isatty.IsTerminal(os.Stdout.Fd()) {
		if ext == "md" {
			str = internal.Must(Render(str))
		} else {
			b := strings.Builder{}
			internal.MustNoErr(quick.Highlight(&b, str, ext, "terminal", MustStyle(cfg.style)))
			str = b.String()
		}
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) || cfg.noPager {
		cmd.Print(str)
		return
	}

	internal.MustNoErr(pager.Options{
		Style:               style.Styles{Border: "rounded", Padding: "0 1", BorderForeground: "7"},
		HelpStyle:           style.Styles{Foreground: "241"},
		Content:             str,
		ShowLineNumbers:     ext != "md",
		LineNumberStyle:     style.Styles{Foreground: "237"},
		SoftWrap:            true,
		MatchStyle:          style.Styles{Foreground: "226", Bold: true},
		MatchHighlightStyle: style.Styles{Foreground: "235", Background: "226", Bold: true},
	}.Run())
}

func MustStyle(n string) string {
	s := styles.Get(n)
	if s == styles.Fallback {
		ns := maps.Keys(styles.Registry)
		slices.Sort(ns)
		internal.MustOkMsgf(n, s != styles.Fallback, "Style '%s' does not exist. Valid styles are: %s", n, strings.Join(ns, " "))
	}
	return n
}
