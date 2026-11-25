import { ArrowLeft } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import { useUpdateCurrentParokiProfile } from "../api/actions";
import { useCurrentIdentity, useCurrentParoki } from "../api/queries";
import { isParokiAdmin } from "../model";

export function meta() {
  return [{ title: "Profil Paroki — Domus" }];
}

export default function ParokiProfilePage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const canManage = isParokiAdmin(me?.roles);
  const {
    data: paroki,
    isLoading,
    error: parokiError,
  } = useCurrentParoki(Boolean(me?.activeParokiId));
  const updateProfile = useUpdateCurrentParokiProfile();

  const [nama, setNama] = useState("");
  const [pelindung, setPelindung] = useState("");
  const [pastor, setPastor] = useState("");
  const [telepon, setTelepon] = useState("");
  const [email, setEmail] = useState("");
  const [alamat, setAlamat] = useState("");
  const [website, setWebsite] = useState("");
  const [profileSaved, setProfileSaved] = useState(false);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);

  useEffect(() => {
    if (!paroki) return;
    setNama(paroki.nama ?? "");
    setPelindung(paroki.pelindung ?? "");
    setPastor(paroki.pastor ?? "");
    setTelepon(paroki.telepon ?? "");
    setEmail(paroki.email ?? "");
    setAlamat(paroki.alamat ?? "");
    setWebsite(paroki.website ?? "");
  }, [paroki]);

  async function handleSaveProfile(e: FormEvent) {
    e.preventDefault();
    setFeedbackError(null);
    setProfileSaved(false);
    try {
      await updateProfile.mutateAsync({
        nama: nama.trim(),
        pelindung: pelindung.trim(),
        pastor: pastor.trim(),
        telepon: telepon.trim(),
        email: email.trim(),
        alamat: alamat.trim(),
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
          to="/ref"
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

      {!me?.activeParokiId ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.noActiveParish")}
        </div>
      ) : !canManage ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenParishAdmin")}
        </div>
      ) : isLoading ? (
        <p className="text-sm text-muted">{t("auth.loading")}</p>
      ) : parokiError ? (
        <p className="text-sm text-danger">{parokiError.message}</p>
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
                {paroki?.nama} ({paroki?.kode})
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
                  value={nama}
                  onChange={(e) => setNama(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.patronSaintLabel")}
                </span>
                <input
                  value={pelindung}
                  onChange={(e) => setPelindung(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.pastorNameLabel")}
                </span>
                <input
                  value={pastor}
                  onChange={(e) => setPastor(e.target.value)}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-fg"
                />
              </label>
              <label className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-fg">
                  {t("ref.phoneLabel")}
                </span>
                <input
                  value={telepon}
                  onChange={(e) => setTelepon(e.target.value)}
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
                value={alamat}
                onChange={(e) => setAlamat(e.target.value)}
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
