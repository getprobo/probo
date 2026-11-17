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

import "fmt"

type Error struct {
	Pos     int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

type parser struct {
	lexer *Lexer
	tok   Token
}

func Parse(raw string) (*Query, error) {
	p := &parser{lexer: NewLexer(raw)}
	p.next()

	query := &Query{Clauses: make([]Clause, 0)}

	for p.tok.Kind != EOF {
		clause, err := p.parseClause()
		if err != nil {
			return nil, err
		}

		query.Clauses = append(query.Clauses, clause)
	}

	return query, nil
}

func (p *parser) next() {
	p.tok = p.lexer.Next()
}

func (p *parser) parseClause() (Clause, error) {
	if p.tok.Kind == Illegal {
		return nil, &Error{Pos: p.tok.Pos, Message: p.tok.Literal}
	}

	if p.tok.Kind == String {
		literal := p.tok.Literal
		p.next()

		return TextClause{Value: literal}, nil
	}

	if p.tok.Kind == Ident {
		qualifier := p.tok.Literal
		qualifierPos := p.tok.Pos
		p.next()

		if p.tok.Kind != Colon {
			return TextClause{Value: qualifier}, nil
		}

		p.next()

		switch qualifier {
		case "tag":
			return p.parseTagClause()
		case "type":
			return p.parseTypeClause()
		default:
			return nil, &Error{
				Pos:     qualifierPos,
				Message: fmt.Sprintf("unknown qualifier %q", qualifier),
			}
		}
	}

	return nil, &Error{Pos: p.tok.Pos, Message: "unexpected token"}
}

func (p *parser) parseTagClause() (Clause, error) {
	name, err := p.parseTerm("tag name is required")
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, &Error{Pos: p.tok.Pos, Message: "tag name is required"}
	}

	return TagClause{Name: name}, nil
}

func (p *parser) parseTypeClause() (Clause, error) {
	name, err := p.parseTerm("type name is required")
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, &Error{Pos: p.tok.Pos, Message: "type name is required"}
	}

	return TypeClause{Name: name}, nil
}

func (p *parser) parseTerm(emptyMessage string) (string, error) {
	if p.tok.Kind == Illegal {
		return "", &Error{Pos: p.tok.Pos, Message: p.tok.Literal}
	}

	switch p.tok.Kind {
	case Ident, String:
		literal := p.tok.Literal
		p.next()

		return literal, nil
	default:
		return "", &Error{Pos: p.tok.Pos, Message: emptyMessage}
	}
}
