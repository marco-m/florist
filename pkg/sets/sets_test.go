package sets_test

import (
	"testing"

	"github.com/marco-m/florist/pkg/sets"
	"github.com/marco-m/rosina/assert"
)

func TestFromInt(t *testing.T) {
	type testCase struct {
		name       string
		items      []int
		wantSize   int
		wantList   []int
		wantString string
	}

	test := func(t *testing.T, tc testCase) {
		s := sets.From(tc.items...)
		sorted := s.OrderedList()

		assert.Equal(t, s.Size(), tc.wantSize, "set size")
		assert.DeepEqual(t, sorted, tc.wantList, "set ordered list")
		assert.Equal(t, s.String(), tc.wantString, "set string")
	}

	testCases := []testCase{
		{
			name:       "nil",
			items:      nil,
			wantSize:   0,
			wantList:   []int{},
			wantString: "[]",
		},
		{
			name:       "empty",
			items:      []int{},
			wantSize:   0,
			wantList:   []int{},
			wantString: "[]",
		},
		{
			name:       "non empty",
			items:      []int{2, 3, 1},
			wantSize:   3,
			wantList:   []int{1, 2, 3},
			wantString: "[1 2 3]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestFromString(t *testing.T) {
	type testCase struct {
		name       string
		items      []string
		wantSize   int
		wantList   []string
		wantString string
	}

	test := func(t *testing.T, tc testCase) {
		s := sets.From(tc.items...)
		sorted := s.OrderedList()

		assert.Equal(t, s.Size(), tc.wantSize, "set size")
		assert.DeepEqual(t, sorted, tc.wantList, "set ordered list")
		assert.Equal(t, s.String(), tc.wantString, "set string")
	}

	testCases := []testCase{
		{
			name:       "non empty",
			items:      []string{"b", "c", "a"},
			wantSize:   3,
			wantList:   []string{"a", "b", "c"},
			wantString: "[a b c]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestDifference(t *testing.T) {
	type testCase struct {
		name     string
		s        *sets.Set[int]
		x        *sets.Set[int]
		wantList []int
	}

	test := func(t *testing.T, tc testCase) {
		result := tc.s.Difference(tc.x)
		sorted := result.OrderedList()

		assert.DeepEqual(t, sorted, tc.wantList, "set ordered list")
	}

	testCases := []testCase{
		{
			name:     "both empty",
			s:        sets.From[int](),
			x:        sets.From[int](),
			wantList: []int{},
		},
		{
			name:     "empty x returns s",
			s:        sets.From(1, 2, 3),
			x:        sets.From[int](),
			wantList: []int{1, 2, 3},
		},
		{
			name:     "nothing in common returns s",
			s:        sets.From(1, 2, 3),
			x:        sets.From(4, 5),
			wantList: []int{1, 2, 3},
		},
		{
			name:     "one in common",
			s:        sets.From(1, 2, 3),
			x:        sets.From(4, 2),
			wantList: []int{1, 3},
		},
		{
			name:     "all in common returns empty set",
			s:        sets.From(1, 2, 3),
			x:        sets.From(1, 2, 3, 12),
			wantList: []int{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestIntersection(t *testing.T) {
	type testCase struct {
		name     string
		s        *sets.Set[int]
		x        *sets.Set[int]
		wantList []int
	}

	test := func(t *testing.T, tc testCase) {
		result := tc.s.Intersection(tc.x)
		sorted := result.OrderedList()

		assert.DeepEqual(t, sorted, tc.wantList, "set ordered list")
	}

	testCases := []testCase{
		{
			name:     "both empty",
			s:        sets.From[int](),
			x:        sets.From[int](),
			wantList: []int{},
		},
		{
			name:     "empty x returns empty",
			s:        sets.From(1, 2, 3),
			x:        sets.From[int](),
			wantList: []int{},
		},
		{
			name:     "nothing in common returns empty",
			s:        sets.From(1, 2, 3),
			x:        sets.From(4, 5),
			wantList: []int{},
		},
		{
			name:     "one in common",
			s:        sets.From(1, 2, 3),
			x:        sets.From(4, 2),
			wantList: []int{2},
		},
		{
			name:     "s subset of x returns s",
			s:        sets.From(1, 2, 3),
			x:        sets.From(1, 2, 3, 12),
			wantList: []int{1, 2, 3},
		},
		{
			name:     "x subset of s returns x",
			s:        sets.From(1, 2, 3, 12),
			x:        sets.From(1, 2, 3),
			wantList: []int{1, 2, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestUnion(t *testing.T) {
	type testCase struct {
		name     string
		s        *sets.Set[int]
		x        *sets.Set[int]
		wantList []int
	}

	test := func(t *testing.T, tc testCase) {
		result := tc.s.Union(tc.x)
		sorted := result.OrderedList()

		assert.DeepEqual(t, sorted, tc.wantList, "set ordered list")
	}

	testCases := []testCase{
		{
			name:     "both empty",
			s:        sets.From[int](),
			x:        sets.From[int](),
			wantList: []int{},
		},
		{
			name:     "empty x",
			s:        sets.From(1, 2),
			x:        sets.From[int](),
			wantList: []int{1, 2},
		},
		{
			name:     "empty s",
			s:        sets.From[int](),
			x:        sets.From(1, 2),
			wantList: []int{1, 2},
		},
		{
			name:     "identical",
			s:        sets.From(1, 2),
			x:        sets.From(1, 2),
			wantList: []int{1, 2},
		},
		{
			name:     "all different",
			s:        sets.From(1, 3),
			x:        sets.From(2, 4),
			wantList: []int{1, 2, 3, 4},
		},
		{
			name:     "partial overlap",
			s:        sets.From(1, 3),
			x:        sets.From(3, 5),
			wantList: []int{1, 3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestRemoveFound(t *testing.T) {
	type testCase struct {
		name     string
		items    []int
		remove   int
		wantList []int
	}

	test := func(t *testing.T, tc testCase) {
		s := sets.From(tc.items...)

		found := s.Remove(tc.remove)

		assert.DeepEqual(t, s.OrderedList(), tc.wantList, "set ordered list")
		assert.True(t, found, "found")
	}

	testCases := []testCase{
		{
			name:     "set with one element",
			items:    []int{42},
			remove:   42,
			wantList: []int{},
		},
		{
			name:     "set with multiple elements",
			items:    []int{-5, 100, 42},
			remove:   42,
			wantList: []int{-5, 100},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestRemoveNotFound(t *testing.T) {
	type testCase struct {
		name   string
		items  []int
		remove int
	}

	test := func(t *testing.T, tc testCase) {
		s := sets.From(tc.items...)

		found := s.Remove(tc.remove)

		assert.DeepEqual(t, s.OrderedList(), tc.items, "set ordered list")
		assert.False(t, found, "found")
	}

	testCases := []testCase{
		{
			name:   "empty set",
			items:  []int{},
			remove: 42,
		},
		{
			name:   "non empty set",
			items:  []int{10, 50},
			remove: 42,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}

func TestAdd(t *testing.T) {
	type testCase struct {
		name     string
		items    []int
		wantList []int
	}

	test := func(t *testing.T, tc testCase) {
		s := sets.New[int](5)
		for _, item := range tc.items {
			s.Add(item)
		}
		assert.DeepEqual(t, s.OrderedList(), tc.wantList, "set ordered list")
	}

	testCases := []testCase{
		{
			name:     "one item",
			items:    []int{3},
			wantList: []int{3},
		},
		{
			name:     "multiple items",
			items:    []int{3, 0, 42},
			wantList: []int{0, 3, 42},
		},
		{
			name:     "duplicates",
			items:    []int{10, 5, 5, 10, 1},
			wantList: []int{1, 5, 10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { test(t, tc) })
	}
}
