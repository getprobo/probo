// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package searchquery

import "strings"

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) Next() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Kind: EOF, Pos: l.pos}
	}

	start := l.pos
	ch := l.input[l.pos]

	switch ch {
	case ':':
		l.pos++
		return Token{Kind: Colon, Literal: ":", Pos: start}
	case '=':
		l.pos++
		return Token{Kind: Equal, Literal: "=", Pos: start}
	case '"':
		return l.scanString(start)
	default:
		return l.scanIdent(start)
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		switch l.input[l.pos] {
		case ' ', '\t', '\r', '\n':
			l.pos++
		default:
			return
		}
	}
}

func (l *Lexer) scanIdent(start int) Token {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		switch ch {
		case ' ', '\t', '\r', '\n', ':', '=', '"':
			return Token{Kind: Ident, Literal: l.input[start:l.pos], Pos: start}
		default:
			l.pos++
		}
	}

	return Token{Kind: Ident, Literal: l.input[start:l.pos], Pos: start}
}

func (l *Lexer) scanString(start int) Token {
	l.pos++ // skip opening quote

	var b strings.Builder

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		switch ch {
		case '\\':
			if l.pos+1 >= len(l.input) {
				l.pos = len(l.input)
				return Token{Kind: Illegal, Literal: "unterminated quote", Pos: start}
			}

			escaped := l.input[l.pos+1]
			if escaped != '\\' && escaped != '"' {
				return Token{Kind: Illegal, Literal: "invalid escape sequence", Pos: l.pos}
			}

			b.WriteByte(escaped)

			l.pos += 2
		case '"':
			l.pos++
			return Token{Kind: String, Literal: b.String(), Pos: start}
		default:
			b.WriteByte(ch)

			l.pos++
		}
	}

	return Token{Kind: Illegal, Literal: "unterminated quote", Pos: start}
}
