import { CheckOutlined, GlobalOutlined } from "@ant-design/icons";
import { Button, Dropdown } from "antd";
import { useTranslation } from "react-i18next";
import { getCurrentLanguage, type Language, SUPPORTED_LANGUAGES } from "~/i18n";

/** Nama bahasa ditulis dalam bahasanya sendiri, tidak diterjemahkan. */
const LANGUAGE_NAMES: Record<Language, string> = {
  id: "Bahasa Indonesia",
  en: "English",
};

/** Pilih bahasa tampilan. Pilihan disimpan di browser (localStorage). */
export function LanguageSwitcher() {
  const { t, i18n } = useTranslation();
  const current = getCurrentLanguage();

  return (
    <Dropdown
      trigger={["click"]}
      placement="bottomRight"
      menu={{
        selectedKeys: [current],
        items: SUPPORTED_LANGUAGES.map((lang) => ({
          key: lang,
          lang,
          label: LANGUAGE_NAMES[lang],
          icon: lang === current ? <CheckOutlined /> : <span />,
        })),
        onClick: ({ key }) => void i18n.changeLanguage(key),
      }}
    >
      <Button
        type="text"
        icon={<GlobalOutlined />}
        aria-label={t("common.language")}
        aria-haspopup="menu"
      >
        {current.toUpperCase()}
      </Button>
    </Dropdown>
  );
}
