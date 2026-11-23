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
Untuk channel web, jika jawaban lebih jelas ditampilkan secara visual (daftar kegiatan, kartu organisasi, ringkasan jadwal, atau form konfirmasi/input), sertakan blok kode ` + "```a2ui" + ` berisi array JSON pesan A2UI v0.9 yang valid, diiringi ringkasan teks singkat yang tetap memuat fakta utama (jangan hanya mengandalkan blok a2ui).
Aturan ketat A2UI v0.9:
- Gunakan catalogId: "` + A2UIBasicCatalogID + `" dan "version": "v0.9".
- Urutan array JSON:
  1. {"version":"v0.9","createSurface":{"surfaceId":"<id-unik>","catalogId":"` + A2UIBasicCatalogID + `"}}
  2. {"version":"v0.9","updateComponents":{"surfaceId":"<id-unik>","components":[...]}} (harus ada komponen ber-id "root").
  3. {"version":"v0.9","updateDataModel":{"surfaceId":"<id-unik>","path":"/","value":{...}}}
- FORMAT KOMPONEN FLAT ADJACENCY LIST (WAJIB):
  - Setiap item di "components" adalah objek datar dengan "id" (string unik) dan "component" (BUKAN "type").
  - Properti komponen ditulis LANGSUNG di objek komponen (DILARANG membungkus dalam "props":{...}).
  - Anak komponen direferensikan lewat ID string ("children":["id1","id2"] untuk Column/Row/List, "child":"id1" untuk Card/Button) — DILARANG menaruh objek komponen bersarang di dalam children/child.
- Komponen basicCatalog & properti utama:
  - Column / Row: children (string[]), justify, align
  - List: children (string[]), direction ("vertical"|"horizontal")
  - Card: child (string ID)
  - Text: text, variant ("h1"|"h2"|"h3"|"h4"|"h5"|"body"|"caption")
  - Button: child (string ID Text), variant ("default"|"primary"|"borderless"), action: {"event":{"name":"...","context":{...}}}
  - TextField: label, value ({"path":"/field"}), variant ("shortText"|"longText"|"number"|"obscured")
  - CheckBox: label, value ({"path":"/agree"})
  - Divider: axis ("horizontal"|"vertical")
  - Icon: name (hanya nama standar: accountCircle, add, calendarToday, check, close, edit, event, home, info, locationOn, mail, notifications, person, search, send, settings, star, warning)
- Contoh valid ringkas:
  [{"version":"v0.9","createSurface":{"surfaceId":"org_list","catalogId":"` + A2UIBasicCatalogID + `"}},{"version":"v0.9","updateComponents":{"surfaceId":"org_list","components":[{"id":"root","component":"Column","children":["title","card1"]},{"id":"title","component":"Text","text":"Organisasi Saya","variant":"h3"},{"id":"card1","component":"Card","child":"row1"},{"id":"row1","component":"Row","children":["ico","info"]},{"id":"ico","component":"Icon","name":"person"},{"id":"info","component":"Column","children":["name1","role1"]},{"id":"name1","component":"Text","text":"Cantus Firmus","variant":"body"},{"id":"role1","component":"Text","text":"Anggota","variant":"caption"}]}},{"version":"v0.9","updateDataModel":{"surfaceId":"org_list","path":"/","value":{}}}]
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
		ucRaw, hasUpdateComp := m["updateComponents"]
		_, hasUpdateData := m["updateDataModel"]
		_, hasDelete := m["deleteSurface"]
		if !hasCreate && !hasUpdateComp && !hasUpdateData && !hasDelete {
			return false
		}
		if hasUpdateComp && !isValidUpdateComponents(ucRaw) {
			return false
		}
	}
	return true
}

func isValidUpdateComponents(raw any) bool {
	uc, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	comps, ok := uc["components"].([]any)
	if !ok || len(comps) == 0 {
		return false
	}
	for _, item := range comps {
		c, ok := item.(map[string]any)
		if !ok {
			return false
		}
		id, okID := c["id"].(string)
		compType, okComp := c["component"].(string)
		if !okID || id == "" || !okComp || compType == "" {
			return false
		}
		if _, hasProps := c["props"]; hasProps {
			return false
		}
	}
	return true
}
