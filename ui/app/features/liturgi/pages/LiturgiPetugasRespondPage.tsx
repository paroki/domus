import { ArrowLeft } from "lucide-react";
import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useRespondLiturgiPetugas } from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import {
  useActiveParoki,
  useLiturgiOrgs,
  useLiturgiParokiMembers,
  useLiturgiSchedule,
} from "../api/queries";
import {
  DEFAULT_PAROKI_TIMEZONE,
  formatLiturgiRange,
  getLiturgiMinistryMeta,
} from "../model";
import type { Route } from "./+types/LiturgiPetugasRespondPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.petugasRespondTitle")} · Domus` }];
}

type RespondStatus = "confirmed" | "declined";
type SubstituteMode = "none" | "user" | "org" | "manual";

export default function LiturgiPetugasRespondPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { eventId = "", petugasId = "" } = useParams();
  const { data: me } = useMe();
  const parokiId = me?.activeParokiId ?? "";

  const activeParokiQuery = useActiveParoki();
  const parokiTimezone =
    activeParokiQuery.data?.timezone || DEFAULT_PAROKI_TIMEZONE;

  const scheduleQuery = useLiturgiSchedule(eventId, Boolean(parokiId));
  const schedule = scheduleQuery.data;
  const petugas = useMemo(
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
  const respondMutation = useRespondLiturgiPetugas(eventId);

  const [status, setStatus] = useState<RespondStatus>("confirmed");
  const [note, setNote] = useState("");
  const [subMode, setSubMode] = useState<SubstituteMode>("none");
  const [substituteUserId, setSubstituteUserId] = useState("");
  const [substituteOrgId, setSubstituteOrgId] = useState("");
  const [substituteName, setSubstituteName] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!petugas) return;
    setStatus(petugas.status === "declined" ? "declined" : "confirmed");
    setNote(petugas.note ?? "");
    if (petugas.substituteUserId) {
      setSubMode("user");
      setSubstituteUserId(petugas.substituteUserId);
    } else if (petugas.substituteOrgId) {
      setSubMode("org");
      setSubstituteOrgId(petugas.substituteOrgId);
    } else if (petugas.substituteName) {
      setSubMode("manual");
      setSubstituteName(petugas.substituteName);
    } else {
      setSubMode("none");
    }
  }, [petugas]);

  const members = membersQuery.data ?? [];
  const orgs = orgsQuery.data ?? [];

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    if (status === "declined") {
      if (subMode === "user" && !substituteUserId) {
        setErrorMsg(t("liturgi.validationMemberRequired"));
        return;
      }
      if (subMode === "org" && !substituteOrgId) {
        setErrorMsg(t("liturgi.validationOrgRequired"));
        return;
      }
      if (subMode === "manual" && !substituteName.trim()) {
        setErrorMsg(t("liturgi.validationNameRequired"));
        return;
      }
    }

    try {
      await respondMutation.mutateAsync({
        petugasId,
        body: {
          status,
          note: note.trim() || undefined,
          substituteUserId:
            status === "declined" && subMode === "user"
              ? substituteUserId
              : undefined,
          substituteOrgId:
            status === "declined" && subMode === "org"
              ? substituteOrgId
              : undefined,
          substituteName:
            status === "declined" && subMode === "manual"
              ? substituteName.trim()
              : undefined,
        },
      });
      navigate(`/liturgi/${eventId}`);
    } catch (err) {
      setErrorMsg(resolveLiturgiError(err));
    }
  }

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
          { label: t("liturgi.petugasRespondTitle") },
        ]}
        title={t("liturgi.petugasRespondTitle")}
        subtitle={t("liturgi.petugasRespondSubtitle")}
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

      {schedule && petugas && (
        <Card className="max-w-2xl space-y-2 p-4 sm:p-5">
          <p className="text-sm font-semibold text-fg">{schedule.title}</p>
          <p className="text-xs text-muted">
            {formatLiturgiRange(
              schedule.startAt,
              schedule.endAt,
              parokiTimezone,
            )}
          </p>
          <div className="flex flex-wrap items-center gap-2 pt-1">
            <Badge tone="neutral">
              {getLiturgiMinistryMeta(petugas.role)?.label ?? petugas.role}
            </Badge>
            {petugas.slotLabel && (
              <Badge tone="neutral">{petugas.slotLabel}</Badge>
            )}
            <span className="text-sm font-medium text-fg">{petugas.name}</span>
          </div>
        </Card>
      )}

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-5">
        {errorMsg && (
          <Card className="border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
            {errorMsg}
          </Card>
        )}

        <InductiveFormSection title={t("liturgi.petugasRespondStatusLabel")}>
          <div className="space-y-4">
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="button"
                variant={status === "confirmed" ? "primary" : "outline"}
                onClick={() => setStatus("confirmed")}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.petugasRespondReadyOption")}
              </Button>
              <Button
                type="button"
                variant={status === "declined" ? "danger" : "outline"}
                onClick={() => setStatus("declined")}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.petugasRespondDeclineOption")}
              </Button>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.petugasRespondNoteLabel")}
              <textarea
                rows={3}
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder={t("liturgi.petugasRespondNotePlaceholder")}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>
          </div>
        </InductiveFormSection>

        {status === "declined" && (
          <InductiveFormSection
            title={t("liturgi.petugasSubstituteSectionTitle")}
          >
            <div className="space-y-4">
              <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                {(
                  [
                    { key: "none", label: t("liturgi.petugasSubstituteNone") },
                    { key: "user", label: t("liturgi.petugasModeUser") },
                    { key: "org", label: t("liturgi.petugasModeOrg") },
                    { key: "manual", label: t("liturgi.petugasModeManual") },
                  ] as const
                ).map((item) => (
                  <Button
                    key={item.key}
                    type="button"
                    variant={subMode === item.key ? "primary" : "outline"}
                    onClick={() => setSubMode(item.key)}
                    className="h-11 w-full sm:w-auto"
                  >
                    {item.label}
                  </Button>
                ))}
              </div>

              {subMode === "user" && (
                <div className="space-y-2">
                  <input
                    type="search"
                    value={memberSearch}
                    onChange={(e) => setMemberSearch(e.target.value)}
                    placeholder={t("liturgi.petugasSearchMemberPlaceholder")}
                    className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                  />
                  <select
                    value={substituteUserId}
                    onChange={(e) => setSubstituteUserId(e.target.value)}
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

              {subMode === "org" && (
                <select
                  value={substituteOrgId}
                  onChange={(e) => setSubstituteOrgId(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                >
                  <option value="">{t("liturgi.petugasSelectOrg")}</option>
                  {orgs.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name}
                    </option>
                  ))}
                </select>
              )}

              {subMode === "manual" && (
                <input
                  type="text"
                  value={substituteName}
                  onChange={(e) => setSubstituteName(e.target.value)}
                  placeholder={t("liturgi.petugasFieldManualPlaceholder")}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                />
              )}
            </div>
          </InductiveFormSection>
        )}

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Link
            to={`/liturgi/${eventId}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            {t("liturgi.ministriesCancel")}
          </Link>
          <Button
            type="submit"
            disabled={respondMutation.isPending}
            className="h-11 w-full sm:w-auto"
          >
            {respondMutation.isPending
              ? t("liturgi.creating")
              : t("liturgi.petugasRespondSubmit")}
          </Button>
        </div>
      </form>
    </div>
  );
}
