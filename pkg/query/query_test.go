package query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		expected Query
	}{
		{
			name: "equal string",
			raw:  `equal("email","a@b.com")`,
			expected: Query{
				Filters: []Filter{{Op: "equal", Attribute: "email", Values: []string{"a@b.com"}}},
			},
		},
		{
			name: "equal array",
			raw:  `equal("status",["active","pending"])`,
			expected: Query{
				Filters: []Filter{{Op: "equal", Attribute: "status", Values: []string{"active", "pending"}}},
			},
		},
		{
			name: "between",
			raw:  `between("age",18,65)`,
			expected: Query{
				Filters: []Filter{{Op: "between", Attribute: "age", Values: []string{"18", "65"}}},
			},
		},
		{
			name: "order desc",
			raw:  `orderDesc("createdAt")`,
			expected: Query{
				Orders: []Order{{Attribute: "createdAt", Desc: true}},
			},
		},
		{
			name: "limit",
			raw:  `limit(25)`,
			expected: Query{
				Limit: 25,
			},
		},
		{
			name: "select",
			raw:  `select(["name","email"])`,
			expected: Query{
				Selects: []string{"name", "email"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := Parse(tc.raw)
			require.NoError(t, err)
			require.Equal(t, &tc.expected, q)
		})
	}
}

func TestParseMany(t *testing.T) {
	q, err := ParseMany([]string{
		`equal("status","active")`,
		`greaterThan("age",18)`,
		`orderDesc("createdAt")`,
		`limit(10)`,
		`offset(20)`,
	})
	require.NoError(t, err)
	require.Len(t, q.Filters, 2)
	require.Len(t, q.Orders, 1)
	require.Equal(t, 10, q.Limit)
	require.Equal(t, 20, q.Offset)
}
