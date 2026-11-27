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
  USER_TOKEN_NOT_FOUND: "Token akses pribadi tidak ditemukan.",
  USER_ACCOUNT_NOT_FOUND: "Akun tertaut tidak ditemukan.",
  USER_LAST_ACCOUNT_REQUIRED:
    "Minimal satu metode login harus tetap tertaut pada akun Anda.",
  USER_ACCOUNT_CONFLICT:
    "Akun WhatsApp atau Telegram ini sudah terhubung dengan akun Domus lain.",
  AUTH_CHANNEL_DISABLED:
    "Kanal verifikasi yang dipilih belum diaktifkan oleh administrator.",
  REF_PAROKI_NOT_FOUND: "Paroki tidak ditemukan.",
  REF_PAROKI_MEMBER_NOT_FOUND: "Anggota paroki tidak ditemukan.",
  REF_LAST_PAROKI_ADMIN_REQUIRED:
    "Minimal harus ada satu Admin Paroki yang tersisa.",
  REF_PAROKI_MEMBER_ALREADY_EXISTS:
    "Pengguna sudah terdaftar sebagai anggota paroki ini.",
  REF_FORBIDDEN: "Anda tidak memiliki izin untuk melakukan tindakan ini.",
  ORG_NOT_FOUND: "Organisasi tidak ditemukan.",
  ORG_FORBIDDEN: "Anda tidak memiliki izin untuk melakukan tindakan ini.",
  ORG_MEMBER_NOT_FOUND: "Anggota organisasi tidak ditemukan.",
  ORG_MEMBER_ALREADY_EXISTS:
    "Pengguna sudah terdaftar sebagai anggota organisasi ini.",
  ORG_LAST_ADMIN_REQUIRED:
    "Minimal harus ada satu Admin Organisasi aktif yang tersisa.",
  TENANT_REQUIRED: "Pilih paroki aktif terlebih dahulu.",
  UMAT_USER_ALREADY_LINKED:
    "Data umat ini sudah terhubung dengan akun Domus lain. Hubungi admin paroki jika ini keliru.",
  VERIFY_EXTRACTION_FAILED:
    "Gagal membaca dokumen. Pastikan foto jelas, tidak buram, dan coba lagi.",
  VERIFY_RATE_LIMITED:
    "Terlalu banyak percobaan verifikasi. Tunggu beberapa saat sebelum mencoba lagi.",
  UMAT_NIK_CONFLICT: "NIK sudah terdaftar pada data umat lain.",
  UMAT_KK_CONFLICT: "Nomor Kartu Keluarga sudah terdaftar pada keluarga lain.",
  VALIDATION_ERROR:
    "Data atau file tidak valid. Pastikan file berupa gambar (maks. 2 MiB) dan kolom wajib terisi.",
};
