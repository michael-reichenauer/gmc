package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	var names []string
	assert.Nil(t, names)

	names2 := []string{}
	assert.NotNil(t, names2)

	assert.Equal(t, len(names), len(names2))

	names3 := []string{}
	assert.True(t, &names2 != &names3)
	assert.NotEqual(t, names, names2)
	assert.Equal(t, names2, names3)
	names2 = []string{"name"}
	names3 = []string{"name"}
	assert.Equal(t, names2, names3)
	names3 = []string{"name3"}
	assert.NotEqual(t, names2, names3)
}
