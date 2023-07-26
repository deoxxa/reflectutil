package reflectutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseTagTestCase struct {
	name  string
	input string
	tag   *Tag
	error string
}

var parseTagTestCases []parseTagTestCase

func init() {
	for _, value := range []struct {
		description, value string
		result             string
	}{
		{"no value", "", ""},
		{"value", "foo", "foo"},
	} {
		for _, parameters := range []struct {
			description, value string
			result             ParameterList
		}{
			{"no parameters", "", ParameterList{}},
			{"one parameter with key and no value", "p", ParameterList{{"p", ""}}},
			{"one parameter with key and value", "p:v", ParameterList{{"p", "v"}}},
			{"two parameters with different keys and no values", "p1,p2", ParameterList{{"p1", ""}, {"p2", ""}}},
			{"two parameters with different keys and the same values", "p1:v,p2:v", ParameterList{{"p1", "v"}, {"p2", "v"}}},
			{"two parameters with different keys and different values", "p1:v1,p2:v2", ParameterList{{"p1", "v1"}, {"p2", "v2"}}},
			{"two parameters with the same key and no value", "p,p", ParameterList{{"p", ""}, {"p", ""}}},
			{"two parameters with the same key and the same values", "p:v,p:v", ParameterList{{"p", "v"}, {"p", "v"}}},
			{"two parameters with the same key and no value", "p:v1,p:v2", ParameterList{{"p", "v1"}, {"p", "v2"}}},
		} {
			input := value.value
			if parameters.value != "" {
				input = input + "," + parameters.value
			}

			parseTagTestCases = append(parseTagTestCases, parseTagTestCase{
				name:  value.description + " with " + parameters.description,
				input: input,
				tag: &Tag{
					name:       "x",
					value:      value.result,
					parameters: parameters.result,
				},
			})
		}
	}
}

func TestParseTag(t *testing.T) {
	for _, tc := range parseTagTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			tag, err := ParseTag("x", tc.input)

			if tc.error != "" {
				a.ErrorContains(err, tc.error)
			} else {
				a.NoError(err)
			}

			a.Equal(tc.tag, tag)
		})
	}
}

func BenchmarkParseTag(b *testing.B) {
	for _, tc := range parseTagTestCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseTag("x", tc.input)
			}
		})
	}
}

