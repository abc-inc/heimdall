// Copyright 2025 The Heimdall authors
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

package parse

import (
	"io"
	"maps"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type dotenvCfg struct {
	cli.OutCfg
}

func NewDotEnvCmd() *cobra.Command {
	cfg := dotenvCfg{}
	cmd := &cobra.Command{
		Use:     "dotenv [flags] <file>...",
		Short:   "Load dotenv files and process them",
		GroupID: cli.FileGroup,
		Example: heredoc.Doc(`
			heimdall parse .env
		`),
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m := make(map[string]string)
			for _, f := range args {
				maps.Copy(m, processDotEnv(f))
			}
			cli.Fmtln(m)
		},
	}

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func processDotEnv(name string) map[string]string {
	r := internal.Must(res.Open(name))
	defer func() { _ = r.Close() }()
	return internal.Must(godotenv.Parse(r))
}

func init() {
	Decoders["env"] = func(r io.Reader) (any, error) {
		return godotenv.Parse(r)
	}
}
