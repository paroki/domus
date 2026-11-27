import { ApiError } from "~/shared/api/client";
import { resolveErrorMessage } from "~/shared/api/errors";

export const EVENT_ERRORS: Record<string, string> = {
  EVENT_NOT_FOUND: "Kegiatan tidak ditemukan.",
  EVENT_FORBIDDEN:
    "Anda tidak memiliki izin untuk mengakses atau mengelola kegiatan ini.",
  EVENT_INVALID_TIME_RANGE:
    "Waktu selesai harus setelah waktu mulai, dan batas RSVP tidak boleh melewati waktu selesai.",
  EVENT_RSVP_CLOSED:
    "Konfirmasi kehadiran (RSVP) untuk kegiatan ini sudah ditutup.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
};

export function resolveEventError(err: unknown, fallback?: string): string {
  if (err instanceof ApiError) {
    return resolveErrorMessage(err.code, err.message, EVENT_ERRORS);
  }
  return fallback ?? "Terjadi kesalahan yang tidak terduga.";
}
