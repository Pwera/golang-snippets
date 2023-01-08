package fuego

import (
	"fmt"
	"github.com/seborama/fuego/v11"
	"github.com/stretchr/testify/assert"
	"hash/crc32"
	"hash/fnv"
	"strings"
	"testing"
)

const PanicDuplicateKey = "duplicate key"

var stringLength = func(el string) int {
	return len(el)
}

var stringToUpper = func(el string) string {
	return strings.ToUpper(el)
}
var stringLengthGreaterThan = func(length int) fuego.Predicate[string] {
	return func(el string) bool {
		return len(el) > length
	}
}
var intGreaterThanPredicate = func(rhs int) fuego.Predicate[int] {
	return func(lhs int) bool {
		return lhs > rhs
	}
}

var isString = func(val string) fuego.Predicate[string] {
	return func(val string) bool {
		//_, ok := val.(string)
		return true
	}
}

var stringHash = func(s string) uint32 {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	return h.Sum32()
}

var concatenateStringsBiFunc = func(i, j string) string {
	iStr := i
	jStr := j
	return iStr + jStr
}

var toStringList = func(e string) []string {
	var r []string
	for _, c := range e {
		r = append(r, string(c))
	}
	return r
}

var flattenStringListToDistinct = func(e []string) fuego.Stream[string] {
	return fuego.NewStreamFromSlice(e, 0).
		Distinct(func(s string) uint32 { return crc32.ChecksumIEEE([]byte(s)) })
}

func TestCollector_GroupingBy_Mapping_Filtering_ToEntrySlice(t *testing.T) {

	tt := map[string]struct {
		inputData []string
		expected  map[int][]string
	}{
		"1:": {
			inputData: []string{
				"a",
				"b",
				"bb",
				"bb",
				"cc",
				"ddd",
			},
			expected: map[int][]string{
				1: {},
				2: {"BB", "CC"},
				3: {"DDD"},
			},
		},
		"2:": {
			inputData: []string{
				"a",
				"b",
				"c",
				"bb",
				"bb",
				"cc",
				"ddd",
				"eee",
				"fff",
			},
			expected: map[int][]string{
				1: {},
				2: {"BB", "CC"},
				3: {"DDD", "EEE", "FFF"},
			},
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := fuego.Collect(
				fuego.NewStreamFromSlice[string](tc.inputData, 100).
					Distinct(stringHash),
				fuego.GroupingBy(
					stringLength,
					fuego.Mapping(
						stringToUpper,
						fuego.Filtering(
							stringLengthGreaterThan(1),
							fuego.ToSlice[string](),
						),
					),
				),
			)
			fmt.Println(got)

			if !assert.Equal(t, tc.expected, got) {
				t.Fatalf("After stream processing got: %v\nexpected: %v", got, tc.expected)
			}

		})
	}
}

func TestCollector_Collect_ToEntryMap(t *testing.T) {
	tt := map[string]struct {
		inputData     []employee
		expected      map[string]int
		expectedPanic string
	}{
		"panics when key exists": {
			inputData: []employee{
				{
					id:   1,
					name: "One",
				},
				{
					id:   1000,
					name: "One",
				},
			},
			expectedPanic: PanicDuplicateKey + ": 'One'",
		},
		"returns a map of employee (name, id)": {
			inputData: getEmployeesSample(),
			expected: map[string]int{
				"One":   1,
				"Two":   2,
				"Three": 3,
				"Four":  4,
				"Five":  5,
			},
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			employeeNameByID := func() map[string]int {
				return fuego.Collect(
					fuego.NewStreamFromSlice(tc.inputData, 0),
					fuego.ToMap(employee.Name, employee.ID),
				)
			}

			if tc.expectedPanic != "" {
				assert.PanicsWithValue(t, tc.expectedPanic, func() { _ = employeeNameByID() })
				return
			}

			assert.EqualValues(t, tc.expected, employeeNameByID())
		})
	}
}

func TestCollector_Collect_ToEntryMapWithKeyMerge(t *testing.T) {
	employees := getEmployeesSample()

	overwriteKeyMergeFn := func(v1, v2 int) int {
		return v2
	}

	employeeNameByID :=
		fuego.Collect(
			fuego.NewStreamFromSlice(employees, 0),
			fuego.ToMapWithMerge(employee.Department, employee.ID, overwriteKeyMergeFn))

	expected := map[string]int{
		"HR":        5,
		"IT":        3,
		"Marketing": 1,
	}

	assert.EqualValues(t, expected, employeeNameByID)
}

func TestCollector_Filtering(t *testing.T) {
	employees := getEmployeesSample()

	downstreamFc := fuego.Filtering(func(e employee) bool {
		return e.Salary() > 2000
	}, fuego.ToSlice[employee]())
	highestPaidEmployeesByDepartment :=
		fuego.Collect(
			fuego.NewStreamFromSlice(employees, 0),
			fuego.GroupingBy(employee.Department, downstreamFc))

	expected := map[string][]employee{
		"HR": {
			{
				id:         5,
				name:       "Five",
				department: "HR",
				salary:     2300,
			}},
		"IT": {
			{
				id:         2,
				name:       "Two",
				department: "IT",
				salary:     2500,
			},
			{
				id:         3,
				name:       "Three",
				department: "IT",
				salary:     2200,
			}},
		"Marketing": {},
	}

	assert.EqualValues(t, expected, highestPaidEmployeesByDepartment)
}

