package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_Empty(t *testing.T) {
	qb := NewQueryBuilder()
	where, args := qb.Build()
	assert.Equal(t, "", where)
	assert.Empty(t, args)
}

func TestQueryBuilder_Build(t *testing.T) {
	qb := NewQueryBuilder()
	qb.Add("a = $%d", 1)
	qb.Add("b = $%d", "x")
	where, args := qb.Build()
	assert.True(t, strings.HasPrefix(where, " AND "), where)
	assert.Contains(t, where, "a = $1")
	assert.Contains(t, where, "b = $2")
	assert.Equal(t, []any{1, "x"}, args)
}
