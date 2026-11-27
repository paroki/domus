package prompts

import _ "embed"

// Default Markdown system prompts for Domus AI agents.
var (
	//go:embed gabriel.md
	Gabriel string
	//go:embed mikael.md
	Mikael string
	//go:embed rafael.md
	Rafael string
	//go:embed yosef.md
	Yosef string
	//go:embed gregorius.md
	Gregorius string
	//go:embed sesilia.md
	Sesilia string
	//go:embed memory_extractor.md
	MemoryExtractor string
)
