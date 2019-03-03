package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	}

	for _, tt := range tests {
		name, email := nameEmail(tt.ident)
		assert.Equal(t, tt.name, name)
		assert.Equal(t, tt.email, email)
	}
}
