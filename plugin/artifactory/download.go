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
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/jfrog/build-info-go/entities"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/spf13/cobra"
)

type downloadCfg struct {
	console.OutCfg
	pattern string
	target  string
	noFlat  bool
}

func NewDownloadCmd() *cobra.Command {
	cfg := downloadCfg{}
	cmd := &cobra.Command{
		Use:   "download <subcommand>",
		Short: "Download files from Artifactory",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(download(cfg))
		},
	}

	cmd.Flags().BoolVar(&cfg.noFlat, "no-flat", cfg.noFlat, "Do not flatten the target directory")
	cmd.Flags().StringVarP(&cfg.pattern, "pattern", "p", cfg.pattern, "Download only assets that match a glob pattern")
	cmd.Flags().StringVarP(&cfg.target, "target", "t", cfg.target, "Target path")
	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func download(cfg downloadCfg) map[string]entities.Artifact {
	rtManager := newRtManager()

	params := services.NewDownloadParams()
	params.Flat = !cfg.noFlat
	params.Pattern, params.Target = cfg.pattern, cfg.target
	params.Recursive = true

	sum := internal.Must(rtManager.DownloadFilesWithSummary(params))
	defer func() { _ = sum.Close() }()

	m := make(map[string]string)
	for item := new(clientutils.FileTransferDetails); sum.TransferDetailsReader.NextRecord(item) == nil; item = new(clientutils.FileTransferDetails) {
		m[item.SourcePath] = item.TargetPath
	}

	as := make(map[string]entities.Artifact)
	for item := new(utils.ArtifactDetails); sum.ArtifactsDetailsReader.NextRecord(item) == nil; item = new(utils.ArtifactDetails) {
		as[m[item.ArtifactoryPath]] = internal.Must(item.ToBuildInfoArtifact())
	}
	return as
}
