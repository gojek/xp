package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type data struct {
	Devs  map[string]*dev  `json:"devs"`
	Repos map[string]*repo `json:"repos"`
}

func load(r io.Reader) (*data, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}

	var d data
	if err := yaml.Unmarshal(bytes, &d); err != nil {
		return nil, errors.Wrap(err, "unmarshall failed")
	}

	return &d, nil
}

func (d *data) String() string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic(err)
	}

	return string(b)
}

func (d *data) store(w io.Writer) error {
	if _, err := io.Copy(w, strings.NewReader(d.String())); err != nil {
		return errors.Wrap(err, "store failed")
	}

	return nil
}

type dev struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (d *dev) String() string {
	return d.Name + " <" + d.Email + ">"
}

func (d *data) addDev(id, name, email string) {
	if d.Devs == nil {
		d.Devs = make(map[string]*dev)
	}
	d.Devs[id] = &dev{Name: name, Email: email}
}

func (d *data) lookupDev(id string) *dev {
	if d.Devs == nil {
		return nil
	}
	return d.Devs[id]
}

type repo struct {
	Devs    []string `json:"devs"`
	StoryID string   `json:"storyId"`
}

func (d *data) validateDevs(devIDs []string) error {
	for _, did := range devIDs {
		if d.lookupDev(did) == nil {
			return errors.Errorf("no dev with id %s found", did)
		}
	}
	return nil
}

func (d *data) addRepo(path string, devIDs []string, storyID string) error {
	if d.Repos == nil {
		d.Repos = make(map[string]*repo)
	}

	if err := d.validateDevs(devIDs); err != nil {
		return errors.Wrap(err, "dev ids validation failed")
	}

	d.Repos[path] = &repo{
		Devs:    devIDs,
		StoryID: storyID,
	}

	return nil
}

func initRepo(pathStr string, overwrite bool) error {
	gitPath := path.Join(pathStr, ".git")

	if _, err := os.Stat(gitPath); err != nil {
		return errors.Wrapf(err, ".git not found in %s", pathStr)
	}

	hookFile := path.Join(gitPath, "hooks/prepare-commit-msg")

	if !overwrite {
		if _, err := os.Stat(hookFile); err == nil {
			// TODO: Check if it is our prepare-commit-msg hook.
			return errors.Errorf("prepare-commit-msg hook (%s) is already defined", hookFile)
		}
	}

	f, err := os.Create(hookFile)
	if err != nil {
		return errors.Wrapf(err, "create hook file %s failed", hookFile)
	}
	defer f.Close()

	f.WriteString("#!/bin/sh\n")
	f.WriteString("/Users/kidoman/dev/personal/xp/_bin/xp add-info $1\n")

	return nil
}

func (d *data) lookupRepo(pathStr string) (string, *repo) {
	if d.Repos == nil {
		return "", nil
	}

	r := d.Repos[pathStr]
	if r != nil {
		return pathStr, r
	}

	for k, v := range d.Repos {
		matched, err := path.Match(k+"/*", pathStr)
		if err != nil {
			log.Printf("match failed for %s", pathStr)
			continue
		}
		if matched {
			return k, v
		}
	}

	return "", nil
}

func (d *data) updateRepoDevs(wd string, devIDs []string) error {
	_, repo := d.lookupRepo(wd)
	if repo == nil {
		return errors.Errorf("no repo with path %s found", wd)
	}

	if err := d.validateDevs(devIDs); err != nil {
		return errors.Wrap(err, "dev ids validation failed")
	}

	repo.Devs = devIDs

	return nil
}

func (d *data) appendInfo(wd, msgFile string) error {
	repoPath, repo := d.lookupRepo(wd)
	if repo == nil {
		return errors.Errorf("no repo with path %s found", wd)
	}

	// GIT_COMMITTER_IDENT can be used to get committer info.
	author, err := gitVar("GIT_AUTHOR_IDENT")
	if err != nil {
		return errors.Wrap(err, "get author info failed")
	}
	authorName, authorEmail := nameEmail(author)

	f, err := os.OpenFile(msgFile, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return errors.Wrapf(err, "open on commit msg file %s failed", msgFile)
	}
	defer f.Close()

	// TODO: Add story info.

	fmt.Fprintf(f, "\n\n")
	for _, devID := range repo.Devs {
		dev := d.lookupDev(devID)
		if dev == nil {
			return errors.Errorf("non-existing dev %s marked as working for repo %s", devID, repoPath)
		}

		if dev.Email == authorEmail || dev.Name == authorName {
			log.Printf("skipping %s (same as author)", dev)
			continue
		}

		fmt.Fprintf(f, "Co-authored-by: %s <%s>\n", dev.Name, dev.Email)
	}

	return nil
}

func gitVar(varStr string) (string, error) {
	output, err := exec.Command("git", "var", varStr).Output()
	if err != nil {
		return "", errors.Wrap(err, "git exec failed")
	}
	return string(output), nil
}

func nameEmail(ident string) (string, string) {
	idx := strings.Index(ident, "<")
	name := ident[:idx-1]
	email := ident[idx+1 : strings.Index(ident, ">")]
	return name, email
}
