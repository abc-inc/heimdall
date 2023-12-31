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

package http

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/square/certigo/lib"
	"github.com/square/certigo/starttls"
	"golang.org/x/term"
)

type certCfg struct {
	console.OutCfg
	connectTo    string
	sni          string
	startTLS     string
	identity     string
	timeout      time.Duration
	pem          bool
	first        bool
	expectedName string
	verbose      bool
	expiresIn    uint32
}

func NewCertificateCmd() *cobra.Command {
	cfg := certCfg{timeout: 5 * time.Second}
	cmd := &cobra.Command{
		Use:   "certificate [flags] <host>[:<port>]",
		Short: "Verify SSL certificates",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.connectTo = args[0]
			res := verifyIt(cfg)
			if res != nil {
				console.Fmtln(res)
				if time.Now().AddDate(0, 0, int(cfg.expiresIn)).After(res.Certificates[0].NotAfter) {
					os.Exit(2)
				}
			}
		},
	}

	cmd.Flags().StringVar(&cfg.sni, "name", cfg.sni, "Override the server name used for Server Name Indication (SNI).")
	cmd.Flags().StringVar(&cfg.startTLS, "start-tls", cfg.startTLS, fmt.Sprintf("Enable StartTLS protocol; one of: %v.", starttls.Protocols))
	cmd.Flags().StringVar(&cfg.identity, "identity", cfg.identity, "With --start-tls, sets the DB user or SMTP EHLO name")
	cmd.Flags().DurationVar(&cfg.timeout, "timeout", cfg.timeout, "Timeout for connecting to remote server (can be '5m', '1s', etc).")
	cmd.Flags().BoolVar(&cfg.pem, "pem", cfg.pem, "Write output as PEM blocks instead of human-readable format.")
	cmd.Flags().BoolVar(&cfg.first, "first", cfg.first, "Only display the first certificate. This flag can be paired with --json or --pem.")
	cmd.Flags().StringVar(&cfg.expectedName, "expected-name", cfg.expectedName, "Name expected in the server TLS certificate. Defaults to name from SNI or, if SNI not overridden, the hostname to connect to.")
	cmd.Flags().Uint32Var(&cfg.expiresIn, "expires-in", 30, "Exit with status 2 if the certificate will expire within that many days.")

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func verifyIt(cfg certCfg) *Result {
	result := Result{}
	if cfg.startTLS == "" && cfg.identity != "" {
		log.Fatal().Msg("--identity can only be used with --start-tls")
	}
	connState, cri, err := starttls.GetConnectionState(
		cfg.startTLS, cfg.sni, cfg.connectTo, cfg.identity,
		"", "", nil, cfg.timeout)
	if err != nil {
		log.Fatal().Str("server", cfg.connectTo).Dur("timeout", cfg.timeout).Err(err).Msg("error connecting to server")
	}

	result.TLSConnectionState = connState
	result.CertificateRequestInfo = cri
	for _, cert := range connState.PeerCertificates {
		if cfg.pem {
			internal.MustNoErr(pem.Encode(os.Stdout, lib.EncodeX509ToPEM(cert, nil)))
			if cfg.first {
				break
			}
		} else {
			result.Certificates = append(result.Certificates, cert)
		}
	}
	if cfg.pem {
		return nil
	}

	// Determine what name the server's certificate should match
	var expectedNameInCertificate string
	switch {
	case cfg.expectedName != "":
		// Use the explicitly provided name
		expectedNameInCertificate = cfg.expectedName
	case cfg.sni != "":
		// Use the provided SNI
		expectedNameInCertificate = cfg.sni
	default:
		// Use the hostname/IP from the connect string
		expectedNameInCertificate = strings.Split(cfg.connectTo, ":")[0]
	}
	verifyResult := lib.VerifyChain(connState.PeerCertificates, connState.OCSPResponse, expectedNameInCertificate, "")
	result.VerifyResult = &verifyResult

	certList := result.Certificates
	if cfg.first && len(certList) > 0 {
		certList = certList[:1]
		result.Certificates = certList
	}

	if !strings.HasPrefix(cfg.OutCfg.Output, "text") {
		return &result
	}

	if !cfg.pem {
		fmt.Printf("%s\n\n", lib.EncodeTLSInfoToText(result.TLSConnectionState, result.CertificateRequestInfo))

		termWidth, _, _ := term.GetSize(0)
		for i, cert := range result.Certificates {
			fmt.Printf("** CERTIFICATE %d **\n", i+1)
			fmt.Printf("%s\n\n", lib.EncodeX509ToText(cert, termWidth, cfg.verbose))
		}
		lib.PrintVerifyResult(os.Stdout, *result.VerifyResult)
		if time.Now().AddDate(0, 0, int(cfg.expiresIn)).After(result.Certificates[0].NotAfter) {
			os.Exit(2)
		}
	}
	return nil
}

type Validity struct {
	Days int `json:"days" yaml:"days"`
}

type Result struct {
	lib.SimpleResult
	Validity Validity `json:"validity" yaml:"validity"`
}

func (r Result) MarshalJSON() ([]byte, error) {
	certs := make([]interface{}, len(r.Certificates))
	var expDate time.Time
	for i, c := range r.Certificates {
		certs[i] = lib.EncodeX509ToObject(c)
		if expDate.IsZero() || c.NotAfter.Before(expDate) {
			expDate = c.NotAfter
		}
	}

	out := map[string]interface{}{}
	out["certificates"] = certs
	if r.VerifyResult != nil {
		out["verify_result"] = r.VerifyResult
	}
	if r.TLSConnectionState != nil {
		out["tls_connection"] = lib.EncodeTLSToObject(r.TLSConnectionState)
	}
	if r.CertificateRequestInfo != nil {
		encoded, err := lib.EncodeCRIToObject(r.CertificateRequestInfo)
		if err != nil {
			return nil, err
		}
		out["certificate_request_info"] = encoded
	}

	maxDays := int(time.Until(expDate).Truncate(24*time.Hour).Hours()) / 24
	out["validity"] = Validity{maxDays}

	return json.Marshal(out)
}

func (r Result) MarshalYAML() (interface{}, error) {
	j, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	return m, json.Unmarshal(j, &m)
}
