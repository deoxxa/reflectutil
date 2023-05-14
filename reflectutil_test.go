package reflectutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var getDescriptionTestCases = []struct {
	name   string
	get    func() interface{}
	result *StructDescription
	error  string
}{
	{
		name: "not a struct",
		get: func() interface{} {
			return "x"
		},
		result: nil,
		error:  "input should be struct or pointer to struct",
	},
	{
		name: "struct with a name and no fields",
		get: func() interface{} {
			type S struct{}
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{}},
	},
	{
		name: "struct with no name and no fields",
		get: func() interface{} {
			return struct{}{}
		},
		result: &StructDescription{name: "", fields: FieldList{}},
	},
	{
		name: "struct with one field",
		get: func() interface{} {
			type S struct{ A string }
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{
			{name: "A", typ: reflect.TypeOf(""), tags: []Tag{}},
		}},
	},
	{
		name: "struct with two fields",
		get: func() interface{} {
			type S struct{ A, B string }
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{
			{name: "A", typ: reflect.TypeOf(""), tags: []Tag{}},
			{name: "B", typ: reflect.TypeOf(""), tags: []Tag{}},
		}},
	},
	{
		name: "struct with two fields each having complex struct tags",
		get: func() interface{} {
			type S struct {
				F1 string `t1:"v1,p1,p2k:p2v" t2:",p3,p4k:p4v"`
				F2 string `t1:"v1,p1,p2k:p2v" t2:",p3,p4k:p4v"`
			}
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{
			{name: "F1", typ: reflect.TypeOf(""), tags: []Tag{
				{"t1", "v1", ParameterList{{"p1", ""}, {"p2k", "p2v"}}},
				{"t2", "", ParameterList{{"p3", ""}, {"p4k", "p4v"}}},
			}},
			{name: "F2", typ: reflect.TypeOf(""), tags: []Tag{
				{"t1", "v1", ParameterList{{"p1", ""}, {"p2k", "p2v"}}},
				{"t2", "", ParameterList{{"p3", ""}, {"p4k", "p4v"}}},
			}},
		}},
	},
	{
		name: "sql tags",
		get: func() interface{} {
			type S struct {
				ID   int    `sql:"id,table:t"`
				Name string `sql:"name"`
			}
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{
			{name: "ID", typ: reflect.TypeOf(int(1)), tags: []Tag{
				{"sql", "id", ParameterList{{"table", "t"}}},
			}},
			{name: "Name", typ: reflect.TypeOf(""), tags: []Tag{
				{"sql", "name", ParameterList{}},
			}},
		}},
	},
	{
		name: "json tags",
		get: func() interface{} {
			type S struct {
				ID   int    `json:"id"`
				Name string `json:"name,omitempty"`
			}
			return S{}
		},
		result: &StructDescription{name: "S", fields: FieldList{
			{name: "ID", typ: reflect.TypeOf(int(1)), tags: []Tag{
				{"json", "id", ParameterList{}},
			}},
			{name: "Name", typ: reflect.TypeOf(""), tags: []Tag{
				{"json", "name", ParameterList{{"omitempty", ""}}},
			}},
		}},
	},
}

func TestGetDescription(t *testing.T) {
	for _, tc := range getDescriptionTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			d, err := GetDescription(tc.get())

			if tc.result == nil {
				a.Nil(d)
			} else {
				if a.NotNil(d) {
					a.Equal(tc.result.name, d.Name())
					a.Equal(tc.result.fields, d.Fields())
				}
			}

			if tc.error != "" {
				a.ErrorContains(err, tc.error)
			} else {
				a.NoError(err)
			}
		})
	}
}

func BenchmarkGetDescription(b *testing.B) {
	for _, tc := range getDescriptionTestCases {
		b.Run(tc.name, func(b *testing.B) {
			v := tc.get()
			for i := 0; i < b.N; i++ {
				GetDescription(v)
			}
		})
	}
}

