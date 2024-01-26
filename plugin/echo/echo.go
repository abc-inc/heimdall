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

//go:build !no_echo

package echo

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/charmbracelet/gum/style"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewEchoCmd() *cobra.Command {
	cfg := style.Styles{Border: "none", Align: "left", Margin: "0 0", Padding: "0 0"}
	cmd := &cobra.Command{
		Use:   "echo",
		Short: "Apply coloring, borders, spacing to text",
		Example: heredoc.Doc(`
			heimdall echo --foreground=#ff0000 --bold "Error: something bad happened"
		`),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return echo(cfg, args...)
		},
	}

	viper.SetDefault("BORDER", cfg.Border)
	viper.SetDefault("ALIGN", cfg.Align)
	viper.SetDefault("MARGIN", cfg.Margin)
	viper.SetDefault("PADDING", cfg.Padding)

	cmd.Flags().StringVar(&cfg.Background, "background", cfg.Background, "Background color")
	cmd.Flags().StringVar(&cfg.Foreground, "foreground", cfg.Foreground, "Foreground color")
	cmd.Flags().StringVar(&cfg.Border, "border", cfg.Border, "Border style")
	cmd.Flags().StringVar(&cfg.BorderBackground, "border-background", cfg.BorderBackground, "Border background color")
	cmd.Flags().StringVar(&cfg.BorderForeground, "border-foreground", cfg.BorderForeground, "Border foreground color")
	cmd.Flags().StringVar(&cfg.Align, "align", cfg.Align, "Text alignment")
	cmd.Flags().IntVar(&cfg.Height, "height", cfg.Height, "Text height")
	cmd.Flags().IntVar(&cfg.Width, "width", cfg.Width, "Text width")
	cmd.Flags().StringVar(&cfg.Margin, "margin", cfg.Margin, "Text margin")
	cmd.Flags().StringVar(&cfg.Padding, "padding", cfg.Padding, "Text padding")
	cmd.Flags().BoolVar(&cfg.Bold, "bold", cfg.Bold, "Bold text")
	cmd.Flags().BoolVar(&cfg.Faint, "faint", cfg.Faint, "Faint text")
	cmd.Flags().BoolVar(&cfg.Italic, "italic", cfg.Italic, "Italicize text")
	cmd.Flags().BoolVar(&cfg.Strikethrough, "strikethrough", cfg.Strikethrough, "Strikethrough text")
	cmd.Flags().BoolVar(&cfg.Underline, "underline", cfg.Underline, "Underline text")
	return cmd
}

func echo(cfg style.Styles, args ...string) error {
	return style.Options{Text: args, Style: cfg}.Run()
}
