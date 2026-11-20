import { ApiError } from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";

export const ORG_ERRORS: Record<string, string> = {
  ORG_NOT_FOUND: "Organisasi tidak ditemukan.",
  ORG_CODE_CONFLICT: "Kode organisasi sudah digunakan di paroki ini.",
  ORG_IN_USE:
    "Organisasi masih memiliki sub-organisasi atau data aktif sehingga tidak dapat dihapus.",
  ORG_FORBIDDEN: "Anda tidak memiliki izin untuk melakukan tindakan ini.",
  ORG_MEMBER_NOT_FOUND: "Anggota organisasi tidak ditemukan.",
  ORG_MEMBER_ALREADY_EXISTS:
    "Pengguna sudah terdaftar sebagai anggota organisasi ini.",
  ORG_LAST_ADMIN_REQUIRED:
    "Minimal harus ada satu Admin Organisasi aktif yang tersisa.",
  ORG_INVITE_NOT_FOUND: "Tautan undangan organisasi tidak ditemukan.",
  ORG_INVITE_EXPIRED: "Tautan undangan organisasi telah kedaluwarsa.",
  ORG_INVITE_REVOKED: "Tautan undangan organisasi telah dicabut.",
  ORG_INVITE_MAX_USES_REACHED:
    "Tautan undangan organisasi sudah mencapai batas penggunaan.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
  AUTH_OTP_INVALID: "Kode OTP tidak valid.",
  AUTH_OTP_EXPIRED: "Kode OTP sudah kedaluwarsa.",
  AUTH_OTP_RATE_LIMITED: "Terlalu banyak permintaan. Coba lagi sebentar.",
  USER_DISABLED: "Akun pengguna sedang dinonaktifkan.",
};

export function resolveOrgError(err: unknown, fallback?: string): string {
  if (err instanceof ApiError) {
    return resolveErrorMessage(err.code, err.message, ORG_ERRORS);
  }
  return fallback ?? "Terjadi kesalahan yang tidak terduga.";
}
