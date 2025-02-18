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

//go:build !no_github

package github

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/abc-inc/goava/base/casefmt"
	"github.com/abc-inc/heimdall"
	"github.com/abc-inc/heimdall/cli"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/plugin/root"
	"github.com/google/go-github/v69/github"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
)

type Inv func(cfg *ghCfg, cmd *cobra.Command) (any, error)

type ghCfg struct {
	// Common
	client *github.Client
	token  string
	host   string
	owner  string
	repo   string
	// Misc
	branch  *string
	name    string
	id      int64
	sarifID string
	// Page
	page    int
	perPage int
	// CodeScan
	state    *string
	severity *string
	// Dependabot
	ecosystem *string
	pkg       *string
	scope     *string
	sort      *string
	direction *string
	toolName  *string
	// PR
	reviewID int64
	// Repos
	affiliation string
	author      string
	base        string
	head        string
	path        string
	protected   *bool
	ref         string
	sha         string
	sinceID     int64
	sinceTime   time.Time
	tag         string
	typ         string
	user        string
	visibility  string
	// SecretScan
	// state      string
	secretType []string
	resolution []string
	// Teams
	slug             string
	discussionNumber int
	commentID        int
	groupID          int64
	orgID            int64
	teamID           int64
	displayName      *string
	role             string
	// Actions
	workflowFileName string
	inputs           map[string]interface{}
	// Enterprise
	enterprise string
	// Output
	cli.OutCfg
}

func newGHCfg() *ghCfg {
	direction, ecosystem, pkg, scope, severity, sort, state, typ := "", "", "", "", "", "created", "open", "all"
	return &ghCfg{
		direction: &direction,
		ecosystem: &ecosystem,
		pkg:       &pkg,
		scope:     &scope,
		severity:  &severity,
		sort:      &sort,
		state:     &state,
		typ:       typ,
	}
}

const envHelp = `
GH_TOKEN        <PERSONAL_ACCESS_TOKEN>
GH_HOST         https://api.github.com
GH_OWNER        <OWNER>
GH_REPO         <REPO>
GITHUB_TOKEN    <PERSONAL_ACCESS_TOKEN>
GITHUB_API_URL  https://api.github.com
`

func NewGitHubCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "github <subcommand>",
		Short:       "Query information from GitHub",
		GroupID:     cli.ServiceGroup,
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"help:environment": envHelp},
	}

	cmd.AddCommand(
		NewCodeScanCmd(),
		NewDependabotCmd(),
		NewDepGraphCmd(),
		NewMarkdownCmd(),
		NewPRCmd(),
		NewRepoCmd(),
		NewSecretScanCmd(),
		NewTeamsCmd(),
		NewUsersCmd(),
	)

	return cmd
}

func init() {
	root.GetRootCmd().AddCommand(NewGitHubCmd())
}

func addRepoFlags(cfg *ghCfg, cmd *cobra.Command) {
	cfg.token = os.Getenv("GH_TOKEN")
	if cfg.token == "" {
		cfg.token = os.Getenv("GITHUB_TOKEN")
	}

	cfg.host, cfg.owner, cfg.repo = os.Getenv("GH_HOST"), os.Getenv("GH_OWNER"), os.Getenv("GH_REPO")
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)

	cmd.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "Owner/org of the repository")
	cmd.Flags().StringVar(&cfg.repo, "repo", cfg.repo, "Repository name")
}

func setHostOwnerRepo(cfg *ghCfg, host, owner, repo string) {
	repoParts := strings.SplitN(repo, "/", 3)
	h, o, r := host, owner, repoParts[len(repoParts)-1]
	if h != "" {
		cfg.host = h
	}
	if o != "" {
		cfg.owner = o
	}
	if r != "" {
		cfg.repo = r
	}

	if len(repoParts) >= 2 {
		cfg.owner = repoParts[len(repoParts)-2]
		if len(repoParts) == 3 {
			cfg.host = repoParts[0]
		}
	}
}

func addListFlags(cfg *ghCfg, cmd *cobra.Command) {
	cmd.Flags().IntVar(&cfg.page, "page", cfg.page, "Page number")
	cmd.Flags().IntVar(&cfg.perPage, "per-page", cfg.perPage, "Items per page")
}

func addItemFlags(cfg *ghCfg, cmd *cobra.Command) {
	cmd.Flags().Int64Var(&cfg.id, "id", cfg.id, "Unique ID of the item")
	internal.MustNoErr(cmd.MarkFlagRequired("id"))
}

