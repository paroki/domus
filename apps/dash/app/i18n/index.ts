import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import { en } from "./locales/en";
import { id } from "./locales/id";

export const SUPPORTED_LANGUAGES = ["id", "en"] as const;
export type Language = (typeof SUPPORTED_LANGUAGES)[number];

/** Bahasa bawaan: dipakai kalau bahasa browser tidak didukung. */
export const DEFAULT_LANGUAGE: Language = "id";

export const LANGUAGE_STORAGE_KEY = "domus.lang";

export const resources = {
  id: { translation: id },
  en: { translation: en },
} as const;

/**
 * Urutan deteksi: pilihan user (localStorage) lalu bahasa browser.
 * `en-US`, `id-ID`, dst. dipetakan ke `en` / `id`; bahasa lain jatuh ke `id`.
 */
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    supportedLngs: SUPPORTED_LANGUAGES,
    fallbackLng: DEFAULT_LANGUAGE,
    nonExplicitSupportedLngs: true,
    load: "languageOnly",
    detection: {
      order: ["localStorage", "navigator"],
      lookupLocalStorage: LANGUAGE_STORAGE_KEY,
      caches: ["localStorage"],
    },
    interpolation: { escapeValue: false }, // React sudah meng-escape
    react: { useSuspense: false },
  });

/** Bahasa aktif yang sudah dinormalisasi ke salah satu bahasa yang didukung. */
export function getCurrentLanguage(): Language {
  const code = (i18n.resolvedLanguage ?? i18n.language ?? "").slice(0, 2);
  return (SUPPORTED_LANGUAGES as readonly string[]).includes(code)
    ? (code as Language)
    : DEFAULT_LANGUAGE;
}

export default i18n;
