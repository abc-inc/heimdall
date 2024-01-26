// This file contains code of CycloneDX GoMod
// Copyright 2024 The Heimdall authors, OWASP Foundation
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
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func addCommonMetadata(logger zerolog.Logger, bom *cdx.BOM) error {
	if bom.Metadata == nil {
		bom.Metadata = &cdx.Metadata{}
	}

	tool, err := buildToolMetadata(logger)
	if err != nil {
		return fmt.Errorf("failed to build tool metadata: %w", err)
	}

	bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	bom.Metadata.Tools = &[]cdx.Tool{*tool}

	return nil
}

// setSerialNumber sets the serial number of a given BOM according to the provided SBOMOptions.
func setSerialNumber(bom *cdx.BOM, cfg goModCfg) error {
	if cfg.noserial {
		return nil
	}

	if cfg.serial == "" {
		bom.SerialNumber = uuid.New().URN()
	} else {
		serial, err := uuid.Parse(cfg.serial)
		if err != nil {
			return err
		}
		bom.SerialNumber = serial.URN()
	}

	return nil
}

// writeBOM writes the given bom according to the provided OutputOptions.
func writeBOM(bom *cdx.BOM, cfg goModCfg) error {
	var outputFormat cdx.BOMFileFormat
	if cfg.json {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	outputVersion, err := parseSpecVersion(cfg.outputVersion)
	if err != nil {
		return fmt.Errorf("failed to parse output version: %w", err)
	}

	encoder := cdx.NewBOMEncoder(os.Stdout, outputFormat)
	encoder.SetPretty(true)

	if err = encoder.EncodeVersion(bom, outputVersion); err != nil {
		return fmt.Errorf("failed to encode sbom: %w", err)
	}

	return err
}

func parseSpecVersion(specVersion string) (sv cdx.SpecVersion, err error) {
	switch specVersion {
	case cdx.SpecVersion1_0.String():
		sv = cdx.SpecVersion1_0
	case cdx.SpecVersion1_1.String():
		sv = cdx.SpecVersion1_1
	case cdx.SpecVersion1_2.String():
		sv = cdx.SpecVersion1_2
	case cdx.SpecVersion1_3.String():
		sv = cdx.SpecVersion1_3
	case cdx.SpecVersion1_4.String():
		sv = cdx.SpecVersion1_4
	default:
		err = cdx.ErrInvalidSpecVersion
	}

	return
}
