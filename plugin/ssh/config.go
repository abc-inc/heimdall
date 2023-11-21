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

package ssh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/kevinburke/ssh_config"
	"github.com/spf13/cobra"
)

type configCfg struct {
	console.OutCfg
	file string
	host string
	kws  []string
	def  bool
}

func NewConfigCmd() *cobra.Command {
	cfg := configCfg{file: filepath.Join(internal.Must(os.UserHomeDir()), ".ssh", "config")}

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Obtain parameters from SSH configuration",
		Example: heredoc.Doc(`
			heimdall ssh config -f ~/.ssh/config --defaults --hostname web00 --keys HostName,IdentityFile,User"
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(loadConfig(cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.host, "hostname", "n", "", "Name of the host")
	cmd.Flags().StringSliceVarP(&cfg.kws, "keys", "k", nil, "List of keywords. May be provided multiple times.")
	cmd.Flags().BoolVar(&cfg.def, "defaults", false, "Output default values for unspecified parameters")

	console.AddFileFlag(cmd, &cfg.file, "Path to the SSH configuration file", console.Optional)
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("hostname"))
	return cmd
}

func loadConfig(cfg configCfg) map[string]string {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()

	c := internal.Must(ssh_config.Decode(r))
	if len(cfg.kws) == 0 {
		return internal.Must(GetAll(c, cfg.host))
	}

	all := make(map[string]string)
	for _, k := range cfg.kws {
		v := internal.Must(c.Get(cfg.host, k))
		if v == "" {
			v = ssh_config.Default(k)
		}
		all[k] = v
	}
	return all
}

// GetAll returns all values in the configuration that match the alias.
func GetAll(c *ssh_config.Config, alias string) (map[string]string, error) {
	all := make(map[string]string)
	for _, host := range c.Hosts {
		if !host.Matches(alias) {
			continue
		}
		for _, node := range host.Nodes {
			switch t := node.(type) {
			case *ssh_config.Empty:
				continue
			case *ssh_config.KV:
				all[t.Key] = t.Value
			case *ssh_config.Include:
				continue
			default:
				return nil, fmt.Errorf("unknown node type %v", t)
			}
		}
	}

	return all, nil
}
