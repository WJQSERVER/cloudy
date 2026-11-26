package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// The updates released, the issues solved and the pull requests merged, as
// the support page's head shows them: asked of GitHub's search once a run, the first time the
// page shows, and where GitHub cannot be reached (its API has no mirror),
// as supporters.json last wrote them down.

// supportCounts are the numbers GitHub gave this run, 0 where it gave none.
type supportCounts struct {
	asked                bool
	updates, issues, prs int
}

// supportIssueQuery and supportPRQuery are what the numbers count: issues
// closed as done, and pull requests merged, by anyone.
const (
	supportIssueQuery = "repo:" + releaseRepo + " is:issue is:closed reason:completed"
	supportPRQuery    = "repo:" + releaseRepo + " is:pr is:merged"
)

// supportIssuesPage and supportPRsPage are where a click on a number goes.
const (
	supportIssuesPage = "https://github.com/" + releaseRepo + "/issues?q=is%3Aissue+is%3Aclosed+reason%3Acompleted"
	supportPRsPage    = "https://github.com/" + releaseRepo + "/pulls?q=is%3Apr+is%3Amerged"
)

// supportWorkCounts are the issues and pull requests to show: GitHub's
// numbers once read, else the file's, and it asks for GitHub's the first
// time.
func (a *app) supportWorkCounts(d *supportData) (issues, prs int) {
	st := &a.support
	if !st.counts.asked && a.win != nil {
		st.counts.asked = true
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			// Each number on its own: one GitHub won't give leaves the
			// others as it gave them.
			updates, err := fetchReleaseCount(ctx)
			if err != nil {
				log.Printf("support: releases: %v", err)
			}
			issues, err := fetchSupportCount(ctx, supportIssueQuery)
			if err != nil {
				log.Printf("support: issues: %v", err)
			}
			prs, err := fetchSupportCount(ctx, supportPRQuery)
			if err != nil {
				log.Printf("support: pull requests: %v", err)
			}
			a.post(func() {
				a.support.counts.updates, a.support.counts.issues, a.support.counts.prs = updates, issues, prs
				a.redraw()
			})
		}()
	}
	issues, prs = d.Issues, d.PRs
	if st.counts.issues > 0 {
		issues = st.counts.issues
	}
	if st.counts.prs > 0 {
		prs = st.counts.prs
	}
	return issues, prs
}

// supportUpdates is how many updates were released: GitHub's number once
// read, else the file's. supportWorkCounts asks GitHub for it, with the
// others.
func (a *app) supportUpdates(d *supportData) int {
	if n := a.support.counts.updates; n > 0 {
		return n
	}
	return d.Updates
}

// fetchReleaseCount is how many releases GitHub lists, not counting drafts
// or prereleases (the captions libraries are one); tests swap it.
var fetchReleaseCount = func(ctx context.Context) (int, error) {
	n := 0
	for page := 1; page <= 10; page++ {
		u := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100&page=%d", releaseRepo, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return 0, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		var out []struct {
			Draft      bool `json:"draft"`
			Prerelease bool `json:"prerelease"`
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return 0, fmt.Errorf("%s", res.Status)
		}
		err = json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&out)
		res.Body.Close()
		if err != nil {
			return 0, err
		}
		for _, r := range out {
			if !r.Draft && !r.Prerelease {
				n++
			}
		}
		if len(out) < 100 {
			break
		}
	}
	return n, nil
}

// fetchSupportCount is how many GitHub's search finds for query; tests
// swap it.
var fetchSupportCount = func(ctx context.Context, query string) (int, error) {
	u := "https://api.github.com/search/issues?per_page=1&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%s", res.Status)
	}
	var out struct {
		Total int `json:"total_count"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&out); err != nil {
		return 0, err
	}
	return out.Total, nil
}
