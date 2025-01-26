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
	"slices"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/parse"
	"github.com/clbanning/mxj/v2"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type webXMLCfg struct {
	console.OutCfg
	mode string
}

func NewWebXMLCmd() *cobra.Command {
	cfg := webXMLCfg{mode: "raw"}
	cmd := &cobra.Command{
		Use:     "webxml",
		Aliases: []string{"web.xml"},
		Short:   "Process web.xml files",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			app := processWebXML(cfg, args[0])
			console.Fmtln(app)
		},
	}

	cmd.Flags().StringVar(&cfg.mode, "mode", cfg.mode, `Output mode ("raw", "servlet-mappings")`)

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func processWebXML(cfg webXMLCfg, uri string) any {
	m := parse.ReadXML(uri)

	switch cfg.mode {
	case "raw":
		return m.Old()
	case "servlet-mappings":
		return listServlets(m)
	default:
		log.Fatal().Str("mode", cfg.mode).Msg("unsupported operation")
		return nil
	}
}

func listServlets(m mxj.Map) (app WebApp) {
	for _, smEl := range internal.Must(m.ValuesForPath("web-app.servlet-mapping")) {
		sn := getVal[string](smEl, ServletName)
		sp := getVal[string](smEl, URLPattern)

		app.Mapping = append(app.Mapping, ServletMapping{Pattern: sp})
		sm := &app.Mapping[len(app.Mapping)-1]

		for _, s := range internal.Must(m.ValuesForPath("web-app.servlet")) {
			if sEl := s.(map[string]any); sEl[ServletName] == sn {
				sm.Servlet = sEl
				break
			}
		}

		for _, fmEl := range internal.Must(m.ValuesForPath("web-app.filter-mapping")) {
			fn := getVal[string](fmEl, FilterName)
			fp := getVal[string](fmEl, URLPattern)

			// For servlet-mapping, it is permitted to refer to a servlet instead of defining url-pattern.
			if fsn := getVal[string](fmEl, ServletName); fsn != "" {
				for _, sm := range internal.Must(m.ValuesForPath("web-app.servlet-mapping")) {
					if smEl := sm.(map[string]any); getVal[string](smEl, ServletName) == fsn {
						fp = getVal[string](smEl, URLPattern)
					}
				}
			}

			if !glob.MustCompile(fp).Match(sp) {
				continue
			}

			for _, f := range internal.Must(m.ValuesForPath("web-app.filter")) {
				if fEl := f.(map[string]any); getVal[string](f, FilterName) == fn {
					fIdx := slices.IndexFunc(sm.Chain, func(fm FilterMapping) bool {
						return fm.Pattern == fp
					})
					if fIdx == -1 {
						sm.Chain = append(sm.Chain, FilterMapping{Pattern: fp})
						fIdx = len(sm.Chain) - 1
					}
					sm.Chain[fIdx].Filters = append(sm.Chain[fIdx].Filters, fEl)
				}
			}
		}
	}

	return
}

func getVal[T any](m any, name string) (t T) {
	child := m.(map[string]any)[name]
	if child == nil {
		return t
	}
	if r, ok := child.(T); ok {
		return r
	}
	return child.(map[string]any)["#text"].(T)
}

const (
	AsyncSupported = "async-supported"
	FilterName     = "filter-name"
	FilterClass    = "filter-class"
	ServletName    = "servlet-name"
	ServletClass   = "servlet-class"
	URLPattern     = "url-pattern"
)

type WebApp struct {
	Mapping []ServletMapping `json:"servlet-mappings"`
}

type ServletMapping struct {
	Pattern string          `json:"url-pattern"`
	Servlet Servlet         `json:"servlet"`
	Chain   []FilterMapping `json:"filter-mappings,omitempty"`
}

func (sm ServletMapping) String() string {
	return sm.Pattern
}

type FilterMapping struct {
	Pattern string   `json:"url-pattern"`
	Filters []Filter `json:"filters"`
}

func (fm FilterMapping) String() string {
	return fm.Pattern
}

type Filter map[string]any

func (f Filter) String() string {
	return f[FilterName].(string)
}

type Servlet map[string]any

func (s Servlet) String() string {
	return s[ServletName].(string)
}
