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

//go:build !no_cyclonedx

package cyclonedx

import (
	"os"
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/spf13/cobra"
)

type readCfg struct {
	name   string
	format string
	pretty bool
}

func NewReadCmd() *cobra.Command {
	cfg := readCfg{format: "json", pretty: true}
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a CycloneDX SBOM",
		Example: heredoc.Doc(`
			heimdall cyclonedx read --format json --pretty sbom.xml
		`),
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.name = args[0]
			readBom(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.format, "format", cfg.format, "Target format of the SBOM file (xml or json)")
	cmd.Flags().BoolVar(&cfg.pretty, "pretty", cfg.pretty, "Pretty print the SBOM")

	return cmd
}

func readBom(cfg readCfg) (b cdx.BOM) {
	f := internal.Must(os.Open(cfg.name))
	defer func() { _ = f.Close() }()

	format := cdx.BOMFileFormatJSON
	if strings.ToLower(filepath.Ext(cfg.name)) == ".xml" {
		format = cdx.BOMFileFormatXML
	}
	d := cdx.NewBOMDecoder(f, format)
	internal.MustNoErr(d.Decode(&b))

	if cfg.format == "json" {
		internal.MustNoErr(cdx.NewBOMEncoder(console.Output, cdx.BOMFileFormatJSON).SetPretty(cfg.pretty).EncodeVersion(&b, b.SpecVersion))
	} else {
		internal.MustNoErr(cdx.NewBOMEncoder(console.Output, cdx.BOMFileFormatXML).SetPretty(cfg.pretty).EncodeVersion(&b, b.SpecVersion))
	}
	if !cfg.pretty {
		_ = internal.Must(console.Msg("\n"))
	}

	return
}
