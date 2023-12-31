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

package res

import (
	"io"
	"net/http"
	"os"
	"strings"
)

var client http.Client

func init() {
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
	client = http.Client{Transport: t}
}

func Open(uri string) (io.ReadCloser, error) {
	uri = os.ExpandEnv(uri)
	if IsURL(uri) {
		resp, err := client.Get(uri)
		if err != nil {
			return nil, err
		}
		return resp.Body, err
	} else if uri == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(uri)
}

func IsURL(uri string) bool {
	return strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "http://")
}
