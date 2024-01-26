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

//go:build !no_java

package java

import (
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/res"
	"github.com/stretchr/testify/require"
)

func TestReadJUnit(t *testing.T) {
	f := filepath.Join(res.GetRootDir(), "testdata", "TEST-com.example.project.CalculatorTests.xml")
	cfg := junitCfg{file: f, summary: true, withOutput: false}

	ss := readJUnit(cfg)
	require.Equal(t, 1, len(ss))
	require.Equal(t, 0, len(ss[0].Tests))
	require.Equal(t, 1, len(ss[0].Suites))
	require.Equal(t, 5, len(ss[0].Suites[0].Tests))
	require.Equal(t, 0, len(ss[0].Suites[0].Suites))
	require.Equal(t, "", ss[0].Suites[0].SystemOut)
}
