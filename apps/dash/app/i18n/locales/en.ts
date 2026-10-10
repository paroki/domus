import type { id } from "./id";

/** English translation. Must mirror the structure of `id`. */
export const en: typeof id = {
  common: {
    appName: "Domus",
    back: "Back",
    login: "Sign in",
    logout: "Sign out",
    loggingOut: "Signing out",
    loadingAccount: "Loading account",
    allModules: "All modules",
    language: "Language",
  },
  roles: {
    admin: "Admin",
  },
  error: {
    title: "Oops!",
    unexpected: "An unexpected error occurred.",
    notFoundTitle: "404",
    notFound: "The requested page could not be found.",
    generic: "Error",
  },
  underConstruction: {
    title: "Under construction",
    description: "This feature is still under development.",
  },
  launcher: {
    pageTitle: "Home",
    title: "All modules",
    subtitle: "Pick a module to get started.",
  },
  userMenu: {
    accountMenu: "Account menu for {{name}}",
  },
  login: {
    pageTitle: "Sign in",
    title: "Sign in to Domus",
    subtitle: "Manage dioceses, parishes, and neighborhoods in one place.",
    continueWithGoogle: "Continue with Google",
    continueWithGithub: "Continue with GitHub",
    noAccess:
      "Don't have access yet? Contact your diocese or parish administrator.",
    agreement:
      "By signing in, you agree to the <terms>Terms of Service</terms> and <privacy>Privacy Policy</privacy>.",
    errors: {
      access_denied: "Sign-in was cancelled. Try again whenever you're ready.",
      account_not_linked:
        "This email is already registered with another provider. Sign in with the provider you used before.",
      oauth: "Sign-in failed. Please try again in a moment.",
    },
  },
  modules: {
    website: {
      name: "Website",
      description: "The parish public site: pages, news, and announcements.",
      menu: {
        halaman: "Pages",
        berita: "News",
        pengumuman: "Announcements",
        galeri: "Gallery",
        pengaturan: "Site settings",
      },
    },
    sakramen: {
      name: "Sacraments",
      description: "Records and archive of parishioners' sacraments.",
      menu: {
        baptis: "Baptism",
        "komuni-pertama": "First communion",
        krisma: "Confirmation",
        perkawinan: "Marriage",
        laporan: "Reports",
      },
    },
    keuangan: {
      name: "Finance",
      description: "Cash, collections, and the parish budget.",
      groups: {
        transaksi: "Transactions",
      },
      menu: {
        penerimaan: "Income",
        pengeluaran: "Expenses",
        kolekte: "Collections",
        anggaran: "Budget",
        laporan: "Reports",
      },
    },
    umat: {
      name: "Parishioners",
      description: "Parishioner, family, and neighborhood data.",
      menu: {
        daftar: "Parishioner list",
        keluarga: "Families",
        lingkungan: "Neighborhoods",
        pengurus: "Committee",
      },
    },
    kegiatan: {
      name: "Activities",
      description: "Calendar, Mass schedule, and parish events.",
      menu: {
        kalender: "Calendar",
        "jadwal-misa": "Mass schedule",
        acara: "Events",
        kepanitiaan: "Committees",
      },
    },
  },
};
