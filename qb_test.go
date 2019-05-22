package qb_test

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/haleyrc/qb"
	"github.com/jmoiron/sqlx"
)

func TestQuery(t *testing.T) {
	type output struct {
		query string
		vals  []interface{}
	}

	type testcase struct {
		name  string
		query qb.Query
		want  output
	}

	testcases := []testcase{
		testcase{
			name:  "simple query",
			query: qb.Select("dealerships"),
			want: output{
				query: `SELECT * FROM dealerships;`,
			},
		},
		testcase{
			name:  "simple query with fields",
			query: qb.Select("dealerships", "id", "name"),
			want: output{
				query: `SELECT id, name FROM dealerships;`,
			},
		},
		testcase{
			name: "simple query with fields and where",
			query: qb.
				Select("dealerships", "id", "name").
				Where(qb.Equal("state", "NY")),
			want: output{
				query: `SELECT id, name FROM dealerships WHERE state = ?;`,
				vals:  []interface{}{"NY"},
			},
		},
		testcase{
			name: "simple query with greater than",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Greater("cost", 10)),
			want: output{
				query: `SELECT id FROM vehicles WHERE cost > ?;`,
				vals:  []interface{}{10},
			},
		},
		testcase{
			name: "simple query with greater or equal",
			query: qb.
				Select("vehicles", "id").
				Where(qb.GreaterEqual("cost", 10)),
			want: output{
				query: `SELECT id FROM vehicles WHERE cost >= ?;`,
				vals:  []interface{}{10},
			},
		},
		testcase{
			name: "simple query with less than",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Less("cost", 10)),
			want: output{
				query: `SELECT id FROM vehicles WHERE cost < ?;`,
				vals:  []interface{}{10},
			},
		},
		testcase{
			name: "simple query with less or equal",
			query: qb.
				Select("vehicles", "id").
				Where(qb.LessEqual("cost", 10)),
			want: output{
				query: `SELECT id FROM vehicles WHERE cost <= ?;`,
				vals:  []interface{}{10},
			},
		},
		testcase{
			name: "simple query with and",
			query: qb.
				Select("vehicles", "id").
				Where(qb.And(
					qb.Greater("cost", 10),
					qb.Less("dol", 3),
				)),
			want: output{
				query: `SELECT id FROM vehicles WHERE (cost > ? AND dol < ?);`,
				vals:  []interface{}{10, 3},
			},
		},
		testcase{
			name: "simple query with or",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Or(
					qb.Greater("cost", 10),
					qb.Less("dol", 3),
				)),
			want: output{
				query: `SELECT id FROM vehicles WHERE (cost > ? OR dol < ?);`,
				vals:  []interface{}{10, 3},
			},
		},
		testcase{
			name: "query with nested boolean",
			query: qb.
				Select("vehicles", "id").
				Where(
					qb.And(
						qb.Equal("model", "Honda"),
						qb.Or(
							qb.Greater("cost", 10),
							qb.Less("dol", 3),
						))),
			want: output{
				query: `SELECT id FROM vehicles WHERE (model = ? AND (cost > ? OR dol < ?));`,
				vals:  []interface{}{"Honda", 10, 3},
			},
		},
		testcase{
			name: "simple query with in",
			query: qb.
				Select("vehicles", "id").
				Where(qb.In("make")),
			want: output{
				query: `SELECT id FROM vehicles WHERE make IN (?);`,
			},
		},
		testcase{
			name: "join query",
			query: qb.Join(
				qb.Select("employees", "id", "role"),
				qb.Select("dealerships", "name"),
			).On("employees.dealership_id", "dealerships.id"),
			want: output{
				query: `SELECT employees.id, employees.role, dealerships.name FROM employees, dealerships WHERE employees.dealership_id = dealerships.id;`,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotQuery := tc.query.Build()
			gotVals := tc.query.Values()

			t.Logf("query:\n%s", spew.Sdump(tc.query))
			t.Logf("original:\n\t%s", gotQuery)
			t.Logf("rebound:\n\t%s", sqlx.Rebind(sqlx.DOLLAR, gotQuery))
			t.Logf("values:\n\t%v", gotVals)

			if gotQuery != tc.want.query {
				t.Errorf("\n\twanted:\n%s\n\tgot:\n%s", tc.want.query, gotQuery)
			}

			if !reflect.DeepEqual(gotVals, tc.want.vals) {
				t.Errorf("\n\twanted:\n%v\n\tgot:\n%v", tc.want.vals, gotVals)
			}
		})
	}
}
