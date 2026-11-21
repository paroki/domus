import { ApiError } from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";

export const AGENT_ERRORS: Record<string, string> = {
  AGENT_SESSION_NOT_FOUND: "Sesi percakapan tidak ditemukan.",
  AGENT_MEMORY_NOT_FOUND: "Memori tidak ditemukan.",
  AGENT_FORBIDDEN: "Anda tidak memiliki izin untuk aksi ini.",
  AGENT_EMPTY_MESSAGE: "Pesan tidak boleh kosong.",
  AGENT_SESSION_BUSY:
    "Asisten masih memproses pesan sebelumnya pada sesi ini. Silakan tunggu sebentar.",
  AGENT_PROVIDER_UNAVAILABLE:
    "Penyedia AI belum dikonfigurasi atau sedang tidak tersedia.",
  AGENT_MAX_STEPS_EXCEEDED:
    "Asisten mencapai batas maksimum langkah pemanggilan alat.",
  AGENT_MEMORY_LIMIT_EXCEEDED:
    "Batas maksimum memori (5 disematkan / 100 aktif) telah tercapai.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
};

export function resolveAgentError(err: unknown, fallback?: string): string {
  if (err instanceof ApiError) {
    return resolveErrorMessage(err.code, err.message, AGENT_ERRORS);
  }
  if (err instanceof Error && err.message) {
    return err.message;
  }
  return fallback ?? "Terjadi kesalahan yang tidak terduga.";
}