func TestCollector_GroupingBy_Mapping_FlatMapping_Filtering_Mapping_Reducing(t *testing.T) {
	strs := []string{
		"a",
		"bb",
		"cc",
		"ee",
		"ddd",
	}

	got :=
		fuego.Collect(
			fuego.NewStreamFromSlice(strs, 0),
			fuego.GroupingBy(
				stringLength,
				fuego.Mapping(
					toStringList,
					fuego.FlatMapping(flattenStringListToDistinct,
						fuego.Mapping(stringToUpper,
							fuego.Reducing(concatenateStringsBiFunc),
						),
					),
				),
			),
		)

	expected := map[int]string{
		1: "A",
		2: "BCE",
		3: "D",
	}

	assert.EqualValues(t, expected, got)
}

func TestNotPredicate(t *testing.T) {
	type args struct {
		p fuego.Predicate[int]
		t int
	}
	tt := map[string]struct {
		args args
		want bool
	}{
		"Should negate the predicate": {
			args: args{
				p: intGreaterThanPredicate(5),
				t: 7,
			},
			want: false,
		},
		"Should confirm the predicate": {
			args: args{
				p: intGreaterThanPredicate(10),
				t: 7,
			},
			want: true,
		},
		"Should return true when nil predicate": /* TODO: is that correct? */ {
			args: args{
				p: nil,
				t: 2,
			},
			want: true,
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := tc.args.p.Negate()(tc.args.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFalsePredicate(t *testing.T) {
	type args struct {
		t any
	}
	tt := map[string]struct {
		args args
		want bool
	}{
		"Should return false when '123'": {
			args: args{
				t: 123,
			},
			want: false,
		},
		"Should return false when 'Hello World'": {
			args: args{
				t: "Hello World",
			},
			want: false,
		},
		"Should return false when 'true'": {
			args: args{
				t: true,
			},
			want: false,
		},
		"Should return false when 'false'": {
			args: args{
				t: false,
			},
			want: false,
		},
		"Should return false when 'nil'": {
			args: args{
				t: nil,
			},
			want: false,
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := fuego.False[any]()(tc.args.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTruePredicate(t *testing.T) {
	type args struct {
		t any
	}
	tt := map[string]struct {
		args args
		want bool
	}{
		"Should return true when '123'": {
			args: args{
				t: 123,
			},
			want: true,
		},
		"Should return true when 'Hello World'": {
			args: args{
				t: "Hello World",
			},
			want: true,
		},
		"Should return true when 'true'": {
			args: args{
				t: true,
			},
			want: true,
		},
		"Should return true when 'false'": {
			args: args{
				t: false,
			},
			want: true,
		},
		"Should return true when 'nil'": {
			args: args{
				t: nil,
			},
			want: true,
		},
	}

	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := fuego.True[any]()(tc.args.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAndPredicate(t *testing.T) {
	tt := map[string]struct {
		p1   fuego.Predicate[int]
		p2   fuego.Predicate[int]
		want bool
	}{
		"Should return true for: true AND true": {
			p1:   fuego.True[int](),
			p2:   fuego.True[int](),
			want: true,
		},
		"Should return false for: true AND false": {
			p1:   fuego.True[int](),
			p2:   fuego.False[int](),
			want: false,
		},
		"Should return false for: false AND true": {
			p1:   fuego.False[int](),
			p2:   fuego.True[int](),
			want: false,
		},
		"Should return false for: false AND false": {
			p1:   fuego.False[int](),
			p2:   fuego.False[int](),
			want: false,
		},
		"Should return false for: nil AND true": {
			p1:   nil,
			p2:   fuego.True[int](),
			want: false,
		},
		"Should return false for: true AND nil": {
			p1:   fuego.True[int](),
			p2:   nil,
			want: false,
		},
		"Should return false for: nil AND false": {
			p1:   nil,
			p2:   fuego.False[int](),
			want: false,
		},
		"Should return false for: false AND nil": {
			p1:   fuego.False[int](),
			p2:   nil,
			want: false,
		},
		"Should return false for: nil AND nil": {
			p1:   nil,
			p2:   nil,
			want: false,
		},
	}
	for name, tc := range tt {
		tc := tc
		t.Run(name, func(t *testing.T) {
			got := tc.p1.And(tc.p2)(0)
			assert.Equal(t, tc.want, got)
		})
	}
}

type employee struct {
	id         int
	name       string
	department string
	salary     float32
}

func (e employee) ID() int {
	return e.id
}

func (e employee) Name() string {
	return e.name
}

func (e employee) Department() string {
	return e.department
}

func (e employee) Salary() float32 {
	return e.salary
}

func getEmployeesSample() []employee {
	return []employee{
		{
			id:         1,
			name:       "One",
			department: "Marketing",
			salary:     1500,
		},
		{
			id:         2,
			name:       "Two",
			department: "IT",
			salary:     2500,
		},
		{
			id:         3,
			name:       "Three",
			department: "IT",
			salary:     2200,
		},
		{
			id:         4,
			name:       "Four",
			department: "HR",
			salary:     1800,
		},
		{
			id:         5,
			name:       "Five",
			department: "HR",
			salary:     2300,
		},
	}
}
