package qb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Query is the primary interface our components must implement.
type Query interface {
	fmt.Stringer

	// Build returns an unbound query string. Compiling the string may involve
	// nested calls to Build for the query's subcomponents.
	Build() string

	// Values returns a slice of values that must be in the same order as their
	// respective locations in the query string.
	Values() []interface{}
}

// In returns a new IN clause that resolves to the form `field IN (?)`.
func In(field string) InClause {
	return InClause(field)
}

// InClause represents an SQL query where a column value can be one of multiple
// potential values. Currently this is the only query type that doesn't retain
// any comparison values, but since we have to rebind the query as a client
// anyway, we can probably extend this to include them.
//
// TODO (RCH): Add the comparison values
type InClause string

// Build returns an IN clause of the form `field IN (?)`.
func (c InClause) Build() string {
	return fmt.Sprintf("%s IN (?)", string(c))
}

func (c InClause) String() string {
	return c.Build()
}

// Values returns nil since we don't store the comparison values for IN clauses
// in the current implementation.
func (c InClause) Values() []interface{} {
	return nil
}

// Greater returns a boolean clause that resolves to the form `(field > value)`.
func Greater(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    ">",
		Field: field,
		Value: value,
	}
}

// GreaterEqual returns a boolean clause that resolves to the form
// `(field >= value)`.
func GreaterEqual(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    ">=",
		Field: field,
		Value: value,
	}
}

// Less returns a boolean clause that resolves to the form `(field < value)`.
func Less(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "<",
		Field: field,
		Value: value,
	}
}

// LessEqual returns a boolean clause that resolves to the form
// `(field <= value)`.
func LessEqual(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "<=",
		Field: field,
		Value: value,
	}
}

// Equal returns a boolean clause that resolves to the form `(field = value)`.
func Equal(field string, value interface{}) ComparisonClause {
	return ComparisonClause{
		Op:    "=",
		Field: field,
		Value: value,
	}
}

// ComparisonClause represents a binary boolean expression. Comparison clauses
// are automatically surrounded by parentheses to prevent order-of-operations
// issues in the resulting query.
//
// TODO (RCH): We might be able to combine this with BooleanQuery, but I think
// we'd either have to introduce additional Query implementations that just
// resolve to a string, for example, with no values. Otherwise, we'll have to do
// the type check anywhere it matters, which might actually be the simplest way,
// but may or may not be worth it.
type ComparisonClause struct {
	// Op is a boolean operator e.g. =, <=, etc.
	Op string

	// Field is the LHS of the boolean expression.
	Field string

	// Value is the RHS of the boolean expression. Value can also be a Query which
	// will be built and injected appropriately.
	Value interface{}
}

// Build returns a binary binary boolean expression of the form
// `(field op value)` in the case of simple values, or `(field op (subquery))`
// if the value is a Query.
func (c ComparisonClause) Build() string {
	if q, ok := c.Value.(Query); ok {
		return fmt.Sprintf("%s %s (%s)", c.Field, c.Op, q.Build())
	}
	return fmt.Sprintf("%s %s ?", c.Field, c.Op)
}

func (c ComparisonClause) String() string {
	return c.Build()
}

// Values returns the RHS value in the case of simple expressions. If the value
// is a query, it returns the values for that subquery instead.
func (c ComparisonClause) Values() []interface{} {
	if q, ok := c.Value.(Query); ok {
		return q.Values()
	}
	return []interface{}{c.Value}
}

// Or returns a boolean query that resolves to the form `(expr OR expr)`.
func Or(comp1, comp2 Query) BooleanQuery {
	return BooleanQuery{
		Op:          "OR",
		Comparison1: comp1,
		Comparison2: comp2,
	}
}

// And returns a boolean query that resolves to the form `(expr AND expr)`.
func And(comp1, comp2 Query) BooleanQuery {
	return BooleanQuery{
		Op:          "AND",
		Comparison1: comp1,
		Comparison2: comp2,
	}
}

// BooleanQuery represents a binary boolean expression using logic operators.
// The primary distinction between this and ComparisonClause (apart from the
// supported operations) is that we allow Queries for both sides of the
// operation.
type BooleanQuery struct {
	Op          string
	Comparison1 Query
	Comparison2 Query
}

// Build returns a binary boolean expression of the form `(expr op expr)`. Where
// the `expr`s are the result of building the subqueries.
func (q BooleanQuery) Build() string {
	return fmt.Sprintf("(%s %s %s)", q.Comparison1.Build(), q.Op, q.Comparison2.Build())
}

func (q BooleanQuery) String() string {
	return q.Build()
}

// Values returns the aggregate of the values for the LHS and RHS subqueries.
func (q BooleanQuery) Values() []interface{} {
	vals := q.Comparison1.Values()
	return append(vals, q.Comparison2.Values()...)
}

