import {
  ApartmentOutlined,
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
    id: "sys",
    name: "modules.sys.name",
    description: "modules.sys.description",
    icon: ApartmentOutlined,
    color: "var(--domus-app-sys)",
    menu: [
      { label: "modules.sys.menu.diocese", slug: "diocese" },
      { label: "modules.sys.menu.parish", slug: "parish" },
      { label: "modules.sys.menu.territorial", slug: "territorial" },
    ],
  },
  {
    id: "web",
    name: "modules.web.name",
    description: "modules.web.description",
    icon: GlobalOutlined,
    color: "var(--domus-app-web)",
    menu: [
      { label: "modules.web.menu.halaman", slug: "halaman" },
      { label: "modules.web.menu.berita", slug: "berita" },
      { label: "modules.web.menu.pengumuman", slug: "pengumuman" },
      { label: "modules.web.menu.galeri", slug: "galeri" },
      { label: "modules.web.menu.pengaturan", slug: "pengaturan" },
    ],
  },
  {
    id: "sacra",
    name: "modules.sacra.name",
    description: "modules.sacra.description",
    icon: BookOutlined,
    color: "var(--domus-app-sacra)",
    menu: [
      { label: "modules.sacra.menu.baptis", slug: "baptis" },
      { label: "modules.sacra.menu.komuni-pertama", slug: "komuni-pertama" },
      { label: "modules.sacra.menu.krisma", slug: "krisma" },
      { label: "modules.sacra.menu.perkawinan", slug: "perkawinan" },
      { label: "modules.sacra.menu.laporan", slug: "laporan" },
    ],
  },
  {
    id: "fin",
    name: "modules.fin.name",
    description: "modules.fin.description",
    icon: WalletOutlined,
    color: "var(--domus-app-fin)",
    menu: [
      {
        label: "modules.fin.groups.transaksi",
        children: [
          { label: "modules.fin.menu.penerimaan", slug: "penerimaan" },
          { label: "modules.fin.menu.pengeluaran", slug: "pengeluaran" },
        ],
      },
      { label: "modules.fin.menu.kolekte", slug: "kolekte" },
      { label: "modules.fin.menu.anggaran", slug: "anggaran" },
      { label: "modules.fin.menu.laporan", slug: "laporan" },
    ],
  },
  {
    id: "par",
    name: "modules.par.name",
    description: "modules.par.description",
    icon: TeamOutlined,
    color: "var(--domus-app-par)",
    menu: [
      { label: "modules.par.menu.daftar", slug: "daftar" },
      { label: "modules.par.menu.keluarga", slug: "keluarga" },
      { label: "modules.par.menu.lingkungan", slug: "lingkungan" },
      { label: "modules.par.menu.pengurus", slug: "pengurus" },
    ],
  },
  {
    id: "act",
    name: "modules.act.name",
    description: "modules.act.description",
    icon: CalendarOutlined,
    color: "var(--domus-app-act)",
    menu: [
      { label: "modules.act.menu.kalender", slug: "kalender" },
      { label: "modules.act.menu.jadwal-misa", slug: "jadwal-misa" },
      { label: "modules.act.menu.acara", slug: "acara" },
      { label: "modules.act.menu.kepanitiaan", slug: "kepanitiaan" },
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
