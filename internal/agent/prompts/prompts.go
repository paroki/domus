package prompts

import (
	_ "embed"
	"strings"
)

//go:embed gabriel.md
var rawGabriel string

//go:embed mikael.md
var rawMikael string

//go:embed rafael.md
var rawRafael string

// Embedded default system instructions for Domus AI agent personas.
var (
	GabrielInstruction = strings.TrimSpace(rawGabriel)
	MikaelInstruction  = strings.TrimSpace(rawMikael)
	RafaelInstruction  = strings.TrimSpace(rawRafael)
)
