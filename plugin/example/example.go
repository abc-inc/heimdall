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

//go:build !no_example

package example

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/spf13/cobra"
)

type exampleCfg struct {
	name    string
	style   string
	list    bool
	noPager bool
}

func NewExampleCmd() *cobra.Command {
	cfg := exampleCfg{style: "github"}

	cmd := &cobra.Command{
		Use:   "example [<name>]",
		Short: "Print examples and detailed information",
		Example: heredoc.Doc(`
			heimdall example java
		`),
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cfg.name = args[0]
			}
			render(cfg, cmd, args)
		},
	}

	cmd.Flags().BoolVarP(&cfg.list, "list", "l", cfg.list, "List built-in examples")
	cmd.Flags().BoolVar(&cfg.noPager, "no-pager", cfg.noPager, "Do not pipe output into a pager")
	cmd.Flags().StringVar(&cfg.style, "style", cfg.style,
		fmt.Sprintf("Style to use for formatting (%s)", strings.Join(styles.Names(), ", ")))

	return cmd
}

func render(cfg exampleCfg, cmd *cobra.Command, args []string) {
	_, cfg.name = filepath.Split(filepath.ToSlash(cfg.name))
	if cfg.name == "" {
		cfg.name = "*"
		cfg.list = true
	}

	es := internal.Must(fs.Glob(heimdall.StaticFS, "docs/examples/"+cfg.name+".yaml"))
	if len(es) == 0 {
		log.Fatalf("Example '%s' does not exist.", cfg.name)
	} else if cfg.list {
		for i := range es {
			es[i] = filepath.Base(strings.TrimSuffix(es[i], filepath.Ext(es[i])))
			_ = internal.Must(console.Msg(es[i] + "\n"))
		}
		return
	}
	if len(es) > 1 {
		log.Fatalf("Multiple examples found for '%s'.", cfg.name)
		return
	}

	if cmd == nil {
		return
	}

	args2 := []string{"help:examples/" + args[0]}
	if cfg.noPager {
		args2 = append(args2, "--no-pager")
	}
	if cfg.style != "" {
		args2 = append(args2, "--style", cfg.style)
	}
	args2 = append(args2, args[1:]...)
	cmd.Root().SetArgs(args2)
	internal.MustNoErr(cmd.Root().Execute())
}
