package pair

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	expectedData := Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
		Repos: map[string]*repo{
			"/path/to/repo": &repo{
				Devs: []string{"ak"},
			},
		},
	}

	r := strings.NewReader(`devs:
  ak:
    email: akshat@beef.com
    name: akshat
repos:
  /path/to/repo:
    devs:
    - ak
    issueId: ""
`)

	data, err := Load(r)
	assert.NoError(t, err)

	assert.Equal(t, &expectedData, data)
}

func TestDataString(t *testing.T) {
	data := Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
		Repos: map[string]*repo{
			"/path/to/repo": &repo{
				Devs: []string{"ak"},
			},
		},
	}

	expectedString := `devs:
  ak:
    email: akshat@beef.com
    name: akshat
repos:
  /path/to/repo:
    devs:
    - ak
    issueId: ""
`

	assert.Equal(t, expectedString, data.String())
}

func TestDataStore(t *testing.T) {
	data := Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
		Repos: map[string]*repo{
			"/path/to/repo": &repo{
				Devs: []string{"ak"},
			},
		},
	}

	var buf bytes.Buffer
	assert.NoError(t, data.Store(&buf))

	expectedString := `devs:
  ak:
    email: akshat@beef.com
    name: akshat
repos:
  /path/to/repo:
    devs:
    - ak
    issueId: ""
`

	assert.Equal(t, expectedString, buf.String())
}

func TestDevString(t *testing.T) {
	dev := dev{
		Name:  "akshat",
		Email: "akshat@beef.com",
	}

	assert.Equal(t, "akshat <akshat@beef.com>", dev.String())
}

func TestDataAddDev(t *testing.T) {
	var d Data

	d.AddDev("km", "Karan Misra", "karan@beef.com")

	assert.Equal(t, &dev{Name: "Karan Misra", Email: "karan@beef.com"}, d.Devs["km"])

	d = Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
	}

	d.AddDev("km", "Karan Misra", "karan@beef.com")

	assert.Equal(t, &dev{Name: "Karan Misra", Email: "karan@beef.com"}, d.Devs["km"])
	assert.Equal(t, &dev{Name: "akshat", Email: "akshat@beef.com"}, d.Devs["ak"])
}

func TestDataLookupDev(t *testing.T) {
	var d Data

	assert.Equal(t, (*dev)(nil), d.lookupDev("km"))

	d = Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
	}

	assert.Equal(t, (*dev)(nil), d.lookupDev("km"))
	assert.Equal(t, &dev{Name: "akshat", Email: "akshat@beef.com"}, d.lookupDev("ak"))
}

func TestDataValidateDevs(t *testing.T) {
	d := Data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
			"km": &dev{
				Name:  "Karan Misra",
				Email: "karan@beef.com",
			},
		},
	}

	tests := []struct {
		ids    []string
		errMsg string
	}{
		{
			ids:    []string{"ak"},
			errMsg: "",
		},
		{
			ids:    []string{"ak", "km"},
			errMsg: "",
		},
		{
			ids:    []string{"anand"},
			errMsg: "no dev with id anand found",
		},
	}

	for _, tt := range tests {
		err := d.validateDevs(tt.ids)

		if tt.errMsg == "" {
			assert.NoError(t, err)
			continue
		}

		if assert.Error(t, err) {
			assert.Equal(t, tt.errMsg, err.Error())
		}
	}
}

