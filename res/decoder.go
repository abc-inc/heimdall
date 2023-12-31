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
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"

	"github.com/abc-inc/heimdall/internal"
	"github.com/magiconair/properties"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var Decoders = make(map[string]func(r io.Reader) (map[string]any, error))

func init() {
	Decoders["json"] = func(r io.Reader) (m map[string]any, err error) {
		all, err := io.ReadAll(r)
		if err != nil || len(all) == 0 {
			return nil, err
		}

		in := strings.NewReader(string(all))
		d := json.NewDecoder(in)

		switch all[0] {
		case '[':
			data := make([]any, 0)
			internal.MustNoErr(d.Decode(&data))
			m = map[string]any{"data": data}
		case '{':
			err = d.Decode(&m)
		default:
			log.Fatal().Msgf(`expected JSON but got "%b"`, all[0])
		}
		return
	}

	Decoders["properties"] = func(r io.Reader) (map[string]any, error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		props, err := properties.Load(b, properties.UTF8)
		if err != nil {
			return nil, err
		}

		m := make(map[string]any, props.Len())
		for k, v := range props.Map() {
			m[k] = v
		}
		return m, nil
	}

	Decoders["xml"] = func(r io.Reader) (m map[string]any, err error) {
		d := xml.NewDecoder(r)
		d.Strict = true
		err = d.Decode(&m)
		return
	}

	Decoders["yaml"] = func(r io.Reader) (m map[string]any, err error) {
		d := yaml.NewDecoder(r)
		err = d.Decode(&m)
		return
	}

	Decoders["yml"] = Decoders["yaml"]
}
