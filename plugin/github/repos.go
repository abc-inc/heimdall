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

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/repos\*.go ../../docs

package github

import (
	"reflect"
	"time"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewRepoCmd() *cobra.Command {
	var branch string
	var protected bool
	cfg := newGHCfg()
	cfg.branch = &branch
	cfg.protected = &protected
	cfg.visibility = "all"
	cfg.affiliation = "owner,collaborator,organization_member"

	cmd := &cobra.Command{
		Use:   "repositories",
		Short: "Handles communication with the repository related methods.",
		Example: heredoc.Doc(`
		`),
		Args: cobra.ExactArgs(0),
	}

	ops := []string{
		"actions-access-level",
		"actions-allowed",
		"actions-permissions",
		"admin-enforcement",
		"all",
		"all-topics",
		"archive-link",
		"automated-security-fixes",
		"branch",
		"branch-protection",
		"branches",
		"branches-head-commit",
		"collaborators",
		"combined-status",
		"comment",
		"comments",
		"commit",
		"commit-comments",
		"commit-raw",
		"commit-sha1",
		"commits",
		"compare-commits",
		"compare-commits-raw",
		"contents",
		"contributors",
		"contributors-stats",
		"download-contents",
		"download-contents-with-meta",
		"environment",
		"environments",
		"generate-release-notes",
		"get",
		"languages",
		"latest-release",
		"license",
		"list",
		"participation",
		"permission-level",
		"pre-receive-hook",
		"pre-receive-hooks",
		"pull-request-review-enforcement",
		"readme",
		"release",
		"release-by-tag",
		"release-asset",
		"release-assets",
		"releases",
		"required-status-checks",
		"rules-for-branch",
		"statuses",
		"tag-protection",
		"tag-protections",
		"tags",
		"team-restrictions",
		"teams",
		"topics",
		"user-restrictions",
		"vulnerability-alerts",
	}

	svcTyp := reflect.TypeOf(&github.RepositoriesService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execRepos, ops)...)
	for _, sub := range cmd.Commands() {
		if sub.Name() != "all" {
			addRepoFlags(cfg, sub)
		}

		switch sub.Name() {
		case "admin-enforcement":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "all":
			sub.Flags().Int64Var(&cfg.sinceID, "since", cfg.sinceID, "A repository ID. Only return repositories with an ID greater than this ID.")
		case "archive-link":
			addItemFlags(cfg, sub)
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "branch":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "branch-protection":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "branches-head-commit":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "branches":
			sub.Flags().BoolVar(cfg.protected, "protected", *cfg.protected, "Whether to return protected, unprotected, or all branches")
		case "combined-status":
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
		case "comment":
			addItemFlags(cfg, sub)
		case "commit-comments":
			sub.Flags().StringVar(&cfg.sha, "sha", cfg.sha, "SHA of the commit")
		case "commit-raw":
			sub.Flags().StringVar(&cfg.sha, "sha", cfg.sha, "SHA of the commit")
		case "commit-sha1":
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
			sub.Flags().StringVar(&cfg.sha, "sha", cfg.sha, "SHA of the commit")
		case "commits":
			sub.Flags().StringVar(&cfg.author, "author", cfg.author, "GitHub username or email address")
			sub.Flags().StringVar(&cfg.path, "path", cfg.path, "Path to file")
			sub.Flags().StringVar(&cfg.sha, "sha", cfg.sha, "SHA of the commit")
		case "compare-commits":
			sub.Flags().StringVar(&cfg.base, "base", cfg.base, "Base branch to compare")
			sub.Flags().StringVar(&cfg.head, "head", cfg.head, "Head branch to compare")
		case "contents":
			sub.Flags().StringVar(&cfg.path, "path", cfg.path, "Path to file")
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
		case "download-contents", "download-contents-with-meta":
			sub.Flags().StringVar(&cfg.path, "path", cfg.path, "Path to file")
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
		case "environment":
			sub.Flags().StringVar(&cfg.name, "name", cfg.name, "Name of the environment")
		case "generate-release-notes":
			sub.Flags().StringVar(&cfg.tag, "tag", cfg.tag, "Tag name for the release. This can be an existing tag or a new one.")
		case "list":
			sub.Flags().StringVar(&cfg.visibility, "visibility", cfg.visibility, "Visibility (all, public, private)")
			sub.Flags().StringVar(&cfg.affiliation, "affiliation", cfg.affiliation, "Comma-separated list of (owner, collaborator, organization_member)")
			sub.Flags().StringVar(&cfg.typ, "typ", cfg.typ, "Tag name for the release. This can be an existing tag or a new one.")
			sub.Flags().StringVar(cfg.sort, "sort", *cfg.sort, "Tag name for the release. This can be an existing tag or a new one.")
			sub.Flags().StringVar(cfg.direction, "direction", *cfg.direction, "Tag name for the release. This can be an existing tag or a new one.")
		case "permission-level":
			sub.Flags().StringVar(&cfg.user, "username", cfg.user, "Handle for the GitHub user account.")
		case "pre-receive-hook":
			addItemFlags(cfg, sub)
		case "pull-request-review-enforcement":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "readme":
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
		case "release":
			addItemFlags(cfg, sub)
		case "release-by-tag":
			sub.Flags().StringVar(&cfg.tag, "tag", cfg.tag, "Tag name for the release. This can be an existing tag or a new one.")
		case "release-asset":
			addItemFlags(cfg, sub)
		case "release-assets":
			addItemFlags(cfg, sub)
		case "required-status-checks":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "rules-for-branch":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "statuses":
			sub.Flags().StringVar(&cfg.ref, "ref", cfg.ref, "Commit reference")
		case "team-restrictions":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		case "user-restrictions":
			sub.Flags().StringVar(cfg.branch, "branch", *cfg.branch, "Branch name")
		}
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execRepos(cfg *ghCfg, cmd *cobra.Command) (x any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.Repositories

	if cmd.Flag("branch") != nil {
		branchOrDefault(cfg, cfg.client.Repositories)
	}

	switch cmd.Name() {
	case "actions-access-level":
		x, _, err = svc.GetActionsAccessLevel(getCtx(cfg), cfg.owner, cfg.repo)
	case "actions-allowed":
		x, _, err = svc.GetActionsAllowed(getCtx(cfg), cfg.owner, cfg.repo)
	case "actions-permissions":
		x, _, err = svc.GetActionsPermissions(getCtx(cfg), cfg.owner, cfg.repo)
	case "admin-enforcement":
		x, _, err = svc.GetAdminEnforcement(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "all":
		x, _, err = svc.ListAll(getCtx(cfg), &github.RepositoryListAllOptions{Since: cfg.sinceID})
	case "all-topics":
		x, _, err = svc.ListAllTopics(getCtx(cfg), cfg.owner, cfg.repo)
	case "archive-link":
		x, _, err = svc.GetArchiveLink(getCtx(cfg), cfg.owner, cfg.repo, github.Tarball,
			&github.RepositoryContentGetOptions{Ref: *cfg.branch}, int(cfg.id))
	case "automated-security-fixes":
		x, _, err = svc.GetAutomatedSecurityFixes(getCtx(cfg), cfg.owner, cfg.repo)
	case "branch":
		x, _, err = svc.GetBranch(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch, 3)
	case "branch-protection":
		x, _, err = svc.GetBranchProtection(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "branches-head-commit":
		x, _, err = svc.ListBranchesHeadCommit(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "branches":
		x, _, err = svc.ListBranches(getCtx(cfg), cfg.owner, cfg.repo, &github.BranchListOptions{
			Protected: cfg.protected, ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "collaborators":
		x, _, err = svc.ListCollaborators(getCtx(cfg), cfg.owner, cfg.repo, &github.ListCollaboratorsOptions{
			Affiliation: "",
			Permission:  "",
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "combined-status":
		x, _, err = svc.GetCombinedStatus(getCtx(cfg), cfg.owner, cfg.repo, cfg.ref,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "comment":
		x, _, err = svc.GetComment(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "comments":
		x, _, err = svc.ListComments(getCtx(cfg), cfg.owner, cfg.repo,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "commit-comments":
		x, _, err = svc.ListCommitComments(getCtx(cfg), cfg.owner, cfg.repo, cfg.sha,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "commit-raw":
		x, _, err = svc.GetCommitRaw(getCtx(cfg), cfg.owner, cfg.repo, cfg.sha, github.RawOptions{Type: github.Diff})
	case "commit-sha1":
		x, _, err = svc.GetCommitSHA1(getCtx(cfg), cfg.owner, cfg.repo, cfg.ref, cfg.sha)
	case "commits":
		x, _, err = svc.ListCommits(getCtx(cfg), cfg.owner, cfg.repo, &github.CommitsListOptions{
			SHA:         cfg.sha,
			Path:        cfg.path,
			Author:      cfg.author,
			Since:       cfg.sinceTime,
			Until:       time.Time{},
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "compare-commits":
		x, _, err = svc.CompareCommits(getCtx(cfg), cfg.owner, cfg.repo, cfg.base, cfg.head,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "compare-commits-raw":
		x, _, err = svc.CompareCommitsRaw(getCtx(cfg), cfg.owner, cfg.repo, cfg.base, cfg.head, github.RawOptions{Type: github.Diff})
	case "contents":
		var dcs []*github.RepositoryContent
		x, dcs, _, err = svc.GetContents(getCtx(cfg), cfg.owner, cfg.repo, cfg.path, &github.RepositoryContentGetOptions{Ref: cfg.ref})
		if dcs != nil {
			x = dcs
		}
	case "contributors":
		x, _, err = svc.ListContributors(getCtx(cfg), cfg.owner, cfg.repo, &github.ListContributorsOptions{
			Anon:        "",
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "contributors-stats":
		x, _, err = svc.ListContributorsStats(getCtx(cfg), cfg.owner, cfg.repo)
	case "download-contents":
		x, _, err = svc.DownloadContents(getCtx(cfg), cfg.owner, cfg.repo, cfg.path, &github.RepositoryContentGetOptions{Ref: cfg.ref})
	case "download-contents-with-meta":
		_, x, _, err = svc.DownloadContentsWithMeta(getCtx(cfg), cfg.owner, cfg.repo, cfg.path, &github.RepositoryContentGetOptions{Ref: cfg.ref})
	case "environment":
		x, _, err = svc.GetEnvironment(getCtx(cfg), cfg.owner, cfg.repo, cfg.name)
	case "environments":
		x, _, err = svc.ListEnvironments(getCtx(cfg), cfg.owner, cfg.repo, &github.EnvironmentListOptions{
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "generate-release-notes":
		x, _, err = svc.GenerateReleaseNotes(getCtx(cfg), cfg.owner, cfg.repo, &github.GenerateNotesOptions{TagName: cfg.tag})
	case "get":
		x, _, err = svc.Get(getCtx(cfg), cfg.owner, cfg.repo)
	case "languages":
		x, _, err = svc.ListLanguages(getCtx(cfg), cfg.owner, cfg.repo)
	case "latest-release":
		x, _, err = svc.GetLatestRelease(getCtx(cfg), cfg.owner, cfg.repo)
	case "license":
		x, _, err = svc.License(getCtx(cfg), cfg.owner, cfg.repo)
	case "list":
		x, _, err = svc.List(getCtx(cfg), cfg.owner, &github.RepositoryListOptions{
			Visibility:  "",
			Affiliation: "",
			Type:        "",
			Sort:        "",
			Direction:   "",
			ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "participation":
		x, _, err = svc.ListParticipation(getCtx(cfg), cfg.owner, cfg.repo)
	case "permission-level":
		x, _, err = svc.GetPermissionLevel(getCtx(cfg), cfg.owner, cfg.repo, cfg.user)
	case "pre-receive-hook":
		x, _, err = svc.GetPreReceiveHook(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "pre-receive-hooks":
		x, _, err = svc.ListPreReceiveHooks(getCtx(cfg), cfg.owner, cfg.repo,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "pull-request-review-enforcement":
		x, _, err = svc.GetPullRequestReviewEnforcement(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "readme":
		x, _, err = svc.GetReadme(getCtx(cfg), cfg.owner, cfg.repo, &github.RepositoryContentGetOptions{Ref: cfg.ref})
	case "release":
		x, _, err = svc.GetRelease(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "release-by-tag":
		x, _, err = svc.GetReleaseByTag(getCtx(cfg), cfg.owner, cfg.repo, cfg.tag)
	case "release-asset":
		x, _, err = svc.GetReleaseAsset(getCtx(cfg), cfg.owner, cfg.repo, cfg.id)
	case "release-assets":
		x, _, err = svc.ListReleaseAssets(getCtx(cfg), cfg.owner, cfg.repo, cfg.id,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		)
	case "releases":
		x, _, err = svc.ListReleases(getCtx(cfg), cfg.owner, cfg.repo,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "required-status-checks":
		x, _, err = svc.GetRequiredStatusChecks(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "rules-for-branch":
		x, _, err = svc.GetRulesForBranch(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "statuses":
		x, _, err = svc.ListStatuses(getCtx(cfg), cfg.owner, cfg.repo, cfg.ref,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "tag-protection":
		x, _, err = svc.ListTagProtection(getCtx(cfg), cfg.owner, cfg.repo)
	case "tag-protections":
		x, _, err = svc.ListTagProtection(getCtx(cfg), cfg.owner, cfg.repo)
	case "tags":
		x, _, err = svc.ListTags(getCtx(cfg), cfg.owner, cfg.repo, &github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "team-restrictions":
		x, _, err = svc.ListTeamRestrictions(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "teams":
		x, _, err = svc.ListTeams(getCtx(cfg), cfg.owner, cfg.repo, &github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "topics":
		x, _, err = svc.ListAllTopics(getCtx(cfg), cfg.owner, cfg.repo)
	case "user-restrictions":
		x, _, err = svc.ListUserRestrictions(getCtx(cfg), cfg.owner, cfg.repo, *cfg.branch)
	case "vulnerability-alerts":
		x, _, err = svc.GetVulnerabilityAlerts(getCtx(cfg), cfg.owner, cfg.repo)
	default:
		panic("Unsupported operation: " + cmd.Name())
	}
	return x, err
}
