import { ApiError } from "~/shared/api/client";

const LITURGI_ERROR_MESSAGES: Record<string, string> = {
  LITURGI_MINISTRY_NOT_FOUND: "Bidang petugas liturgi tidak ditemukan.",
  LITURGI_MINISTRY_INVALID_ROLE: "Peran bidang liturgi tidak valid.",
  LITURGI_SCHEDULE_NOT_FOUND: "Jadwal Misa tidak ditemukan.",
  LITURGI_SCHEDULE_INVALID_RANGE: "Waktu selesai harus setelah waktu mulai.",
  LITURGI_PETUGAS_NOT_FOUND: "Penugasan petugas liturgi tidak ditemukan.",
  LITURGI_PETUGAS_FORBIDDEN:
    "Anda tidak memiliki izin mengelola petugas untuk bidang ini.",
  LITURGI_PETUGAS_INVALID_STATUS: "Status penugasan petugas tidak valid.",
  LITURGI_LAGU_NOT_FOUND: "Entri lagu liturgi tidak ditemukan.",
  LITURGI_LAGU_FORBIDDEN:
    "Anda tidak memiliki izin mengelola daftar lagu untuk Misa ini.",
  EVENT_RSVP_CLOSED: "Batas waktu konfirmasi kehadiran (RSVP) sudah berakhir.",
  EVENT_RSVP_DISABLED: "Konfirmasi kehadiran tidak diaktifkan untuk Misa ini.",
};

export function resolveLiturgiError(err: unknown): string {
  if (err instanceof ApiError) {
    return (
      LITURGI_ERROR_MESSAGES[err.code] ??
      err.details[0]?.message ??
      err.message ??
      "Terjadi kesalahan pada layanan Liturgi."
    );
  }
  if (err instanceof Error) return err.message;
  return "Terjadi kesalahan yang tidak diketahui.";
}
