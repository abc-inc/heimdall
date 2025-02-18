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
	"encoding/csv"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/abc-inc/heimdall/internal"
	"github.com/spf13/viper"
)

func decodeCSV(r io.Reader) (any, error) {
	m := map[string][]string{}
	c := csv.NewReader(r)
	if d := viper.GetString("csv-delimiter"); len(d) >= 1 {
		ru, _ := utf8.DecodeRuneInString(d)
		c.Comma = ru
	}

	fs := internal.Must(c.Read())
	for _, f := range fs {
		m[f] = []string{}
	}

	c.ReuseRecord = true
	for {
		rec, err := c.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return m, err
		}
		for i, v := range rec {
			m[fs[i]] = append(m[fs[i]], v)
		}
	}

	return m, nil
}

func decodeCSVRecords(r io.Reader) (any, error) {
	var m [][]string
	c := csv.NewReader(r)
	if d := viper.GetString("csv-delimiter"); len(d) >= 1 {
		ru, _ := utf8.DecodeRuneInString(d)
		c.Comma = ru
	}

	fs := internal.Must(c.Read())
	m = append(m, fs)

	for {
		rec, err := c.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return m, err
		}
		m = append(m, rec)
	}

	return m, nil
}

func init() {
	Decoders["csv"] = decodeCSVRecords
}
