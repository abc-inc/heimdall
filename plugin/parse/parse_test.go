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

package parse_test

import (
	"testing"

	"github.com/abc-inc/heimdall/plugin/parse"
	"github.com/stretchr/testify/require"
)

func TestSplitNamePrefixType(t *testing.T) {
	require.Equal(t, parse.Input{File: "a.xml", Typ: "xml"}, parse.SplitNamePrefixType("a.xml"))
	require.Equal(t, parse.Input{File: "a.xml", Typ: "xml"}, parse.SplitNamePrefixType("a.xml:"))
	require.Equal(t, parse.Input{File: "a.xml", Typ: ""}, parse.SplitNamePrefixType("a.xml::"))
	require.Equal(t, parse.Input{File: "a.xml", Alias: "d", Typ: "xml"}, parse.SplitNamePrefixType("a.xml:d"))
	require.Equal(t, parse.Input{File: "a.config", Alias: "d", Typ: "xml"}, parse.SplitNamePrefixType("a.config:d:xml"))
	require.Equal(t, parse.Input{File: "a/b/c.config", Alias: "", Typ: "xml"}, parse.SplitNamePrefixType("a/b/c.config::xml"))
	require.Equal(t, parse.Input{File: "/a/b/c.config", Alias: "", Typ: "xml"}, parse.SplitNamePrefixType("/a/b/c.config::xml"))
	require.Equal(t, parse.Input{File: "file:///a/b/c.config", Alias: "", Typ: "xml"}, parse.SplitNamePrefixType("file:///a/b/c.config::xml"))
	require.Equal(t, parse.Input{File: `C:\a\b\c.config`, Alias: "", Typ: "xml"}, parse.SplitNamePrefixType(`C:\a\b\c.config::xml`))
}
