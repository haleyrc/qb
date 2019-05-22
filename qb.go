package qb

import (
	"encoding/json"
	"fmt"
	"strings"
)

func In(field string) InClause {
	return InClause(field)
}

type InClause string

func (c InClause) Build() string {
	return fmt.Sprintf("%s IN (?)", string(c))
}

func (c InClause) String() string {
	return c.Build()
}

type ComparisonClause struct {
	Op    string
	Field string
	Value interface{}
}

func (c ComparisonClause) Build() string {
	switch v := c.Value.(type) {
	case string:
		return fmt.Sprintf("%s %s '%s'", c.Field, c.Op, v)
	case int:
		return fmt.Sprintf("%s %s %d", c.Field, c.Op, v)
	default:
		return fmt.Sprintf("%s %s %s", c.Field, c.Op, c.Value)
	}
}

func (c ComparisonClause) String() string {
	return c.Build()
}

func Greater(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    ">",
		Field: field,
		Value: value,
	}
}

func GreaterEqual(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    ">=",
		Field: field,
		Value: value,
	}
}

func Less(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "<",
		Field: field,
		Value: value,
	}
}

func LessEqual(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "<=",
		Field: field,
		Value: value,
	}
}

func Equal(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "=",
		Field: field,
		Value: value,
	}
}

func Or(comp1, comp2 Query) BooleanQuery {
	return BooleanQuery{
		Op:          "OR",
		Comparison1: comp1,
		Comparison2: comp2,
	}
}

func And(comp1, comp2 Query) BooleanQuery {
	return BooleanQuery{
		Op:          "AND",
		Comparison1: comp1,
		Comparison2: comp2,
	}
}

type BooleanQuery struct {
	Op          string
	Comparison1 Query
	Comparison2 Query
}

func (q BooleanQuery) Build() string {
	return fmt.Sprintf("(%s %s %s)", q.Comparison1.Build(), q.Op, q.Comparison2.Build())
}

func (q BooleanQuery) String() string {
	return q.Build()
}

type Query interface {
	fmt.Stringer
	Build() string
}

type SelectQuery struct {
	Table       string
	Fields      []string
	WhereClause Query
}

func (q SelectQuery) Where(wq Query) SelectQuery {
	q.WhereClause = wq
	return q
}

func (q SelectQuery) Build() string {
	var stmt string
	if len(q.Fields) == 0 {
		stmt = fmt.Sprintf("SELECT * FROM %s", q.Table)
	} else {
		fields := strings.Join(q.Fields, ", ")
		stmt = fmt.Sprintf("SELECT %s FROM %s", fields, q.Table)
	}
	if q.WhereClause != nil {
		stmt += fmt.Sprintf(" WHERE %s", q.WhereClause.Build())
	}
	return stmt + ";"
}

func (q SelectQuery) String() string {
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

func Select(table string, fields ...string) SelectQuery {
	return SelectQuery{
		Table:  table,
		Fields: fields,
	}
}

func Join(sq1, sq2 SelectQuery) JoinQuery {
	return JoinQuery{
		Query1: sq1,
		Query2: sq2,
	}
}

type On struct {
	Field1 string
	Field2 string
}

func (o On) Build() string {
	return fmt.Sprintf("%s = %s", o.Field1, o.Field2)
}

func (o On) String() string {
	return o.Build()
}

type JoinQuery struct {
	Query1   SelectQuery
	Query2   SelectQuery
	OnClause Query
}

func (q JoinQuery) Build() string {
	fields := make([]string, 0)
	for _, field := range q.Query1.Fields {
		fields = append(fields, q.Query1.Table+"."+field)
	}
	for _, field := range q.Query2.Fields {
		fields = append(fields, q.Query2.Table+"."+field)
	}

	stmt := fmt.Sprintf("SELECT %s FROM %s, %s", strings.Join(fields, ", "), q.Query1.Table, q.Query2.Table)
	stmt += fmt.Sprintf(" WHERE %s", q.OnClause.Build())
	return stmt + ";"
}

func (q JoinQuery) String() string {
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

func (q JoinQuery) On(field1, field2 string) JoinQuery {
	q.OnClause = On{
		Field1: field1,
		Field2: field2,
	}
	return q
}
