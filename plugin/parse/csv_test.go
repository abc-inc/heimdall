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

//go:build !no_parse && !no_csv

package parse

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestDecodeCSV(t *testing.T) {
	r := internal.Must(res.Open(filepath.Join(test.GetRootDir(), "testdata", "jacoco.csv")))
	defer func(r io.ReadCloser) { _ = r.Close() }(r)
	csv := internal.Must(decodeCSVRecords(r)).([][]string)
	require.Equal(t, "METHOD_COVERED", csv[0][12])
	require.Equal(t, 300, len(csv))
}
