package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func ctrlKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t, Runes: nil, Alt: false, Paste: false}
}

func TestHandleKeyModeWrapsAndBlocksWhenBusy(t *testing.T) {
	t.Parallel()

	initial := initialModel("http://localhost:11434", "", "test-model")

	next, _, handled := initial.handleKey(ctrlKey(tea.KeyCtrlB))
	if !handled || next.modeIdx != len(modes)-1 {
		t.Fatalf("ctrl+b from index 0 should wrap to last mode, got modeIdx=%d handled=%v", next.modeIdx, handled)
	}

	next, _, handled = next.handleKey(ctrlKey(tea.KeyCtrlN))
	if !handled || next.modeIdx != 0 {
		t.Fatalf("ctrl+n should wrap back to 0, got modeIdx=%d handled=%v", next.modeIdx, handled)
	}

	next.busy = true

	before := next.modeIdx

	next, cmd, handled := next.handleKey(ctrlKey(tea.KeyCtrlN))
	if !handled || cmd != nil || next.modeIdx != before {
		t.Fatalf("mode switching must be blocked while busy, got modeIdx=%d cmd=%v", next.modeIdx, cmd)
	}
}