func TestDataAddRepo(t *testing.T) {
	tests := []struct {
		path    string
		devIDs  []string
		issueID string
		errMsg  string
	}{
		{
			path:    "/some/path",
			devIDs:  []string{"ak"},
			issueID: "o-1",
			errMsg:  "",
		},
		{
			devIDs: []string{"anand"},
			errMsg: "dev ids validation failed: no dev with id anand found",
		},
		{
			path:    "/some/path",
			devIDs:  []string{"ak", "km"},
			issueID: "o-1",
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		d := Data{
			Devs: map[string]*dev{
				"ak": &dev{
					Name:  "akshat",
					Email: "akshat@beef.com",
				},
				"km": &dev{
					Name:  "Karan Misra",
					Email: "karan@beef.com",
				},
			},
		}

		err := d.AddRepo(tt.path, tt.devIDs, tt.issueID)

		if tt.errMsg != "" {
			if assert.Error(t, err) {
				assert.Equal(t, tt.errMsg, err.Error())
			}

			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, &repo{Devs: tt.devIDs, IssueID: tt.issueID}, d.Repos[tt.path])
	}
}

func TestInitRepo(t *testing.T) {
	tests := []struct {
		desc      string
		prepareFn func(string) error
		overwrite bool
		errMsg    string
	}{

		{
			desc: "happy path",
			prepareFn: func(dir string) error {
				return os.MkdirAll(path.Join(dir, ".git", "hooks"), 0700|os.ModeDir)
			},
		},
		{
			desc: "not a repo",
			prepareFn: func(dir string) error {
				return nil
			},
			errMsg: ".git not found in %s: stat %[1]s/.git: no such file or directory",
		},
		{
			desc: "hooks folder does not exist",
			prepareFn: func(dir string) error {
				return os.MkdirAll(path.Join(dir, ".git"), 0700|os.ModeDir)
			},
			errMsg: "create hook file %s/.git/hooks/prepare-commit-msg failed: open %[1]s/.git/hooks/prepare-commit-msg: no such file or directory",
		},
		{
			desc: "hook already exists",
			prepareFn: func(dir string) error {
				hooksDir := path.Join(dir, ".git", "hooks")
				if err := os.MkdirAll(hooksDir, 0700|os.ModeDir); err != nil {
					return err
				}
				f, err := os.Create(path.Join(hooksDir, "prepare-commit-msg"))
				if err != nil {
					return err
				}
				return f.Close()
			},
			errMsg: "hooks/prepare-commit-msg is already defined",
		},
		{
			desc: "overwrite existing hook",
			prepareFn: func(dir string) error {
				hooksDir := path.Join(dir, ".git", "hooks")
				if err := os.MkdirAll(hooksDir, 0700|os.ModeDir); err != nil {
					return err
				}
				f, err := os.Create(path.Join(hooksDir, "prepare-commit-msg"))
				if err != nil {
					return err
				}
				return f.Close()
			},
			overwrite: true,
		},
	}

	for _, tt := range tests {
		t.Logf("case: %s", tt.desc)

		repoDir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(repoDir)

		if !assert.NoError(t, tt.prepareFn(repoDir)) {
			continue
		}

		err = InitRepo(repoDir, tt.overwrite, "/path/to/xp")

		if tt.errMsg != "" {
			if assert.Error(t, err) {
				var args []interface{}
				if strings.Contains(tt.errMsg, "%s") {
					args = []interface{}{repoDir}
				}
				assert.Equal(t, fmt.Sprintf(tt.errMsg, args...), err.Error())
			}
			continue
		}

		if !assert.NoError(t, err) {
			continue
		}

		for _, hookFile := range hookFiles {
			validateHook(t, path.Join(repoDir, ".git", hookFile))
		}
	}
}

func validateHook(t *testing.T, hook string) {
	data, err := ioutil.ReadFile(hook)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "#!/bin/sh\n/path/to/xp add-info $1\n", string(data))
}

func TestLookupRepo(t *testing.T) {
	repo1, repo2 := new(repo), new(repo)

	d := Data{
		Repos: map[string]*repo{
			"/a/b/c": repo1,
			"/x/y/z": repo2,
		},
	}

	tests := []struct {
		pathStr  string
		repoPath string
		repo     *repo
	}{
		{
			pathStr:  "/a/b/c/d",
			repoPath: "/a/b/c",
			repo:     repo1,
		},
		{
			pathStr:  "/x/y/z",
			repoPath: "/x/y/z",
			repo:     repo2,
		},
		{
			pathStr:  "/a/b",
			repoPath: "",
		},
		{
			pathStr:  "",
			repoPath: "",
		},
	}

	for _, tt := range tests {
		repoPath, repo := d.lookupRepo(tt.pathStr)

		if !assert.Equal(t, tt.repoPath, repoPath) {
			continue
		}

		assert.Equal(t, unsafe.Pointer(tt.repo), unsafe.Pointer(repo))
	}
}

