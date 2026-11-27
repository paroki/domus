import { ArrowLeft } from "lucide-react";
import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import {
  useCreateLiturgiPetugas,
  useUpdateLiturgiPetugas,
} from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import {
  useLiturgiOrgs,
  useLiturgiParokiMembers,
  useLiturgiSchedule,
} from "../api/queries";
import { getLiturgiMinistryMeta, LITURGI_MINISTRIES } from "../model";
import type { Route } from "./+types/LiturgiPetugasFormPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.petugasFormCreateTitle")} · Domus` }];
}

type AssignmentMode = "user" | "org" | "manual";

export default function LiturgiPetugasFormPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { eventId = "", petugasId } = useParams();
  const [searchParams] = useSearchParams();
  const isEdit = Boolean(petugasId);

  const { data: me } = useMe();
  const parokiId = me?.activeParokiId ?? "";
  const scheduleQuery = useLiturgiSchedule(eventId, Boolean(parokiId));
  const schedule = scheduleQuery.data;

  const existingPetugas = useMemo(
    () => schedule?.petugas?.find((p) => p.id === petugasId),
    [schedule?.petugas, petugasId],
  );

  const [memberSearch, setMemberSearch] = useState("");
  const membersQuery = useLiturgiParokiMembers(
    parokiId,
    memberSearch,
    Boolean(parokiId),
  );
  const orgsQuery = useLiturgiOrgs(Boolean(parokiId));

  const createPetugas = useCreateLiturgiPetugas(eventId);
  const updatePetugas = useUpdateLiturgiPetugas(eventId);

  const roleOptions = useMemo(() => {
    const allowed = schedule?.manageableRoles ?? [];
    if (allowed.length === 0) return LITURGI_MINISTRIES;
    return LITURGI_MINISTRIES.filter((m) => allowed.includes(m.slug));
  }, [schedule?.manageableRoles]);

  const [role, setRole] = useState("");
  const [slotLabel, setSlotLabel] = useState("");
  const [mode, setMode] = useState<AssignmentMode>("user");
  const [userId, setUserId] = useState("");
  const [orgId, setOrgId] = useState("");
  const [manualName, setManualName] = useState("");
  const [urutan, setUrutan] = useState(0);
  const [note, setNote] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (isEdit && existingPetugas) {
      setRole(existingPetugas.role ?? "");
      setSlotLabel(existingPetugas.slotLabel ?? "");
      if (existingPetugas.userId) {
        setMode("user");
        setUserId(existingPetugas.userId);
      } else if (existingPetugas.orgId) {
        setMode("org");
        setOrgId(existingPetugas.orgId);
      } else {
        setMode("manual");
        setManualName(existingPetugas.name ?? "");
      }
      setUrutan(existingPetugas.urutan ?? 0);
      setNote(existingPetugas.note ?? "");
      return;
    }

    if (!isEdit) {
      const requestedRole = searchParams.get("role") ?? "";
      const validRequested = roleOptions.find((r) => r.slug === requestedRole);
      setRole(validRequested?.slug ?? roleOptions[0]?.slug ?? "lektor");
    }
  }, [isEdit, existingPetugas, searchParams, roleOptions]);

  const members = membersQuery.data ?? [];
  const orgs = orgsQuery.data ?? [];

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    if (mode === "user" && !userId) {
      setErrorMsg(t("liturgi.validationMemberRequired"));
      return;
    }
    if (mode === "org" && !orgId) {
      setErrorMsg(t("liturgi.validationOrgRequired"));
      return;
    }
    if (mode === "manual" && !manualName.trim()) {
      setErrorMsg(t("liturgi.validationNameRequired"));
      return;
    }

    const selectedMember = members.find((m) => m.userId === userId);
    const selectedOrg = orgs.find((o) => o.id === orgId);
    const resolvedName =
      mode === "user"
        ? selectedMember?.name
        : mode === "org"
          ? selectedOrg?.name
          : manualName.trim();

    try {
      if (isEdit && petugasId) {
        await updatePetugas.mutateAsync({
          petugasId,
          body: {
            role,
            slotLabel: slotLabel.trim(),
            urutan,
            note: note.trim(),
            userId: mode === "user" ? userId : undefined,
            orgId: mode === "org" ? orgId : undefined,
            clearUser: mode !== "user" ? true : undefined,
            clearOrg: mode !== "org" ? true : undefined,
            name: resolvedName,
          },
        });
      } else {
        await createPetugas.mutateAsync({
          role,
          slotLabel: slotLabel.trim() || undefined,
          urutan,
          note: note.trim() || undefined,
          userId: mode === "user" ? userId : undefined,
          orgId: mode === "org" ? orgId : undefined,
          name: mode === "manual" ? manualName.trim() : resolvedName,
        });
      }
      navigate(`/liturgi/${eventId}`);
    } catch (err) {
      setErrorMsg(resolveLiturgiError(err));
    }
  }

  const pageTitle = isEdit
    ? t("liturgi.petugasFormEditTitle")
    : t("liturgi.petugasFormCreateTitle");
  const pending = createPetugas.isPending || updatePetugas.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi"), to: "/liturgi" },
          {
            label: schedule?.title ?? t("liturgi.detailTitle"),
            to: `/liturgi/${eventId}`,
          },
          { label: pageTitle },
        ]}
        title={pageTitle}
        subtitle={t("liturgi.petugasFormSubtitle")}
        action={
          <Link
            to={`/liturgi/${eventId}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            <ArrowLeft aria-hidden className="size-4" />
            {t("liturgi.backToDetail")}
          </Link>
        }
      />

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-5">
        {errorMsg && (
          <Card className="border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
            {errorMsg}
          </Card>
        )}

        <InductiveFormSection title={pageTitle}>
          <div className="space-y-4">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.petugasFieldRole")}
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              >
                {roleOptions.map((opt) => (
                  <option key={opt.slug} value={opt.slug}>
                    {opt.label}
                  </option>
                ))}
                {role && !roleOptions.some((r) => r.slug === role) && (
                  <option value={role}>
                    {getLiturgiMinistryMeta(role)?.label ?? role}
                  </option>
                )}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.petugasFieldSlotLabel")}
              <input
                type="text"
                value={slotLabel}
                onChange={(e) => setSlotLabel(e.target.value)}
                placeholder={t("liturgi.petugasFieldSlotPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <div className="space-y-2">
              <span className="block text-sm font-medium text-fg">
                {t("liturgi.petugasFieldMode")}
              </span>
              <div className="flex flex-col gap-2 sm:flex-row">
                {(
                  [
                    { key: "user", label: t("liturgi.petugasModeUser") },
                    { key: "org", label: t("liturgi.petugasModeOrg") },
                    { key: "manual", label: t("liturgi.petugasModeManual") },
                  ] as const
                ).map((item) => (
                  <Button
                    key={item.key}
                    type="button"
                    variant={mode === item.key ? "primary" : "outline"}
                    onClick={() => setMode(item.key)}
                    className="h-11 w-full sm:w-auto"
                  >
                    {item.label}
                  </Button>
                ))}
              </div>
            </div>

            {mode === "user" && (
              <div className="space-y-2">
                <input
                  type="search"
                  value={memberSearch}
                  onChange={(e) => setMemberSearch(e.target.value)}
                  placeholder={t("liturgi.petugasSearchMemberPlaceholder")}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                />
                <select
                  value={userId}
                  onChange={(e) => setUserId(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                >
                  <option value="">{t("liturgi.petugasSelectMember")}</option>
                  {members.map((m) => (
                    <option key={m.userId} value={m.userId}>
                      {m.name} {m.email ? `(${m.email})` : ""}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {mode === "org" && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.petugasModeOrg")}
                <select
                  value={orgId}
                  onChange={(e) => setOrgId(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                >
                  <option value="">{t("liturgi.petugasSelectOrg")}</option>
                  {orgs.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name}
                    </option>
                  ))}
                </select>
              </label>
            )}

            {mode === "manual" && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.petugasFieldManualName")}
                <input
                  type="text"
                  value={manualName}
                  onChange={(e) => setManualName(e.target.value)}
                  placeholder={t("liturgi.petugasFieldManualPlaceholder")}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                />
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.petugasFieldUrutan")}
              <input
                type="number"
                value={urutan}
                onChange={(e) => setUrutan(Number(e.target.value) || 0)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.petugasFieldNote")}
              <textarea
                rows={3}
                value={note}
                onChange={(e) => setNote(e.target.value)}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-base text-fg sm:text-sm"
              />
            </label>
          </div>
        </InductiveFormSection>

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Link
            to={`/liturgi/${eventId}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            {t("liturgi.ministriesCancel")}
          </Link>
          <Button
            type="submit"
            disabled={pending}
            className="h-11 w-full sm:w-auto"
          >
            {pending ? t("liturgi.creating") : t("liturgi.petugasFormSave")}
          </Button>
        </div>
      </form>
    </div>
  );
}
