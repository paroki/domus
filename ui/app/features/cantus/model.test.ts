import assert from "node:assert/strict";
import test from "node:test";
import {
  buildLaguMeta,
  buildMazmurMeta,
  cantusJenisLabel,
  isCantusAdmin,
  parseCantusDirectoryFilters,
  parseLaguMeta,
  parseMazmurMeta,
  ROLE_CANTUS_ADMIN,
  slugifyCantus,
} from "./model";

test("isCantusAdmin checks isCantusAdmin flag and cantus:admin role", () => {
  assert.equal(isCantusAdmin({ isCantusAdmin: true }), true);
  assert.equal(isCantusAdmin({ roles: [ROLE_CANTUS_ADMIN] }), true);
  assert.equal(isCantusAdmin({ roles: ["user:super-admin"] }), false);
  assert.equal(isCantusAdmin(null), false);
});

test("cantusJenisLabel formats known jenis values", () => {
  assert.equal(cantusJenisLabel("lagu"), "Lagu Liturgi");
  assert.equal(cantusJenisLabel("mazmur"), "Mazmur & BPI");
  assert.equal(cantusJenisLabel("profan"), "Profan");
  assert.equal(cantusJenisLabel(""), "-");
});

test("slugifyCantus builds deterministic lowercase slug <= 64 chars", () => {
  assert.equal(
    slugifyCantus("PS", "PS-320", "Tuhan Kasihanilah Kami!"),
    "ps-320-tuhan-kasihanilah-kami",
  );
  assert.equal(
    slugifyCantus("MB", "", "Madah Syukur Agung"),
    "mb-madah-syukur-agung",
  );
  const longSlug = slugifyCantus(
    "PS",
    "PS-999",
    "Lagu Yang Sangat Panjang Sekali Untuk Menguji Batas Maksimum Enam Puluh Empat Karakter",
  );
  assert.ok(longSlug.length <= 64);
  assert.match(longSlug, /^[a-z0-9][a-z0-9_-]{1,63}$/);
});

test("parseMazmurMeta and buildMazmurMeta preserve unknown keys", () => {
  const existing = { custom_key: "keep-me", masa: "Adven" };
  const built = buildMazmurMeta(
    {
      perayaan: "Minggu Adven I",
      siklusTahun: "A",
      refAlkitab: "Mzm 25:4-5",
      ulangan: "Kepada-Mu ya Tuhan",
      bpi: "Alleluya",
    },
    existing,
  );
  assert.equal(built.custom_key, "keep-me");
  assert.equal(built.siklus_tahun, "A");
  assert.equal(built.masa, undefined);

  const parsed = parseMazmurMeta(built);
  assert.equal(parsed.perayaan, "Minggu Adven I");
  assert.equal(parsed.siklusTahun, "A");
  assert.equal(parsed.refAlkitab, "Mzm 25:4-5");
  assert.equal(parsed.ulangan, "Kepada-Mu ya Tuhan");
  assert.equal(parsed.bpi, "Alleluya");
});

test("parseLaguMeta and buildLaguMeta preserve unknown keys", () => {
  const built = buildLaguMeta(
    { birama: "4/4", tempo: "Andante", notAwal: "1=D" },
    { source_note: "edisi 2020" },
  );
  assert.equal(built.source_note, "edisi 2020");
  assert.equal(built.not_awal, "1=D");

  const parsed = parseLaguMeta(built);
  assert.equal(parsed.birama, "4/4");
  assert.equal(parsed.tempo, "Andante");
  assert.equal(parsed.notAwal, "1=D");
});

test("parseCantusDirectoryFilters parses URLSearchParams with active default", () => {
  const defaults = parseCantusDirectoryFilters(new URLSearchParams());
  assert.equal(defaults.active, "active");
  assert.equal(defaults.jenis, "");

  const custom = parseCantusDirectoryFilters(
    new URLSearchParams("q=syukur&jenis=lagu&tag=Adven&active=false"),
  );
  assert.equal(custom.q, "syukur");
  assert.equal(custom.jenis, "lagu");
  assert.equal(custom.tag, "Adven");
  assert.equal(custom.active, "inactive");

  const all = parseCantusDirectoryFilters(new URLSearchParams("active=all"));
  assert.equal(all.active, "all");
});
