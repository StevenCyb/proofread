# Proofread

A small terminal app that proofreads/rewrites text using any OpenAI-compatible chat-completions endpoint — [Ollama](https://ollama.com), [LiteLLM](https://www.litellm.ai/), vLLM, OpenAI itself, etc. Edit in a text area, pick a mode, hit a shortcut, get corrected text back in place with changes highlighted.

## Modes

- **Proofread** — fix spelling/grammar, keep tone and length as-is
- **Professional** — rewrite in a business-appropriate tone
- **Casual** — relaxed, conversational tone
- **Polite** — extra courteous tone
- **#social** — punchy social post with hashtags

Output stays in the same language as the input.

## Install

Requires Go 1.26+ and [Ollama](https://ollama.com) running locally with a model pulled:

```sh
ollama pull gemma4:e4b   # or any model you like
ollama serve             # if not already running
```

Build:

```sh
go build -o proofread .
```

Or install straight from the repo (puts `proofread` in `$(go env GOPATH)/bin`):

```sh
go install github.com/StevenCyb/proofread@latest
```

## Usage

```sh
./proofread                                          # uses gemma4:e4b on localhost:11434 (Ollama)
./proofread -model gemma4:26b                        # bigger model, better quality, slower
./proofread -host https://my-litellm.example.com -model gpt-4o -api-key sk-...
```

Any backend that speaks the OpenAI `/v1/chat/completions` API works — point `-host` at it and pass `-api-key` (or set `PROOFREAD_API_KEY`) if it requires auth. Ollama serves this same API locally, so no separate mode is needed.

### Shortcuts

| Key | Action |
|---|---|
| `ctrl+n` / `ctrl+b` | next / previous mode |
| `ctrl+p` | proofread current text |
| any key | dismiss highlight, resume editing |
| `esc` / `ctrl+c` | quit |

While a request is in flight the border pulses and input is blocked. When it finishes, the text area is replaced with the corrected text; changed words are underlined until you start typing again.

## Configuration

Each setting resolves in this order, highest priority last:

1. built-in fallback
2. `.env` file **next to the binary** (not the working directory)
3. real environment variable
4. CLI flag

| Flag | Env var | Fallback | Description |
|---|---|---|---|
| `-host` | `PROOFREAD_HOST` | `http://localhost:11434` | OpenAI-compatible API host (Ollama, LiteLLM, etc.) |
| `-model` | `PROOFREAD_MODEL` | `gemma4:e4b` | Model name |
| `-api-key` | `PROOFREAD_API_KEY` | *(empty)* | API key, if the endpoint requires one |

Example `.env` dropped alongside the `proofread` binary:

```
PROOFREAD_HOST=https://my-litellm.example.com
PROOFREAD_MODEL=gpt-4o
PROOFREAD_API_KEY=sk-...
```
