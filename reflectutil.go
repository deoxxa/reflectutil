package reflectutil

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// main entry point

func GetDescription(input interface{}) (*StructDescription, error) {
	d, err := getDescriptionFromType(reflect.TypeOf(input))
	if err != nil {
		return nil, fmt.Errorf("reflectutil.GetDescription: could not get description: %w", err)
	}

	return d, nil
}

func GetDescriptionFromType(typ reflect.Type) (*StructDescription, error) {
	d, err := getDescriptionFromType(typ)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.GetDescriptionFromType: could not get description: %w", err)
	}

	return d, nil
}

// struct

type StructDescription struct {
	name   string
	typ    reflect.Type
	fields FieldList
}

func (s *StructDescription) Name() string       { return s.name }
func (s *StructDescription) Type() reflect.Type { return s.typ }
func (s *StructDescription) Fields() FieldList  { return s.fields }

func (s *StructDescription) Field(name string) *Field { return s.fields.Get(name) }

// field

type Field struct {
	name  string
	index []int
	typ   reflect.Type
	tags  TagList
}

func (f *Field) Name() string       { return f.name }
func (f *Field) Index() []int       { return f.index }
func (f *Field) Type() reflect.Type { return f.typ }
func (f *Field) Tags() TagList      { return f.tags }

func (f *Field) Tag(name string) *Tag { return f.tags.Get(name) }

// field list

type FieldList []Field

func (l FieldList) Names() []string {
	r := make([]string, len(l))
	for i, e := range l {
		r[i] = e.name
	}
	return r
}
func (l FieldList) Get(name string) *Field {
	for _, e := range l {
		if e.name == name {
			return &e
		}
	}

	return nil
}
func (l FieldList) Has(name string) bool {
	for _, e := range l {
		if e.name == name {
			return true
		}
	}

	return false
}

func (l FieldList) WithTag(name string) FieldList {
	r := make(FieldList, 0, len(l))

loop:
	for _, f := range l {
		for _, t := range f.tags {
			if t.name == name {
				r = append(r, f)

				continue loop
			}
		}
	}

	return r
}
func (l FieldList) WithoutTag(name string) FieldList {
	r := make(FieldList, 0, len(l))

loop:
	for _, f := range l {
		for _, t := range f.tags {
			if t.name == name {
				continue loop
			}
		}

		r = append(r, f)
	}

	return r
}

func (l FieldList) WithTagValue(name, value string) FieldList {
	r := make(FieldList, 0, len(l))

loop:
	for _, f := range l {
		for _, t := range f.tags {
			if t.name == name && t.value == value {
				r = append(r, f)

				continue loop
			}
		}
	}

	return r
}
func (l FieldList) WithoutTagValue(name, value string) FieldList {
	r := make(FieldList, 0, len(l))

loop:
	for _, f := range l {
		for _, t := range f.tags {
			if t.name == name && t.value == value {
				continue loop
			}
		}

		r = append(r, f)
	}

	return r
}

// tag

type Tag struct {
	name       string
	value      string
	parameters ParameterList
}

func (t *Tag) Name() string              { return t.name }
func (t *Tag) Value() string             { return t.value }
func (t *Tag) Parameters() ParameterList { return t.parameters }

func (t *Tag) Parameter(name string) *Parameter { return t.parameters.Get(name) }

// tag list

type TagList []Tag

func (l TagList) Names() []string {
	r := make([]string, len(l))
	for i, e := range l {
		r[i] = e.name
	}
	return r
}
func (l TagList) Get(name string) *Tag {
	for _, e := range l {
		if e.name == name {
			return &e
		}
	}

	return nil
}
func (l TagList) Has(name string) bool {
	for _, e := range l {
		if e.name == name {
			return true
		}
	}

	return false
}

func (l TagList) WithName(name string) TagList {
	r := make(TagList, 0, len(l))

	for _, e := range l {
		if e.name == name {
			r = append(r, e)
		}
	}

	return r
}

