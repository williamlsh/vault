package main

import "testing"

func TestPop(t *testing.T) {
	args := []string{"one", "two", "three"}
	var s string
	s, args = pop(args)
	if s != "one" {
		t.Errorf("unexpected: %q", s)
	}
	s, args = pop(args)
	if s != "two" {
		t.Errorf("unexpected: %q", s)
	}
	s, args = pop(args)
	if s != "three" {
		t.Errorf("unexpected: %q", s)
	}
}
