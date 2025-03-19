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
	"cmp"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/andygrunwald/go-jira"
	"github.com/spf13/cobra"
)

type jiraAttCfg struct {
	keyOrID string
	attId   string
	attName string
	file    string
	dir     string
	jiraCfg
}

func NewAttachmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Work with Jira attachments",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(
		NewAttachmentDownloadCmd(),
		NewAttachmentsDownloadCmd(),
		NewAttachmentsListCmd(),
		NewAttachmentUploadCmd(),
	)

	return cmd
}

func NewAttachmentDownloadCmd() *cobra.Command {
	cfg := jiraAttCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download an attachment from a Jira issue",
		Example: heredoc.Doc(`
			heimdall jira attachment download --key ABC-123 --name users.csv
			# or write to standard output (-) directly
			heimdall jira attachment download --file - --id 12345
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			download(client, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.dir, "directory", "C", cfg.dir, "Change to the directory before downloading attachments")
	cmd.Flags().StringVarP(&cfg.file, "file", "O", cfg.file, "File to save the attachment to (use '-' for standard output, default is the name of the attachment)")
	cmd.Flags().StringVar(&cfg.attId, "id", cfg.attId, "Attachment ID to download")
	cmd.Flags().StringVar(&cfg.keyOrID, "key", cfg.keyOrID, "Issue key or ID")
	cmd.Flags().StringVar(&cfg.attName, "name", cfg.attName, "Attachment name to download")

	cmd.MarkFlagsMutuallyExclusive("id", "name")
	cmd.MarkFlagsOneRequired("id", "name")
	cmd.MarkFlagsRequiredTogether("key", "name")
	return cmd
}

func NewAttachmentsDownloadCmd() *cobra.Command {
	cfg := jiraAttCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "downloads",
		Short: "Downloads all matching attachments from a Jira issue",
		Example: heredoc.Doc(`
			# wildcard match with a glob pattern
			heimdall jira attachment downloads --key ABC-123 --name '*.csv'
			# regular expressions must start with ^
			heimdall jira attachment downloads --directory /tmp --key ABC-123 --name '^.*2025\.png'
		`),
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			downloadMatching(client, cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.dir, "directory", "C", cfg.dir, "Change to the directory before downloading attachments")
	cmd.Flags().StringVar(&cfg.keyOrID, "key", cfg.keyOrID, "Issue key or ID")
	cmd.Flags().StringVar(&cfg.attName, "name", "*", "Attachment name(s) to download (can be a wildcard or regular expression)")
	internal.MustNoErr(cmd.MarkFlagRequired("key"))
	return cmd
}

func NewAttachmentsListCmd() *cobra.Command {
	cfg := jiraAttCfg{jiraCfg: *newJiraCfg(), attName: "*"}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List attachments from a Jira issue",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			cli.Fmtln(list(client, cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.keyOrID, "key", cfg.keyOrID, "Issue key or ID")

	cli.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("key"))
	return cmd
}

func NewAttachmentUploadCmd() *cobra.Command {
	cfg := jiraAttCfg{jiraCfg: *newJiraCfg()}
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload an attachment to a Jira issue",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			client := internal.Must(newClient(cfg.apiURL, cfg.token))
			cli.Fmtln(upload(client, cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.keyOrID, "key", cfg.keyOrID, "Issue key or ID")
	cmd.Flags().StringVar(&cfg.file, "file", cfg.file, "File to upload")

	internal.MustNoErr(cmd.MarkFlagRequired("key"))
	internal.MustNoErr(cmd.MarkFlagRequired("file"))
	return cmd
}

// download downloads an attachment from a Jira issue.
func download(client *jira.Client, cfg jiraAttCfg) {
	if cfg.dir != "" {
		internal.MustNoErr(os.Chdir(cfg.dir))
	}

	if cfg.attId == "" {
		atts := list(client, cfg)
		filename := func(e *jira.Attachment) string { return e.Filename }
		internal.MustOkMsgf(atts, len(atts) == 1, "expected exactly one attachment, found: %s", res.Strings(filename, atts...))
		cfg.attId, cfg.attName = atts[0].ID, atts[0].Filename
	}

	resp := internal.Must(client.Issue.DownloadAttachment(cfg.attId))
	defer func() { _ = resp.Body.Close() }()

	if cfg.file == "-" {
		_, _ = io.Copy(os.Stdout, resp.Body)
		return
	}
	if cfg.file == "" {
		cfg.file = cfg.attName
	}

	f := internal.Must(os.OpenFile(cfg.file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640))
	defer func() { _ = f.Close() }()
	_ = internal.Must(io.Copy(f, resp.Body))
}

// downloadMatching downloads all attachments matching a given name, glob, or regex from a Jira issue.
func downloadMatching(client *jira.Client, cfg jiraAttCfg) {
	if cfg.dir != "" {
		internal.MustNoErr(os.Chdir(cfg.dir))
	}

	atts := list(client, cfg)
	internal.MustOkMsgf(atts, len(atts) > 0, "")
	for _, att := range atts {
		fileCfg := cfg
		fileCfg.attId, fileCfg.file = att.ID, att.Filename
		download(client, fileCfg)
	}
}

// list returns a list of files attached to a Jira issue.
func list(client *jira.Client, cfg jiraAttCfg) []*jira.Attachment {
	filename := func(e *jira.Attachment) string { return e.Filename }
	atts := handle(client.Issue.Get(cfg.keyOrID, &jira.GetQueryOptions{Fields: "attachment"})).Fields.Attachments
	matches := slices.Collect(res.Seq(res.MatchAny[*jira.Attachment](filename, cfg.attName), atts...))
	if len(matches) > 1 {
		slices.SortFunc(matches, func(a, b *jira.Attachment) int {
			return cmp.Or(strings.Compare(a.Filename, b.Filename),
				internal.Must(strconv.Atoi(b.ID))-internal.Must(strconv.Atoi(a.ID)))
		})
		matches = slices.CompactFunc(matches, func(a *jira.Attachment, b *jira.Attachment) bool { return a.Filename == b.Filename })
	}
	return matches
}

// upload uploads an attachment to a Jira issue.
func upload(client *jira.Client, cfg jiraAttCfg) *[]jira.Attachment {
	r := internal.Must(res.Open(cfg.file))
	defer func() { _ = r.Close() }()

	return handle(client.Issue.PostAttachment(cfg.keyOrID, r, filepath.Base(cfg.file)))
}
