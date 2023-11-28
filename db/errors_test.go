package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnownErrorsMatchKnownErrorsPrecedence(t *testing.T) {
	assert.Equal(t, len(KnownErrors), len(knownErrorsPrecedence))

	for _, errStr := range knownErrorsPrecedence {
		_, found := KnownErrors[errStr]
		assert.True(t, found, "%s not in KnownErrors", errStr)
	}
}
