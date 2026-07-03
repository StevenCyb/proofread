package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const maxErrorBodyPreview = 500

var modes = []string{"Proofread", "Professional", "Casual", "Polite", "#social"}

var modePrompts = map[string]string{
	"Proofread":    "Proofread and correct spelling/grammar only. Preserve tone, style and length exactly.",
	"Professional": "Proofread and rewrite in a professional, business-appropriate tone.",
	"Casual":       "Proofread and rewrite in a relaxed, casual, conversational tone.",
	"#social":      "Proofread and rewrite as a punchy social media post, concise, with fitting hashtags.",
	"Polite":       "Proofread and rewrite to sound extra polite and courteous.",
}

var (
	errEmptyInput = errors.New("nothing to proofread")
	errAPI        = errors.New("api error")
)

type proofreadResult struct {
	text string
	err  error
}

func buildPrompt(mode, text string) string {
	instruction := modePrompts[mode] + " Keep the same language as the input text."
	return instruction + "\n\nReturn ONLY the corrected text, no preamble, no quotes.\n\nText:\n" + text
}

// callOllama talks to any OpenAI-compatible chat-completions endpoint —
// Ollama, LiteLLM, vLLM, OpenAI itself, etc. all speak this same shape.
func callOllama(host, apiKey, mdl, mode, text string) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(text) == "" {
			return proofreadResult{err: errEmptyInput, text: ""}
		}

		req, err := buildRequest(host, apiKey, mdl, mode, text)
		if err != nil {
			return proofreadResult{err: err, text: ""}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return proofreadResult{err: fmt.Errorf("request failed (is the server running at %s?): %w", host, err), text: ""}
		}
		defer func() { _ = resp.Body.Close() }()

		return parseChatResponse(resp)
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

func buildRequest(host, apiKey, mdl, mode, text string) (*http.Request, error) {
	body, err := json.Marshal(chatRequest{
		Model:    mdl,
		Stream:   false,
		Messages: []chatMessage{{Role: "user", Content: buildPrompt(mode, text)}},
	})
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}

	endpoint := strings.TrimSuffix(strings.TrimSuffix(host, "/"), "/v1") + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return req, nil
}

func parseChatResponse(resp *http.Response) proofreadResult {
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return proofreadResult{err: fmt.Errorf("read response: %w", err), text: ""}
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Detail string `json:"detail"` // FastAPI-style error shape used by LiteLLM and others
	}

	if err := json.Unmarshal(rawBody, &out); err != nil {
		return apiError(resp.StatusCode, rawBody)
	}

	switch {
	case out.Error.Message != "":
		return proofreadResult{err: fmt.Errorf("%w (%d): %s", errAPI, resp.StatusCode, out.Error.Message), text: ""}
	case out.Detail != "":
		return proofreadResult{err: fmt.Errorf("%w (%d): %s", errAPI, resp.StatusCode, out.Detail), text: ""}
	case resp.StatusCode != http.StatusOK || len(out.Choices) == 0:
		return apiError(resp.StatusCode, rawBody)
	}

	return proofreadResult{text: strings.TrimSpace(out.Choices[0].Message.Content), err: nil}
}

func apiError(statusCode int, rawBody []byte) proofreadResult {
	preview := strings.TrimSpace(string(rawBody))
	if len(preview) > maxErrorBodyPreview {
		preview = preview[:maxErrorBodyPreview] + "..."
	}

	return proofreadResult{err: fmt.Errorf("%w (%d): %s", errAPI, statusCode, preview), text: ""}
}
