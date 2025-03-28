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

//go:build !no_html

package html_test

import (
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/plugin/html"
	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestNewHTMLCmd(t *testing.T) {
	cli.SetFormat(map[string]any{"output": "json"})
	got := test.Run("",
		html.NewHTMLCmd(), []string{filepath.Join(test.GetRootDir(), "testdata", "index.html"), "-o", "text", "--find", "div b", "--text"})
	require.Equal(t, "bold", got)
}
