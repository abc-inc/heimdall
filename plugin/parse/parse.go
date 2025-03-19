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

//go:build !no_parse

package parse

import (
	"io"
	"path"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/imdario/mergo"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Decoder func(io.Reader) (any, error)

var Decoders = make(map[string]Decoder)

type parseCfg struct {
	cli.OutCfg
	defType string
}

func NewParseCmd() *cobra.Command {
	cfg := parseCfg{}

	cmd := &cobra.Command{
		Use:     "parse [flags] <file>...",
		Short:   "Parse various files formats and process them",
		GroupID: cli.FileGroup,
		Example: heredoc.Doc(`
			heimdall parse --query sys_id app_svc.json
			heimdall parse --query libraries.junit.version gradle/libs.versions.toml
			heimdall parse --query 'to_number("web-app"."-version")' WEB-INF/web.xml"
			heimdall parse --query 'to_number("web-app"."-version")' config.yaml
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				v := internal.Must(processFile(f, cfg.defType))
				i := SplitNamePrefixType(f)
				internal.MustNoErr(mergo.Map(&m, toAnyMap(v, i.Alias), mergo.WithOverride))
			}
			if len(args) == 1 && len(m) == 1 && m[""] != nil {
				cli.Fmtln(m[""])
			} else {
				cli.Fmtln(m)
			}
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processFile(name string, defType string) (m any, err error) {
	i := SplitNamePrefixType(name)
	if i.Type == "" || i.Type == "auto" {
		i.Type = defType
	}
	if d, ok := Decoders[i.Type]; ok {
		r := internal.Must(res.Open(i.File))
		defer func() { _ = r.Close() }()
		return d(r)
	}
	log.Fatal().Msgf("unsupported file type: %s", i.Type)
	return nil, err
}

type Input struct {
	File  string
	Alias string
	Type  string
}

func (i Input) String() string {
	return i.File + ":" + i.Alias + ":" + i.Type
}

func SplitNamePrefixType(name string) Input {
	idx := strings.LastIndexAny(name, `/\`) + 1
	parts := strings.SplitN(name[idx:], ":", 3)
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(parts[0]), "."))
	switch len(parts) {
	case 3:
		return Input{File: name[:idx] + parts[0], Alias: parts[1], Type: strings.ToLower(parts[2])}
	case 2:
		return Input{File: name[:idx] + parts[0], Alias: parts[1], Type: ext}
	default:
		return Input{File: name[:idx] + parts[0], Alias: "", Type: ext}
	}
}
