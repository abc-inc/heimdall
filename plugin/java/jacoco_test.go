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

	"github.com/abc-inc/heimdall/test"
	"github.com/stretchr/testify/require"
)

func TestProcessJaCoCo(t *testing.T) {
	cfg := jaCoCoCfg{files: []string{filepath.Join(test.GetRootDir(), "testdata", "jacoco.csv")}, summary: false}

	crs := processJaCoCo(cfg)
	require.Equal(t, 299, len(crs))

	tot := aggregateJaCoCo(crs...)
	require.Less(t, 0.94, float64(tot.LineCov)/float64(tot.LineCov+tot.LineMis))
}
