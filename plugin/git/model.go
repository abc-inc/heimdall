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

//go:build !no_git

package git

import (
	"encoding/json"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Hash is a marshalling-friendly plumbing.Hash.
type Hash plumbing.Hash

func (h Hash) String() string {
	return plumbing.Hash(h).String()
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

func (h Hash) MarshalYAML() (any, error) {
	return h.String(), nil
}

// Commit is a marshalling-friendly object.Commit.
type Commit struct {
	// Hash of the commit object.
	Hash Hash `json:"hash" yaml:"hash"`
	// Author is the original author of the commit.
	Author Signature `json:"author" yaml:"author"`
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer Signature `json:"committer" yaml:"committer"`
	// PGPSignature is the PGP signature of the commit.
	PGPSignature string `json:"pgpSignature,omitempty" yaml:"pgp_signature,omitempty"`
	// Message is the commit message, contains arbitrary text.
	Message string `json:"message,omitempty" yaml:"message"`
	// TreeHash is the hash of the root tree of the commit.
	TreeHash Hash `json:"treeHash" yaml:"tree_hash"`
	// ParentHashes are the hashes of the parent commits of the commit.
	ParentHashes []Hash `json:"parentHashes,omitempty" yaml:"parent_hashes,omitempty"`
}

func NewCommit(c *object.Commit) *Commit {
	var pHs []Hash
	for _, p := range c.ParentHashes {
		pHs = append(pHs, Hash(p))
	}
	return &Commit{
		Hash:         Hash(c.Hash),
		Author:       Signature(c.Author),
		Committer:    Signature(c.Committer),
		PGPSignature: c.PGPSignature,
		Message:      c.Message,
		TreeHash:     Hash(c.TreeHash),
		ParentHashes: pHs,
	}
}

type Signature struct {
	// Name represents a person name. It is an arbitrary string.
	Name string `json:"name" yaml:"name"`
	// Email is an email, but it cannot be assumed to be well-formed.
	Email string `json:"email" yaml:"email"`
	// When is the timestamp of the signature.
	When time.Time `json:"when" yaml:"when"`
}
