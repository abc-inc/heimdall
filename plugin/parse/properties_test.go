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

//go:build !no_parse && !no_properties

package parse

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestProcessProps(t *testing.T) {
	gradleExp := heredoc.Doc(`
		# Generated today
		distributionBase=/gradle
		zipStoreBase=GRADLE_USER_HOME
		zipstoreBase=/gradle`)

	f := filepath.Join(test.GetRootDir(), "testdata", "gradle", "wrapper", "gradle-wrapper.properties")
	tests := []struct {
		name string
		cfg  propsCfg
		exp  string
	}{
		{name: "getWithComment", cfg: propsCfg{file: f, get: []string{"distributionBase"}, set: nil,
			unset: nil, sep: "", comment: "# ", OutCfg: console.OutCfg{Output: "text"}}, exp: "GRADLE_USER_HOME"},
		{name: "setNewWithSep", cfg: propsCfg{file: f, get: []string{"a"}, set: []string{"a:b"},
			unset: nil, sep: ":", comment: "", OutCfg: console.OutCfg{Output: "text"}}, exp: "b"},
		{name: "setMultiple", cfg: propsCfg{file: f, get: nil, set: []string{"distributionBase=/gradle", "zipstoreBase=/gradle"},
			unset: []string{"distributionPath", "distributionUrl", "zipStorePath"}, sep: "", comment: "# ", OutCfg: console.OutCfg{Output: "text"}}, exp: gradleExp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console.SetFormat(map[string]any{"output": tt.cfg.Output})
			require.Equal(t, tt.exp, strings.Join(processProps(tt.cfg).([]string), "\n"))
		})
	}
}