// parameter

type Parameter struct {
	name  string
	value string
}

func (p *Parameter) Name() string  { return p.name }
func (p *Parameter) Value() string { return p.value }

// parameter list

type ParameterList []Parameter

func (l ParameterList) Names() []string {
	r := make([]string, len(l))
	for i, e := range l {
		r[i] = e.name
	}
	return r
}
func (l ParameterList) Get(name string) *Parameter {
	for _, e := range l {
		if e.name == name {
			return &e
		}
	}

	return nil
}
func (l ParameterList) Has(name string) bool {
	for _, e := range l {
		if e.name == name {
			return true
		}
	}

	return false
}

// struct tag parser

func getDescriptionFromType(typ reflect.Type) (*StructDescription, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("reflectutil.getDescriptionFromType: input should be struct or pointer to struct")
	}

	fields, err := getFields(typ)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.getDescriptionFromType: could not get field descriptions: %w", err)
	}

	return &StructDescription{
		name:   typ.Name(),
		typ:    typ,
		fields: fields,
	}, nil
}

func getFields(typ reflect.Type) (FieldList, error) {
	fields := FieldList{}

	structFields := reflect.VisibleFields(typ)

	for i := range structFields {
		structField := structFields[i]

		tags, err := getTags(structField.Tag)
		if err != nil {
			return nil, fmt.Errorf("reflectutil.getFields: could not get tags for field %s: %w", structField.Name, err)
		}

		fields = append(fields, Field{
			name:  structField.Name,
			index: structField.Index,
			typ:   structField.Type,
			tags:  tags,
		})
	}

	return fields, nil
}

func getTags(structTag reflect.StructTag) (TagList, error) {
	tags := TagList{}

	tagPositionDescriptors, err := parseTags(string(structTag))
	if err != nil {
		return nil, fmt.Errorf("reflectutil.getTags: could not parse struct tags: %w", err)
	}

	for _, rawTag := range tagPositionDescriptors.getNamesAndValues(string(structTag)) {
		unquoted, err := rawTag.unquotedValue()
		if err != nil {
			return nil, fmt.Errorf("reflectutil.getTags: could not unquote value for tag %s: %w", rawTag.name, err)
		}

		tag, err := parseTag(rawTag.name, unquoted)
		if err != nil {
			return nil, fmt.Errorf("reflectutil.getTags: could not parse value for tag %s: %w", rawTag.name, err)
		}

		tags = append(tags, *tag)
	}

	return tags, nil
}

func parseTag(name, tagValue string) (*Tag, error) {
	value, parameters, err := getValueAndParameters(tagValue)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.parseTag: couldn't get value and parameters: %w", err)
	}

	return &Tag{name: name, value: value, parameters: parameters}, nil
}

func getValueAndParameters(tagValue string) (string, ParameterList, error) {
	if tagValue == "" {
		return "", ParameterList{}, nil
	}

	valueAndParameters := strings.SplitN(tagValue, ",", 2)
	if len(valueAndParameters) == 1 {
		return valueAndParameters[0], ParameterList{}, nil
	}

	if valueAndParameters[1] == "" {
		return valueAndParameters[0], ParameterList{}, nil
	}

	parameters, err := getParameters(valueAndParameters[1])
	if err != nil {
		return "", nil, fmt.Errorf("reflectutil.getValueAndParameters: couldn't get parameters: %w", err)
	}

	return valueAndParameters[0], parameters, nil
}

func getParameters(remainingTagValue string) (ParameterList, error) {
	parameters := ParameterList{}

	for _, e := range strings.Split(remainingTagValue, ",") {
		if e == "" {
			continue
		}

		if a := strings.SplitN(e, ":", 2); len(a) == 2 {
			parameters = append(parameters, Parameter{name: a[0], value: a[1]})
		} else {
			parameters = append(parameters, Parameter{name: a[0], value: ""})
		}
	}

	return parameters, nil
}

func validTagNameCharacter(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
}

type parseTagsState int

