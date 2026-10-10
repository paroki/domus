import "i18next";
import type { resources } from "./index";

// Key terjemahan jadi type-safe: `t("login.title")` dicek saat typecheck.
declare module "i18next" {
  interface CustomTypeOptions {
    defaultNS: "translation";
    resources: (typeof resources)["id"];
  }
}
