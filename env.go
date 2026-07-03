package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// parseDotEnv reads simple KEY=VALUE lines, skipping blanks and #-comments.
func parseDotEnv(r io.Reader) map[string]string {
	values := map[string]string{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}

	return values
}

// loadDotEnv reads .env next to the running binary. `go run` builds to a throwaway
// temp path, so it falls back to the working directory when that lookup misses.
// Missing file is not an error — it just yields no overrides.
func loadDotEnv() map[string]string {
	if exe, err := os.Executable(); err == nil {
		if values, ok := readDotEnv(filepath.Join(filepath.Dir(exe), ".env")); ok {
			return values
		}
	}

	if values, ok := readDotEnv(".env"); ok {
		return values
	}

	return map[string]string{}
}

func readDotEnv(path string) (map[string]string, bool) {
	file, err := os.Open(path) //nolint:gosec // path is always one of two fixed local candidates, not user input
	if err != nil {
		return nil, false
	}
	defer func() { _ = file.Close() }()

	return parseDotEnv(file), true
}

// configDefault resolves in priority order: real env var, then .env-next-to-binary, then fallback.
// Whatever this returns is only a flag default — an explicit CLI flag still wins at parse time.
func configDefault(dotenv map[string]string, envVar, fallback string) string {
	if v := os.Getenv(envVar); v != "" {
		return v
	}

	if v, ok := dotenv[envVar]; ok && v != "" {
		return v
	}

	return fallback
}
