const id = {
  home: {
    title: "Selamat datang di Domus",
    subtitle: "Administrasi paroki, terdigitalisasi.",
    menuLabel: "Menu aplikasi",
    empty: 'Menu "{{q}}" tidak ditemukan.',
  },
  header: {
    homeLabel: "Domus, beranda",
    searchLabel: "Cari menu",
    searchPlaceholder: "Cari menu…",
    account: "Akun saya",
  },
  theme: {
    modeLabel: "Mode tampilan: {{mode}}. Klik untuk mengganti.",
    mode: { system: "Ikuti sistem", light: "Terang", dark: "Gelap" },
    colorLabel: "Warna tema: {{theme}}. Klik untuk mengganti.",
    color: { green: "Hijau Sakristi", terracotta: "Terakota" },
  },
  underConstruction: {
    badge: "Segera hadir",
    title: "Halaman {{name}} sedang dibangun",
    description:
      "Fitur ini masih dalam pengerjaan. Kami sedang menyiapkannya supaya segera bisa dipakai.",
    back: "Kembali ke beranda",
  },
  apps: {
    organization: {
      name: "Organisasi",
      description: "Struktur organisasi, wilayah, dan lingkungan paroki.",
    },
    event: {
      name: "Kegiatan",
      description: "Jadwal misa dan kegiatan paroki.",
    },
    sacrament: {
      name: "Sakramen",
      description: "Pencatatan baptis, komuni, krisma, dan perkawinan.",
    },
    parishioner: {
      name: "Data Umat",
      description: "Data umat dan keluarga paroki.",
    },
    users: {
      name: "Pengguna",
      description: "Akun dan hak akses pengguna aplikasi.",
    },
    settings: {
      name: "Pengaturan",
      description: "Preferensi dan konfigurasi aplikasi.",
    },
  },
} as const;

export default id;
