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

package test

import (
	"bytes"
	"os"
	"strings"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/cli/go-gh/v2/pkg/jq"
	"github.com/jfrog/build-info-go/utils"
	"github.com/spf13/cobra"
)

var root string

func GetRootDir() string {
	if root == "" {
		root = internal.Must(utils.GetProjectRoot())
	}
	return root
}

func Run(jqExpr string, cmd *cobra.Command, args []string) string {
	out := &bytes.Buffer{}
	console.Output = out
	defer func() { console.Output = out }()
	if jqExpr != "" {
		console.SetFormat(map[string]any{"output": "json"})
	}

	if cmd.RunE != nil {
		internal.MustNoErr(cmd.RunE(cmd, args))
	} else {
		cmd.Run(cmd, args)
	}
	if jqExpr != "" {
		internal.MustNoErr(jq.Evaluate(out, out, jqExpr))
	}
	return strings.Trim(out.String(), " \t\r\n")
}

func RunStdout(jqExpr string, cmd *cobra.Command, args []string) string {
	oldStdout, newStdout := os.Stdout, internal.Must(os.CreateTemp("", "output"))
	defer func() { _ = os.Remove(newStdout.Name()) }()

	console.Output, os.Stdout = newStdout, newStdout
	defer func() { console.Output, os.Stdout = oldStdout, oldStdout }()

	if jqExpr != "" {
		console.SetFormat(map[string]any{"output": "json"})
	}

	if cmd.RunE != nil {
		internal.MustNoErr(cmd.RunE(cmd, args))
	} else {
		cmd.Run(cmd, args)
	}
	if jqExpr != "" {
		internal.MustNoErr(jq.Evaluate(newStdout, newStdout, jqExpr))
	}
	all := internal.Must(os.ReadFile(newStdout.Name()))
	return strings.Trim(string(all), " \t\r\n")
}
