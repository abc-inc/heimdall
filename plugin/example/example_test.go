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
	"testing"

	"github.com/abc-inc/heimdall/cli"
	"github.com/stretchr/testify/require"
)

func TestRenderListFilter(t *testing.T) {
	io, _, out, _ := cli.Test()
	cli.IO = io

	render(exampleCfg{name: "*sh*", list: true}, nil, nil)
	require.Equal(t, "shell\n", out.String())
}
