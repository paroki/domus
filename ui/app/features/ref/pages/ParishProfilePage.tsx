import { ArrowLeft, Trash2, UserPlus } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useAddParishMember,
  useRemoveParishMember,
  useUpdateCurrentParishProfile,
  useUpdateParishMember,
} from "../api/actions";
import {
  useCurrentIdentity,
  useCurrentParish,
  useParishMembersSearch,
} from "../api/queries";
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
  const addMember = useAddParishMember();
  const updateMember = useUpdateParishMember();
  const removeMember = useRemoveParishMember();

  const [name, setName] = useState("");
  const [patronSaint, setPatronSaint] = useState("");
  const [pastorName, setPastorName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [address, setAddress] = useState("");
  const [website, setWebsite] = useState("");
  const [profileSaved, setProfileSaved] = useState(false);

  const [memberEmail, setMemberEmail] = useState("");
  const [memberName, setMemberName] = useState("");
  const [memberPhone, setMemberPhone] = useState("");
  const [memberRole, setMemberRole] = useState<"admin" | "member">("member");
  const [memberSearch, setMemberSearch] = useState("");
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

  const membersQuery = useParishMembersSearch(
    parish?.id ?? "",
    { q: memberSearch || undefined },
    Boolean(parish?.id),
  );
  const members =
    membersQuery.data?.pages.flatMap((page) => page.data ?? []) ?? [];

  async function handleSaveProfile(e: FormEvent) {
    e.preventDefault();
    setFeedbackError(null);
    setProfileSaved(false);
    try {
      await updateProfile.mutateAsync({
        name: name.trim(),
        patronSaint: patronSaint.trim() || undefined,
        pastorName: pastorName.trim() || undefined,
        phone: phone.trim() || undefined,
        email: email.trim() || undefined,
        address: address.trim() || undefined,
        website: website.trim() || undefined,
      });
      setProfileSaved(true);
    } catch (err) {
      setFeedbackError(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  async function handleAddMember(e: FormEvent) {
    e.preventDefault();
    if (!parish?.id) return;
    setFeedbackError(null);
    try {
      await addMember.mutateAsync({
        parishId: parish.id,
        body: {
          email: memberEmail.trim(),
          name: memberName.trim() || undefined,
          phone: memberPhone.trim() || undefined,
          role: memberRole,
        },
      });
      setMemberEmail("");
      setMemberName("");
      setMemberPhone("");
      setMemberRole("member");
    } catch (err) {
      setFeedbackError(err instanceof Error ? err.message : "Gagal menambah.");
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

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-fg">
                {t("ref.membersTitle")}
              </h2>
              <input
                type="search"
                value={memberSearch}
                onChange={(e) => setMemberSearch(e.target.value)}
                placeholder={t("ref.searchMembersPlaceholder")}
                className="h-9 w-60 rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted"
              />
            </div>

            <form
              onSubmit={handleAddMember}
              className="grid gap-3 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-5"
            >
              <input
                required
                type="email"
                value={memberEmail}
                onChange={(e) => setMemberEmail(e.target.value)}
                placeholder={t("ref.emailLabel")}
                className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <input
                value={memberName}
                onChange={(e) => setMemberName(e.target.value)}
                placeholder={t("ref.nameLabel")}
                className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <input
                value={memberPhone}
                onChange={(e) => setMemberPhone(e.target.value)}
                placeholder={t("ref.phoneLabel")}
                className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              />
              <select
                value={memberRole}
                onChange={(e) =>
                  setMemberRole(e.target.value as "admin" | "member")
                }
                className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
              >
                <option value="member">{t("ref.roleMember")}</option>
                <option value="admin">{t("ref.roleAdmin")}</option>
              </select>
              <button
                type="submit"
                disabled={addMember.isPending}
                className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                <UserPlus aria-hidden className="size-4" />
                {t("ref.addMember")}
              </button>
            </form>

            <div className="divide-y divide-line">
              {members.map((m) => (
                <div
                  key={m.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div>
                    <p className="text-sm font-semibold text-fg">{m.name}</p>
                    <p className="text-xs text-muted">
                      {m.email}
                      {m.phone ? ` · ${m.phone}` : ""}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <select
                      aria-label={`${t("ref.roleLabel")} ${m.name}`}
                      value={m.role}
                      disabled={updateMember.isPending}
                      onChange={async (e) => {
                        if (!parish?.id || !m.userId) return;
                        setFeedbackError(null);
                        try {
                          await updateMember.mutateAsync({
                            parishId: parish.id,
                            userId: m.userId,
                            body: {
                              role: e.target.value as "admin" | "member",
                            },
                          });
                        } catch (err) {
                          setFeedbackError(
                            err instanceof Error
                              ? err.message
                              : "Gagal memperbarui peran.",
                          );
                        }
                      }}
                      className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs font-medium text-fg"
                    >
                      <option value="member">{t("ref.roleMember")}</option>
                      <option value="admin">{t("ref.roleAdmin")}</option>
                    </select>
                    <button
                      type="button"
                      aria-label={`${t("ref.removeMember")} ${m.name}`}
                      disabled={removeMember.isPending}
                      onClick={async () => {
                        if (!parish?.id || !m.userId) return;
                        setFeedbackError(null);
                        try {
                          await removeMember.mutateAsync({
                            parishId: parish.id,
                            userId: m.userId,
                          });
                        } catch (err) {
                          setFeedbackError(
                            err instanceof Error
                              ? err.message
                              : "Gagal menghapus anggota.",
                          );
                        }
                      }}
                      className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10 disabled:opacity-50"
                    >
                      <Trash2 aria-hidden className="size-4" />
                    </button>
                  </div>
                </div>
              ))}
              {members.length === 0 && (
                <p className="py-4 text-center text-sm text-muted">
                  {t("ref.membersEmpty")}
                </p>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  );
}
