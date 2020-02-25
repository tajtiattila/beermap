package main

import "testing"

func TestCondParse(t *testing.T) {
	tests := []struct {
		want bool
		src  string
	}{
		{true, "#foo"},
		{true, "not #foo"},
		{true, "#foo or #bar"},
		{true, "#foo and #bar"},
		{true, "#foo and #bar or #baz"},
		{true, "(#foo and #bar) or #baz"},
		{true, "#foo and (#bar or #baz)"},
		{true, "(#foo) and (#bar or #baz)"},
		{true, "#foo and not #bar"},
		{true, "#foo and (not #bar or not #baz)"},
		{false, "#foo and (not #bar or not #baz"},
		{false, "()"},
		{false, "#foo and (#bar or #baz"},
		{false, "#foo not and #baz"},
	}

	for _, x := range tests {
		_, err := decodeCondString(x.src)
		got := err == nil
		if got != x.want {
			t.Errorf("Parse %s got %v (%v), want %v", x.src, got, err, x.want)
		}
	}
}

func TestCond(t *testing.T) {
	var pubs []Pub
	for _, n := range "abcdef" {
		pubs = append(pubs, Pub{
			Label: string(n),
			Tags:  []string{"#" + string(n)},
		})
	}

	pubs = append(pubs, Pub{
		Label: "x",
		Tags:  []string{"#foo", "#bar", "baz"},
	}, Pub{
		Label: "y",
		Tags:  []string{"#foo", "#baz"},
	}, Pub{
		Label: "z",
		Tags:  []string{"#bar", "#baz"},
	})

	tests := []struct {
		src  string
		want string
	}{
		{"#a", "a"},
		{"#a or #b", "ab"},
		{"#a and #b", ""},
		{"not #a", "bcdefxyz"},
		{"not #a and not #foo", "bcdefz"},
		{"#foo or #bar", "xyz"},
		{"#foo and #bar", "x"},
		{"not #foo or #bar", "abcdefxz"},
		{"not #foo and #bar", "z"},
		{"#foo or not #bar", "abcdefxy"},
	}

	for _, x := range tests {
		cond, err := decodeCondString(x.src)
		if err != nil {
			t.Fatal(err)
		}

		var got string
		for _, p := range pubs {
			if cond.Accept(p) {
				got += p.Label
			}
		}

		if got != x.want {
			t.Errorf("%s accept got %v, want %v", x.src, got, x.want)
		}
	}
}
