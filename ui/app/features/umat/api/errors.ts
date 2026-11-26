import { ApiError } from "~/shared/api/client";

const UMAT_ERROR_MESSAGES: Record<string, string> = {
  UMAT_FORBIDDEN: "Anda tidak memiliki izin Admin Data Umat di paroki ini.",
  UMAT_KELUARGA_NOT_FOUND: "Kartu Keluarga tidak ditemukan.",
  UMAT_NOT_FOUND: "Data umat tidak ditemukan.",
  UMAT_SAKRAMEN_NOT_FOUND: "Catatan sakramen tidak ditemukan.",
  UMAT_SAKRAMEN_CONFLICT:
    "Jenis sakramen tersebut sudah tercatat untuk umat ini.",
  UMAT_KK_CONFLICT: "Nomor KK sudah terdaftar di paroki ini.",
  UMAT_NIK_CONFLICT: "NIK sudah terdaftar di paroki ini.",
  UMAT_USER_ALREADY_LINKED:
    "Akun pengguna tersebut sudah terhubung dengan data umat lain di paroki ini.",
  UMAT_IN_USE:
    "Kartu Keluarga masih memiliki anggota umat aktif dan tidak dapat dihapus.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
};

export function resolveUmatErrorMessage(
  err: unknown,
  fallback = "Terjadi kesalahan saat memproses data umat.",
): string {
  if (err instanceof ApiError) {
    if (err.code && UMAT_ERROR_MESSAGES[err.code]) {
      return UMAT_ERROR_MESSAGES[err.code];
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
