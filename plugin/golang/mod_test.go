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

//go:build !no_golang

package golang

import (
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/res"
	"github.com/stretchr/testify/require"
)

func TestReadMod(t *testing.T) {
	m := readMod(modCfg{file: filepath.Join(res.GetRootDir(), "go.mod")})
	require.Equal(t, "github.com/abc-inc/heimdall",
		m["Module"].(map[string]any)["Mod"].(map[string]any)["Path"])
}
