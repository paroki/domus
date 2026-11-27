import { useState } from "react";
import { useTranslation } from "react-i18next";
import { InductivePageHeader } from "~/shared/ui/iui";
import { useUpdateCurrentParokiProfile } from "../api/actions";
import { useCurrentIdentity, useCurrentParoki } from "../api/queries";
import type { UpdateParokiProfileRequest } from "../api/types";
import {
  ParokiDomisiliCard,
  ParokiIdentitasCard,
  ParokiKontakCard,
  type ParokiSectionKey,
  ParokiStrukturReadOnlyCard,
} from "../components/ParokiSectionCards";
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

  const [editingSection, setEditingSection] = useState<ParokiSectionKey | null>(
    null,
  );
  const [profileSaved, setProfileSaved] = useState(false);
  const [feedbackError, setFeedbackError] = useState<string | null>(null);

  async function handleSaveSection(body: UpdateParokiProfileRequest) {
    setFeedbackError(null);
    setProfileSaved(false);
    try {
      await updateProfile.mutateAsync(body);
      setEditingSection(null);
      setProfileSaved(true);
    } catch (err) {
      setFeedbackError(err instanceof Error ? err.message : "Gagal menyimpan.");
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Administrasi Paroki", to: "/settings/paroki" },
          { label: t("ref.profileTitle") },
        ]}
        title={t("ref.profileTitle")}
        subtitle={t("ref.profileSubtitle")}
      />

      {!canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          {t("ref.forbiddenParishAdmin")}
        </div>
      ) : !me?.activeParokiId ? (
        <div className="rounded-xl border border-line bg-surface p-4 text-sm text-muted sm:p-6">
          {t("ref.noActiveParish")}
        </div>
      ) : isLoading ? (
        <p className="text-sm text-muted">{t("ref.loading")}</p>
      ) : parokiError ? (
        <div
          role="alert"
          className="rounded-xl border border-danger-fg/30 bg-danger-bg p-4 text-sm text-danger-fg"
        >
          {parokiError instanceof Error
            ? parokiError.message
            : "Gagal memuat paroki."}
        </div>
      ) : paroki ? (
        <>
          {feedbackError ? (
            <div
              role="alert"
              className="rounded-xl border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
            >
              {feedbackError}
            </div>
          ) : null}

          {profileSaved ? (
            <div
              role="status"
              className="rounded-xl border border-success-fg/30 bg-success-bg px-4 py-3 text-sm text-success-fg"
            >
              {t("ref.saved")}
            </div>
          ) : null}

          <ParokiStrukturReadOnlyCard paroki={paroki} />

          <ParokiIdentitasCard
            paroki={paroki}
            editing={editingSection === "identitas"}
            pending={updateProfile.isPending}
            onEdit={() => setEditingSection("identitas")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <ParokiDomisiliCard
            paroki={paroki}
            editing={editingSection === "domisili"}
            pending={updateProfile.isPending}
            onEdit={() => setEditingSection("domisili")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />

          <ParokiKontakCard
            paroki={paroki}
            editing={editingSection === "kontak"}
            pending={updateProfile.isPending}
            onEdit={() => setEditingSection("kontak")}
            onCancel={() => setEditingSection(null)}
            onSave={handleSaveSection}
          />
        </>
      ) : null}
    </div>
  );
}
