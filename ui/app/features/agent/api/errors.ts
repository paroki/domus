export const AGENT_ERRORS: Record<string, string> = {
  TENANT_REQUIRED: "Silakan pilih paroki aktif terlebih dahulu.",
  AGENT_PROVIDER_UNAVAILABLE:
    "Model AI belum dikonfigurasi atau penyedia LLM sedang tidak tersedia.",
  AGENT_PROVIDER_RATE_LIMITED:
    "Batas permintaan penyedia LLM tercapai. Coba lagi beberapa saat.",
  AGENT_SESSION_NOT_FOUND: "Percakapan tidak ditemukan.",
  AGENT_SESSION_BUSY:
    "Percakapan ini sedang memproses pesan sebelumnya. Tunggu sebentar.",
  AGENT_MEMORY_NOT_FOUND: "Memori tidak ditemukan.",
  VALIDATION_ERROR: "Periksa kembali data yang Anda masukkan.",
  FORBIDDEN: "Anda tidak memiliki izin untuk mengakses fitur ini.",
};
