package gemini

// geminiSession is the top-level session object in Gemini CLI JSON format.
type geminiSession struct {
	ID       string          `json:"session_id"`
	Messages []geminiMessage `json:"messages"`
	Created  string          `json:"created"`
	Updated  string          `json:"updated"`
	Summary  string          `json:"summary,omitempty"`
	CWD      string          `json:"cwd,omitempty"`
	Branch   string          `json:"branch,omitempty"`
}

// geminiMessage represents a single message within a Gemini session.
type geminiMessage struct {
	ID        string        `json:"id"`
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	Model     string        `json:"model"`
	Tokens    *geminiTokens `json:"tokens,omitempty"`
	Timestamp string        `json:"timestamp,omitempty"`
}

// geminiTokens holds per-message token counts from the Gemini CLI.
// The `cached` field is included inside `input` — subtract before pricing.
type geminiTokens struct {
	Input    int `json:"input"`
	Output   int `json:"output"`
	Cached   int `json:"cached,omitempty"`
	Thoughts int `json:"thoughts,omitempty"`
}

// geminiLine represents a single line in the JSONL variant of Gemini sessions.
type geminiLine struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id,omitempty"`
	Message   *geminiMessage `json:"message,omitempty"`
	Created   string         `json:"created,omitempty"`
	Updated   string         `json:"updated,omitempty"`
	Summary   string         `json:"summary,omitempty"`
	CWD       string         `json:"cwd,omitempty"`
	Branch    string         `json:"branch,omitempty"`
}
