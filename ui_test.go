package main

import "testing"

func TestObfuscateKey(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"":          "(none)",
		"ab":        "**",
		"sk-123456": "sk-***",
	}

	for input, want := range cases {
		if got := obfuscateKey(input); got != want {
			t.Fatalf("obfuscateKey(%q) = %q, want %q", input, got, want)
		}
	}
}
