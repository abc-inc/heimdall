// Copyright 2025 The Heimdall authors
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

//go:build !no_atlassian && !no_jira

package jira

import (
	"io"
	"os"
	"path/filepath"

	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraAttachmentCfg struct {
	issueID string
	attId   string
	file    string
	jiraCfg
}

func NewDownloadAttachmentCmd() *cobra.Command {
	cfg := jiraAttachmentCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "download-attachments",
		Short: "Download attachments from Jira issues",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			downloadAttachment(client, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.attId, "attachment-id", cfg.attId, "Attachment ID")
	cmd.Flags().StringVar(&cfg.file, "file", cfg.file, "File to save the attachment to")
	internal.MustNoErr(cmd.MarkFlagRequired("attachment-id"))
	internal.MustNoErr(cmd.MarkFlagRequired("file"))
	return cmd
}

func NewUploadAttachmentCmd() *cobra.Command {
	cfg := jiraAttachmentCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "upload-attachment",
		Short: "Upload attachments to Jira issues",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			cli.Fmtln(uploadAttachment(client, cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.issueID, "key", cfg.issueID, "Issue ID or key")
	cmd.Flags().StringVar(&cfg.file, "file", cfg.file, "File to upload")
	internal.MustNoErr(cmd.MarkFlagRequired("key"))
	internal.MustNoErr(cmd.MarkFlagRequired("file"))
	return cmd
}

// downloadAttachment downloads an attachment from a Jira issues.
func downloadAttachment(client *jira.Client, cfg jiraAttachmentCfg) {
	resp := internal.Must(client.Issue.DownloadAttachment(cfg.attId))
	defer func() { _ = resp.Body.Close() }()

	f := internal.Must(os.OpenFile(cfg.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640))
	defer func() { _ = f.Close() }()

	_ = internal.Must(io.Copy(f, resp.Body))
}

// uploadAttachment uploads an attachment to a Jira issue.
func uploadAttachment(client *jira.Client, cfg jiraAttachmentCfg) *[]jira.Attachment {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()

	return handle(client.Issue.PostAttachment(cfg.issueID, r, filepath.Base(cfg.file)))
}
