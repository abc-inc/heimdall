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

//go:build !no_html

package html

import (
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type htmlCfg struct {
	file string
	add  []string
}

func NewHTMLCmd() *cobra.Command {
	cfg := htmlCfg{}

	cmd := &cobra.Command{
		Use:     "html [flags] <file>",
		Short:   "Load HTML files and process them",
		GroupID: console.FileGroup,
		Example: heredoc.Doc(`
			heimdall html --query 'h1' index.html
		`),
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.file = args[0]
			_ = internal.Must(console.Msg(processHTML(cfg, cmd, args)))
		},
	}

	cmd.Flags().StringSlice("add", []string{}, "")
	cmd.Flags().StringSlice("add-class", []string{}, "")
	cmd.Flags().StringSlice("after-html", []string{}, "")
	cmd.Flags().StringSlice("append-html", []string{}, "")
	cmd.Flags().StringSlice("before-html", []string{}, "")
	cmd.Flags().StringSlice("prepend-html", []string{}, "")
	cmd.Flags().BoolSlice("remove", []bool{}, "")
	cmd.Flags().StringSlice("remove-attr", []string{}, "")
	cmd.Flags().StringSlice("remove-class", []string{}, "")
	cmd.Flags().StringSlice("replace-with-html", []string{}, "")
	cmd.Flags().StringSlice("set-html", []string{}, "")
	cmd.Flags().StringSlice("set-text", []string{}, "")
	cmd.Flags().BoolSlice("text", []bool{}, "")
	cmd.Flags().StringSlice("set-attr", []string{}, "")
	cmd.Flags().StringSlice("not", []string{}, "")
	cmd.Flags().StringSlice("find", []string{}, "")
	cmd.Flags().StringSlice("filter", []string{}, "")

	cmd.DisableFlagsInUseLine = true
	cmd.DisableFlagParsing = true
	return cmd
}

func processHTML(cfg htmlCfg, cmd *cobra.Command, args []string) string {
	internal.MustNoErr(cmd.ParseFlags(args))
	doc := readHTML(cfg.file)
	sel := doc.Selection
	f, v := "", ""
	for _, arg := range args {
		switch ret := handle(sel, f, v).(type) {
		case *goquery.Selection:
			sel = ret
		case string:
			return ret
		}

		if strings.HasPrefix(arg, "--") {
			f, v = strings.TrimPrefix(arg, "--"), ""
		} else {
			v = arg
		}
	}

	switch ret := handle(sel, f, v).(type) {
	case *goquery.Selection:
		sel = ret
	case string:
		return ret
	}

	return internal.Must(sel.Html())
}

func handle(sel *goquery.Selection, f, v string) any {
	if v == "" && f != "html" && f != "remove" && f != "text" {
		return sel
	}

	switch f {
	case "":
		return sel
	case "add":
		return sel.Add(v)
	case "add-class":
		return sel.AddClass(v)
	case "after-html":
		return sel.AfterHtml(v)
	case "append-html":
		return sel.AppendHtml(v)
	case "before-html":
		return sel.BeforeHtml(v)
	case "html":
		return internal.Must(sel.Html())
	case "prepend-html":
		return sel.PrependHtml(v)
	case "remove":
		return sel.Remove()
	case "remove-attr":
		return sel.RemoveAttr(v)
	case "remove-class":
		return sel.RemoveClass(v)
	case "replace-with-html":
		return sel.ReplaceWithHtml(v)
	case "set-html":
		return sel.SetHtml(v)
	case "set-text":
		return sel.SetText(v)
	case "text":
		return sel.Text()
	case "set-attr":
		key, val, ok := strings.Cut(v, "=")
		internal.MustOkMsgf("", ok, "invalid key-value pair: %s", v)
		return sel.SetAttr(key, val)
	case "not":
		return sel.Not(v)
	case "find":
		return sel.Find(v)
	case "filter":
		return sel.Filter(v)
	default:
		log.Fatal().Str("operation", f).Msg("unknown operation")
	}
	return sel
}

func readHTML(name string) *goquery.Document {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	return internal.Must(goquery.NewDocumentFromReader(r))
}
