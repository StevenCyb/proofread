package main

import (
	"strings"
	"testing"
)

func TestParseDotEnvSkipsBlanksCommentsAndQuotes(t *testing.T) {
	t.Parallel()

	input := "# comment\n\nPROOFREAD_HOST=http://example.com\nPROOFREAD_API_KEY=\"sk-123\"\nnotakeyvalue\n"

	got := parseDotEnv(strings.NewReader(input))

	want := map[string]string{
		"PROOFREAD_HOST":    "http://example.com",
		"PROOFREAD_API_KEY": "sk-123",
	}

	if len(got) != len(want) {
		t.Fatalf("parseDotEnv() = %v, want %v", got, want)
	}

	for k, v := range want {
		if got[k] != v {
			t.Fatalf("parseDotEnv()[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestConfigDefaultPrefersEnvOverDotenvOverFallback(t *testing.T) {
	const dotenvValue = "from-dotenv"

	t.Setenv("PROOFREAD_TEST_VAR", "from-env")

	dotenv := map[string]string{"PROOFREAD_TEST_VAR": dotenvValue, "DOTENV_ONLY": dotenvValue}

	if got := configDefault(dotenv, "PROOFREAD_TEST_VAR", "fallback"); got != "from-env" {
		t.Fatalf("configDefault = %q, want env var to win", got)
	}

	if got := configDefault(dotenv, "DOTENV_ONLY", "fallback"); got != dotenvValue {
		t.Fatalf("configDefault = %q, want dotenv value", got)
	}

	if got := configDefault(dotenv, "MISSING_VAR", "fallback"); got != "fallback" {
		t.Fatalf("configDefault = %q, want fallback", got)
	}
}
