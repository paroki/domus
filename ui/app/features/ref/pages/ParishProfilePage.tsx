import { ArrowLeft } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import { useUpdateCurrentParishProfile } from "../api/actions";
import { useCurrentIdentity, useCurrentParish } from "../api/queries";
import { isParishAdmin } from "../model";

export function meta() {
  return [{ title: "Profil Paroki — Domus" }];
}

export default function ParishProfilePage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const canManage = isParishAdmin(me?.roles);
  const {
    data: parish,
    isLoading,
    error: parishError,
  } = useCurrentParish(Boolean(me?.activeParishId));
  const updateProfile = useUpdateCurrentParishProfile();

  const [name, setName] = useState("");
  const [patronSaint, setPatronSaint] = useState("");
  const [pastorName, setPastorName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [address, setAddress] = useState("");
  const [website, setWebsite] = useState("");
  const [profileSaved, setProfileSaved] = useState(false);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);

  useEffect(() => {
    if (!parish) return;
    setName(parish.name ?? "");
    setPatronSaint(parish.patronSaint ?? "");
    setPastorName(parish.pastorName ?? "");
    setPhone(parish.phone ?? "");
    setEmail(parish.email ?? "");
    setAddress(parish.address ?? "");
    setWebsite(parish.website ?? "");
  }, [parish]);

  async function handleSaveProfile(e: FormEvent) {
    e.preventDefault();
    setFeedbackError(null);
    setProfileSaved(false);
    try {
      await updateProfile.mutateAsync({
        name: name.trim(),
        patronSaint: patronSaint.trim(),
        pastorName: pastorName.trim(),
        phone: phone.trim(),
        email: email.trim(),
        address: address.trim(),
        website: website.trim(),
      });
      setProfileSaved(true);
    } catch (err) {
      setFeedbackError(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-8 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/settings"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("ref.backToSettings")}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">
          {t("ref.profileTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.profileSubtitle")}</p>
      </div>

      {!me?.activeParishId ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.noActiveParish")}
        </div>
      ) : !canManage ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenParishAdmin")}
        </div>
      ) : isLoading ? (
        <p className="text-sm text-muted">{t("auth.loading")}</p>
      ) : parishError ? (
        <p className="text-sm text-danger">{parishError.message}</p>
      ) : (
        <>
          {feedbackError && (
            <div
              role="alert"
              className="rounded-xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger"
            >
              {feedbackError}
            </div>
          )}

          <form
            onSubmit={handleSaveProfile}
            className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-fg">
                {parish?.name} ({parish?.code})
              </h2>
              {profileSaved && (
                <span className="text-xs font-medium text-primary">
                  {t("ref.saved")}
                </span>
              )}
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.nameLabel")}
                </span>
                <input
                  required
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.patronSaintLabel")}
                </span>
                <input
                  value={patronSaint}
                  onChange={(e) => setPatronSaint(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.pastorNameLabel")}
                </span>
                <input
                  value={pastorName}
                  onChange={(e) => setPastorName(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.phoneLabel")}
                </span>
                <input
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  placeholder="+62812..."
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.emailLabel")}
                </span>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.websiteLabel")}
                </span>
                <input
                  value={website}
                  onChange={(e) => setWebsite(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
            </div>

            <label className="flex flex-col gap-1 text-sm">
              <span className="font-medium text-fg">
                {t("ref.addressLabel")}
              </span>
              <textarea
                rows={2}
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                className="rounded-lg border border-line bg-canvas p-3 text-fg"
              />
            </label>

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={updateProfile.isPending}
                className="h-10 rounded-lg bg-primary px-5 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                {updateProfile.isPending
                  ? t("ref.saving")
                  : t("ref.saveProfile")}
              </button>
            </div>
          </form>
        </>
      )}
    </div>
  );
}
