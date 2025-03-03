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

package cli

import (
	"bytes"
	stdlog "log"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const ErrKey = "ERROR"

// init sets up the global log configuration.
func init() {
	stdlog.SetFlags(0)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.TimeFieldFormat = "15:04:05"
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:           os.Stderr,
		NoColor:       !isatty.IsTerminal(os.Stderr.Fd()),
		TimeFormat:    zerolog.TimeFieldFormat,
		FieldsExclude: []string{ErrKey},
		FormatExtra: func(m map[string]interface{}, buffer *bytes.Buffer) error {
			if v, ok := m[ErrKey]; ok {
				buffer.WriteRune('\n')
				buffer.WriteString(IO.ColorScheme().Redf("%v", v))
			}
			return nil
		},
	})
}
