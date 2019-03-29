package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
