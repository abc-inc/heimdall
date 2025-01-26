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

//go:build !no_cyclonedx

package cyclonedx

import (
	"github.com/abc-inc/heimdall/console"
	"github.com/spf13/cobra"
)

func NewCycloneDXCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cyclonedx",
		Short:   "Process CycloneDX SBOM files",
		GroupID: console.SoftwareGroup,
		Args:    cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewReadCmd(),
		NewGoModCmd(),
	)

	return cmd
}
