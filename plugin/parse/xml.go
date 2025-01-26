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

//go:build !no_parse && !no_xml

package parse

import (
	"io"
	"maps"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/clbanning/mxj/v2"
	"github.com/spf13/cobra"
)

type xmlCfg struct {
	console.OutCfg
}

func NewXMLCmd() *cobra.Command {
	cfg := xmlCfg{}

	cmd := &cobra.Command{
		Use:     "xml [flags] <file>...",
		Short:   "Load XML files and process them",
		GroupID: console.FileGroup,
		Example: heredoc.Doc(`
			heimdall xml --query 'to_number("web-app"."-version")' WEB-INF/web.xml"
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				maps.Copy(m, processXML(f))
			}
			console.Fmtln(m)
		},
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processXML(f string) map[string]any {
	return ReadXML(f).Old()
}

func ReadXML(name string) mxj.Map {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	return internal.Must(mxj.NewMapXmlReader(r))
}

func init() {
	Decoders["xml"] = func(r io.Reader) (any, error) {
		m, err := mxj.NewMapXmlReader(r)
		if m != nil {
			return m.Old(), err
		}
		return nil, err
	}
}