func (s parseTagsState) String() string {
	switch s {
	case parseTagsStateInitial:
		return "Initial"
	case parseTagsStateReadingName:
		return "ReadingName"
	case parseTagsStateExpectValue:
		return "ExpectValue"
	case parseTagsStateReadingValue:
		return "ReadingValue"
	case parseTagsStateReadingEscapedCharacter:
		return "ReadingEscapedCharacter"
	default:
		return fmt.Sprintf("[UNKNOWN STATE %d]", int(s))
	}
}

const (
	parseTagsStateInitial parseTagsState = iota
	parseTagsStateReadingName
	parseTagsStateExpectValue
	parseTagsStateReadingValue
	parseTagsStateReadingEscapedCharacter
)

type tagNameAndValue struct{ name, value string }

func (t tagNameAndValue) unquotedValue() (string, error) {
	if t.value == "" {
		return "", nil
	}

	s, err := strconv.Unquote(t.value)
	if err != nil {
		return "", fmt.Errorf("reflectutil.tagNameAndValue.unquotedValue: %w", err)
	}

	return s, nil
}

type tagPositionDescriptor struct{ nameStart, nameEnd, colon, valueStart, valueEnd int }

func (t tagPositionDescriptor) getName(s string) string {
	return s[t.nameStart : t.nameEnd+1]
}

func (t tagPositionDescriptor) getValue(s string) string {
	if t.colon == 0 {
		return ""
	}

	return s[t.valueStart : t.valueEnd+1]
}

func (t tagPositionDescriptor) getNameAndValue(s string) tagNameAndValue {
	return tagNameAndValue{t.getName(s), t.getValue(s)}
}

type tagPositionDescriptorList []tagPositionDescriptor

func (l tagPositionDescriptorList) getNamesAndValues(s string) []tagNameAndValue {
	if l == nil {
		return nil
	}

	r := make([]tagNameAndValue, len(l))
	for i := range l {
		r[i] = l[i].getNameAndValue(s)
	}
	return r
}

func parseTags(tag string) (tagPositionDescriptorList, error) {
	positions := make(tagPositionDescriptorList, 0)

	state := parseTagsStateInitial

	var current tagPositionDescriptor

	for i, c := range tag {
	start:
		switch state {
		case parseTagsStateInitial:
			switch {
			case c == ' ':
				continue // skip
			case validTagNameCharacter(c):
				state = parseTagsStateReadingName
				current.nameStart = i
				goto start
			}
		case parseTagsStateReadingName:
			switch {
			case validTagNameCharacter(c):
				continue
			case c == ':':
				current.nameEnd = i - 1
				current.colon = i
				state = parseTagsStateExpectValue
				continue
			case c == ' ':
				current.nameEnd = i - 1
				positions = append(positions, current)
				current = tagPositionDescriptor{}
				state = parseTagsStateInitial
				goto start
			}
		case parseTagsStateExpectValue:
			switch {
			case c == ' ':
				continue // skip
			case c == '"':
				current.valueStart = i
				state = parseTagsStateReadingValue
				continue
			}
		case parseTagsStateReadingValue:
			switch {
			case c == '"':
				current.valueEnd = i
				positions = append(positions, current)
				current = tagPositionDescriptor{}
				state = parseTagsStateInitial
				continue
			case c == '\\':
				state = parseTagsStateReadingEscapedCharacter
				continue
			default:
				continue
			}
		case parseTagsStateReadingEscapedCharacter:
			state = parseTagsStateReadingValue
			continue
		}

		return nil, fmt.Errorf("reflectutil.parseTags: unexpected '%c' at %d in state %s", c, i, state)
	}

	switch state {
	case parseTagsStateInitial:
		// nothing
	case parseTagsStateReadingName:
		current.nameEnd = len(tag) - 1
		positions = append(positions, current)
	default:
		return nil, fmt.Errorf("reflectutil.parseTags: unexpected eof in state %s", state)
	}

	return positions, nil
}
