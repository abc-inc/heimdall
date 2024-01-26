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

	"github.com/abc-inc/heimdall/plugin/xml"
	"github.com/abc-inc/heimdall/res"
	"github.com/stretchr/testify/require"
)

func TestListServlets(t *testing.T) {
	app := listServlets(xml.ReadXML(filepath.Join(res.GetRootDir(), "testdata", "web.xml")))
	require.Equal(t, 2, len(app.Mapping))

	require.Equal(t, "SpnegoServlet", app.Mapping[0].Servlet.String())
	require.Equal(t, "com.rbinternational.security.spnego.SpnegoServlet", app.Mapping[0].Servlet[ServletClass])
	require.Equal(t, 2, len(app.Mapping[0].Chain))
	require.Equal(t, 2, len(app.Mapping[0].Chain[0].Filters))
	require.Equal(t, "/SpnegoServlet", app.Mapping[0].Chain[0].Pattern)
	require.Equal(t, "SpnegoHttpFilter", app.Mapping[0].Chain[0].Filters[1].String())

	require.Equal(t, "/*", app.Mapping[0].Chain[1].Pattern)
	require.Equal(t, "true", app.Mapping[0].Chain[1].Filters[0][AsyncSupported])
}
