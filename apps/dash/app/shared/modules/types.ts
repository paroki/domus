import type { ParseKeys } from "i18next";
import type { ComponentType, CSSProperties } from "react";

/** Key terjemahan (type-safe). Tampilkan lewat `t(key)`. */
export type TKey = ParseKeys;

export type ModuleIcon = ComponentType<{
  className?: string;
  style?: CSSProperties;
}>;

/**
 * Satu entri menu modul. `label` adalah key terjemahan.
 * - Item biasa punya `slug` (segmen path relatif ke `/<module>/`).
 * - Grup punya `children` dan tidak punya `slug`.
 */
export interface ModuleMenuItem {
  label: TKey;
  slug?: string;
  children?: ModuleMenuItem[];
}

export interface DomusModule {
  /** Dipakai sebagai segmen path pertama: `/<id>`. */
  id: string;
  /** Key terjemahan. */
  name: TKey;
  /** Key terjemahan. */
  description: TKey;
  icon: ModuleIcon;
  /** Warna ikon di launcher. Isi dengan CSS variable `--domus-app-*` dari app.css. */
  color: string;
  menu: ModuleMenuItem[];
}

/** Menu yang sudah diratakan: hanya item yang bisa dibuka (punya slug). */
export interface FlatMenuItem {
  slug: string;
  /** Key terjemahan. */
  label: TKey;
  /** Key terjemahan grup induk, kalau ada. */
  group?: TKey;
}
