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

package json

import (
	"encoding/json"
	"maps"
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type jsonCfg struct {
	console.OutCfg
}

func NewJSONCmd() *cobra.Command {
	cfg := jsonCfg{}

	cmd := &cobra.Command{
		Use:   "json [flags] <file>...",
		Short: "Load JSON files and process them",
		Example: heredoc.Doc(`
			heimdall json --query 'sys_id' app_svc.json"
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]any)
			for _, f := range args {
				maps.Copy(m, processJSON(f))
			}
			console.Fmtln(m)
		},
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processJSON(f string) map[string]any {
	return ReadJSON(f)
}

func ReadJSON(name string) (m map[string]any) {
	var in any
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	internal.MustNoErr(json.NewDecoder(r).Decode(&in))
	switch k := reflect.TypeOf(in).Kind(); k {
	case reflect.Map:
		return in.(map[string]any)
	case reflect.Array, reflect.Slice:
		return map[string]any{"data": in}
	default:
		log.Fatal().Stringer("kind", k).Msg("cannot unmarshall data")
	}
	return
}
