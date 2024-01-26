// Copyright 2024 The Heimdall authors
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

//go:generate go run github.com/abc-inc/heimdall/cmd/cmddoc github.com/google/go-github/v56@v56.0.0/github/teams\*.go ../../docs

package github

import (
	"context"
	"reflect"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"
)

func NewTeamsCmd() *cobra.Command {
	var direction, displayName string
	cfg := newGHCfg()
	cfg.direction = &direction
	cfg.displayName = &displayName

	cmd := &cobra.Command{
		Use:   "teams",
		Short: "Handles communication with the teams related methods.",
		Example: heredoc.Doc(`
			heimdall github teams user-teams
		`),
		Args: cobra.ExactArgs(0),
	}

	ops := []string{
		"comment-by-slug",
		"child-teams-by-parent-id",
		"child-teams-by-parent-slug",
		"comments-by-slug",
		"discussion-by-slug",
		"discussions-by-slug",
		"external-group",
		"external-groups",
		"external-groups-for-team-by-slug",
		"idp-groups-for-team-by-slug",
		"idp-groups-in-organization",
		"team-by-slug",
		"team-members-by-slug",
		"team-membership-by-slug",
		"team-repos-by-slug",
		"teams",
		"user-teams",
	}

	svcTyp := reflect.TypeOf(&github.TeamsService{})
	cmd.AddCommand(createCmds(cfg, svcTyp, execTeams, ops)...)
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "child-teams-by-parent-id":
			sub.Flags().Int64Var(&cfg.orgID, "org-id", cfg.orgID, "organization id")
			sub.Flags().Int64Var(&cfg.teamID, "team-id", cfg.teamID, "team id")
		case "child-teams-by-parent-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
		case "comment-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().IntVar(&cfg.discussionNumber, "discussion-number", cfg.discussionNumber, "discussion number")
			sub.Flags().IntVar(&cfg.commentID, "comment-id", cfg.commentID, "comment id")
		case "comments-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().IntVar(&cfg.discussionNumber, "discussion-number", cfg.discussionNumber, "discussion number")
			sub.Flags().StringVar(cfg.direction, "direction", defVal(cfg.direction), "comment sort direction")
		case "discussion-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().IntVar(&cfg.discussionNumber, "discussion-number", cfg.discussionNumber, "discussion number")
		case "discussions-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().StringVar(cfg.direction, "direction", defVal(cfg.direction), "discussion sort direction")
		case "external-group":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().Int64Var(&cfg.groupID, "group-id", cfg.groupID, "group id")
		case "external-groups":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(cfg.displayName, "display-name", defVal(cfg.displayName), "display name")
		case "external-groups-for-team-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
		case "idp-groups-for-team-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
		case "idp-groups-in-organization":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(cfg.displayName, "display-name", defVal(cfg.displayName), "display name")
		case "team-members-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().StringVar(&cfg.role, "role", cfg.role, "team role")
		case "team-membership-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			sub.Flags().StringVar(&cfg.user, "user", cfg.user, "user login")
		case "team-repos-by-slug":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
			sub.Flags().StringVar(&cfg.slug, "slug", cfg.slug, "team slug")
			internal.MustNoErr(sub.MarkFlagRequired("slug"))
		case "teams":
			sub.Flags().StringVar(&cfg.owner, "owner", cfg.owner, "owner")
		}

		if f := sub.Flag("owner"); f != nil {
			internal.MustNoErr(sub.MarkFlagRequired(f.Name))
		}
	}

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	return cmd
}

func execTeams(cfg *ghCfg, cmd *cobra.Command) (x any, err error) {
	setHostOwnerRepo(cfg, cfg.host, cfg.owner, cfg.repo)
	cfg.client = newClient()
	svc := cfg.client.Teams

	switch cmd.Name() {
	case "child-teams-by-parent-id":
		x, _, err = svc.ListChildTeamsByParentID(context.Background(), cfg.orgID, cfg.teamID,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "child-teams-by-parent-slug":
		x, _, err = svc.ListChildTeamsByParentSlug(context.Background(), cfg.owner, cfg.slug,
			&github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "comment-by-slug":
		x, _, err = svc.GetCommentBySlug(context.Background(), cfg.owner, cfg.slug, cfg.discussionNumber, cfg.commentID)
	case "comments-by-slug":
		x, _, err = svc.ListCommentsBySlug(context.Background(), cfg.owner, cfg.slug, cfg.discussionNumber,
			&github.DiscussionCommentListOptions{Direction: defVal(cfg.direction), ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage}})
	case "discussion-by-slug":
		x, _, err = svc.GetDiscussionBySlug(context.Background(), cfg.owner, cfg.slug, cfg.discussionNumber)
	case "discussions-by-slug":
		x, _, err = svc.ListDiscussionsBySlug(context.Background(), cfg.owner, cfg.slug, &github.DiscussionListOptions{
			Direction: defVal(cfg.direction), ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage},
		})
	case "external-group":
		x, _, err = svc.GetExternalGroup(context.Background(), cfg.owner, cfg.groupID)
	case "external-groups":
		x, _, err = svc.ListExternalGroups(context.Background(), cfg.owner, &github.ListExternalGroupsOptions{
			DisplayName: cfg.displayName, ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage}})
	case "external-groups-for-team-by-slug":
		x, _, err = svc.ListExternalGroupsForTeamBySlug(context.Background(), cfg.owner, cfg.slug)
	case "idp-groups-for-team-by-slug":
		x, _, err = svc.ListIDPGroupsForTeamBySlug(context.Background(), cfg.owner, cfg.slug)
	case "idp-groups-in-organization":
		x, _, err = svc.ListIDPGroupsInOrganization(context.Background(), cfg.owner,
			&github.ListCursorOptions{Page: defVal(cfg.displayName), PerPage: cfg.perPage})
	case "team-members-by-slug":
		x, _, err = svc.ListTeamMembersBySlug(context.Background(), cfg.owner, cfg.slug, &github.TeamListTeamMembersOptions{
			Role: cfg.role, ListOptions: github.ListOptions{Page: cfg.page, PerPage: cfg.perPage}})
	case "team-membership-by-slug":
		x, _, err = svc.GetTeamMembershipBySlug(context.Background(), cfg.owner, cfg.slug, cfg.user)
	case "team-repos-by-slug":
		x, _, err = svc.ListTeamReposBySlug(context.Background(), cfg.owner, cfg.slug, &github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "teams":
		x, _, err = svc.ListTeams(context.Background(), cfg.owner, &github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	case "user-teams":
		x, _, err = svc.ListUserTeams(context.Background(), &github.ListOptions{Page: cfg.page, PerPage: cfg.perPage})
	}
	return x, err
}
