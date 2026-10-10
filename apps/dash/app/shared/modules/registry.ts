import {
  BookOutlined,
  CalendarOutlined,
  GlobalOutlined,
  TeamOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import type { DomusModule, FlatMenuItem, ModuleMenuItem } from "./types";

/**
 * Registry modul Domus (gaya Odoo).
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
    name: "Website",
    description: "Situs publik paroki: halaman, berita, dan pengumuman.",
    icon: GlobalOutlined,
    menu: [
      { label: "Halaman", slug: "halaman" },
      { label: "Berita", slug: "berita" },
      { label: "Pengumuman", slug: "pengumuman" },
      { label: "Galeri", slug: "galeri" },
      { label: "Pengaturan situs", slug: "pengaturan" },
    ],
  },
  {
    id: "sakramen",
    name: "Sakramen",
    description: "Pencatatan dan arsip sakramen umat.",
    icon: BookOutlined,
    menu: [
      { label: "Baptis", slug: "baptis" },
      { label: "Komuni pertama", slug: "komuni-pertama" },
      { label: "Krisma", slug: "krisma" },
      { label: "Perkawinan", slug: "perkawinan" },
      { label: "Laporan", slug: "laporan" },
    ],
  },
  {
    id: "keuangan",
    name: "Keuangan",
    description: "Kas, kolekte, dan anggaran paroki.",
    icon: WalletOutlined,
    menu: [
      {
        label: "Transaksi",
        children: [
          { label: "Penerimaan", slug: "penerimaan" },
          { label: "Pengeluaran", slug: "pengeluaran" },
        ],
      },
      { label: "Kolekte", slug: "kolekte" },
      { label: "Anggaran", slug: "anggaran" },
      { label: "Laporan", slug: "laporan" },
    ],
  },
  {
    id: "umat",
    name: "Umat",
    description: "Data umat, keluarga, dan lingkungan.",
    icon: TeamOutlined,
    menu: [
      { label: "Daftar umat", slug: "daftar" },
      { label: "Keluarga", slug: "keluarga" },
      { label: "Lingkungan", slug: "lingkungan" },
      { label: "Pengurus", slug: "pengurus" },
    ],
  },
  {
    id: "kegiatan",
    name: "Kegiatan",
    description: "Kalender, jadwal misa, dan acara paroki.",
    icon: CalendarOutlined,
    menu: [
      { label: "Kalender", slug: "kalender" },
      { label: "Jadwal misa", slug: "jadwal-misa" },
      { label: "Acara", slug: "acara" },
      { label: "Kepanitiaan", slug: "kepanitiaan" },
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
  group?: string,
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
