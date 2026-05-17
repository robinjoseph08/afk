package afk

import "errors"

var (
	ErrDraining    = errors.New("afk: draining, no new runs accepted")
	ErrOutputParse = errors.New("afk: failed to parse structured output")
	ErrOutputMissing = errors.New("afk: output tag not found in agent output")
	ErrPromptConflict = errors.New("afk: cannot specify both Prompt and PromptFile")
	ErrPromptEmpty = errors.New("afk: either Prompt or PromptFile must be specified")
)