func createCmds(cfg *ghCfg, svcTyp reflect.Type, inv Inv, list []string) (cmds []*cobra.Command) {
	ctxTyp := reflect.TypeOf(context.Background())
	respType := reflect.TypeOf(&github.Response{})

	for i := 0; i < svcTyp.NumMethod(); i++ {
		m := svcTyp.Method(i)
		mTyp := m.Type

		if strings.HasSuffix(m.Name, "ForEnterprise") || strings.HasSuffix(m.Name, "ForOrg") ||
			strings.Contains(m.Name, "CodeQL") ||
			strings.Contains(m.Name, "OrgAlerts") || strings.Contains(m.Name, "OrgPublicKey") || strings.Contains(m.Name, "OrgSecret") ||
			strings.Contains(m.Name, "RepoPublicKey") || strings.Contains(m.Name, "Ruleset") || strings.Contains(m.Name, "CodeOfConduct") ||
			strings.HasSuffix(m.Name, "CodeownersErrors") || strings.HasSuffix(m.Name, "Forks") ||
			strings.HasSuffix(m.Name, "ByID") || strings.HasSuffix(m.Name, "ByOrg") ||
			strings.Contains(m.Name, "CommunityHealthMetrics") || strings.Contains(m.Name, "CodeFrequency") || strings.Contains(m.Name, "PunchCard") ||
			strings.Contains(m.Name, "Deployment") || strings.HasSuffix(m.Name, "Activity") ||
			m.Name == "GetHook" || m.Name == "ListHooks" || strings.Contains(m.Name, "HookDeliver") || strings.Contains(m.Name, "HookConfiguration") ||
			strings.Contains(m.Name, "Key") || strings.Contains(m.Name, "Invitation") ||
			strings.Contains(m.Name, "Page") || strings.Contains(m.Name, "Signatures") ||
			strings.Contains(m.Name, "App") || strings.Contains(m.Name, "Autolink") || strings.Contains(m.Name, "Project") ||
			strings.Contains(m.Name, "Context") || strings.Contains(m.Name, "Traffic") ||
			(svcTyp == reflect.TypeOf(&github.UsersService{}) && m.Name != "ListAll" && m.Name != "Get") {

			continue
		}

		if (strings.HasPrefix(m.Name, "Get") || strings.HasPrefix(m.Name, "List")) &&
			(mTyp.NumIn() > 2 && ctxTyp.Implements(mTyp.In(1))) &&
			(mTyp.NumOut() > 2 && mTyp.Out(mTyp.NumOut()-2).AssignableTo(respType) && mTyp.Out(mTyp.NumOut()-1).Name() == "error") {

			name := cmdName(m)
			ok := slices.Contains(list, name)
			internal.MustOkMsgf(ok, ok, "cannot find handler for '%s' '%s'", svcTyp, name)

			p := path.Join("docs", "github.com", "google", "go-github", "v69@v69.2.0", "github", svcTyp.Elem().Name(), m.Name+".txt")
			desc := string(internal.Must(heimdall.StaticFS.ReadFile(p)))
			_, desc, _ = strings.Cut(desc, " ")
			desc = string(unicode.ToUpper(rune(desc[0]))) + desc[1:]
			line, _, _ := strings.Cut(desc, ".")
			line = strings.ReplaceAll(line, "\n", " ") + "."

			kind, fTyp := "", m.Func.Type()
			if last := fTyp.In(fTyp.NumIn() - 1); last.AssignableTo(reflect.TypeOf(&github.ListOptions{})) {
				kind = "list"
			} else if last.Kind() == reflect.Ptr && last.Elem().Kind() == reflect.Struct {
				if lo, okList := last.Elem().FieldByName("ListOptions"); okList && lo.Type.AssignableTo(reflect.TypeOf(github.ListOptions{})) {
					kind = "list"
				}
			}

			cmds = append(cmds, &cobra.Command{
				Use:         name,
				Short:       line,
				Long:        desc,
				Args:        cobra.ExactArgs(0),
				Annotations: map[string]string{"method": m.Name, "kind": kind},
				Run: func(cmd *cobra.Command, args []string) {
					a, err := inv(cfg, cmd)
					var z *github.ErrorResponse
					if errors.As(err, &z) && z.Response.StatusCode == http.StatusNotFound {
						log.Warn().Err(err).Send()
						return
					} else if err != nil {
						log.Err(err).Send()
					} else {
						cli.Fmtln(a)
					}
				},
			})

			if kind == "list" {
				addListFlags(cfg, cmds[len(cmds)-1])
			}
		}
	}
	return
}

func cmdName(m reflect.Method) string {
	n := casefmt.LowerCamel{}.To(casefmt.LowerHyphen{}, m.Name)
	if n == "get" {
		return n
	}

	n = strings.TrimPrefix(n, "get-")
	n = strings.TrimPrefix(n, "list-")
	n = strings.ReplaceAll(n, "-i-d", "-id")
	n = strings.ReplaceAll(n, "i-d-p-", "idp-")
	n = strings.ReplaceAll(n, "s-a-r-i-f", "sarif")
	n = strings.ReplaceAll(n, "s-b-o-m", "sbom")
	n = strings.ReplaceAll(n, "-s-h-a1", "-sha1")
	return n
}

func newClient() *github.Client {
	t := os.Getenv("GITHUB_TOKEN")
	if t == "" {
		log.Fatal().Str("name", "GITHUB_TOKEN").Msg("undefined environment variable")
		return nil
	}
	url := strings.TrimSuffix(os.Getenv("GITHUB_API_URL"), "/")
	if url == "" {
		log.Fatal().Str("name", "GITHUB_API_URL").Msg("undefined environment variable")
		return nil
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t})
	httpClient := oauth2.NewClient(context.Background(), src)
	if strings.HasPrefix(url, "https://api.github.com") {
		return github.NewClient(httpClient)
	}
	return internal.Must(github.NewClient(httpClient).WithEnterpriseURLs(url, strings.TrimSuffix(strings.TrimSuffix(url, "/v3"), "/api")))
}

func getCtx(_ *ghCfg) context.Context {
	return context.Background()
}

func branchOrDefault(cfg *ghCfg, svc *github.RepositoriesService) {
	if *cfg.branch != "" {
		return
	}
	r, _, err := svc.Get(getCtx(cfg), cfg.owner, cfg.repo)
	internal.MustNoErr(err)
	cfg.branch = r.DefaultBranch
}

func defVal[T any](v *T) (t T) {
	if v == nil {
		return
	}
	return *v
}
