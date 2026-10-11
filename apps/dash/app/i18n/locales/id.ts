/**
 * Terjemahan bahasa Indonesia: bahasa utama dan sumber kebenaran untuk tipe key.
 * Locale lain wajib punya struktur yang sama (dicek lewat `typeof id`).
 */
export const id = {
  common: {
    appName: "Domus",
    back: "Kembali",
    login: "Masuk",
    logout: "Keluar",
    loggingOut: "Sedang keluar",
    loadingAccount: "Memuat akun",
    allModules: "Semua modul",
    language: "Bahasa",
  },
  roles: {
    admin: "Admin",
  },
  error: {
    title: "Ups!",
    unexpected: "Terjadi kesalahan yang tidak terduga.",
    notFoundTitle: "404",
    notFound: "Halaman yang dicari tidak ditemukan.",
    generic: "Terjadi kesalahan",
  },
  underConstruction: {
    title: "Sedang dibangun",
    description: "Fitur ini masih dalam tahap pengembangan.",
  },
  launcher: {
    pageTitle: "Beranda",
    title: "Semua modul",
    subtitle: "Pilih modul untuk mulai bekerja.",
  },
  userMenu: {
    accountMenu: "Menu akun {{name}}",
  },
  login: {
    pageTitle: "Masuk",
    title: "Masuk ke Domus",
    subtitle: "Kelola keuskupan, paroki, dan lingkungan di satu tempat.",
    continueWithGoogle: "Lanjutkan dengan Google",
    continueWithGithub: "Lanjutkan dengan GitHub",
    noAccess: "Belum punya akses? Hubungi admin keuskupan atau paroki kamu.",
    agreement:
      "Dengan masuk, kamu menyetujui <terms>Ketentuan Layanan</terms> dan <privacy>Kebijakan Privasi</privacy>.",
    errors: {
      access_denied: "Izin login dibatalkan. Coba lagi kalau mau lanjut.",
      account_not_linked:
        "Email ini sudah terdaftar lewat penyedia lain. Masuk dengan penyedia yang dulu dipakai.",
      oauth: "Login gagal. Coba lagi sebentar lagi.",
    },
  },
  modules: {
    sys: {
      name: "Sistem",
      description: "Data induk: keuskupan, paroki, dan wilayah teritorial.",
      menu: {
        diocese: "Keuskupan",
        parish: "Paroki",
        territorial: "Wilayah teritorial",
      },
    },
    web: {
      name: "Website",
      description: "Situs publik paroki: halaman, berita, dan pengumuman.",
      menu: {
        halaman: "Halaman",
        berita: "Berita",
        pengumuman: "Pengumuman",
        galeri: "Galeri",
        pengaturan: "Pengaturan situs",
      },
    },
    sacra: {
      name: "Sakramen",
      description: "Pencatatan dan arsip sakramen umat.",
      menu: {
        baptis: "Baptis",
        "komuni-pertama": "Komuni pertama",
        krisma: "Krisma",
        perkawinan: "Perkawinan",
        laporan: "Laporan",
      },
    },
    fin: {
      name: "Keuangan",
      description: "Kas, kolekte, dan anggaran paroki.",
      groups: {
        transaksi: "Transaksi",
      },
      menu: {
        penerimaan: "Penerimaan",
        pengeluaran: "Pengeluaran",
        kolekte: "Kolekte",
        anggaran: "Anggaran",
        laporan: "Laporan",
      },
    },
    par: {
      name: "Umat",
      description: "Data umat, keluarga, dan lingkungan.",
      menu: {
        daftar: "Daftar umat",
        keluarga: "Keluarga",
        lingkungan: "Lingkungan",
        pengurus: "Pengurus",
      },
    },
    act: {
      name: "Kegiatan",
      description: "Kalender, jadwal misa, dan acara paroki.",
      menu: {
        kalender: "Kalender",
        "jadwal-misa": "Jadwal misa",
        acara: "Acara",
        kepanitiaan: "Kepanitiaan",
      },
    },
  },
};
