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
    "Data masih digunakan oleh anggota keluarga atau jabatan kepengurusan aktif dan tidak dapat dihapus.",
  UMAT_REFERENSI_NOT_FOUND: "Item referensi umat tidak ditemukan.",
  UMAT_REFERENSI_CONFLICT:
    "Kode referensi sudah digunakan pada kategori dan cakupan ini.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
};

const UMAT_IN_USE_CONTEXT_MESSAGES = {
  keluarga:
    "Kartu Keluarga masih memiliki anggota umat aktif dan tidak dapat dihapus.",
  umat: "Data umat masih terhubung dengan jabatan kepengurusan aktif dan tidak dapat dihapus.",
} as const;

export type UmatErrorContext = keyof typeof UMAT_IN_USE_CONTEXT_MESSAGES;

const UMAT_REFERENSI_ERROR_MESSAGES: Record<string, string> = {
  ...UMAT_ERROR_MESSAGES,
  UMAT_FORBIDDEN:
    "Anda tidak memiliki izin untuk mengelola kamus referensi umat.",
  UMAT_IN_USE:
    "Item referensi masih digunakan oleh data umat atau keluarga dan tidak dapat dihapus.",
};

export function resolveUmatReferensiErrorMessage(
  err: unknown,
  fallback = "Terjadi kesalahan saat memproses kamus referensi umat.",
): string {
  if (
    err instanceof ApiError &&
    err.code &&
    UMAT_REFERENSI_ERROR_MESSAGES[err.code]
  ) {
    return UMAT_REFERENSI_ERROR_MESSAGES[err.code];
  }
  return resolveUmatErrorMessage(err, fallback);
}

export function resolveUmatErrorMessage(
  err: unknown,
  fallback = "Terjadi kesalahan saat memproses data umat.",
  context?: UmatErrorContext,
): string {
  if (err instanceof ApiError) {
    if (err.code === "UMAT_IN_USE" && context) {
      return UMAT_IN_USE_CONTEXT_MESSAGES[context];
    }
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
