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

package echo_test

import (
	"testing"

	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/echo"
	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestNewEchoCmd(t *testing.T) {
	cmd := echo.NewEchoCmd()
	internal.MustNoErr(cmd.Flags().Set("foreground", "#ff0000"))
	got := test.RunStdout(``, cmd, []string{"error"})
	require.Equal(t, "error", got)
}
