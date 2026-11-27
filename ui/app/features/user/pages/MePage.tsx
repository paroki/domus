import { LogOut, Save } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { DEFAULT_TIMEZONE, getTimezoneOptions } from "~/shared/util/timezone";
import { useLogout, useUpdateMe } from "../api/actions";
import { useMe } from "../api/queries";
import { AccountList } from "../components/AccountList";
import { PwaDeviceCard } from "../components/PwaDeviceCard";
import { SessionList } from "../components/SessionList";
import { TokenList } from "../components/TokenList";
import { getBirthDateBounds, USER_GENDERS } from "../model";
import type { Route } from "./+types/MePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("me.title")} · Domus` }];
}

export default function MePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const updateMe = useUpdateMe();
  const logout = useLogout();

  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [gender, setGender] = useState("");
  const [birthDate, setBirthDate] = useState("");
  const [timezone, setTimezone] = useState(DEFAULT_TIMEZONE);
  const [saved, setSaved] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (me) {
      setName(me.name ?? "");
      setPhone(me.phone ?? "");
      setGender(me.gender ?? "");
      setBirthDate(me.birthDate ?? "");
      setTimezone(me.timezone || DEFAULT_TIMEZONE);
    }
  }, [me]);

  const birthDateBounds = getBirthDateBounds();
  const timezoneOptions = getTimezoneOptions(timezone);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaved(false);
    setErrorMsg(null);
    try {
      await updateMe.mutateAsync({
        name: name.trim(),
        phone: phone.trim(),
        gender,
        birthDate,
        timezone,
      });
      setSaved(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setErrorMsg(err.details[0]?.message ?? err.message);
      }
    }
  }

  async function handleLogout() {
    await logout.mutateAsync();
    navigate("/login", { replace: true });
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg">{t("me.title")}</h1>
          <p className="text-sm text-muted">{t("me.subtitle")}</p>
        </div>
        <Button
          variant="outline"
          disabled={logout.isPending}
          onClick={handleLogout}
        >
          <LogOut aria-hidden className="size-4" />
          {t("me.logout")}
        </Button>
      </div>

      <Card className="p-6">
        <h2 className="mb-4 text-lg font-bold text-fg">
          {t("me.profileTitle")}
        </h2>

        {errorMsg && (
          <div
            role="alert"
            className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
          >
            {errorMsg}
          </div>
        )}
        {saved && (
          <div
            role="status"
            className="mb-4 rounded-lg bg-success-bg px-3.5 py-2.5 text-sm font-medium text-success-fg"
          >
            {t("me.profileSaved")}
          </div>
        )}

        <form onSubmit={handleSave} className="flex flex-col gap-4">
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.emailLabel")}
            <input
              type="email"
              disabled
              value={me?.email ?? ""}
              className="h-11 rounded-lg border border-line bg-neutral-bg px-3.5 text-sm text-muted"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.nameLabel")}
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.phoneLabel")}
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder={t("me.phonePlaceholder")}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.genderLabel")}
            <select
              required
              value={gender}
              onChange={(e) => setGender(e.target.value)}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
            >
              <option value="" disabled>
                {t("me.genderPlaceholder")}
              </option>
              {USER_GENDERS.map((g) => (
                <option key={g} value={g}>
                  {g === "pria" ? t("me.genderPria") : t("me.genderWanita")}
                </option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.birthDateLabel")}
            <input
              type="date"
              required
              min={birthDateBounds.min}
              max={birthDateBounds.max}
              value={birthDate}
              onChange={(e) => setBirthDate(e.target.value)}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
            />
          </label>
          <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
            {t("me.timezoneLabel")}
            <select
              required
              value={timezone}
              onChange={(e) => setTimezone(e.target.value)}
              className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
            >
              {timezoneOptions.map((tz) => (
                <option key={tz.value} value={tz.value}>
                  {tz.label}
                </option>
              ))}
            </select>
          </label>
          <div>
            <Button type="submit" disabled={updateMe.isPending}>
              <Save aria-hidden className="size-4" />
              {updateMe.isPending ? t("me.savingProfile") : t("me.saveProfile")}
            </Button>
          </div>
        </form>
      </Card>

      <PwaDeviceCard />
      <SessionList />
      <AccountList />
      <TokenList />
    </div>
  );
}
