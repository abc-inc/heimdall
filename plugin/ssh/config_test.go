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

//go:build !no_ssh

package ssh

import (
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	f := filepath.Join(test.GetRootDir(), "testdata", "ssh_config")
	all := loadConfig(configCfg{file: f, host: "code.company.corp"})
	require.Equal(t, 4, len(all))
	require.Equal(t, "gateway.company.corp", all["HostName"])
	require.Equal(t, "publickey", all["PreferredAuthentications"])
}
