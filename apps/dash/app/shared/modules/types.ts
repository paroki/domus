import type { ComponentType, CSSProperties } from "react";

export type ModuleIcon = ComponentType<{
  className?: string;
  style?: CSSProperties;
}>;

/**
 * Satu entri menu modul.
 * - Item biasa punya `slug` (segmen path relatif ke `/<module>/`).
 * - Grup punya `children` dan tidak punya `slug`.
 */
export interface ModuleMenuItem {
  label: string;
  slug?: string;
  children?: ModuleMenuItem[];
}

export interface DomusModule {
  /** Dipakai sebagai segmen path pertama: `/<id>`. */
  id: string;
  name: string;
  description: string;
  icon: ModuleIcon;
  /** Warna ikon di launcher. Isi dengan CSS variable `--domus-app-*` dari app.css. */
  color: string;
  menu: ModuleMenuItem[];
}

/** Menu yang sudah diratakan: hanya item yang bisa dibuka (punya slug). */
export interface FlatMenuItem {
  slug: string;
  label: string;
  /** Label grup induk, kalau ada. */
  group?: string;
}
