import { ApiError } from "~/shared/api/client";

const CANTUS_ERROR_MESSAGES: Record<string, string> = {
  CANTUS_FORBIDDEN: "Anda tidak memiliki izin Admin Nyanyian (cantus:admin).",
  CANTUS_NOT_FOUND: "Data nyanyian tidak ditemukan.",
  CANTUS_CODE_CONFLICT: "Slug nyanyian sudah digunakan oleh entri lain.",
  CANTUS_IN_USE:
    "Nyanyian ini sedang digunakan pada jadwal kegiatan dan tidak dapat dihapus.",
  INVALID_CURSOR: "Kursor halaman tidak valid.",
};

export function resolveCantusErrorMessage(
  err: unknown,
  fallback = "Terjadi kesalahan saat memproses data nyanyian.",
): string {
  if (err instanceof ApiError) {
    if (err.code && CANTUS_ERROR_MESSAGES[err.code]) {
      return CANTUS_ERROR_MESSAGES[err.code];
    }
    if (err.details.length > 0 && err.details[0]?.message) {
      return err.details[0].message;
    }
    if (err.message) return err.message;
  }
  if (err instanceof Error && err.message) {
    return err.message;
  }
  return fallback;
}