var parseTagListTestCases = []struct {
	name         string
	input        string
	positionList tagPositionList
	tags         []tagNameAndValue
	error        string
}{
	{
		name:         "no keys",
		input:        ``,
		positionList: tagPositionList{},
		tags:         []tagNameAndValue{},
		error:        "",
	},
	{
		name:  "single key",
		input: `k1`,
		positionList: tagPositionList{
			{0, 1, 0, 0, 0},
		},
		tags: []tagNameAndValue{
			{"k1", ""},
		},
		error: "",
	},
	{
		name:  "multiple keys",
		input: `k1 k2 k3`,
		positionList: tagPositionList{
			{0, 1, 0, 0, 0},
			{3, 4, 0, 0, 0},
			{6, 7, 0, 0, 0},
		},
		tags: []tagNameAndValue{
			{"k1", ""},
			{"k2", ""},
			{"k3", ""},
		},
		error: "",
	},
	{
		name:  "multiple keys with extra spaces",
		input: `k1    k2    k3`,
		positionList: tagPositionList{
			{0, 1, 0, 0, 0},
			{6, 7, 0, 0, 0},
			{12, 13, 0, 0, 0},
		},
		tags: []tagNameAndValue{
			{"k1", ""},
			{"k2", ""},
			{"k3", ""},
		},
		error: "",
	},
	{
		name:  "key with value",
		input: `k1 k2:"v2"`,
		positionList: tagPositionList{
			{0, 1, 0, 0, 0},
			{3, 4, 5, 6, 9},
		},
		tags: []tagNameAndValue{
			{"k1", ""},
			{"k2", `"v2"`},
		},
		error: "",
	},
	{
		name:  "key with multi-word value",
		input: `k1 k2:"v2" k3:"v3a v3b"`,
		positionList: tagPositionList{
			{0, 1, 0, 0, 0},
			{3, 4, 5, 6, 9},
			{11, 12, 13, 14, 22},
		},
		tags: []tagNameAndValue{
			{"k1", ""},
			{"k2", `"v2"`},
			{"k3", `"v3a v3b"`},
		},
		error: "",
	},
	{
		name:  "key with multiple instances",
		input: `k1:"v1a" k1:"v1b"`,
		positionList: tagPositionList{
			{0, 1, 2, 3, 7},
			{9, 10, 11, 12, 16},
		},
		tags: []tagNameAndValue{
			{"k1", `"v1a"`},
			{"k1", `"v1b"`},
		},
		error: "",
	},
	{
		name:  "multiple keys with multiple instances",
		input: `k1:"v1a" k2:"v2a" k1:"v1b" k2:"v2b"`,
		positionList: tagPositionList{
			{0, 1, 2, 3, 7},
			{9, 10, 11, 12, 16},
			{18, 19, 20, 21, 25},
			{27, 28, 29, 30, 34},
		},
		tags: []tagNameAndValue{
			{"k1", `"v1a"`},
			{"k2", `"v2a"`},
			{"k1", `"v1b"`},
			{"k2", `"v2b"`},
		},
		error: "",
	},
	{
		name:  "one escaped value with multiple unescaped values",
		input: `k1:"v1\n" k2:"v2" k3:"v3" k4:"v4"`,
		positionList: tagPositionList{
			{0, 1, 2, 3, 8},
			{10, 11, 12, 13, 16},
			{18, 19, 20, 21, 24},
			{26, 27, 28, 29, 32},
		},
		tags: []tagNameAndValue{
			{"k1", `"v1\n"`},
			{"k2", `"v2"`},
			{"k3", `"v3"`},
			{"k4", `"v4"`},
		},
		error: "",
	},
	{
		name:  "key with escaped slash",
		input: `k:"a\\b"`,
		positionList: tagPositionList{
			{0, 0, 1, 2, 7},
		},
		tags: []tagNameAndValue{
			{"k", `"a\\b"`},
		},
		error: "",
	},
	{
		name:  "key with escaped quote",
		input: `k:"a\"b"`,
		positionList: tagPositionList{
			{0, 0, 1, 2, 7},
		},
		tags: []tagNameAndValue{
			{"k", `"a\"b"`},
		},
		error: "",
	},
	{
		name:  "key with escaped control characters",
		input: `k:"a\r\n\t"`,
		positionList: tagPositionList{
			{0, 0, 1, 2, 10},
		},
		tags: []tagNameAndValue{
			{"k", `"a\r\n\t"`},
		},
		error: "",
	},
	{
		name:         "missing value",
		input:        `k:`,
		positionList: nil,
		tags:         nil,
		error:        `reflectutil.parseTagPositionList: unexpected eof in state ExpectValue`,
	},
	{
		name:         "missing closing quote",
		input:        `k:"x`,
		positionList: nil,
		tags:         nil,
		error:        `reflectutil.parseTagPositionList: unexpected eof in state ReadingValue`,
	},
	{
		name:         "invalid key",
		input:        `a$`,
		positionList: nil,
		tags:         nil,
		error:        `reflectutil.parseTagPositionList: unexpected '$' at 1 in state ReadingName`,
	},
}

func TestParseTagList(t *testing.T) {
	for _, tc := range parseTagListTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			positionList, err := parseTagPositionList(tc.input)
			namesAndValues := positionList.getNamesAndValues(tc.input)

			if tc.error != "" {
				a.ErrorContains(err, tc.error)
			} else {
				a.NoError(err)
			}

			a.Equal(tc.positionList, positionList)
			a.Equal(tc.tags, namesAndValues)
		})
	}
}

func BenchmarkParseTagList(b *testing.B) {
	for _, tc := range parseTagListTestCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ParseTagList(tc.input)
			}
		})
	}
}

func BenchmarkGetNamesAndValues(b *testing.B) {
	for _, tc := range parseTagListTestCases {
		if tc.error != "" {
			continue
		}

		b.Run(tc.name, func(b *testing.B) {
			positionList, _ := parseTagPositionList(tc.input)

			for i := 0; i < b.N; i++ {
				positionList.getNamesAndValues(tc.input)
			}
		})
	}
}

func BenchmarkParseTagListAndGetNamesAndValues(b *testing.B) {
	for _, tc := range parseTagListTestCases {
		if tc.error != "" {
			continue
		}

		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				positionList, _ := parseTagPositionList(tc.input)
				positionList.getNamesAndValues(tc.input)
			}
		})
	}
}
