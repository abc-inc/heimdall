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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/abc-inc/heimdall/cli"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/sha3"
)

func assertLicenses(bom *cdx.BOM) {
	if bom == nil {
		return
	}

	if bom.Metadata != nil {
		assertComponentLicenses(bom.Metadata.Component)
	}

	if bom.Components != nil {
		for i := range *bom.Components {
			assertComponentLicenses(&(*bom.Components)[i])
		}
	}
}

func assertComponentLicenses(c *cdx.Component) {
	if c == nil {
		return
	}

	if c.Evidence != nil && c.Evidence.Licenses != nil {
		c.Licenses = c.Evidence.Licenses

		if c.Evidence.Copyright != nil {
			c.Evidence.Licenses = nil
		} else {
			c.Evidence = nil
		}
	}

	if c.Components != nil {
		for i := range *c.Components {
			assertComponentLicenses(&(*c.Components)[i])
		}
	}
}

func buildToolMetadata(logger zerolog.Logger) (*cdx.Tool, error) {
	toolExePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	toolHashes, err := calculateFileHashes(logger, toolExePath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tool hashes: %w", err)
	}

	return &cdx.Tool{
		Vendor:  "Heimdall",
		Name:    "heimdall",
		Version: cli.Version,
		Hashes:  &toolHashes,
		ExternalReferences: &[]cdx.ExternalReference{
			{
				Type: cdx.ERTypeVCS,
				URL:  "https://github.com/abc-inc/heimdall",
			},
			{
				Type: cdx.ERTypeWebsite,
				URL:  "https://abc-inc.github.io/heimdall-web/",
			},
		},
	}, nil
}

func calculateFileHashes(logger zerolog.Logger, filePath string, algos ...cdx.HashAlgorithm) ([]cdx.Hash, error) {
	if len(algos) == 0 {
		return make([]cdx.Hash, 0), nil
	}

	logger.Debug().
		Str("file", filePath).
		Interface("algos", algos).
		Msg("calculating file hashes")

	hashMap := make(map[cdx.HashAlgorithm]hash.Hash)
	hashWriters := make([]io.Writer, 0)

	for _, algo := range algos {
		var hashWriter hash.Hash

		switch algo { //nolint:exhaustive
		case cdx.HashAlgoMD5:
			hashWriter = md5.New()
		case cdx.HashAlgoSHA1:
			hashWriter = sha1.New()
		case cdx.HashAlgoSHA256:
			hashWriter = sha256.New()
		case cdx.HashAlgoSHA384:
			hashWriter = sha512.New384()
		case cdx.HashAlgoSHA512:
			hashWriter = sha512.New()
		case cdx.HashAlgoSHA3_256:
			hashWriter = sha3.New256()
		case cdx.HashAlgoSHA3_512:
			hashWriter = sha3.New512()
		default:
			return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
		}

		hashWriters = append(hashWriters, hashWriter)
		hashMap[algo] = hashWriter
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	multiWriter := io.MultiWriter(hashWriters...)
	if _, err = io.Copy(multiWriter, file); err != nil {
		return nil, err
	}

	cdxHashes := make([]cdx.Hash, 0, len(hashMap))
	for _, algo := range algos { // Don't iterate over hashMap, as it doesn't retain order
		cdxHashes = append(cdxHashes, cdx.Hash{
			Algorithm: algo,
			Value:     fmt.Sprintf("%x", hashMap[algo].Sum(nil)),
		})
	}

	return cdxHashes, nil
}
