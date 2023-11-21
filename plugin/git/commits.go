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

package git

import (
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/abc-inc/heimdall/console"
	"github.com/abc-inc/heimdall/internal"
	"github.com/abc-inc/heimdall/res"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type gitCfg struct {
	console.OutCfg
	repo      string
	user      string
	pass      string
	startRef  string
	endRef    string
	mergeBase bool
}

func NewCommitsCmd() *cobra.Command {
	cfg := gitCfg{endRef: "HEAD"}
	cmd := &cobra.Command{
		Use:   "commits [flags] [<repository>]",
		Short: "List the commits between two arbitrary commits",
		Example: heredoc.Doc(`
			heimdall git commits --merge-base --start-ref release/3.14
		`),
		Args: cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 && res.IsURL(args[0]) {
				if cfg.user = os.Getenv("GIT_USERNAME"); cfg.user == "" {
					log.Fatal().Str("name", "GIT_USERNAME").Msg("undefined environment variable")
				}
				if cfg.pass = os.Getenv("GIT_PASSWORD"); cfg.pass == "" {
					log.Fatal().Str("name", "GIT_PASSWORD").Msg("undefined environment variable")
				}
			}

			if len(args) == 1 {
				cfg.repo = args[0]
			} else {
				cfg.repo = internal.Must(os.Getwd())
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			console.Fmtln(listCommits(cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.startRef, "start-ref", cfg.endRef, "Include revisions after this ref")
	cmd.Flags().StringVar(&cfg.endRef, "end-ref", cfg.endRef, "Last revision to include")
	cmd.Flags().BoolVar(&cfg.mergeBase, "merge-base", cfg.mergeBase, `Use the merge base of the two commits for the "start" side`)

	console.AddOutputFlags(cmd, &cfg.OutCfg)
	internal.MustNoErr(cmd.MarkFlagRequired("start-ref"))
	cmd.DisableFlagsInUseLine = true
	return cmd
}

func listCommits(cfg gitCfg) []*Commit {
	r := internal.Must(open(cfg.repo, cfg.user, cfg.pass))
	if cfg.mergeBase {
		cfg.startRef = mergeBase(r, cfg.startRef, cfg.endRef).Hash.String()
	}
	return loadCommits(r, resolve(r, cfg.startRef), resolve(r, cfg.endRef))
}

func open(uri, user, pass string) (*git.Repository, error) {
	if strings.HasPrefix(uri, "https://") {
		log.Info().Str("URL", uri).Msg("Cloning Git repository")
		return git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:  uri,
			Auth: &http.BasicAuth{Username: user, Password: pass},
		})
	}

	log.Info().Str("dir", uri).Msg("Opening Git repository")
	return git.PlainOpen(uri)
}

func mergeBase(r *git.Repository, revsOrHashes ...string) *object.Commit {
	var hs []plumbing.Hash
	for _, rh := range revsOrHashes {
		hs = append(hs, resolve(r, rh))
	}

	var cs []*object.Commit
	for _, h := range hs {
		cs = append(cs, internal.Must(r.CommitObject(h)))
	}
	internal.MustOkMsgf[any](nil, len(cs) == 2, "expected 2 commits, got %d", len(cs))

	base := internal.Must(cs[0].MergeBase(cs[1]))
	internal.MustOkMsgf[any](nil, len(base) == 1, "expected 1 common ancestor, got %d", len(base))
	log.Debug().Stringer("old", cs[0].Hash).Stringer("new", cs[1].Hash).
		Stringer("hash", base[0].Hash).Msg("Resolved Git merge base")
	return base[0]
}

func resolve(r *git.Repository, revOrHash string) (h plumbing.Hash) {
	rev, err := r.ResolveRevision(plumbing.Revision(revOrHash))
	if err != nil {
		log.Debug().Str("hash", revOrHash).Msg("Failed to lookup revision - assuming hash prefix")
		iter := internal.Must(r.CommitObjects())
		defer iter.Close()
		internal.MustNoErr(iter.ForEach(func(c *object.Commit) error {
			if strings.HasPrefix(c.Hash.String(), revOrHash) {
				h = c.Hash
				return storer.ErrStop
			}
			return nil
		}))
		if h.IsZero() {
			iter.Close()
			log.Fatal().Err(err).Str("rev", revOrHash).Msg("Failed to resolve hash")
		}
	} else {
		h = *internal.Must(rev, err)
	}
	log.Debug().Str("ref", revOrHash).Str("hash", h.String()).Msg("Resolved commit")
	return
}

func loadCommits(r *git.Repository, startHash, endHash plumbing.Hash) []*Commit {
	start := internal.Must(r.CommitObject(startHash))
	end := internal.Must(r.CommitObject(endHash))
	if ok, err := start.IsAncestor(end); err != nil || !ok {
		log.Fatal().Stringer("startHash", startHash).Stringer("endHash", endHash).Msg("Start commit is not an ancestor of end commit")
	}

	log.Debug().Stringer("start-ref", startHash).Stringer("end-ref", endHash).Msg("Loading commits")
	l := internal.Must(r.Log(&git.LogOptions{From: endHash, Order: git.LogOrderCommitterTime}))
	defer l.Close()

	cs := make([]*Commit, 0)
	internal.MustNoErr(l.ForEach(func(c *object.Commit) error {
		if c.Hash == startHash {
			return storer.ErrStop
		}
		cs = append(cs, NewCommit(c))
		return nil
	}))
	return cs
}
