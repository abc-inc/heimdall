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
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type htmlCfg struct {
	cli.OutCfg
	file string
}

func NewHTMLCmd() *cobra.Command {
	cfg := htmlCfg{}

	cmd := &cobra.Command{
		Use:     "html <file> [flags]",
		Short:   "Load HTML files and process them",
		GroupID: cli.FileGroup,
		Example: heredoc.Doc(`
			heimdall html --query 'h1' index.html
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.SetFormat(map[string]any{
				"jq":     cmd.Flag("jq"),
				"output": cmd.Flag("output"),
				"pretty": cmd.Flag("pretty"),
				"query":  cmd.Flag("query")},
			)
			cfg.file = args[0]
			cli.Fmtln(processHTML(cfg, cmd, os.Args))
		},
	}

	cmd.Flags().StringSlice("add", []string{}, "adds the selector string's matching nodes to those in the current selection.")
	cmd.Flags().StringSlice("add-class", []string{}, "adds the given class(es) to each element in the set of matched elements.")
	cmd.Flags().StringSlice("after", []string{}, "applies the selector from the root document and inserts the matched elements after the elements in the set of matched elements.")
	cmd.Flags().StringSlice("after-html", []string{}, "parses the html and inserts it after the set of matched elements.")
	cmd.Flags().StringSlice("all-attr", []string{}, "gets the specified attribute's value for all elements in the selection.")
	cmd.Flags().Bool("all-html", true, "gets the HTML contents of the each element in the set of matched elements, including text and comment nodes.")
	cmd.Flags().Bool("all-text", true, "gets the text contents of each element in the set of matched elements, including their descendants.")
	cmd.Flags().StringSlice("append", []string{}, "appends the elements specified by the selector to the end of each element in the set of matched elements.")
	cmd.Flags().StringSlice("append-html", []string{}, "parses the html and appends it to the set of matched elements.")
	cmd.Flags().StringSlice("attr", []string{}, "gets the specified attribute's value for the first element in the selection.")
	cmd.Flags().StringSlice("before", []string{}, "inserts the matched elements before each element in the set of matched elements.")
	cmd.Flags().StringSlice("before-html", []string{}, "parses the html and inserts it before the set of matched elements.")
	cmd.Flags().Bool("children", true, "gets the child elements of each element in the selection.")
	cmd.Flags().StringSlice("children-filtered", []string{}, "gets the child elements of each element in the selection, filtered by the specified selector.")
	cmd.Flags().StringSlice("closest", []string{}, "gets the first element that matches the selector by testing the element itself and traversing up through its ancestors in the DOM tree.")
	cmd.Flags().Bool("contents", true, "gets the children of each element in the selection, including text and comment nodes.")
	cmd.Flags().StringSlice("contents-filtered", []string{}, "gets the children of each element in the selection, filtered by the specified selector.")
	cmd.Flags().Bool("empty", true, "removes all children nodes from the set of matched elements.")
	cmd.Flags().Bool("end", false, "ends the most recent filtering operation in the current chain and returns the set of matched elements to its previous state.")
	cmd.Flags().StringSlice("filter", []string{}, "reduces the set of matched elements to those that match the selector string.")
	cmd.Flags().StringSlice("find", []string{}, "gets the descendants of each element in the current set of matched elements, filtered by a selector.")
	cmd.Flags().Bool("first", true, "reduces the set of matched elements to the first in the set.")
	cmd.Flags().StringSlice("has", []string{}, "reduces the set of matched elements to those that have a descendant that matches the selector.")
	cmd.Flags().StringSlice("has-class", []string{}, "determines whether any of the matched elements are assigned the given class.")
	cmd.Flags().Bool("html", true, "gets the HTML contents of the first element in the set of matched elements. It includes text and comment nodes.")
	cmd.Flags().Bool("index", true, "returns the position of the first element within the selection object relative to its sibling elements.")
	cmd.Flags().StringSlice("index-selector", []string{}, "returns the position of the first element within the selection object relative to the elements matched by the selector, or -1 if not found.")
	cmd.Flags().StringSlice("is", []string{}, "checks the current matched set of elements against a selector and returns true if at least one of these elements matches.")
	cmd.Flags().Bool("last", true, "reduces the set of matched elements to the last in the set.")
	cmd.Flags().Bool("length", true, "returns the number of elements in the selection object.")
	cmd.Flags().Bool("next", true, "gets the immediately following sibling of each element in the selection.")
	cmd.Flags().Bool("next-all", true, "gets all the following siblings of each element in the selection.")
	cmd.Flags().StringSlice("next-all-filtered", []string{}, "gets all the following siblings of each element in the selection filtered by a selector.")
	cmd.Flags().StringSlice("next-filtered", []string{}, "gets the immediately following sibling of each element in the selection filtered by a selector.")
	cmd.Flags().StringSlice("next-until", []string{}, "gets all following siblings of each element up to but not including the element matched by the selector.")
	cmd.Flags().StringSlice("not", []string{}, "removes elements from the selection that match the selector string.")
	cmd.Flags().Bool("parent", true, "gets the parent of each element in the selection.")
	cmd.Flags().StringSlice("parent-filtered", []string{}, "gets the parent of each element in the selection filtered by a selector.")
	cmd.Flags().Bool("parents", true, "gets the ancestors of each element in the current selection.")
	cmd.Flags().StringSlice("parents-filtered", []string{}, "gets the ancestors of each element in the current selection.")
	cmd.Flags().StringSlice("parents-until", []string{}, "gets the ancestors of each element in the selection, up to but not including the element matched by the selector.")
	cmd.Flags().StringSlice("prepend", []string{}, "prepends the elements specified by the selector to each element in the set of matched elements.")
	cmd.Flags().StringSlice("prepend-html", []string{}, "parses the html and prepends it to the set of matched elements.")
	cmd.Flags().Bool("prev", true, "gets the immediately preceding sibling of each element in the selection.")
	cmd.Flags().Bool("prev-all", true, "gets all the preceding siblings of each element in the selection.")
	cmd.Flags().StringSlice("prev-all-filtered", []string{}, "gets all the preceding siblings of each element in the selection filtered by a selector.")
	cmd.Flags().StringSlice("prev-filtered", []string{}, "gets the immediately preceding sibling of each element in the selection filtered by a selector.")
	cmd.Flags().StringSlice("prev-until", []string{}, "gets all preceding siblings of each element up to but not including the element matched by the selector.")
	cmd.Flags().Bool("remove", true, "removes the set of matched elements from the document.")
	cmd.Flags().StringSlice("remove-attr", []string{}, "removes the named attribute from each element in the set of matched elements.")
	cmd.Flags().StringSlice("remove-class", []string{}, "removes the given class(es) from each element in the set of matched elements. Multiple class names can be specified, separated by a space or via multiple arguments. If no class name is provided, all classes are removed.")
	cmd.Flags().StringSlice("remove-filtered", []string{}, "removes from the current set of matched elements those that match the selector filter. It returns the selection of removed nodes.")
	cmd.Flags().StringSlice("replace-with", []string{}, "replaces each element in the set of matched elements with the nodes matched by the given selector. It returns the removed elements.")
	cmd.Flags().StringSlice("replace-with-html", []string{}, "replaces each element in the set of matched elements with the parsed HTML. It returns the removed elements.")
	cmd.Flags().StringSlice("set-attr", []string{}, "sets the given attribute on each element in the set of matched elements.")
	cmd.Flags().StringSlice("set-html", []string{}, "sets the html content of each element in the selection to specified html string.")
	cmd.Flags().StringSlice("set-text", []string{}, "sets the content of each element in the selection to specified content.")
	cmd.Flags().Bool("siblings", true, "gets the siblings of each element in the selection.")
	cmd.Flags().StringSlice("siblings-filtered", []string{}, "gets the siblings of each element in the selection filtered by a selector.")
	cmd.Flags().Bool("text", true, "gets the combined text contents of each element in the set of matched elements, including their descendants.")
	cmd.Flags().StringSlice("toggle-class", []string{}, "adds or removes the given class(es) for each element in the set of matched elements.")
	cmd.Flags().Bool("unwrap", true, "removes the parents of the set of matched elements, leaving the matched elements (and their siblings, if any) in their place.")
	cmd.Flags().StringSlice("wrap", []string{}, "wraps each element in the set of matched elements inside the first element matched by the given selector.")
	cmd.Flags().StringSlice("wrap-all", []string{}, "wraps a single HTML structure, matched by the given selector, around all elements in the set of matched elements.")
	cmd.Flags().StringSlice("wrap-all-html", []string{}, "wraps the given HTML structure around all elements in the set of matched elements.")
	cmd.Flags().StringSlice("wrap-html", []string{}, "wraps each element in the set of matched elements inside the inner-most child of the given HTML.")
	cmd.Flags().StringSlice("wrap-inner", []string{}, "wraps an HTML structure, matched by the given selector, around the content of element in the set of matched elements.")
	cmd.Flags().StringSlice("wrap-inner-html", []string{}, "wraps an HTML structure, matched by the given selector, around the content of element in the set of matched elements.")

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processHTML(cfg htmlCfg, cmd *cobra.Command, args []string) any {
	doc := readHTML(cfg.file)
	sel := doc.Selection

	n, v := "", ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Skip non-flag arguments, such as the filename.
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		n, v = strings.TrimLeft(arg, "-"), ""
		if len(arg) == 2 || cmd.PersistentFlags().Lookup(n) != nil {
			// This command does not have short flags, so it must be a persistent flag.
			continue
		}

		if fl := cmd.Flag(n); fl != nil && fl.Value.Type() == "stringSlice" {
			i += 1
			v = args[i]
		}

		log.Trace().Str(n, v).Msg("performing DOM operation")
		switch ret := handle(sel, n, v).(type) {
		case *goquery.Selection:
			sel = ret
		case []string:
			return ret
		case string, int, bool:
			return ret
		case nil:
		}
	}

	return sel.Map(func(i int, selection *goquery.Selection) string {
		return internal.Must(selection.Html())
	})
}

func handle(sel *goquery.Selection, f, v string) any {
	switch f {
	case "add":
		return sel.Add(v)
	case "add-class":
		return sel.AddClass(v)
	case "after":
		return sel.After(v)
	case "after-html":
		return sel.AfterHtml(v)
	case "all-attr":
		return sel.Map(func(i int, selection *goquery.Selection) string {
			return selection.AttrOr(v, "")
		})
	case "all-html":
		return sel.Map(func(i int, selection *goquery.Selection) string {
			return internal.Must(selection.Html())
		})
	case "all-text":
		return sel.Map(func(i int, selection *goquery.Selection) string {
			return selection.Text()
		})
	case "append":
		return sel.Append(v)
	case "append-html":
		return sel.AppendHtml(v)
	case "attr":
		return sel.AttrOr(v, "")
	case "before":
		return sel.Before(v)
	case "before-html":
		return sel.BeforeHtml(v)
	case "children":
		return sel.Children()
	case "children-filtered":
		return sel.ChildrenFiltered(v)
	case "closest":
		return sel.Closest(v)
	case "contents":
		return sel.Contents()
	case "contents-filtered":
		return sel.ContentsFiltered(v)
	case "empty":
		return sel.Empty()
	case "end":
		return sel.End()
	case "filter":
		return sel.Filter(v)
	case "find":
		return sel.Find(v)
	case "first":
		return sel.First()
	case "has":
		return sel.Has(v)
	case "has-class":
		return sel.HasClass(v)
	case "html":
		return internal.Must(sel.Html())
	case "index":
		return sel.Index()
	case "index-selector":
		return sel.IndexSelector(v)
	case "is":
		return sel.Is(v)
	case "last":
		return sel.Last()
	case "length":
		return sel.Length()
	case "next":
		return sel.Next()
	case "next-all":
		return sel.NextAll()
	case "next-all-filtered":
		return sel.NextAllFiltered(v)
	case "next-filtered":
		return sel.NextFiltered(v)
	case "next-until":
		return sel.NextUntil(v)
	case "not":
		return sel.Not(v)
	case "parent":
		return sel.Parent()
	case "parent-filtered":
		return sel.ParentFiltered(v)
	case "parents":
		return sel.Parents()
	case "parents-filtered":
		return sel.ParentsFiltered(v)
	case "parents-until":
		return sel.ParentsUntil(v)
	case "prepend":
		return sel.Prepend(v)
	case "prepend-html":
		return sel.PrependHtml(v)
	case "prev":
		return sel.Prev()
	case "prev-all":
		return sel.PrevAll()
	case "prev-all-filtered":
		return sel.PrevAllFiltered(v)
	case "prev-filtered":
		return sel.PrevFiltered(v)
	case "prev-until":
		return sel.PrevUntil(v)
	case "remove":
		return sel.Remove()
	case "remove-attr":
		return sel.RemoveAttr(v)
	case "remove-class":
		return sel.RemoveClass(v)
	case "remove-filtered":
		return sel.RemoveFiltered(v)
	case "replace-with":
		return sel.ReplaceWith(v)
	case "replace-with-html":
		return sel.ReplaceWithHtml(v)
	case "set-attr":
		key, val, ok := strings.Cut(v, "=")
		internal.MustOkMsgf("", ok, "invalid key-value pair: %s", v)
		return sel.SetAttr(key, val)
	case "set-html":
		return sel.SetHtml(v)
	case "set-text":
		return sel.SetText(v)
	case "siblings":
		return sel.Siblings()
	case "siblings-filtered":
		return sel.SiblingsFiltered(v)
	case "text":
		return sel.Text()
	case "toggle-class":
		return sel.ToggleClass(v)
	case "unwrap":
		return sel.Unwrap()
	case "wrap":
		return sel.Wrap(v)
	case "wrap-all":
		return sel.WrapAll(v)
	case "wrap-all-html":
		return sel.WrapAllHtml(v)
	case "wrap-html":
		return sel.WrapHtml(v)
	case "wrap-inner":
		return sel.WrapInner(v)
	case "wrap-inner-html":
		return sel.WrapInnerHtml(v)
	default:
		return nil
	}
}

func readHTML(name string) *goquery.Document {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	return internal.Must(goquery.NewDocumentFromReader(r))
}
