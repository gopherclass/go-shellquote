package shellquote

import (
	"reflect"
	"testing"
)

func TestSimpleSplit(t *testing.T) {
	for _, elem := range simpleSplitTest {
		output, err := Split(elem.input)
		if err != nil {
			t.Errorf("Input %q, got error %#v", elem.input, err)
		} else if !reflect.DeepEqual(output, elem.output) {
			t.Errorf("Input %q, got %q, expected %q", elem.input, output, elem.output)
		}
	}
}

func TestErrorSplit(t *testing.T) {
	for _, elem := range errorSplitTest {
		_, err := Split(elem.input)
		if err != elem.error {
			t.Errorf("Input %q, got error %#v, expected error %#v", elem.input, err, elem.error)
		}
	}
}

func TestToken(t *testing.T) {
	buf := []byte("bytes is fiiled with garbage contents")
	for _, tc := range simpleSplitTest {
		input := tc.input
		expects := tc.output
		for i := 1; input != "" && len(expects) > 0; i++ {
			var token string
			var err error
			token, input, buf, err = Token(input, buf)
			if err != nil {
				t.Errorf("input %q, got error %#v", tc.input, err)
			}
			if token != expects[0] {
				t.Errorf("in iteration %d, input %q, got %q, expected %q",
					i, tc.input, token, expects[0])
			}
			expects = expects[1:]
		}
		if input != "" && len(expects) == 0 {
			t.Errorf("input %q is not consumed fully", tc.input)
		}
		if input == "" && len(expects) > 0 {
			t.Errorf("input %q is consumed fully, but %q expected",
				tc.input, expects)
		}
	}
}

var simpleSplitTest = []struct {
	input  string
	output []string
}{
	{"hello", []string{"hello"}},
	{"hello goodbye", []string{"hello", "goodbye"}},
	{"hello   goodbye", []string{"hello", "goodbye"}},
	{"glob* test?", []string{"glob*", "test?"}},
	{"don\\'t you know the dewey decimal system\\?", []string{"don't", "you", "know", "the", "dewey", "decimal", "system?"}},
	{"'don'\\''t you know the dewey decimal system?'", []string{"don't you know the dewey decimal system?"}},
	{"one '' two", []string{"one", "", "two"}},
	{"text with\\\na backslash-escaped newline", []string{"text", "witha", "backslash-escaped", "newline"}},
	{"text \"with\na\" quoted newline", []string{"text", "with\na", "quoted", "newline"}},
	{"\"quoted\\d\\\\\\\" text with\\\na backslash-escaped newline\"", []string{"quoted\\d\\\" text witha backslash-escaped newline"}},
	{"text with an escaped \\\n newline in the middle", []string{"text", "with", "an", "escaped", "newline", "in", "the", "middle"}},
	{"foo\"bar\"baz", []string{"foobarbaz"}},
}

var errorSplitTest = []struct {
	input string
	error error
}{
	{"don't worry", UnterminatedSingleQuoteError},
	{"'test'\\''ing", UnterminatedSingleQuoteError},
	{"\"foo'bar", UnterminatedDoubleQuoteError},
	{"foo\\", UnterminatedEscapeError},
	{"   \\", UnterminatedEscapeError},
}
