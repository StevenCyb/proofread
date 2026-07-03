package main

import (
	"slices"
	"strings"
)

// renderDiff highlights words in `output` that were added/changed relative to `input`.
func renderDiff(input, output string) string {
	before := strings.Fields(input)
	after := strings.Fields(output)
	lcs := wordLCS(before, after)

	inLCS := make([]bool, len(after))
	beforeIdx, afterIdx := 0, 0

	for _, word := range lcs {
		for beforeIdx < len(before) && before[beforeIdx] != word {
			beforeIdx++
		}

		for afterIdx < len(after) && after[afterIdx] != word {
			afterIdx++
		}

		if afterIdx < len(after) {
			inLCS[afterIdx] = true
			beforeIdx++
			afterIdx++
		}
	}

	words := make([]string, len(after))
	for idx, word := range after {
		if inLCS[idx] {
			words[idx] = word
		} else {
			words[idx] = addedStyle.Render(word)
		}
	}

	return strings.Join(words, " ")
}

// lcsTable builds the dynamic-programming table for the longest common subsequence length.
func lcsTable(before, after []string) [][]int {
	table := make([][]int, len(before)+1)
	for i := range table {
		table[i] = make([]int, len(after)+1)
	}

	for beforeIdx, wordBefore := range slices.Backward(before) {
		for afterIdx, wordAfter := range slices.Backward(after) {
			switch {
			case wordBefore == wordAfter:
				table[beforeIdx][afterIdx] = table[beforeIdx+1][afterIdx+1] + 1
			case table[beforeIdx+1][afterIdx] >= table[beforeIdx][afterIdx+1]:
				table[beforeIdx][afterIdx] = table[beforeIdx+1][afterIdx]
			default:
				table[beforeIdx][afterIdx] = table[beforeIdx][afterIdx+1]
			}
		}
	}

	return table
}

// wordLCS returns the longest common subsequence of words between before and after.
func wordLCS(before, after []string) []string {
	table := lcsTable(before, after)

	var lcs []string

	beforeIdx, afterIdx := 0, 0
	for beforeIdx < len(before) && afterIdx < len(after) {
		switch {
		case before[beforeIdx] == after[afterIdx]:
			lcs = append(lcs, before[beforeIdx])
			beforeIdx++
			afterIdx++
		case table[beforeIdx+1][afterIdx] >= table[beforeIdx][afterIdx+1]:
			beforeIdx++
		default:
			afterIdx++
		}
	}

	return lcs
}
