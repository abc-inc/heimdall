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

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/abc-inc/heimdall/docs"
	"github.com/abc-inc/heimdall/internal"
	"github.com/cli/go-gh/v2/pkg/text"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func rootUsageFunc(w io.Writer, cmd *cobra.Command) {
	cmd.Printf("\nUsage:  %s", cmd.UseLine())

	var subCmds []*cobra.Command
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() {
			continue
		}
		subCmds = append(subCmds, c)
	}

	if len(subCmds) > 0 {
		_, _ = fmt.Fprint(w, "\n\nAvailable commands:\n")
		for _, c := range subCmds {
			_, _ = fmt.Fprintf(w, "  %s\n", c.Name())
		}
		return
	}

	if usgs := cmd.LocalFlags().FlagUsages(); usgs != "" {
		_, _ = fmt.Fprintln(w, "\n\nFlags:")
		_, _ = fmt.Fprint(w, text.Indent(dedent(usgs), "  "))
	}
}

// nestedSuggestFunc displays a helpful error message in case subcommand name was mistyped.
// This matches Cobra's behavior for root cmd, which Cobra
// confusingly doesn't apply to nested commands.
func nestedSuggestFunc(w io.Writer, cmd *cobra.Command, arg string) {
	_, _ = fmt.Fprintf(w, "unknown command %q for %q\n", arg, cmd.CommandPath())

	var candidates []string
	if arg == "help" {
		candidates = []string{"--help"}
	} else {
		if cmd.SuggestionsMinimumDistance <= 0 {
			cmd.SuggestionsMinimumDistance = 2
		}
		candidates = cmd.SuggestionsFor(arg)
	}

	if len(candidates) > 0 {
		_, _ = fmt.Fprint(w, "\nDid you mean this?\n")
		for _, c := range candidates {
			_, _ = fmt.Fprintf(w, "\t%s\n", c)
		}
	}

	_, _ = fmt.Fprint(w, "\n")
	rootUsageFunc(w, cmd)
}

func isRootCmd(cmd *cobra.Command) bool {
	return cmd != nil && !cmd.HasParent()
}

func RootHelpFunc(cmd *cobra.Command, _ []string) {
	flags := cmd.Flags()

	if help, _ := flags.GetBool("help"); !help && !cmd.Runnable() && len(flags.Args()) > 0 {
		nestedSuggestFunc(os.Stderr, cmd, flags.Args()[0])
		return
	}

	nPad := 18

	type helpEntry struct {
		Title string
		Body  string
	}

	longText := cmd.Long
	if longText == "" {
		longText = cmd.Short
	}
	if longText != "" && cmd.LocalFlags().Lookup("jq") != nil {
		longText = strings.TrimRight(longText, "\n") +
			"\n\nFor more information about output formatting flags, see `heimdall help:formatting`."
	}

	es := []helpEntry{}
	if longText != "" {
		es = append(es, helpEntry{"", longText})
	}
	es = append(es, helpEntry{"USAGE", cmd.UseLine()})

	var subCmds []string
	for _, c := range cmd.Commands() {
		if c.IsAvailableCommand() {
			subCmds = append(subCmds, rpad(c.Name(), nPad)+"  "+c.Short)
		}
	}
	if len(subCmds) > 0 {
		es = append(es, helpEntry{"COMMANDS", strings.Join(subCmds, "\n")})
	}

	if isRootCmd(cmd) {
		var ts []string
		for _, t := range docs.Topics {
			n := strings.TrimSuffix(t.File, path.Ext(t.File))
			ts = append(ts, rpad(n, nPad)+"  "+t.Desc)
		}
		sort.Strings(ts)
		es = append(es, helpEntry{"HELP TOPICS (heimdall help:<topic>)", strings.Join(ts, "\n")})
	}

	if usgs := cmd.LocalFlags().FlagUsages(); usgs != "" {
		es = append(es, helpEntry{"FLAGS", dedent(usgs)})
	}
	if usgs := cmd.InheritedFlags().FlagUsages(); usgs != "" {
		es = append(es, helpEntry{"INHERITED FLAGS", dedent(usgs)})
	}
	if cmd.Example != "" {
		es = append(es, helpEntry{"EXAMPLES", dedent(cmd.Example)})
	}

	for p := cmd; p != nil; p = p.Parent() {
		if p.Annotations["help:environment"] != "" {
			e := helpEntry{"ENVIRONMENT VARIABLES", ""}
			for _, l := range strings.Split(p.Annotations["help:environment"], "\n") {
				if l != "" {
					n, d, _ := strings.Cut(l, "  ")
					e.Body += "  " + rpad(n, nPad) + "  " + strings.TrimSpace(d) + "\n"
				}
			}
			e.Body = dedent(e.Body)
			es = append(es, e)
			break
		}
	}

	es = append(es, helpEntry{"LEARN MORE", `
Use 'heimdall <command> [<subcommand>] --help' for more information about a command.`})
	if _, ok := cmd.Annotations["help:feedback"]; ok {
		es = append(es, helpEntry{"FEEDBACK", cmd.Annotations["help:feedback"]})
	}

	out := os.Stderr
	for _, e := range es {
		if e.Title != "" {
			// If there is a title, add indentation to each line in the body
			_, _ = fmt.Fprintln(out, color.New(color.Bold).Sprint(e.Title))
			_, _ = fmt.Fprintln(out, text.Indent(strings.Trim(e.Body, "\r\n"), "  "))
		} else {
			// If there is no title print the body as is
			_, _ = fmt.Fprintln(out, e.Body)
		}
		_, _ = fmt.Fprintln(out)
	}
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	if len(s) >= padding {
		return s
	}
	return s + strings.Repeat(" ", padding-len(s))
}

// dedent determines the shortest amount of leading whitespaces from every line and trims that many spaces.
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	var buf bytes.Buffer
	for _, l := range lines {
		_ = internal.Must(fmt.Fprintln(&buf, strings.TrimPrefix(l, strings.Repeat(" ", minIndent))))
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
