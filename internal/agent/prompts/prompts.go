package prompts

import _ "embed"

// Default English Markdown system prompts for Domus AI agents.
var (
	//go:embed gabriel.md
	Gabriel string
	//go:embed mikael.md
	Mikael string
	//go:embed rafael.md
	Rafael string
	//go:embed memory_extractor.md
	MemoryExtractor string
)
