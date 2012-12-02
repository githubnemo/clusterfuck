package parser

import (
	"testing"
)


func TestSummary(t *testing.T) {

	l := Tokenize("+++")

	if len(l.Nodes) > 1 {
		t.Fatalf("Not summarized: len(l.Nodes) > 1 (%d)\n", len(l.Nodes))
	}

	count := l.Nodes[0].(Summarizable).Count()

	if count != 3 {
		t.Fatalf("Count mismatched: Count != 3 but %d\n", count)
	}

	l = Tokenize("[[[")

	if len(l.Nodes) != 3 {
		t.Fatalf("Wrongly summarized (not allowed): len(d.Nodes) = %d\n", len(l.Nodes))
	}

	l = Tokenize("+-+")

	if len(l.Nodes) != 3 {
		t.Fatalf("Mixed types summarized: len(d.Nodes) = %d\n", len(l.Nodes))
	}

}