func TestGetDescriptionFromType(t *testing.T) {
	for _, tc := range getDescriptionTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			d, err := GetDescriptionFromType(reflect.TypeOf(tc.get()))

			if tc.result == nil {
				a.Nil(d)
			} else {
				if a.NotNil(d) {
					a.Equal(tc.result.name, d.Name())
					a.Equal(tc.result.fields, d.Fields())
				}
			}

			if tc.error != "" {
				a.ErrorContains(err, tc.error)
			} else {
				a.NoError(err)
			}
		})
	}
}

func BenchmarkGetDescriptionFromType(b *testing.B) {
	for _, tc := range getDescriptionTestCases {
		b.Run(tc.name, func(b *testing.B) {
			v := reflect.TypeOf(tc.get())
			for i := 0; i < b.N; i++ {
				GetDescriptionFromType(v)
			}
		})
	}
}

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

			tag, err := parseTag("x", tc.input)

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
				parseTag("x", tc.input)
			}
		})
	}
}

var parseTagsTestCases = []struct {
	name      string
	input     string
	positions tagPositionDescriptorList
	tags      []tagNameAndValue
	error     string
}{
	{
		name:      "no keys",
		input:     ``,
		positions: tagPositionDescriptorList{},
		tags:      []tagNameAndValue{},
		error:     "",
	},
	{
		name:  "single key",
		input: `k1`,
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
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
		positions: tagPositionDescriptorList{
			{0, 0, 1, 2, 10},
		},
		tags: []tagNameAndValue{
			{"k", `"a\r\n\t"`},
		},
		error: "",
	},
	{
		name:      "missing value",
		input:     `k:`,
		positions: nil,
		tags:      nil,
		error:     `reflectutil.parseTags: unexpected eof in state ExpectValue`,
	},
	{
		name:      "missing closing quote",
		input:     `k:"x`,
		positions: nil,
		tags:      nil,
		error:     `reflectutil.parseTags: unexpected eof in state ReadingValue`,
	},
	{
		name:      "invalid key",
		input:     `a$`,
		positions: nil,
		tags:      nil,
		error:     `reflectutil.parseTags: unexpected '$' at 1 in state ReadingName`,
	},
}

func TestParseTags(t *testing.T) {
	for _, tc := range parseTagsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)

			positions, err := parseTags(tc.input)
			namesAndValues := positions.getNamesAndValues(tc.input)

			if tc.error != "" {
				a.ErrorContains(err, tc.error)
			} else {
				a.NoError(err)
			}

			a.Equal(tc.positions, positions)
			a.Equal(tc.tags, namesAndValues)
		})
	}
}

func BenchmarkParseTags(b *testing.B) {
	for _, tc := range parseTagsTestCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				parseTags(tc.input)
			}
		})
	}
}

func BenchmarkGetNamesAndValues(b *testing.B) {
	for _, tc := range parseTagsTestCases {
		if tc.error != "" {
			continue
		}

		b.Run(tc.name, func(b *testing.B) {
			positions, _ := parseTags(tc.input)

			for i := 0; i < b.N; i++ {
				positions.getNamesAndValues(tc.input)
			}
		})
	}
}

func BenchmarkParseTagsAndGetNamesAndValues(b *testing.B) {
	for _, tc := range parseTagsTestCases {
		if tc.error != "" {
			continue
		}

		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				positions, _ := parseTags(tc.input)
				positions.getNamesAndValues(tc.input)
			}
		})
	}
}

