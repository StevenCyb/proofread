package main

import "testing"

const (
	wordThis = "this"
	wordText = "text"
)

func TestWordLCSFindsUnchangedWords(t *testing.T) {
	t.Parallel()

	//nolint:misspell // intentional typo: this is proofreading test data
	before := []string{wordThis, "is", "teh", wordText}
	after := []string{wordThis, "is", "the", wordText}
	got := wordLCS(before, after)

	want := []string{wordThis, "is", wordText}
	if len(got) != len(want) {
		t.Fatalf("wordLCS = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("wordLCS = %v, want %v", got, want)
		}
	}
}
