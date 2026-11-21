package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFormatUserMemoryXML(t *testing.T) {
	assert.Equal(t, "", FormatUserMemoryXML(nil))
	assert.Equal(t, "", FormatUserMemoryXML([]Memory{}))

	id1 := uuid.MustParse("01933e5b-1b2c-7000-8000-000000000001")
	id2 := uuid.MustParse("01933e5b-1b2c-7000-8000-000000000002")

	got := FormatUserMemoryXML([]Memory{
		{
			ID:       id1,
			Scope:    MemoryScopeUser,
			Category: MemoryCategoryMinistry,
			Content:  `Aktif di "Koor St. Cecilia" & Lektor <Wilayah 2>`,
			IsPinned: true,
		},
		{
			ID:       id2,
			Scope:    MemoryScopeParish,
			Category: MemoryCategoryParishFAQ,
			Content:  "Misa Minggu pukul 07:00 & 09:00",
			IsPinned: false,
		},
	})

	expected := `<user_memory>
  <fact id="01933e5b-1b2c-7000-8000-000000000001" category="ministry" scope="user" pinned="true">Aktif di &quot;Koor St. Cecilia&quot; &amp; Lektor &lt;Wilayah 2&gt;</fact>
  <fact id="01933e5b-1b2c-7000-8000-000000000002" category="parish_faq" scope="parish" pinned="false">Misa Minggu pukul 07:00 &amp; 09:00</fact>
</user_memory>`
	assert.Equal(t, expected, got)
}