func TestAccessors(t *testing.T) {
	type S struct {
		Populated string `sql:"populated,table:t" json:"populated,omitempty"`
		SQLEmpty  string `sql:"" json:"sqlEmpty"`
		SQLDash   string `sql:"-" json:"sqlDash"`
		JSONEmpty string `sql:"json_empty" json:""`
		JSONDash  string `sql:"json_dash" json:"-"`
		Repeated  string `z:"x,x:1,x:2" z:"y,y:1,y:2"`
	}

	get := func(t *testing.T) (*assert.Assertions, *StructDescription) {
		a := assert.New(t)

		d, err := GetDescription(S{})

		if !a.NoError(err) {
			t.FailNow()
			return nil, nil
		}

		if !a.NotNil(d) {
			t.FailNow()
			return nil, nil
		}

		return a, d
	}

	// struct description

	t.Run("StructDescription.Name", func(t *testing.T) {
		a, d := get(t)
		a.Equal("S", d.Name())
	})

	t.Run("StructDescription.Type", func(t *testing.T) {
		a, d := get(t)
		a.Equal(reflect.TypeOf(S{}), d.Type())
	})

	t.Run("StructDescription.Fields", func(t *testing.T) {
		a, d := get(t)

		a.Equal(
			[]string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash", "Repeated"},
			d.Fields().Names(),
		)
	})

	t.Run("StructDescription.Field", func(t *testing.T) {
		a, d := get(t)

		a.Equal(&Field{"Populated", reflect.TypeOf(""), []Tag{
			{"sql", "populated", ParameterList{{"table", "t"}}},
			{"json", "populated", ParameterList{{"omitempty", ""}}},
		}}, d.Field("Populated"))
	})

	// field list

	for _, tc := range []struct {
		tag           string
		with, without []string
	}{
		{"z", []string{"Repeated"}, []string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash"}},
	} {
		t.Run("FieldList.WithTag "+tc.tag, func(t *testing.T) {
			a, d := get(t)
			a.Equal(tc.with, d.Fields().WithTag(tc.tag).Names())
		})

		t.Run("FieldList.WithoutTag "+tc.tag, func(t *testing.T) {
			a, d := get(t)
			a.Equal(tc.without, d.Fields().WithoutTag(tc.tag).Names())
		})
	}

	for _, tc := range []struct {
		tag, value    string
		with, without []string
	}{
		{"sql", "populated", []string{"Populated"}, []string{"SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash", "Repeated"}},
		{"z", "x", []string{"Repeated"}, []string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash"}},
		{"z", "y", []string{"Repeated"}, []string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash"}},
		{"z", "z", []string{}, []string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "JSONDash", "Repeated"}},
		{"sql", "-", []string{"SQLDash"}, []string{"Populated", "SQLEmpty", "JSONEmpty", "JSONDash", "Repeated"}},
		{"json", "-", []string{"JSONDash"}, []string{"Populated", "SQLEmpty", "SQLDash", "JSONEmpty", "Repeated"}},
	} {
		t.Run("FieldList.WithTagValue "+tc.tag+"="+tc.value, func(t *testing.T) {
			a, d := get(t)
			a.Equal(tc.with, d.Fields().WithTagValue(tc.tag, tc.value).Names())
		})

		t.Run("FieldList.WithoutTagValue "+tc.tag+"="+tc.value, func(t *testing.T) {
			a, d := get(t)
			a.Equal(tc.without, d.Fields().WithoutTagValue(tc.tag, tc.value).Names())
		})
	}

	// field

	for _, tc := range []struct {
		field, tag string
		result     *Tag
	}{
		{"Populated", "sql", &Tag{"sql", "populated", ParameterList{{"table", "t"}}}},
		{"Populated", "json", &Tag{"json", "populated", ParameterList{{"omitempty", ""}}}},
		{"SQLEmpty", "sql", &Tag{"sql", "", ParameterList{}}},
		{"SQLEmpty", "json", &Tag{"json", "sqlEmpty", ParameterList{}}},
		{"SQLDash", "sql", &Tag{"sql", "-", ParameterList{}}},
		{"SQLDash", "json", &Tag{"json", "sqlDash", ParameterList{}}},
		{"JSONEmpty", "json", &Tag{"json", "", ParameterList{}}},
		{"JSONEmpty", "sql", &Tag{"sql", "json_empty", ParameterList{}}},
		{"JSONDash", "json", &Tag{"json", "-", ParameterList{}}},
		{"JSONDash", "sql", &Tag{"sql", "json_dash", ParameterList{}}},
		{"Repeated", "z", &Tag{"z", "x", ParameterList{{"x", "1"}, {"x", "2"}}}},
	} {
		t.Run("Field.Tag "+tc.field+"/"+tc.tag, func(t *testing.T) {
			a, d := get(t)
			a.Equal(tc.result, d.Field(tc.field).Tag(tc.tag))
		})
	}

	// tag list

	t.Run("Tags.WithName", func(t *testing.T) {
		a, d := get(t)

		a.Equal(TagList{
			{"z", "x", ParameterList{{"x", "1"}, {"x", "2"}}},
			{"z", "y", ParameterList{{"y", "1"}, {"y", "2"}}},
		}, d.Field("Repeated").Tags().WithName("z"))
	})
}

func BenchmarkAccessors(b *testing.B) {
	type S struct {
		Populated string `sql:"populated,table:t" json:"populated,omitempty"`
		SQLEmpty  string `sql:"" json:"sqlEmpty"`
		SQLDash   string `sql:"-" json:"sqlDash"`
		JSONEmpty string `sql:"json_empty" json:""`
		JSONDash  string `sql:"json_dash" json:"-"`
		Repeated  string `z:"x,x:1,x:2" z:"y,y:1,y:2"`
	}

	get := func() *StructDescription {
		d, _ := GetDescription(S{})
		return d
	}

	// struct description

	b.Run("StructDescription.Name", func(b *testing.B) {
		d := get()

		for i := 0; i < b.N; i++ {
			d.Name()
		}
	})

	b.Run("StructDescription.Type", func(b *testing.B) {
		d := get()

		for i := 0; i < b.N; i++ {
			d.Type()
		}
	})

	b.Run("StructDescription.Fields", func(b *testing.B) {
		d := get()

		for i := 0; i < b.N; i++ {
			d.Fields()
		}
	})

	b.Run("StructDescription.Field", func(b *testing.B) {
		d := get()

		for i := 0; i < b.N; i++ {
			d.Field("Populated")
		}
	})

	// field list

	for _, tc := range []struct{ tag string }{
		{"z"},
		{"sql"},
		{"json"},
	} {
		b.Run("FieldList.WithTag "+tc.tag, func(b *testing.B) {
			a := get().Fields()

			for i := 0; i < b.N; i++ {
				a.WithTag(tc.tag)
			}
		})

		b.Run("FieldList.WithoutTag "+tc.tag, func(b *testing.B) {
			a := get().Fields()

			for i := 0; i < b.N; i++ {
				a.WithoutTag(tc.tag)
			}
		})
	}

	for _, tc := range []struct{ tag, value string }{
		{"sql", "populated"},
		{"z", "x"},
		{"z", "y"},
		{"z", "z"},
		{"sql", "-"},
		{"json", "-"},
	} {
		b.Run("FieldList.WithTagValue "+tc.tag+"="+tc.value, func(b *testing.B) {
			a := get().Fields()

			for i := 0; i < b.N; i++ {
				a.WithTagValue(tc.tag, tc.value)
			}
		})

		b.Run("FieldList.WithoutTagValue "+tc.tag+"="+tc.value, func(b *testing.B) {
			a := get().Fields()

			for i := 0; i < b.N; i++ {
				a.WithoutTagValue(tc.tag, tc.value)
			}
		})
	}

	// field

	for _, tc := range []struct{ field, tag string }{
		{"Populated", "sql"},
		{"Populated", "json"},
		{"SQLEmpty", "sql"},
		{"SQLEmpty", "json"},
		{"SQLDash", "sql"},
		{"SQLDash", "json"},
		{"JSONEmpty", "json"},
		{"JSONEmpty", "sql"},
		{"JSONDash", "json"},
		{"JSONDash", "sql"},
		{"Repeated", "z"},
	} {
		b.Run("Field.Tag "+tc.field+"/"+tc.tag, func(b *testing.B) {
			f := get().Field(tc.field)

			for i := 0; i < b.N; i++ {
				f.Tag(tc.tag)
			}
		})
	}

	// tag list

	b.Run("Tags.WithName", func(b *testing.B) {
		a := get().Field("Repeated").Tags()

		for i := 0; i < b.N; i++ {
			a.WithName("z")
		}
	})
}
