package service

import (
	"encoding/json"
	"regexp"
	"strings"
)

// A2UIBasicCatalogID is the canonical A2UI v0.9 basic catalog URI.
const A2UIBasicCatalogID = "https://a2ui.org/specification/v0_9/catalogs/basic/catalog.json"

var a2uiFenceRe = regexp.MustCompile("(?s)```a2ui\\s*\\n(.*?)\\n```")

// A2UIWebPromptInstructions returns compact A2UI v0.9 instructions for web channel turns.
func A2UIWebPromptInstructions() string {
	return `<a2ui_instructions>
Untuk channel web, jika jawaban lebih jelas ditampilkan secara visual (daftar kegiatan, kartu organisasi, ringkasan jadwal, atau form konfirmasi/input), sertakan blok kode ` + "```a2ui" + ` berisi array JSON pesan A2UI v0.9 yang valid, diiringi ringkasan teks singkat.
Aturan A2UI v0.9:
- Gunakan catalogId: "` + A2UIBasicCatalogID + `" dan "version": "v0.9".
- Urutan minimal dalam array JSON:
  1. {"version":"v0.9","createSurface":{"surfaceId":"<id-unik>","catalogId":"` + A2UIBasicCatalogID + `"}}
  2. {"version":"v0.9","updateComponents":{"surfaceId":"<id-unik>","components":[...]}} (komponen akar ber-id "root").
  3. {"version":"v0.9","updateDataModel":{"surfaceId":"<id-unik>","path":"/","value":{...}}}
- Komponen basicCatalog yang tersedia: Column, Row, Card, Text (props: text, variant), Button (props: child, action:{event:{name,context}}), TextField (props: label, value:{path}), CheckBox (props: label, value:{path}), Divider, Icon (props: name).
- Untuk form, ikat nilai input ke path (mis. {"path":"/catatan"}) lalu ulangi path tersebut di action.event.context tombol submit.
- JANGAN pernah menaruh secret, token, atau data sensitif di updateDataModel.
</a2ui_instructions>`
}

// SanitizeA2UIBlocks validates fenced ```a2ui blocks and strips malformed ones so raw broken JSON is not persisted.
func SanitizeA2UIBlocks(content string) string {
	if !strings.Contains(content, "```a2ui") {
		return content
	}
	cleaned := a2uiFenceRe.ReplaceAllStringFunc(content, func(match string) string {
		sub := a2uiFenceRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return ""
		}
		if isValidA2UIMessageArray(sub[1]) {
			return match
		}
		return ""
	})
	return strings.TrimSpace(cleaned)
}

func isValidA2UIMessageArray(raw string) bool {
	var msgs []map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &msgs); err != nil || len(msgs) == 0 {
		return false
	}
	for _, m := range msgs {
		if m["version"] != "v0.9" {
			return false
		}
		_, hasCreate := m["createSurface"]
		_, hasUpdateComp := m["updateComponents"]
		_, hasUpdateData := m["updateDataModel"]
		_, hasDelete := m["deleteSurface"]
		if !hasCreate && !hasUpdateComp && !hasUpdateData && !hasDelete {
			return false
		}
	}
	return true
}