func TestUpdateRepoDevs(t *testing.T) {
	tests := []struct {
		wd     string
		devIDs []string
		errMsg string
	}{
		{
			wd:     "/a",
			devIDs: []string{"anand"},
		},
		{
			wd:     "/b",
			errMsg: "no repo with path /b found",
		},
		{
			wd:     "/a",
			devIDs: []string{"non-existent-dev"},
			errMsg: "dev ids validation failed: no dev with id non-existent-dev found",
		},
	}

	for _, tt := range tests {
		r := new(repo)
		d := Data{
			Devs: map[string]*dev{
				"anand": new(dev),
			},
			Repos: map[string]*repo{
				"/a": r,
			},
		}

		err := d.UpdateRepoDevs(tt.wd, tt.devIDs)

		if tt.errMsg != "" {
			if !assert.Error(t, err) {
				continue
			}

			assert.Equal(t, tt.errMsg, err.Error())
			continue
		}

		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, r.Devs, tt.devIDs)
	}
}

func TestAppendInfo(t *testing.T) {
	d := Data{
		Devs: map[string]*dev{
			"karan": &dev{
				Name: "Karan Misra", Email: "karan@beef.com",
			},
			"anand": &dev{
				Name: "Anand Shankar", Email: "anand@beef.com",
			},
			"akshat": &dev{
				Name: "Akshat Shah", Email: "akshat@beef.com",
			},
		},
		Repos: map[string]*repo{
			"/a": &repo{
				Devs: []string{"karan"},
			},
		},
	}

	tests := []struct {
		desc        string
		wd          string
		author      string
		msg         string
		errMsg      string
		expectedMsg string
	}{
		{
			desc:        "no co authors",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "Line 1\n\nLine 2",
			expectedMsg: "Line 1\n\nLine 2\n\n",
		},
		{
			desc:        "co-author via repo",
			author:      "Anand Shankar <anand@beef.com>",
			msg:         "Line 1",
			expectedMsg: "Line 1\n\nCo-authored-by: Karan Misra <karan@beef.com>\n",
		},
		{
			desc:        "co-author added in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[anand] Line 1\n\nLine 2",
			expectedMsg: "Line 1\n\nLine 2\n\nCo-authored-by: Anand Shankar <anand@beef.com>\n",
		},
		{
			desc:        "co-author in first line is same as author",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[karan] Line 1",
			expectedMsg: "Line 1\n\n",
		},
		{
			desc:   "unknown dev in first line",
			author: "Karan Misra <karan@beef.com>",
			msg:    "[shobhit] Line 1",
			errMsg: "non-existing dev shobhit provided in the first line",
		},
		{
			desc:        "co-author in message",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "Line 1\n\nCo-authored-by: Anand Shankar <anand@beef.com>",
			expectedMsg: "Line 1\n\nCo-authored-by: Anand Shankar <anand@beef.com>\n",
		},
		{
			desc:        "new co-author in first line with existing co-author in message",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[akshat] Line 1\n\nCo-authored-by: Anand Shankar <anand@beef.com>",
			expectedMsg: "Line 1\n\nCo-authored-by: Akshat Shah <akshat@beef.com>\n",
		},
		{
			desc:        "unknown co-author in message",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "Line 1\n\nCo-authored-by: Unknown <unknown@beef.com>",
			expectedMsg: "Line 1\n\nCo-authored-by: Unknown <unknown@beef.com>\n",
		},
		{
			desc:        "simple issue id in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[1337] Line 1",
			expectedMsg: "Line 1\n\nIssue-id: #1337\n\n",
		},
		{
			desc:        "simple issue id with hash in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[#1337] Line 1",
			expectedMsg: "Line 1\n\nIssue-id: #1337\n\n",
		},
		{
			desc:        "complex issue id in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[GOJ-1337] Line 1",
			expectedMsg: "Line 1\n\nIssue-id: GOJ-1337\n\n",
		},
		{
			desc:        "issue id and co-author in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[#1337,anand] Line 1",
			expectedMsg: "Line 1\n\nIssue-id: #1337\n\nCo-authored-by: Anand Shankar <anand@beef.com>\n",
		},
		{
			desc:        "issue id and co-author in first line",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[GOJ-1337|anand|akshat] Line 1",
			expectedMsg: "Line 1\n\nIssue-id: GOJ-1337\n\nCo-authored-by: Akshat Shah <akshat@beef.com>\nCo-authored-by: Anand Shankar <anand@beef.com>\n",
		},
		{
			desc:        "co-author in first line and issue id in message",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "[anand] Line 1\n\nIssue-id: #1337",
			expectedMsg: "Line 1\n\nIssue-id: #1337\n\nCo-authored-by: Anand Shankar <anand@beef.com>\n",
		},
		{
			desc:        "complex issue id in message",
			author:      "Karan Misra <karan@beef.com>",
			msg:         "Line 1\n\nIssue-id: GOJ-1337",
			expectedMsg: "Line 1\n\nIssue-id: GOJ-1337\n\n",
		},
	}

	oldGitVar := gitVar
	defer func() {
		gitVar = oldGitVar
	}()

	for _, tt := range tests {
		t.Logf("case: %s", tt.desc)

		f, err := ioutil.TempFile("", "")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if err := ioutil.WriteFile(f.Name(), []byte(tt.msg), 0700); err != nil {
			panic(err)
		}

		gitVar = func(_ string) (string, error) {
			return tt.author, nil
		}

		err = d.AppendInfo("/a", f.Name())

		if tt.errMsg != "" {
			if !assert.Error(t, err) {
				continue
			}

			assert.Equal(t, tt.errMsg, err.Error())
			continue
		}

		if !assert.NoError(t, err) {
			continue
		}

		msg, err := ioutil.ReadFile(f.Name())
		if !assert.NoError(t, err) {
			continue
		}

		assert.Equal(t, tt.expectedMsg, string(msg))
	}
}

