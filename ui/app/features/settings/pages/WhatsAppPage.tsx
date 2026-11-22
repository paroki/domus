import { useTranslation } from "react-i18next";
import { ChannelSettingView } from "../components/ChannelSettingView";
import { CHANNEL_GROUP_NAMES } from "../model";

export function meta() {
  return [{ title: "Pengaturan WhatsApp — Domus" }];
}

export default function WhatsAppPage() {
  const { t } = useTranslation();
  return (
    <ChannelSettingView
      groupName={CHANNEL_GROUP_NAMES.whatsapp}
      title={t("settings.channels.whatsappTitle")}
      subtitle={t("settings.channels.whatsappSubtitle")}
    />
  );
}
