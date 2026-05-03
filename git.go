package main

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	authorCloak    = `twitch-bot+author@luzifer.io`
	committerName  = `Twitch-Bot %s`
	committerEmail = `twitch-bot+committer@luzifer.io`
)

type gitHelper struct {
	repoDir string
}

func newGitHelper(repoDir string) *gitHelper {
	return &gitHelper{
		repoDir: repoDir,
	}
}

func (g gitHelper) CommitChange(filename, authorName, authorEmail, summary string) error {
	repo, err := git.PlainOpen(g.repoDir)
	if err != nil {
		return fmt.Errorf("opening git repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree: %w", err)
	}

	if _, err = wt.Add(filename); err != nil {
		return fmt.Errorf("adding file to index: %w", err)
	}

	if authorEmail == "" {
		authorEmail = authorCloak
	}

	if _, err = wt.Commit(summary, &git.CommitOptions{
		Author:    g.getSignature(authorName, authorEmail),
		Committer: g.getSignature(fmt.Sprintf(committerName, version), committerEmail),
	}); err != nil {
		return fmt.Errorf("issuing commit: %w", err)
	}

	return nil
}

func (g gitHelper) HasRepo() bool {
	_, err := git.PlainOpen(g.repoDir)
	return err == nil
}

func (gitHelper) getSignature(name, mail string) *object.Signature {
	return &object.Signature{Name: name, Email: mail, When: time.Now()}
}
