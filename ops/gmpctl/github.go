// Copyright 2026 Google LLC
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

package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type prInfo struct {
	URL string `json:"url"`
}

func (p Project) mustEnsurePullRequest(dir, baseBranch, headBranch, title, body string) {
	err := ensurePullRequestGhCommand(dir, baseBranch, headBranch, title, body)
	if err == nil {
		return
	}
	logf("failed to use `gh` to create a pull request: %v", err)
	// https://docs.github.com/en/pull-requests/reference/using-query-parameters-to-create-a-pull-request
	query := url.Values{
		"title":      {title},
		"body":       {body},
		"quick_pull": {"1"},
	}
	u := fmt.Sprintf(
		"https://github.com/%s/%s/compare/%s...%s?%s",
		p.Organization(), p.RepoName(),
		baseBranch,
		headBranch,
		query.Encode(),
	)
	logf("Create a pull request manually at %s", u)
}

func ensurePullRequestGhCommand(dir, baseBranch, headBranch, title, body string) error {
	logf("Checking for existing pull request for %v...", headBranch)

	// gh pr list --head <head> --base <base> --state open --json url
	out, err := runCommand(
		&cmdOpts{Dir: dir, HideOutputs: true},
		"gh", "pr", "list",
		"--head", headBranch,
		"--base", baseBranch,
		"--state", "open",
		"--json", "url",
	)
	if err != nil {
		return fmt.Errorf("failed to check existing PR: %w", err)
	}

	var prs []prInfo
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return fmt.Errorf("failed to parse gh output: %w, output: %q", err, out)
	}

	if len(prs) > 0 {
		logf("Pull request already exists: %v", prs[0].URL)
		return nil
	}

	logf("Creating pull request from %v to %v...", headBranch, baseBranch)

	// gh pr create --title <title> --body <body> --base <base> --head <head>
	prURL, err := runCommand(
		&cmdOpts{Dir: dir, HideOutputs: true},
		"gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", baseBranch,
		"--head", headBranch,
	)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	logf("Pull request created: %v", prURL)
	return nil
}
