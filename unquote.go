package shellquote

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	UnterminatedSingleQuoteError = errors.New("Unterminated single-quoted string")
	UnterminatedDoubleQuoteError = errors.New("Unterminated double-quoted string")
	UnterminatedEscapeError      = errors.New("Unterminated backslash-escape")
)

var (
	splitChars        = " \n\t"
	singleChar        = '\''
	doubleChar        = '"'
	escapeChar        = '\\'
	doubleEscapeChars = "$`\"\n\\"
)

// Token is the same as Split, but just split one token.
// User can provide a buffer used internally by Token to reuse buffer and avoid
// allocation.
func Token(input string, buf []byte) (token string, unparsed string, reused []byte, err error) {
	for input != "" {
		r, size := utf8.DecodeRuneInString(input)
		if strings.ContainsRune(splitChars, r) {
			input = input[size:]
			continue
		}
		if r == escapeChar {
			lookahead := input[size:]
			if len(lookahead) == 0 {
				return "", input, buf, UnterminatedEscapeError
			}
			r1, size1 := utf8.DecodeRuneInString(lookahead)
			if r1 == '\n' {
				input = lookahead[size1:]
				continue
			}
		}
		w := byteWriter(buf)
		token, unparsed, err := splitWord(input, &w)
		return token, unparsed, w.Bytes(), err
	}
	return "", "", buf, nil
}

// Split splits a string according to /bin/sh's word-splitting rules. It
// supports backslash-escapes, single-quotes, and double-quotes. Notably it does
// not support the $'' style of quoting. It also doesn't attempt to perform any
// other sort of expansion, including brace expansion, shell expansion, or
// pathname expansion.
//
// If the given input has an unterminated quoted string or ends in a
// backslash-escape, one of UnterminatedSingleQuoteError,
// UnterminatedDoubleQuoteError, or UnterminatedEscapeError is returned.
func Split(input string) (words []string, err error) {
	var buf byteWriter

	for len(input) > 0 {
		// skip any splitChars at the start
		c, l := utf8.DecodeRuneInString(input)
		if strings.ContainsRune(splitChars, c) {
			input = input[l:]
			continue
		} else if c == escapeChar {
			// Look ahead for escaped newline so we can skip over it
			next := input[l:]
			if len(next) == 0 {
				err = UnterminatedEscapeError
				return
			}
			c2, l2 := utf8.DecodeRuneInString(next)
			if c2 == '\n' {
				input = next[l2:]
				continue
			}
		}

		var word string
		word, input, err = splitWord(input, &buf)
		if err != nil {
			return
		}
		words = append(words, word)
	}
	return
}

type byteWriter []byte

func (w *byteWriter) Reset() {
	*w = (*w)[:0]
}

func (w *byteWriter) Write(p []byte) (int, error) {
	*w = append(*w, p...)
	return len(p), nil
}

func (w *byteWriter) WriteString(s string) (int, error) {
	*w = append(*w, s...)
	return len(s), nil
}

func (w *byteWriter) WriteRune(r rune) (int, error) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	*w = append(*w, buf[:n]...)
	return n, nil
}

func (w *byteWriter) String() string {
	return string(*w)
}

func (w *byteWriter) Bytes() []byte {
	return []byte(*w)
}

func splitWord(input string, buf *byteWriter) (word string, remainder string, err error) {
	buf.Reset()

raw:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == singleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto single
			} else if c == doubleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto double
			} else if c == escapeChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto escape
			} else if strings.ContainsRune(splitChars, c) {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				return buf.String(), cur, nil
			}
		}
		if len(input) > 0 {
			buf.WriteString(input)
			input = ""
		}
		goto done
	}

escape:
	{
		if len(input) == 0 {
			return "", "", UnterminatedEscapeError
		}
		c, l := utf8.DecodeRuneInString(input)
		if c == '\n' {
			// a backslash-escaped newline is elided from the output entirely
		} else {
			buf.WriteString(input[:l])
		}
		input = input[l:]
	}
	goto raw

single:
	{
		i := strings.IndexRune(input, singleChar)
		if i == -1 {
			return "", "", UnterminatedSingleQuoteError
		}
		buf.WriteString(input[0:i])
		input = input[i+1:]
		goto raw
	}

double:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			if c == doubleChar {
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto raw
			} else if c == escapeChar {
				// bash only supports certain escapes in double-quoted strings
				c2, l2 := utf8.DecodeRuneInString(cur)
				cur = cur[l2:]
				if strings.ContainsRune(doubleEscapeChars, c2) {
					buf.WriteString(input[0 : len(input)-len(cur)-l-l2])
					if c2 == '\n' {
						// newline is special, skip the backslash entirely
					} else {
						buf.WriteRune(c2)
					}
					input = cur
				}
			}
		}
		return "", "", UnterminatedDoubleQuoteError
	}

done:
	return buf.String(), input, nil
}
