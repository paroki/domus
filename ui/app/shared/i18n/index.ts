import i18n from "i18next";
import { initReactI18next } from "react-i18next";

import id from "./locales/id";

// ponytail: single locale bundled inline; add locales/<lng>.ts + a language
// switcher when a second language is needed.
void i18n.use(initReactI18next).init({
  resources: { id: { translation: id } },
  lng: "id",
  fallbackLng: "id",
  interpolation: { escapeValue: false },
});

export default i18n;
