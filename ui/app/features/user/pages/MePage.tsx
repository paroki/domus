import { LogOut, Save } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { ApiError } from "~/shared/api/client";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useLogout, useUpdateMe } from "../api/actions";
import { useMe } from "../api/queries";
import { AccountList } from "../components/AccountList";
import { SessionList } from "../components/SessionList";
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
  const [saved, setSaved] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (me) {
      setName(me.name ?? "");
      setPhone(me.phone ?? "");
    }
  }, [me]);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaved(false);
    setErrorMsg(null);
    try {
      await updateMe.mutateAsync({
        name: name.trim(),
        phone: phone.trim(),
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
          <div>
            <Button type="submit" disabled={updateMe.isPending}>
              <Save aria-hidden className="size-4" />
              {updateMe.isPending ? t("me.savingProfile") : t("me.saveProfile")}
            </Button>
          </div>
        </form>
      </Card>

      <SessionList />
      <AccountList />
    </div>
  );
}
