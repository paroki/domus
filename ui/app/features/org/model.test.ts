import assert from "node:assert/strict";
import test from "node:test";
import {
  buildInviteJoinUrl,
  buildJabatanTree,
  computeInviteStatus,
  flattenJabatanTreeForSelect,
  inviteStatusMeta,
  isParokiAdmin,
  kindLabel,
  ROLE_PAROKI_ADMIN,
  slugifyCode,
} from "./model";

test("isParokiAdmin checks ref:paroki-admin role", () => {
  assert.equal(isParokiAdmin([ROLE_PAROKI_ADMIN]), true);
  assert.equal(isParokiAdmin(["user:user"]), false);
  assert.equal(isParokiAdmin(undefined), false);
});

test("slugifyCode normalizes mixed case and punctuation", () => {
  assert.equal(slugifyCode("  Seksi Liturgi 01! "), "seksi-liturgi-01");
});

test("kindLabel maps known org kinds and falls back safely", () => {
  assert.equal(kindLabel("dpp"), "DPP");
  assert.equal(kindLabel("section"), "Bidang / Seksi");
  assert.equal(kindLabel("custom"), "custom");
});

test("inviteStatusMeta and computeInviteStatus reflect invite states", () => {
  const future = new Date(Date.now() + 3600_000).toISOString();
  const past = new Date(Date.now() - 3600_000).toISOString();

  assert.equal(
    computeInviteStatus({ expiresAt: future, usedCount: 0 }),
    "active",
  );
  assert.equal(
    computeInviteStatus({
      revokedAt: past,
      expiresAt: future,
      usedCount: 0,
    }),
    "revoked",
  );
  assert.equal(
    computeInviteStatus({ expiresAt: past, usedCount: 0 }),
    "expired",
  );
  assert.equal(
    computeInviteStatus({ expiresAt: future, maxUses: 2, usedCount: 2 }),
    "exhausted",
  );
  assert.equal(inviteStatusMeta("active").label, "Aktif");
  assert.equal(inviteStatusMeta("revoked").tone, "danger");
});

test("buildInviteJoinUrl formats shareable join URL", () => {
  assert.equal(
    buildInviteJoinUrl("https://domus.example.com/", "Abc123xyz"),
    "https://domus.example.com/org/join/Abc123xyz",
  );
});

test("buildJabatanTree builds sorted multi-level hierarchy and handles orphans", () => {
  const tree = buildJabatanTree([
    {
      id: "child-b",
      parentId: "root",
      jabatan: "Sekretaris",
      urutan: 2,
      active: true,
    },
    {
      id: "child-a",
      parentId: "root",
      jabatan: "Bendahara",
      urutan: 1,
      active: true,
    },
    {
      id: "grandchild",
      parentId: "child-b",
      jabatan: "Wakil Sekretaris",
      urutan: 1,
      active: true,
    },
    {
      id: "root",
      parentId: null,
      jabatan: "Ketua",
      nama: "Bpk. Budi",
      urutan: 1,
      active: true,
    },
    {
      id: "orphan",
      parentId: "missing-parent",
      jabatan: "Penasihat",
      urutan: 2,
      active: true,
    },
  ]);

  assert.equal(tree.length, 2);
  assert.equal(tree[0].item.id, "root");
  assert.equal(tree[0].depth, 0);
  assert.equal(tree[1].item.id, "orphan");
  assert.equal(
    tree[0].children.map((c) => c.item.id).join(","),
    "child-a,child-b",
  );
  assert.equal(tree[0].children[1].depth, 1);
  assert.equal(tree[0].children[1].children[0].item.id, "grandchild");
  assert.equal(tree[0].children[1].children[0].depth, 2);

  const flatAll = flattenJabatanTreeForSelect(tree);
  assert.deepEqual(
    flatAll.map((f) => f.id),
    ["root", "child-a", "child-b", "grandchild", "orphan"],
  );
  assert.equal(flatAll[0].label, "Ketua (Bpk. Budi)");
  assert.equal(flatAll[3].label, "— — Wakil Sekretaris");

  // Excluding child-b must also exclude its descendant grandchild.
  const flatExcl = flattenJabatanTreeForSelect(tree, "child-b");
  assert.deepEqual(
    flatExcl.map((f) => f.id),
    ["root", "child-a", "orphan"],
  );
});

test("buildJabatanTree handles deep nesting beyond 3 levels", () => {
  const tree = buildJabatanTree([
    { id: "l0", parentId: null, jabatan: "Ketua", urutan: 1, active: true },
    { id: "l1", parentId: "l0", jabatan: "Wakil", urutan: 1, active: true },
    { id: "l2", parentId: "l1", jabatan: "Koord", urutan: 1, active: true },
    { id: "l3", parentId: "l2", jabatan: "Sub", urutan: 1, active: true },
    { id: "l4", parentId: "l3", jabatan: "Staf", urutan: 1, active: true },
  ]);
  assert.equal(tree[0].depth, 0);
  assert.equal(
    tree[0].children[0].children[0].children[0].children[0].depth,
    4,
  );
});
