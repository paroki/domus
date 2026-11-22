import { useTranslation } from "react-i18next";
import { ChannelSettingView } from "../components/ChannelSettingView";
import { CHANNEL_GROUP_NAMES } from "../model";

export function meta() {
  return [{ title: "Pengaturan Telegram — Domus" }];
}

export default function TelegramPage() {
  const { t } = useTranslation();
  return (
    <ChannelSettingView
      groupName={CHANNEL_GROUP_NAMES.telegram}
      title={t("settings.channels.telegramTitle")}
      subtitle={t("settings.channels.telegramSubtitle")}
    />
  );
}
