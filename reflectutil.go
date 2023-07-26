package reflectutil

import (
	"fmt"
	"reflect"
)

// main entry point

func GetDescription(input interface{}) (*StructDescription, error) {
	switch input := input.(type) {
	case reflect.Type:
		d, err := getDescriptionFromReflectType(input)
		if err != nil {
			return nil, fmt.Errorf("reflectutil.GetDescription(%T): could not get description: %w", input, err)
		}

		return d, nil
	default:
		d, err := getDescriptionFromReflectType(reflect.TypeOf(input))
		if err != nil {
			return nil, fmt.Errorf("reflectutil.GetDescription(%T): could not get description: %w", input, err)
		}

		return d, nil
	}
}

// GetDescriptionFromType is deprecated - use GetDescriptionFromReflectType
// instead
func GetDescriptionFromType(typ reflect.Type) (*StructDescription, error) {
	return GetDescriptionFromReflectType(typ)
}

func GetDescriptionFromReflectType(typ reflect.Type) (*StructDescription, error) {
	d, err := getDescriptionFromReflectType(typ)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.GetDescriptionFromReflectType: could not get description: %w", err)
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

// reflect implementation

func getDescriptionFromReflectType(typ reflect.Type) (*StructDescription, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("reflectutil.getDescriptionFromReflectType: input should be struct or pointer to struct")
	}

	fields, err := getFieldsFromReflectType(typ)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.getDescriptionFromReflectType: could not get field descriptions: %w", err)
	}

	return &StructDescription{
		name:   typ.Name(),
		typ:    typ,
		fields: fields,
	}, nil
}

func getFieldsFromReflectType(typ reflect.Type) (FieldList, error) {
	fields := FieldList{}

	structFields := reflect.VisibleFields(typ)

	for i := range structFields {
		structField := structFields[i]

		tags, err := ParseTagList(string(structField.Tag))
		if err != nil {
			return nil, fmt.Errorf("reflectutil.getFieldsFromReflectType: could not get tags for field %s: %w", structField.Name, err)
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
