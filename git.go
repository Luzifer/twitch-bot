package main

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "opening git repo")
	}

	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "getting worktree")
	}

	if _, err = wt.Add(filename); err != nil {
		return errors.Wrap(err, "adding file to index")
	}

	if authorEmail == "" {
		authorEmail = authorCloak
	}

	_, err = wt.Commit(summary, &git.CommitOptions{
		Author:    g.getSignature(authorName, authorEmail),
		Committer: g.getSignature(fmt.Sprintf(committerName, version), committerEmail),
	})
	return errors.Wrap(err, "issuing commit")
}

func (g gitHelper) HasRepo() bool {
	_, err := git.PlainOpen(g.repoDir)
	return err == nil
}

func (gitHelper) getSignature(name, mail string) *object.Signature {
	return &object.Signature{Name: name, Email: mail, When: time.Now()}
}
