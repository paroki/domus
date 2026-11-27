import { LogOut, Pencil, Save, X } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { DEFAULT_TIMEZONE, getTimezoneOptions } from "~/shared/util/timezone";
import { useLogout, useUpdateMe } from "../api/actions";
import { useMe } from "../api/queries";
import { AccountList } from "../components/AccountList";
import { MeUmatCard } from "../components/MeUmatCard";
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

  const [isEditing, setIsEditing] = useState(false);
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

  function handleStartEdit() {
    if (me) {
      setName(me.name ?? "");
      setPhone(me.phone ?? "");
      setGender(me.gender ?? "");
      setBirthDate(me.birthDate ?? "");
      setTimezone(me.timezone || DEFAULT_TIMEZONE);
    }
    setErrorMsg(null);
    setSaved(false);
    setIsEditing(true);
  }

  function handleCancelEdit() {
    if (me) {
      setName(me.name ?? "");
      setPhone(me.phone ?? "");
      setGender(me.gender ?? "");
      setBirthDate(me.birthDate ?? "");
      setTimezone(me.timezone || DEFAULT_TIMEZONE);
    }
    setErrorMsg(null);
    setIsEditing(false);
  }

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
      setIsEditing(false);
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

  const genderDisplay =
    me?.gender === "pria"
      ? t("me.genderPria")
      : me?.gender === "wanita"
        ? t("me.genderWanita")
        : t("me.notSet");

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("me.homeBreadcrumb"), to: "/" },
          { label: t("me.title") },
        ]}
        title={t("me.title")}
        subtitle={t("me.subtitle")}
        action={
          <Button
            variant="outline"
            disabled={logout.isPending}
            onClick={handleLogout}
            className="w-full sm:w-auto"
          >
            <LogOut aria-hidden className="size-4" />
            {t("me.logout")}
          </Button>
        }
      />

      <Card className="p-4 sm:p-6">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-bold text-fg">{t("me.profileTitle")}</h2>
          {!isEditing && (
            <Button
              type="button"
              variant="outline"
              onClick={handleStartEdit}
              className="h-11"
            >
              <Pencil aria-hidden className="size-4" />
              {t("me.editProfile")}
            </Button>
          )}
        </div>

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

        {isEditing ? (
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
            <div className="flex flex-col-reverse gap-2 pt-2 sm:flex-row sm:items-center sm:justify-end">
              <Button
                type="button"
                variant="outline"
                disabled={updateMe.isPending}
                onClick={handleCancelEdit}
                className="w-full sm:w-auto"
              >
                <X aria-hidden className="size-4" />
                {t("me.cancelEdit")}
              </Button>
              <Button
                type="submit"
                disabled={updateMe.isPending}
                className="w-full sm:w-auto"
              >
                <Save aria-hidden className="size-4" />
                {updateMe.isPending
                  ? t("me.savingProfile")
                  : t("me.saveProfile")}
              </Button>
            </div>
          </form>
        ) : (
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.emailLabel")}
              </dt>
              <dd
                data-testid="profile-email"
                className="mt-0.5 truncate text-sm font-medium text-fg"
              >
                {me?.email || t("me.notSet")}
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.nameLabel")}
              </dt>
              <dd
                data-testid="profile-name"
                className="mt-0.5 truncate text-sm font-medium text-fg"
              >
                {me?.name || t("me.notSet")}
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.phoneLabel")}
              </dt>
              <dd
                data-testid="profile-phone"
                className="mt-0.5 truncate text-sm font-medium text-fg"
              >
                {me?.phone || t("me.notSet")}
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.genderLabel")}
              </dt>
              <dd
                data-testid="profile-gender"
                className="mt-0.5 text-sm font-medium text-fg"
              >
                {genderDisplay}
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.birthDateLabel")}
              </dt>
              <dd
                data-testid="profile-birthDate"
                className="mt-0.5 font-mono text-sm font-medium text-fg"
              >
                {me?.birthDate || t("me.notSet")}
              </dd>
            </div>
            <div className="min-w-0">
              <dt className="text-xs font-medium text-muted">
                {t("me.timezoneLabel")}
              </dt>
              <dd
                data-testid="profile-timezone"
                className="mt-0.5 truncate text-sm font-medium text-fg"
              >
                {me?.timezone || DEFAULT_TIMEZONE}
              </dd>
            </div>
          </dl>
        )}
      </Card>

      <MeUmatCard verified={me?.verified} />
      <PwaDeviceCard />
      <SessionList />
      <AccountList />
      <TokenList />
    </div>
  );
}
