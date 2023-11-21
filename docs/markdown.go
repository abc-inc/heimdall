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

package docs

import (
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/internal"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"
)

// WithoutIndentation removes margins to produce dense output.
func WithoutIndentation() glamour.TermRendererOption {
	overrides := []byte(heredoc.Doc(`
		{
			"document": { "margin": 0 },
			"code_block": { "margin": 0 }
		}`,
	))

	return glamour.WithStylesFromJSONBytes(overrides)
}

// WithWrap sets the word wrap to w columns.
func WithWrap(w int) glamour.TermRendererOption {
	return glamour.WithWordWrap(w)
}

// WithBaseURL sets the base URL for relative links to the location of the GitHub project.
func WithBaseURL() glamour.TermRendererOption {
	return glamour.WithBaseURL("https://github.com/abc-inc/heimdall/tree/master/docs/")
}

// Render returns the markdown rendered as a string.
// It does a series of checks to detect terminal width, color scheme, etc. to achieve the best possible formatting.
func Render(md string) (string, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return md, nil
	}

	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	if width > 120 {
		width = 0 // terminal width is sufficient - skip wrapping to improve performance
	}
	opts := []glamour.TermRendererOption{glamour.WithAutoStyle(), WithBaseURL(), glamour.WithEmoji(), WithoutIndentation(), WithWrap(width)}
	return internal.Must(glamour.NewTermRenderer(opts...)).Render(md)
}
