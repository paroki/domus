export const SETTING_ERRORS: Record<string, string> = {
  FORBIDDEN: "Hanya Super Admin yang dapat mengelola pengaturan sistem.",
  SETTING_KEY_UNKNOWN: "Kunci pengaturan tidak dikenal.",
  SETTING_VALIDATION_ERROR:
    "Konfigurasi pengaturan tidak valid. Periksa kembali nilai yang dimasukkan.",
  AGENT_PROVIDER_NOT_FOUND: "Konfigurasi penyedia LLM tidak ditemukan.",
  AGENT_PROVIDER_IN_USE:
    "Konfigurasi penyedia ini sedang digunakan oleh agen atau embedding dan tidak dapat dihapus.",
  AGENT_PROVIDER_EMBEDDING_UNSUPPORTED:
    "Penyedia Anthropic tidak mendukung text embedding.",
  AGENT_PROVIDER_UNAVAILABLE:
    "Penyedia LLM belum dikonfigurasi atau gagal dihubungi. Pastikan API Key valid.",
};
