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

package eval_test

import (
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/eval"
	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestNewEvalCmdExpr(t *testing.T) {
	cmd := eval.NewEvalCmd()
	internal.MustNoErr(cmd.Flags().Set("engine", "expr"))
	internal.MustNoErr(cmd.Flags().Set("expression", `g.distributionBase`))
	got := test.Run(``, cmd, []string{filepath.Join(test.GetRootDir(), "testdata", "gradle", "wrapper", "gradle-wrapper.properties:g")})
	require.Equal(t, "GRADLE_USER_HOME", got)
}

func TestNewEvalCmdJavaScript(t *testing.T) {
	cmd := eval.NewEvalCmd()
	internal.MustNoErr(cmd.Flags().Set("engine", "javascript"))
	internal.MustNoErr(cmd.Flags().Set("expression", `upper('a ')+([]+{})`))
	got := test.Run(``, cmd, []string{})
	require.Equal(t, "A [object Object]", got)
}

func TestNewEvalCmdTemplate(t *testing.T) {
	cmd := eval.NewEvalCmd()
	internal.MustNoErr(cmd.Flags().Set("engine", "template"))
	internal.MustNoErr(cmd.Flags().Set("expression", `{{upper "a"}}`))
	got := test.Run(``, cmd, []string{})
	require.Equal(t, "A", got)
}
