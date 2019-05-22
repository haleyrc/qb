package qb_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/haleyrc/qb"
)

func TestQuery(t *testing.T) {
	type testcase struct {
		name  string
		query qb.Query
		want  string
	}
	testcases := []testcase{
		testcase{
			name:  "simple query",
			query: qb.Select("dealerships"),
			want:  `SELECT * FROM dealerships;`,
		},
		testcase{
			name:  "simple query with fields",
			query: qb.Select("dealerships", "id", "name"),
			want:  `SELECT id, name FROM dealerships;`,
		},
		testcase{
			name: "simple query with fields and where",
			query: qb.
				Select("dealerships", "id", "name").
				Where(qb.Equal("state", "NY")),
			want: `SELECT id, name FROM dealerships WHERE state = 'NY';`,
		},
		testcase{
			name: "simple query with greater than",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Greater("cost", 10)),
			want: `SELECT id FROM vehicles WHERE cost > 10;`,
		},
		testcase{
			name: "simple query with greater or equal",
			query: qb.
				Select("vehicles", "id").
				Where(qb.GreaterEqual("cost", 10)),
			want: `SELECT id FROM vehicles WHERE cost >= 10;`,
		},
		testcase{
			name: "simple query with less than",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Less("cost", 10)),
			want: `SELECT id FROM vehicles WHERE cost < 10;`,
		},
		testcase{
			name: "simple query with less or equal",
			query: qb.
				Select("vehicles", "id").
				Where(qb.LessEqual("cost", 10)),
			want: `SELECT id FROM vehicles WHERE cost <= 10;`,
		},
		testcase{
			name: "simple query with and",
			query: qb.
				Select("vehicles", "id").
				Where(qb.And(
					qb.Greater("cost", 10),
					qb.Less("dol", 3),
				)),
			want: `SELECT id FROM vehicles WHERE (cost > 10 AND dol < 3);`,
		},
		testcase{
			name: "simple query with or",
			query: qb.
				Select("vehicles", "id").
				Where(qb.Or(
					qb.Greater("cost", 10),
					qb.Less("dol", 3),
				)),
			want: `SELECT id FROM vehicles WHERE (cost > 10 OR dol < 3);`,
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
			want: `SELECT id FROM vehicles WHERE (model = 'Honda' AND (cost > 10 OR dol < 3));`,
		},
		testcase{
			name: "simple query with in",
			query: qb.
				Select("vehicles", "id").
				Where(qb.In("make")),
			want: `SELECT id FROM vehicles WHERE make IN (?);`,
		},
		testcase{
			name: "join query",
			query: qb.Join(
				qb.Select("employees", "id", "role"),
				qb.Select("dealerships", "name"),
			).On("employees.dealership_id", "dealerships.id"),
			want: `SELECT employees.id, employees.role, dealerships.name FROM employees, dealerships WHERE employees.dealership_id = dealerships.id;`,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.query.Build()
			if got != tc.want {
				t.Errorf("built:\n%s\nwanted:\n%s\ngot:\n%s", indent(tc.query), tc.want, got)
			}
		})
	}
}

func indent(stringer fmt.Stringer) string {
	s := stringer.String()
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = "\t" + lines[i]
	}
	return strings.Join(lines, "\n")
}