func TestFirstLineIDs(t *testing.T) {
	tests := []struct {
		msg string
		ids []string
		idx int
	}{
		{
			msg: "Hello there",
			ids: nil,
			idx: 0,
		},
		{
			msg: "[hello there",
			ids: nil,
			idx: 0,
		},
		{
			msg: "[k]hello there",
			ids: []string{"k"},
			idx: 3,
		},
		{
			msg: " [k]Hello there",
			ids: nil,
			idx: 0,
		},
		{
			msg: "[a,b,c]hello there",
			ids: []string{"a", "b", "c"},
			idx: 7,
		},
		{
			msg: "[a|b|c]hello there",
			ids: []string{"a", "b", "c"},
			idx: 7,
		},
		{
			msg: "[GOJ-1337,a,b,c] hello there",
			ids: []string{"GOJ-1337", "a", "b", "c"},
			idx: 16,
		},
	}

	for _, tt := range tests {
		ids, idx := firstLineIDs(tt.msg)
		assert.Equal(t, tt.ids, ids)
		assert.Equal(t, tt.idx, idx)
	}
}

func TestNameEmail(t *testing.T) {
	tests := []struct {
		ident       string
		name, email string
	}{
		{
			"Karan Misra <kidoman@gmail.com> 1551654611 +0530",
			"Karan Misra", "kidoman@gmail.com",
		},
		{
			"name <name> 1551654611 +0530",
			"name", "name",
		},
		{
			"Co-authored-by: name <email>",
			"name", "email",
		},
	}

	for _, tt := range tests {
		name, email := nameEmail(tt.ident)
		assert.Equal(t, tt.name, name)
		assert.Equal(t, tt.email, email)
	}
}
