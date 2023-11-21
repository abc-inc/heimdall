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
	"encoding/json"
	"fmt"

	"github.com/abc-inc/heimdall/internal"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Hash is a marshalling-friendly plumbing.Hash.
type Hash plumbing.Hash

func (h Hash) String() string {
	return plumbing.Hash(h).String()
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(plumbing.Hash(h).String())
}

func (h Hash) MarshalYAML() (any, error) {
	return plumbing.Hash(h).String(), nil
}

// Commit is a marshalling-friendly object.Commit.
type Commit struct {
	// Hash of the commit object.
	Hash Hash
	// Author is the original author of the commit.
	Author object.Signature
	// Committer is the one performing the commit, might be different from
	// Author.
	Committer object.Signature
	// PGPSignature is the PGP signature of the commit.
	PGPSignature string
	// Message is the commit message, contains arbitrary text.
	Message string
	// TreeHash is the hash of the root tree of the commit.
	TreeHash Hash
	// ParentHashes are the hashes of the parent commits of the commit.
	ParentHashes []Hash
}

func (c Commit) MarshalJSON() ([]byte, error) {
	j := fmt.Sprintf(
		`{"Hash": %s, "Author": %s, "Committer": %s, "PGPSignature": %s, "Message": %s, "TreeHash": %s, "ParentHashes": %s}`,
		internal.Must(json.Marshal(c.Hash.String())), internal.Must(json.Marshal(c.Author)),
		internal.Must(json.Marshal(c.Committer)), internal.Must(json.Marshal(c.PGPSignature)),
		internal.Must(json.Marshal(c.Message)), internal.Must(json.Marshal(c.TreeHash.String())),
		internal.Must(json.Marshal(c.ParentHashes)))
	return []byte(j), nil
}

func NewCommit(c *object.Commit) *Commit {
	var pHs []Hash
	for _, p := range c.ParentHashes {
		pHs = append(pHs, Hash(p))
	}
	return &Commit{
		Hash:         Hash(c.Hash),
		Author:       c.Author,
		Committer:    c.Committer,
		PGPSignature: c.PGPSignature,
		Message:      c.Message,
		TreeHash:     Hash(c.TreeHash),
		ParentHashes: pHs,
	}
}
