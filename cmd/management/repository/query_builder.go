package repository

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	conditions []string
	args       []any
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Add(condFmt string, value any) {
	qb.args = append(qb.args, value)
	qb.conditions = append(qb.conditions, fmt.Sprintf(condFmt, len(qb.args)))
}

func (qb *QueryBuilder) ArgCount() int {
	return len(qb.args)
}

func (qb *QueryBuilder) Build() (string, []any) {
	if len(qb.conditions) == 0 {
		return "", qb.args
	}
	return " AND " + strings.Join(qb.conditions, " AND "), qb.args
}
