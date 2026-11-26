import type { BadgeTone } from "~/shared/ui/Badge";

export const ROLE_PAROKI_ADMIN = "ref:paroki-admin";

export type OrgKind =
  | "dpp"
  | "bgkp"
  | "section"
  | "categorical"
  | "wilayah"
  | "lingkungan"
  | "community";

export const ORG_KINDS: { value: OrgKind; label: string }[] = [
  { value: "dpp", label: "DPP" },
  { value: "bgkp", label: "BGKP" },
  { value: "section", label: "Bidang / Seksi" },
  { value: "categorical", label: "Kategorial" },
  { value: "wilayah", label: "Wilayah / Stasi" },
  { value: "lingkungan", label: "Lingkungan" },
  { value: "community", label: "Komunitas" },
];

export const DEFAULT_POSITION_SUGGESTIONS = [
  "Ketua",
  "Wakil Ketua",
  "Sekretaris",
  "Bendahara",
  "Koordinator",
  "Anggota",
] as const;

export const INVITE_TTL_OPTIONS: { label: string; seconds: number }[] = [
  { label: "24 jam", seconds: 86400 },
  { label: "3 hari", seconds: 259200 },
  { label: "7 hari", seconds: 604800 },
];

export function isParokiAdmin(roles?: string[]): boolean {
  return Boolean(roles?.includes(ROLE_PAROKI_ADMIN));
}

export function slugifyCode(raw: string): string {
  return raw
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 64);
}

export function kindLabel(kind: string): string {
  return ORG_KINDS.find((k) => k.value === kind)?.label ?? kind;
}

export function providerLabel(provider: string): string {
  switch (provider) {
    case "google":
      return "Google";
    case "github":
      return "GitHub";
    default:
      return provider;
  }
}

export function inviteStatusMeta(status: string): {
  label: string;
  tone: BadgeTone;
} {
  switch (status) {
    case "active":
      return { label: "Aktif", tone: "success" };
    case "expired":
      return { label: "Kedaluwarsa", tone: "warning" };
    case "revoked":
      return { label: "Dicabut", tone: "danger" };
    case "exhausted":
      return { label: "Kuota Habis", tone: "neutral" };
    default:
      return { label: status, tone: "neutral" };
  }
}

export function computeInviteStatus(invite: {
  revokedAt?: string | null;
  expiresAt: string;
  maxUses?: number | null;
  usedCount: number;
}): "active" | "expired" | "revoked" | "exhausted" {
  if (invite.revokedAt) return "revoked";
  if (new Date(invite.expiresAt).getTime() <= Date.now()) return "expired";
  if (invite.maxUses != null && invite.usedCount >= invite.maxUses) {
    return "exhausted";
  }
  return "active";
}

export function buildInviteJoinUrl(origin: string, code: string): string {
  const base = origin.replace(/\/+$/, "");
  return `${base}/org/join/${encodeURIComponent(code)}`;
}

export interface JabatanTreeItem {
  id?: string;
  parentId?: string | null;
  jabatan?: string;
  nama?: string | null;
  userId?: string | null;
  umatId?: string | null;
  startDate?: string | null;
  endDate?: string | null;
  urutan?: number;
  active?: boolean;
}

export interface OrgJabatanNode<T extends JabatanTreeItem = JabatanTreeItem> {
  item: T;
  depth: number;
  children: OrgJabatanNode<T>[];
}

export function buildJabatanTree<T extends JabatanTreeItem>(
  items: T[],
): OrgJabatanNode<T>[] {
  const sorted = [...items].sort((a, b) => {
    const au = a.urutan ?? 0;
    const bu = b.urutan ?? 0;
    if (au !== bu) return au - bu;
    return (a.jabatan ?? "").localeCompare(b.jabatan ?? "", "id");
  });

  const nodeMap = new Map<string, OrgJabatanNode<T>>();
  for (const item of sorted) {
    if (!item.id) continue;
    nodeMap.set(item.id, { item, depth: 0, children: [] });
  }

  const roots: OrgJabatanNode<T>[] = [];
  for (const item of sorted) {
    if (!item.id) continue;
    const node = nodeMap.get(item.id);
    if (!node) continue;
    const parentNode = item.parentId ? nodeMap.get(item.parentId) : undefined;
    if (parentNode && parentNode.item.id !== item.id) {
      parentNode.children.push(node);
    } else {
      roots.push(node);
    }
  }

  function assignDepth(nodes: OrgJabatanNode<T>[], depth: number) {
    for (const node of nodes) {
      node.depth = depth;
      assignDepth(node.children, depth + 1);
    }
  }
  assignDepth(roots, 0);

  return roots;
}

export function flattenJabatanTreeForSelect<T extends JabatanTreeItem>(
  tree: OrgJabatanNode<T>[],
  excludeId?: string,
): Array<{ id: string; label: string; depth: number }> {
  const out: Array<{ id: string; label: string; depth: number }> = [];
  function walk(nodes: OrgJabatanNode<T>[]) {
    for (const node of nodes) {
      const id = node.item.id ?? "";
      if (!id || (excludeId && id === excludeId)) {
        continue;
      }
      const prefix = node.depth > 0 ? `${"— ".repeat(node.depth)}` : "";
      const holder = node.item.nama ? ` (${node.item.nama})` : "";
      out.push({
        id,
        label: `${prefix}${node.item.jabatan ?? ""}${holder}`,
        depth: node.depth,
      });
      walk(node.children);
    }
  }
  walk(tree);
  return out;
}
