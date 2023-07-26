package reflectutil

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseTagList(input string) (TagList, error) {
	tags := TagList{}

	tagPositions, err := parseTagPositionList(input)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.ParseTagList: could not parse struct tags: %w", err)
	}

	for _, rawTag := range tagPositions.getNamesAndValues(input) {
		unquoted, err := rawTag.unquotedValue()
		if err != nil {
			return nil, fmt.Errorf("reflectutil.ParseTagList: could not unquote value for tag %s: %w", rawTag.name, err)
		}

		tag, err := ParseTag(rawTag.name, unquoted)
		if err != nil {
			return nil, fmt.Errorf("reflectutil.ParseTagList: could not parse value for tag %s: %w", rawTag.name, err)
		}

		tags = append(tags, *tag)
	}

	return tags, nil
}

func ParseTag(name, tagValue string) (*Tag, error) {
	value, parameters, err := parseValueAndParameterList(tagValue)
	if err != nil {
		return nil, fmt.Errorf("reflectutil.ParseTag: couldn't get value and parameters: %w", err)
	}

	return &Tag{name: name, value: value, parameters: parameters}, nil
}

func parseValueAndParameterList(tagValue string) (string, ParameterList, error) {
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

	parameters, err := parseParameterList(valueAndParameters[1])
	if err != nil {
		return "", nil, fmt.Errorf("reflectutil.parseValueAndParameterList: couldn't get parameters: %w", err)
	}

	return valueAndParameters[0], parameters, nil
}

func parseParameterList(input string) (ParameterList, error) {
	parameters := ParameterList{}

	for _, e := range strings.Split(input, ",") {
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

type parseTagPositionListState int

func (s parseTagPositionListState) String() string {
	switch s {
	case parseTagPositionListStateInitial:
		return "Initial"
	case parseTagPositionListStateReadingName:
		return "ReadingName"
	case parseTagPositionListStateExpectValue:
		return "ExpectValue"
	case parseTagPositionListStateReadingValue:
		return "ReadingValue"
	case parseTagPositionListStateReadingEscapedCharacter:
		return "ReadingEscapedCharacter"
	default:
		return fmt.Sprintf("[UNKNOWN STATE %d]", int(s))
	}
}

const (
	parseTagPositionListStateInitial parseTagPositionListState = iota
	parseTagPositionListStateReadingName
	parseTagPositionListStateExpectValue
	parseTagPositionListStateReadingValue
	parseTagPositionListStateReadingEscapedCharacter
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

type tagPosition struct{ nameStart, nameEnd, colon, valueStart, valueEnd int }

func (t tagPosition) getName(s string) string {
	return s[t.nameStart : t.nameEnd+1]
}

func (t tagPosition) getValue(s string) string {
	if t.colon == 0 {
		return ""
	}

	return s[t.valueStart : t.valueEnd+1]
}

func (t tagPosition) getNameAndValue(s string) tagNameAndValue {
	return tagNameAndValue{
		name:  t.getName(s),
		value: t.getValue(s),
	}
}

type tagPositionList []tagPosition

func (l tagPositionList) getNamesAndValues(s string) []tagNameAndValue {
	if l == nil {
		return nil
	}

	r := make([]tagNameAndValue, len(l))
	for i := range l {
		r[i] = l[i].getNameAndValue(s)
	}
	return r
}

func parseTagPositionList(tag string) (tagPositionList, error) {
	positions := make(tagPositionList, 0)

	state := parseTagPositionListStateInitial

	var current tagPosition

	for i, c := range tag {
	start:
		switch state {
		case parseTagPositionListStateInitial:
			switch {
			case c == ' ':
				continue // skip
			case validTagNameCharacter(c):
				state = parseTagPositionListStateReadingName
				current.nameStart = i
				goto start
			}
		case parseTagPositionListStateReadingName:
			switch {
			case validTagNameCharacter(c):
				continue
			case c == ':':
				current.nameEnd = i - 1
				current.colon = i
				state = parseTagPositionListStateExpectValue
				continue
			case c == ' ':
				current.nameEnd = i - 1
				positions = append(positions, current)
				current = tagPosition{}
				state = parseTagPositionListStateInitial
				goto start
			}
		case parseTagPositionListStateExpectValue:
			switch {
			case c == ' ':
				continue // skip
			case c == '"':
				current.valueStart = i
				state = parseTagPositionListStateReadingValue
				continue
			}
		case parseTagPositionListStateReadingValue:
			switch {
			case c == '"':
				current.valueEnd = i
				positions = append(positions, current)
				current = tagPosition{}
				state = parseTagPositionListStateInitial
				continue
			case c == '\\':
				state = parseTagPositionListStateReadingEscapedCharacter
				continue
			default:
				continue
			}
		case parseTagPositionListStateReadingEscapedCharacter:
			state = parseTagPositionListStateReadingValue
			continue
		}

		return nil, fmt.Errorf("reflectutil.parseTagPositionList: unexpected '%c' at %d in state %s", c, i, state)
	}

	switch state {
	case parseTagPositionListStateInitial:
		// nothing
	case parseTagPositionListStateReadingName:
		current.nameEnd = len(tag) - 1
		positions = append(positions, current)
	default:
		return nil, fmt.Errorf("reflectutil.parseTagPositionList: unexpected eof in state %s", state)
	}

	return positions, nil
}
