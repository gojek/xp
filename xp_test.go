package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	expectedData := data{
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
    storyId: ""
`)

	data, err := load(r)
	assert.NoError(t, err)

	assert.Equal(t, &expectedData, data)
}

func TestDataString(t *testing.T) {
	data := data{
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
    storyId: ""
`

	assert.Equal(t, expectedString, data.String())
}

func TestDataStore(t *testing.T) {
	data := data{
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
	assert.NoError(t, data.store(&buf))

	expectedString := `devs:
  ak:
    email: akshat@beef.com
    name: akshat
repos:
  /path/to/repo:
    devs:
    - ak
    storyId: ""
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
	var d data

	d.addDev("km", "Karan Misra", "karan@beef.com")

	assert.Equal(t, &dev{Name: "Karan Misra", Email: "karan@beef.com"}, d.Devs["km"])

	d = data{
		Devs: map[string]*dev{
			"ak": &dev{
				Name:  "akshat",
				Email: "akshat@beef.com",
			},
		},
	}

	d.addDev("km", "Karan Misra", "karan@beef.com")

	assert.Equal(t, &dev{Name: "Karan Misra", Email: "karan@beef.com"}, d.Devs["km"])
	assert.Equal(t, &dev{Name: "akshat", Email: "akshat@beef.com"}, d.Devs["ak"])
}

func TestDataLookupDev(t *testing.T) {
	var d data

	assert.Equal(t, (*dev)(nil), d.lookupDev("km"))

	d = data{
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
	d := data{
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
		storyID string
		errMsg  string
	}{
		{
			path:    "/some/path",
			devIDs:  []string{"ak"},
			storyID: "o-1",
			errMsg:  "",
		},
		{
			devIDs: []string{"anand"},
			errMsg: "dev ids validation failed: no dev with id anand found",
		},
		{
			path:    "/some/path",
			devIDs:  []string{"ak", "km"},
			storyID: "o-1",
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		d := data{
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

		err := d.addRepo(tt.path, tt.devIDs, tt.storyID)

		if tt.errMsg != "" {
			if assert.Error(t, err) {
				assert.Equal(t, tt.errMsg, err.Error())
			}

			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, &repo{Devs: tt.devIDs, StoryID: tt.storyID}, d.Repos[tt.path])
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
	}

	for _, tt := range tests {
		t.Logf("case: %s", tt.desc)

		repoDir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		defer os.RemoveAll(repoDir)

		if !assert.NoError(t, tt.prepareFn(repoDir)) {
			continue
		}

		err = initRepo(repoDir, tt.overwrite)

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

	assert.Equal(t, hookStr, string(data))
}

func TestFirstLineDevIDs(t *testing.T) {
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
	}

	for _, tt := range tests {
		ids, idx := firstLineDevIDs(tt.msg)
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
