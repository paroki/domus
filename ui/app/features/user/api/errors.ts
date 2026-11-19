export const USER_ERRORS: Record<string, string> = {
  USER_NOT_FOUND:
    "Akun tidak ditemukan. Silakan minta undangan dari pengurus organisasi atau admin.",
  USER_EMAIL_CONFLICT: "Alamat email sudah terdaftar pada akun lain.",
  USER_DISABLED: "Akun Anda dinonaktifkan. Hubungi administrator.",
  USER_LAST_SUPER_ADMIN_REQUIRED:
    "Minimal harus ada satu Super Admin yang aktif.",
  AUTH_OTP_INVALID: "Kode OTP tidak valid. Periksa kembali kode Anda.",
  AUTH_OTP_EXPIRED: "Kode OTP sudah kedaluwarsa. Silakan kirim ulang kode.",
  AUTH_OTP_RATE_LIMITED:
    "Terlalu sering meminta kode. Tunggu sebentar sebelum mencoba lagi.",
  AUTH_OAUTH_PROVIDER_UNKNOWN: "Penyedia login OAuth tidak dikenali.",
  AUTH_OAUTH_EMAIL_UNVERIFIED: "Email dari penyedia OAuth belum terverifikasi.",
  USER_SESSION_NOT_FOUND: "Sesi tidak ditemukan.",
  USER_ACCOUNT_NOT_FOUND: "Akun tertaut tidak ditemukan.",
  USER_LAST_ACCOUNT_REQUIRED:
    "Minimal satu metode login harus tetap tertaut pada akun Anda.",
};
