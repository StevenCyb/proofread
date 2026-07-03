// Proofread: a small terminal proofreading tool backed by a local Ollama model.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	inputWidth  = 80
	inputHeight = 10
)

type model struct {
	input    textarea.Model
	modeIdx  int
	rendered string
	err      error
	busy     bool
	glow     int
	llmHost  string
	llmKey   string
	llmModel string
}

func initialModel(host, apiKey, mdl string) model {
	textInput := textarea.New()
	textInput.Placeholder = "Type or paste text to proofread..."
	textInput.Focus()
	textInput.ShowLineNumbers = false
	textInput.SetWidth(inputWidth)
	textInput.SetHeight(inputHeight)

	return model{
		input:    textInput,
		modeIdx:  0,
		rendered: "",
		err:      nil,
		busy:     false,
		glow:     0,
		llmHost:  host,
		llmKey:   apiKey,
		llmModel: mdl,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

//nolint:ireturn // Update's signature is fixed by the tea.Model interface.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		if next, cmd, handled := m.handleKey(typed); handled {
			return next, cmd
		}
	case glowTickMsg:
		if !m.busy {
			return m, nil
		}

		m.glow = (m.glow + 1) % len(glowColors)

		return m, glowTick()
	case proofreadResult:
		m.busy = false
		m.input.Focus()

		if typed.err != nil {
			m.err = typed.err
		} else {
			m.rendered = renderDiff(m.input.Value(), typed.text)
			m.input.SetValue(typed.text)
		}

		return m, nil
	}

	// Any other keypress resumes normal editing and drops the highlight overlay.
	if _, ok := msg.(tea.KeyMsg); ok {
		m.rendered = ""
	}

	var cmd tea.Cmd

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit, true
	}

	if m.busy {
		return m, nil, true // block editing/mode-switching while a request is in flight
	}

	switch msg.String() {
	case "ctrl+n":
		m.modeIdx = (m.modeIdx + 1) % len(modes)
		return m, nil, true
	case "ctrl+b":
		m.modeIdx = (m.modeIdx - 1 + len(modes)) % len(modes)
		return m, nil, true
	case "ctrl+p":
		m.busy = true
		m.err = nil
		m.input.Blur()

		proofreadCmd := callOllama(m.llmHost, m.llmKey, m.llmModel, modes[m.modeIdx], m.input.Value())

		return m, tea.Batch(proofreadCmd, glowTick()), true
	case "ctrl+k":
		m.input.Reset()
		m.rendered = ""
		m.err = nil

		return m, nil, true
	case "ctrl+y":
		if err := clipboard.WriteAll(m.input.Value()); err != nil {
			m.err = fmt.Errorf("copy to clipboard: %w", err)
		}

		return m, nil, true
	}

	return m, nil, false
}

func main() {
	dotenv := loadDotEnv()

	host := flag.String("host", configDefault(dotenv, "PROOFREAD_HOST", "http://localhost:11434"),
		"OpenAI-compatible API host (Ollama, LiteLLM, etc.)")
	mdl := flag.String("model", configDefault(dotenv, "PROOFREAD_MODEL", "gemma4:e4b"), "Model name")
	apiKey := flag.String("api-key", configDefault(dotenv, "PROOFREAD_API_KEY", ""),
		"API key, if the endpoint requires one (e.g. LiteLLM)")

	flag.Parse()

	p := tea.NewProgram(initialModel(*host, *apiKey, *mdl))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
