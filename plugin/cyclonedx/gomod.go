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
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/app"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/local"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type goModCfg struct {
	assertLicenses bool
	files          bool
	json           bool
	licenses       bool
	main           string
	noserial       bool
	outputVersion  string
	packages       bool
	serial         string
	std            bool
	verbose        bool
}

func NewGoModCmd() *cobra.Command {
	cfg := goModCfg{outputVersion: "1.4"}
	cmd := &cobra.Command{
		Use:   "gomod",
		Short: "Generate an SBOM for a Go application",
		Example: heredoc.Doc(`
			heimdall cyclonedx gomod --assert-licenses --json --licenses --main cmd/heimdall --packages
		`),
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				args = append(args, internal.Must(os.Getwd()))
			}
			internal.Must(writeBom(cfg, args[0]))
		},
	}

	cmd.Flags().BoolVar(&cfg.assertLicenses, "assert-licenses", cfg.assertLicenses, "Assert detected licenses")
	cmd.Flags().BoolVar(&cfg.files, "files", cfg.files, "Include files")
	cmd.Flags().BoolVar(&cfg.json, "json", cfg.json, "Output in JSON")
	cmd.Flags().BoolVar(&cfg.licenses, "licenses", cfg.licenses, "Perform license detection")
	cmd.Flags().StringVar(&cfg.main, "main", cfg.main, "Path to the application's main package, relative to MODULE_PATH")
	cmd.Flags().BoolVar(&cfg.noserial, "no-serial", cfg.noserial, "Omit serial number")
	cmd.Flags().StringVar(&cfg.outputVersion, "output-version", cfg.outputVersion, "Output spec version (1.4, 1.3, 1.2, 1.1, 1.0)")
	cmd.Flags().BoolVar(&cfg.packages, "packages", cfg.packages, "Include packages")
	cmd.Flags().StringVar(&cfg.serial, "serial", cfg.serial, "Serial number")
	cmd.Flags().BoolVar(&cfg.std, "std", cfg.std, "Include Go standard library as component and dependency of the module")

	return cmd
}

func writeBom(cfg goModCfg, name string) (*cdx.BOM, error) {
	var licenseDetector licensedetect.Detector
	if cfg.licenses {
		licenseDetector = local.NewDetector(log.Logger)
	}

	generator, err := app.NewGenerator(name,
		app.WithLogger(log.Logger),
		app.WithIncludeFiles(cfg.files),
		app.WithIncludePackages(cfg.packages),
		app.WithIncludeStdlib(cfg.std),
		app.WithLicenseDetector(licenseDetector),
		app.WithMainDir(cfg.main))
	if err != nil {
		return nil, err
	}

	bom, err := generator.Generate()
	if err != nil {
		return nil, err
	}

	err = setSerialNumber(bom, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to set serial number: %w", err)
	}
	err = addCommonMetadata(log.Logger, bom)
	if err != nil {
		return nil, fmt.Errorf("failed to add common metadata: %w", err)
	}
	if cfg.assertLicenses {
		assertLicenses(bom)
	}

	return bom, writeBOM(bom, cfg)
}
