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

//go:build !no_artifactory

package artifactory

import (
	"os"
	"strings"

	"github.com/abc-inc/heimdall/internal"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/jfrog/jfrog-client-go/config"
	jflog "github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/cobra"
)

func NewArtifactoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifactory <subcommand>",
		Short: "Various Artifactory-related commands",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewDownloadCmd(),
		NewListCmd(),
	)

	return cmd
}

func newRtManager() artifactory.ArtifactoryServicesManager {
	url, ok := os.LookupEnv("ARTIFACTORY_BASE_URL")
	internal.MustOkMsgf(url, ok, "environment variable '%s' must be set", "ARTIFACTORY_BASE_URL")

	tok, ok := os.LookupEnv("ARTIFACTORY_TOKEN")
	internal.MustOkMsgf(tok, ok, "environment variable '%s' must be set", "ARTIFACTORY_TOKEN")

	rtDetails := auth.NewArtifactoryDetails()
	rtDetails.SetUrl(strings.TrimSuffix(url, "/") + "/")
	rtDetails.SetAccessToken(tok)

	svcCfg := internal.Must(config.NewConfigBuilder().SetServiceDetails(rtDetails).Build())
	rtManager := internal.Must(artifactory.New(svcCfg))
	return rtManager
}

func init() {
	jflog.SetLogger(jflog.NewLoggerWithFlags(jflog.WARN, os.Stderr, 0))
}