// Delete returns a query that resolves to the general form `DELETE FROM table
// [WHERE expr]`.
func Delete(table string) DeleteQuery {
	return DeleteQuery{
		Table: table,
	}
}

// DeleteQuery represents a query that resolves to the general form `DELETE FROM
// table [WHERE expr]`.
type DeleteQuery struct {
	Table       string
	Vals        []interface{}
	WhereClause Query
}

// Build returns a query string of the form `DELETE FROM table [WHERE expr]`.
func (q DeleteQuery) Build() string {
	stmt := fmt.Sprintf("DELETE FROM %s", q.Table)
	if q.WhereClause != nil {
		stmt += fmt.Sprintf(" WHERE %s", q.WhereClause.Build())
	}
	return stmt
}

func (q DeleteQuery) String() string {
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

// Values returns the accumulated values for the query and any subqueries.
func (q DeleteQuery) Values() []interface{} {
	return q.Vals
}

// Where adds an additional WHERE clause condition to the query that will be
// evaluated and injected into the final query string.
func (q DeleteQuery) Where(wq Query) DeleteQuery {
	q.WhereClause = wq
	q.Vals = append(q.Vals, wq.Values()...)
	return q
}

// Select returns a query that resolves to the general form `SELECT fields FROM
// table [WHERE expr]`.
func Select(table string, fields ...string) SelectQuery {
	return SelectQuery{
		Table:  table,
		Fields: fields,
	}
}

// SelectQuery represents a query that resolves to the general form `SELECT
// fields FROM table [WHERE expr]`.
type SelectQuery struct {
	Table       string
	Fields      []string
	Vals        []interface{}
	WhereClause Query
}

// Build returns a query string of the general form `SELECT fields FROM table
// [WHERE expr]`.
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
	return stmt
}

func (q SelectQuery) String() string {
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

// Values returns the accumulated values for the query and any subqueries.
func (q SelectQuery) Values() []interface{} {
	return q.Vals
}

// Where adds an additional WHERE clause condition to the query that will be
// evaluated and injected into the final query string.
func (q SelectQuery) Where(wq Query) SelectQuery {
	q.WhereClause = wq
	q.Vals = append(q.Vals, wq.Values()...)
	return q
}

// On represents a specific implementation of a WHERE clause used for joining
// two tables.
type On struct {
	Field1 string
	Field2 string
}

// Build returns a clause of the form `field1 = field2` where the fields
// represent the related key/foreign key used in the table join.
func (o On) Build() string {
	return fmt.Sprintf("%s = %s", o.Field1, o.Field2)
}

func (o On) String() string {
	return o.Build()
}

// Values always returns nil for On.
func (o On) Values() []interface{} {
	return nil
}

// Join returns a query that resolves to the general form `SELECT fields FROM
// table1, table2 WHERE field1 = field2`. In the general form, field1 and field2
// should probably be an id/foreign key pair or you might get interesting
// results. The columns returned are automatically prepended with the related
// table name to prevent accidental collisions.
func Join(sq1, sq2 SelectQuery) JoinQuery {
	return JoinQuery{
		Query1: sq1,
		Query2: sq2,
	}
}

// JoinQuery represents a query that resolves to the general form `SELECT fields
// FROM table1, table2 WHERE field1 = field2`. In the general form, field1 and
// field2 should probably be an id/foreign key pair or you might get interesting
// results. The columns returned are automatically prepended with the related
// table name to prevent accidental collisions.
type JoinQuery struct {
	Query1   SelectQuery
	Query2   SelectQuery
	OnClause Query
}

// Build returns a query string of the general form `SELECT fields FROM table1,
// table2 WHERE field1 = field2`. In the general form, field1 and field2 should
// probably be an id/foreign key pair or you might get interesting results. The
// columns returned are automatically prepended with the related table name to
// prevent accidental collisions.
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
	// This feels pretty hacky, but somehow works
	if q1Where := q.Query1.WhereClause; q1Where != nil {
		stmt += fmt.Sprintf(" AND (%s)", q1Where.Build())
	}
	if q2Where := q.Query2.WhereClause; q2Where != nil {
		stmt += fmt.Sprintf(" AND (%s)", q2Where.Build())
	}
	return stmt
}

// On sets the fields for the WHERE query that is required to join the two
// tables.
func (q JoinQuery) On(field1, field2 string) JoinQuery {
	q.OnClause = On{
		Field1: field1,
		Field2: field2,
	}
	return q
}

func (q JoinQuery) String() string {
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}

// Values returns the aggregate of the values from the two Queries.
func (q JoinQuery) Values() []interface{} {
	vals := q.Query1.Values()
	return append(vals, q.Query2.Values()...)
}
