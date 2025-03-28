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

//go:build !no_java

package java

import (
	"crypto/rsa"
	"crypto/x509"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Entry any

type keystoreCfg struct {
	cli.OutCfg
	file     string
	password []byte
}

const envHelp = `
KEYSTORE_PASSWORD  <PASSWORD>
`

func NewKeystoreCmd() *cobra.Command {
	cfg := keystoreCfg{password: []byte(os.Getenv("KEYSTORE_PASSWORD"))}
	if len(cfg.password) == 0 {
		cfg.password = []byte("")
	}

	cmd := &cobra.Command{
		Use:   "keystore",
		Short: "Read a Java keystore and display its entries",
		Example: heredoc.Doc(`
			heimdall java keystore -f keystore.jks
		`),
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"help:environment": envHelp},
		Run: func(cmd *cobra.Command, args []string) {
			cli.Fmtln(listEntries(cfg))
		},
	}

	cli.AddFileFlag(cmd, &cfg.file, "Path to the keystore file")
	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func listEntries(cfg keystoreCfg) (es []Entry) {
	ks := readKeyStore(cfg)
	for _, a := range ks.Aliases() {
		if ks.IsTrustedCertificateEntry(a) {
			entry := internal.Must(ks.GetTrustedCertificateEntry(a))
			cert := internal.Must(x509.ParseCertificate(entry.Certificate.Content))
			if _, ok := cert.PublicKey.(*rsa.PublicKey); ok {
				// N is a big.Int, which cannot be handled by a lot of tools like JQ, so it is set to nil.
				cert.PublicKey = nil
			}
			es = append(es, entry, cert)
		} else if ks.IsPrivateKeyEntry(a) {
			entry := internal.Must(ks.GetPrivateKeyEntry(a, cfg.password))
			cert := internal.Must(x509.ParseCertificate(entry.CertificateChain[0].Content))
			es = append(es, entry, cert)
		}
	}
	return
}

func readKeyStore(cfg keystoreCfg) keystore.KeyStore {
	f := internal.Must(res.Open(cfg.file))
	defer func() { _ = f.Close() }()

	ks := keystore.New()
	if err := ks.Load(f, cfg.password); err != nil {
		log.Fatal().Err(err).Send()
	}
	return ks
}
