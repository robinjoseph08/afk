package afk

import (
	"testing"
)

func TestParseOutput(t *testing.T) {
	type result struct {
		Score  int    `json:"score"`
		Reason string `json:"reason"`
	}

	tests := []struct {
		name    string
		raw     string
		tag     string
		want    result
		wantErr error
	}{
		{
			name: "basic",
			raw:  `some output <result>{"score": 42, "reason": "good"}</result> more output`,
			tag:  "result",
			want: result{Score: 42, Reason: "good"},
		},
		{
			name: "custom tag",
			raw:  `<review>{"score": 10, "reason": "great"}</review>`,
			tag:  "review",
			want: result{Score: 10, Reason: "great"},
		},
		{
			name: "uses last match",
			raw:  `<result>{"score": 1, "reason": "first"}</result> <result>{"score": 2, "reason": "second"}</result>`,
			tag:  "result",
			want: result{Score: 2, Reason: "second"},
		},
		{
			name:    "missing tag",
			raw:     `no tags here`,
			tag:     "result",
			wantErr: ErrOutputMissing,
		},
		{
			name:    "invalid json",
			raw:     `<result>not json</result>`,
			tag:     "result",
			wantErr: ErrOutputParse,
		},
		{
			name: "default tag",
			raw:  `<result>{"score": 5, "reason": "default"}</result>`,
			tag:  "",
			want: result{Score: 5, Reason: "default"},
		},
		{
			name: "multiline json",
			raw: `<result>
{
  "score": 99,
  "reason": "multiline"
}
</result>`,
			tag:  "result",
			want: result{Score: 99, Reason: "multiline"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutput[result](tt.raw, tt.tag)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errorContains(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResolvePrompt(t *testing.T) {
	t.Run("both set", func(t *testing.T) {
		_, err := resolvePrompt(RunOpts{Prompt: "a", PromptFile: "b"})
		if err != ErrPromptConflict {
			t.Fatalf("expected ErrPromptConflict, got %v", err)
		}
	})

	t.Run("neither set", func(t *testing.T) {
		_, err := resolvePrompt(RunOpts{})
		if err != ErrPromptEmpty {
			t.Fatalf("expected ErrPromptEmpty, got %v", err)
		}
	})

	t.Run("prompt set", func(t *testing.T) {
		got, err := resolvePrompt(RunOpts{Prompt: "hello"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	})
}

func errorContains(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}
		unwrapped, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = unwrapped.Unwrap()
	}
	return false
}
