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

//go:build !no_properties

package properties

import (
	"errors"
	"strings"

	"github.com/magiconair/properties"
	"github.com/spf13/pflag"
)

type Enc properties.Encoding

var _ pflag.Value = (*Enc)(nil)

const (
	def        = Enc(0)
	UTF_8      = Enc(properties.UTF8)
	ISO_8859_1 = Enc(properties.ISO_8859_1)
)

// String is used both by fmt.Print and by Cobra in help text.
func (e *Enc) String() string {
	return []string{"UTF-8", "UTF-8", "ISO-8859-1"}[*e]
}

// Set changes the value of this receiver.
func (e *Enc) Set(v string) error {
	def, utf_8, iso_8859_1 := def, UTF_8, ISO_8859_1
	switch v {
	case def.String(), utf_8.String():
		*e = UTF_8
	case iso_8859_1.String():
		*e = ISO_8859_1
	default:
		return errors.New("must be one of: " + strings.Join([]string{utf_8.String(), iso_8859_1.String()}, ", "))
	}
	return nil
}

// Type is only used in help text.
func (e *Enc) Type() string {
	return "string"
}
