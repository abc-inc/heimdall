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

package internal

import (
	"reflect"

	"github.com/rs/zerolog/log"
)

// Must panics if the error is non-nil. It is intended for use in variable
// initializations such as
// f := internal.MustNoErr(os.OpenFile("notes.txt", os.O_RDONLY, 0600))
func Must[T any](a T, err error) T {
	MustNoErr(err)
	return a
}

// MustNoErr logs the error message and exits, if the error is non-nil. It is
// intended for use in statements that could return an error and continued
// execution is not meaningful.
func MustNoErr(err error) {
	if err != nil {
		log.Fatal().Err(err).Str("type", reflect.TypeOf(err).String()).Send()
	}
}

// MustOkMsgf logs the error message and exits, if the condition is false.
// It is intended for use in statements that could return a flag and continued
// execution is not meaningful.
func MustOkMsgf[T any](a T, ok bool, msg string, args ...any) T {
	if !ok {
		log.Fatal().Msgf(msg, args...)
	}
	return a
}
