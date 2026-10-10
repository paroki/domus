import {
  BookOutlined,
  CalendarOutlined,
  GlobalOutlined,
  TeamOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import type { DomusModule, FlatMenuItem, ModuleMenuItem, TKey } from "./types";

/**
 * Registry modul Domus (gaya Odoo).
 *
 * `name`, `description`, dan `label` adalah key terjemahan (lihat `~/i18n/locales`).
 * Tambah modul baru = tambah entri di sini + key-nya di `id.ts` dan `en.ts`.
 *
 * Tambah modul baru cukup dengan menambah satu entri di sini: launcher,
 * navbar, dan halaman placeholder-nya otomatis ikut. Kalau modul sudah
 * punya halaman sungguhan, bikin route statis `app/routes/_app.<id>.<slug>.tsx`
 * (route statis menang atas route dinamis `$moduleId/$page`).
 *
 * Belum ada filter role: semua modul dan menu tampil untuk semua user.
 */
export const modules: readonly DomusModule[] = [
  {
    id: "website",
    name: "modules.website.name",
    description: "modules.website.description",
    icon: GlobalOutlined,
    color: "var(--domus-app-website)",
    menu: [
      { label: "modules.website.menu.halaman", slug: "halaman" },
      { label: "modules.website.menu.berita", slug: "berita" },
      { label: "modules.website.menu.pengumuman", slug: "pengumuman" },
      { label: "modules.website.menu.galeri", slug: "galeri" },
      { label: "modules.website.menu.pengaturan", slug: "pengaturan" },
    ],
  },
  {
    id: "sakramen",
    name: "modules.sakramen.name",
    description: "modules.sakramen.description",
    icon: BookOutlined,
    color: "var(--domus-app-sakramen)",
    menu: [
      { label: "modules.sakramen.menu.baptis", slug: "baptis" },
      { label: "modules.sakramen.menu.komuni-pertama", slug: "komuni-pertama" },
      { label: "modules.sakramen.menu.krisma", slug: "krisma" },
      { label: "modules.sakramen.menu.perkawinan", slug: "perkawinan" },
      { label: "modules.sakramen.menu.laporan", slug: "laporan" },
    ],
  },
  {
    id: "keuangan",
    name: "modules.keuangan.name",
    description: "modules.keuangan.description",
    icon: WalletOutlined,
    color: "var(--domus-app-keuangan)",
    menu: [
      {
        label: "modules.keuangan.groups.transaksi",
        children: [
          { label: "modules.keuangan.menu.penerimaan", slug: "penerimaan" },
          { label: "modules.keuangan.menu.pengeluaran", slug: "pengeluaran" },
        ],
      },
      { label: "modules.keuangan.menu.kolekte", slug: "kolekte" },
      { label: "modules.keuangan.menu.anggaran", slug: "anggaran" },
      { label: "modules.keuangan.menu.laporan", slug: "laporan" },
    ],
  },
  {
    id: "umat",
    name: "modules.umat.name",
    description: "modules.umat.description",
    icon: TeamOutlined,
    color: "var(--domus-app-umat)",
    menu: [
      { label: "modules.umat.menu.daftar", slug: "daftar" },
      { label: "modules.umat.menu.keluarga", slug: "keluarga" },
      { label: "modules.umat.menu.lingkungan", slug: "lingkungan" },
      { label: "modules.umat.menu.pengurus", slug: "pengurus" },
    ],
  },
  {
    id: "kegiatan",
    name: "modules.kegiatan.name",
    description: "modules.kegiatan.description",
    icon: CalendarOutlined,
    color: "var(--domus-app-kegiatan)",
    menu: [
      { label: "modules.kegiatan.menu.kalender", slug: "kalender" },
      { label: "modules.kegiatan.menu.jadwal-misa", slug: "jadwal-misa" },
      { label: "modules.kegiatan.menu.acara", slug: "acara" },
      { label: "modules.kegiatan.menu.kepanitiaan", slug: "kepanitiaan" },
    ],
  },
];

const byId = new Map(modules.map((m) => [m.id, m]));

export function getModule(id: string): DomusModule | undefined {
  return byId.get(id);
}

export function modulePath(moduleId: string, slug?: string): string {
  return slug ? `/${moduleId}/${slug}` : `/${moduleId}`;
}

export function flattenMenu(
  items: ModuleMenuItem[],
  group?: TKey,
): FlatMenuItem[] {
  return items.flatMap((item) => {
    if (item.children) return flattenMenu(item.children, item.label);
    if (item.slug) return [{ slug: item.slug, label: item.label, group }];
    return [];
  });
}

export function findMenuItem(
  mod: DomusModule,
  slug: string,
): FlatMenuItem | undefined {
  return flattenMenu(mod.menu).find((item) => item.slug === slug);
}
