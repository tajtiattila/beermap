package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type Cond interface {
	Accept(p Pub) bool
}

func decodeCond(i interface{}) (Cond, error) {
	switch x := i.(type) {

	case nil:
		return &trueCond{}, nil

	case map[string]interface{}:
		return decodeCondMap(x)

	case string:
		return decodeCondString(x)
	}

	return nil, fmt.Errorf("Unknown condition type %T", i)
}

type trueCond struct{}

func (*trueCond) Accept(Pub) bool {
	return true
}

type hasTagCond struct {
	tag string
}

func (c *hasTagCond) Accept(p Pub) bool {
	return p.Has(c.tag)
}

func decodeCondMap(m map[string]interface{}) (Cond, error) {
	if len(m) == 0 {
		return &trueCond{}, nil
	}

	t, ok := jsstring(m, "type")
	if !ok {
		return nil, errors.New(`missing or invalid cond key "type"`)
	}

	if t != "tag" {
		return nil, errors.Errorf("unknown cond type %q", t)
	}

	h, ok := jsstring(m, "value")
	if !ok || h == "" || h[0] != '#' {
		return nil, errors.New(`missing or invalid cond tag key "value"`)
	}

	return &hasTagCond{h}, nil
}

func decodeCondString(s string) (Cond, error) {
	t := condTok{src: s}
	c, err := decodeCondExpr(&t)
	if err == nil && !t.done() {
		err = errors.New("garbage after expression")
	}

	if err != nil {
		return nil, fmt.Errorf("Parse condition at %d: %w", t.pos, err)
	}

	return c, nil
}

func decodeCondExpr(t *condTok) (Cond, error) {
	left, err := decodeCondArg(t)
	if err != nil {
		return nil, err
	}

	for {
		op := t.next()
		if op == "" || op == ")" {
			t.back = op
			return left, nil
		}

		if op != "and" && op != "or" {
			return nil, fmt.Errorf("invalid op %q", op)
		}

		right, err := decodeCondArg(t)
		if err != nil {
			return nil, err
		}

		switch op {
		case "and":
			left = &andCond{left, right}
		case "or":
			left = &orCond{left, right}
		default:
			return nil, errors.New("internal condition parser error")
		}
	}
}

func decodeCondArg(t *condTok) (Cond, error) {
	tok := t.next()

	switch {

	case tok == "not":
		not, err := decodeCondArg(t)
		if err != nil {
			return nil, err
		}
		return &notCond{not}, nil

	case tok == "(":
		savepos := t.pos
		c, err := decodeCondExpr(t)
		if err == nil {
			if t.done() {
				t.pos = savepos - 1
				err = errors.New("unclosed parenthesis")
			}
			if t.next() != ")" {
				err = errors.New("internal condition parser error")
			}
		}
		return c, err

	case len(tok) > 1 && tok[0] == '#':
		return &hasTagCond{tok}, nil
	}

	return nil, errors.New("invalid expression")
}

// notCond is Cond representing a logical NOT condition
type notCond struct {
	n Cond
}

func (c *notCond) Accept(p Pub) bool {
	return !c.n.Accept(p)
}

// andCond is Cond representing a logical AND condition
type andCond struct {
	a, b Cond
}

func (c *andCond) Accept(p Pub) bool {
	return c.a.Accept(p) && c.b.Accept(p)
}

// orCond is Cond representing a logical OR condition
type orCond struct {
	a, b Cond
}

func (c *orCond) Accept(p Pub) bool {
	return c.a.Accept(p) || c.b.Accept(p)
}

// condTok is the condition tokenizer
type condTok struct {
	src string
	pos int

	back string // used to yield last token again
}

func (t *condTok) done() bool {
	return t.back == "" && t.pos == len(t.src)
}

func (t *condTok) next() string {
	if t.back != "" {
		r := t.back
		t.back = ""
		return r
	}

	t.skipSpace()

	if t.done() {
		return ""
	}

	start := t.pos
	ch := t.src[t.pos]
	t.pos++
	switch ch {
	case '(', ')':
		return string(ch)
	case '!':
		return "not"
	case '&':
		return "and"
	case '|':
		return "or"
	}

	for !t.done() && !istoksep(t.src[t.pos]) {
		t.pos++
	}

	return t.src[start:t.pos]
}

func (t *condTok) skipSpace() {
	for t.pos < len(t.src) && isspace(t.src[t.pos]) {
		t.pos++
	}
}

func istoksep(b byte) bool {
	return isspace(b) || strings.IndexByte("()!&|", b) >= 0
}

func isspace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}
