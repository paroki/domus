import {
  ArrowLeft,
  CalendarDays,
  Check,
  ExternalLink,
  MapPin,
  Music,
  Pencil,
  Plus,
  Trash2,
  Users,
} from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import { InductiveDangerZone, InductivePageHeader } from "~/shared/ui/iui";
import {
  useApproveLiturgiSubstitute,
  useDeleteLiturgiLagu,
  useDeleteLiturgiPetugas,
  useDeleteLiturgiRSVP,
  useDeleteLiturgiSchedule,
  useRespondLiturgiPetugas,
  useUpdateLiturgiSchedule,
  useUpsertLiturgiRSVP,
} from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import {
  useActiveParoki,
  useLiturgiSchedule,
  useLiturgiSongPartLookups,
} from "../api/queries";
import type { LiturgiLagu, LiturgiPetugas } from "../api/types";
import {
  DEFAULT_PAROKI_TIMEZONE,
  formatLiturgiRange,
  formatParokiTimezoneShort,
  formatRespondedAt,
  formatSubstituteLabel,
  fromLocalDatetimeInputValue,
  groupPetugasByRole,
  hasSubstituteProposal,
  LITURGI_MINISTRIES,
  petugasStatusMeta,
  scheduleStatusMeta,
  songPartLabel,
  toLocalDatetimeInputValue,
} from "../model";
import type { Route } from "./+types/LiturgiDetailPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.detailTitle")} · Domus` }];
}

type RsvpChoice = "going" | "maybe" | "not_going";

type ConfirmTarget =
  | { type: "cancel-toggle"; nextStatus: "cancelled" | "scheduled" }
  | { type: "delete-schedule" }
  | { type: "delete-petugas"; petugas: LiturgiPetugas }
  | { type: "delete-lagu"; lagu: LiturgiLagu }
  | null;

export default function LiturgiDetailPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { eventId = "" } = useParams();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const activeParokiQuery = useActiveParoki();
  const parokiTimezone =
    activeParokiQuery.data?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const tzShort = formatParokiTimezoneShort(parokiTimezone);

  const scheduleQuery = useLiturgiSchedule(eventId, hasParish);
  const schedule = scheduleQuery.data;
  const songPartLookupsQuery = useLiturgiSongPartLookups(hasParish);
  const songPartLookups = songPartLookupsQuery.data ?? [];

  const updateSchedule = useUpdateLiturgiSchedule(eventId);
  const deleteSchedule = useDeleteLiturgiSchedule();
  const upsertRSVP = useUpsertLiturgiRSVP(eventId);
  const deleteRSVP = useDeleteLiturgiRSVP(eventId);
  const respondPetugas = useRespondLiturgiPetugas(eventId);
  const approveSubstitute = useApproveLiturgiSubstitute(eventId);
  const deletePetugas = useDeleteLiturgiPetugas(eventId);
  const deleteLagu = useDeleteLiturgiLagu(eventId);

  const [editingInfo, setEditingInfo] = useState(false);
  const [title, setTitle] = useState("");
  const [location, setLocation] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [description, setDescription] = useState("");
  const [rsvpEnabled, setRsvpEnabled] = useState(true);
  const [rsvpDeadline, setRsvpDeadline] = useState("");
  const [clearRsvpDeadline, setClearRsvpDeadline] = useState(false);

  const [rsvpNote, setRsvpNote] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [confirmTarget, setConfirmTarget] = useState<ConfirmTarget>(null);

  useEffect(() => {
    if (!schedule) return;
    setTitle(schedule.title ?? "");
    setLocation(schedule.location ?? "");
    setStartAt(toLocalDatetimeInputValue(schedule.startAt, parokiTimezone));
    setEndAt(toLocalDatetimeInputValue(schedule.endAt, parokiTimezone));
    setDescription(schedule.description ?? "");
    setRsvpEnabled(schedule.rsvpEnabled ?? true);
    setRsvpDeadline(
      toLocalDatetimeInputValue(schedule.rsvpDeadline, parokiTimezone),
    );
    setClearRsvpDeadline(false);
    setRsvpNote(schedule.myRsvp?.note ?? "");
  }, [schedule, parokiTimezone]);

  if (!hasParish) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <Card className="p-6">
          <h1 className="text-xl font-bold text-fg">
            {t("liturgi.noParishTitle")}
          </h1>
          <p className="mt-2 text-sm text-muted">
            {t("liturgi.noParishDescription")}
          </p>
        </Card>
      </div>
    );
  }

  if (scheduleQuery.isLoading) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8">
        <Card className="p-6">
          <p className="text-sm text-muted">Memuat detail perayaan Misa…</p>
        </Card>
      </div>
    );
  }

  if (scheduleQuery.isError || !schedule) {
    return (
      <div className="mx-auto max-w-5xl space-y-4 px-4 py-8">
        <Link to="/liturgi" className={buttonClass("outline")}>
          <ArrowLeft aria-hidden className="size-4" />
          {t("liturgi.backToSchedules")}
        </Link>
        <Card className="border-danger/30 bg-danger-bg p-4 text-sm text-danger-fg">
          {resolveLiturgiError(scheduleQuery.error)}
        </Card>
      </div>
    );
  }

  const statusMeta = scheduleStatusMeta(schedule.status);
  const groupedPetugas = groupPetugasByRole(schedule.petugas);
  const manageableRoles = new Set(schedule.manageableRoles ?? []);
  const sortedLagu = [...(schedule.lagu ?? [])].sort(
    (a, b) => (a.urutan ?? 0) - (b.urutan ?? 0),
  );
  const myRsvpStatus = schedule.myRsvp?.status as RsvpChoice | undefined;

  async function handleSaveInfo(e: FormEvent) {
    e.preventDefault();
    setActionError(null);

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      setActionError(t("liturgi.validationTitleRequired"));
      return;
    }
    if (!startAt || !endAt) {
      setActionError(t("liturgi.validationStartEndRequired"));
      return;
    }
    const startIso = fromLocalDatetimeInputValue(startAt, parokiTimezone);
    const endIso = fromLocalDatetimeInputValue(endAt, parokiTimezone);
    if (!startIso || !endIso || new Date(endIso) <= new Date(startIso)) {
      setActionError(t("liturgi.validationEndAfterStart"));
      return;
    }

    try {
      await updateSchedule.mutateAsync({
        title: trimmedTitle,
        location: location.trim(),
        startAt: startIso,
        endAt: endIso,
        description: description.trim(),
        rsvpEnabled,
        clearRsvpDeadline: clearRsvpDeadline || undefined,
        rsvpDeadline:
          rsvpEnabled && !clearRsvpDeadline && rsvpDeadline
            ? fromLocalDatetimeInputValue(rsvpDeadline, parokiTimezone)
            : undefined,
      });
      setEditingInfo(false);
    } catch (err) {
      setActionError(resolveLiturgiError(err));
    }
  }

  async function handleSubmitRsvp(status: RsvpChoice) {
    setActionError(null);
    try {
      await upsertRSVP.mutateAsync({
        status,
        note: rsvpNote.trim() || undefined,
      });
    } catch (err) {
      setActionError(resolveLiturgiError(err));
    }
  }

  async function handleDeleteRsvp() {
    setActionError(null);
    try {
      await deleteRSVP.mutateAsync();
      setRsvpNote("");
    } catch (err) {
      setActionError(resolveLiturgiError(err));
    }
  }

  async function handleReadyPetugas(petugasId: string) {
    setActionError(null);
    try {
      await respondPetugas.mutateAsync({
        petugasId,
        body: { status: "confirmed" },
      });
    } catch (err) {
      setActionError(resolveLiturgiError(err));
    }
  }

  async function handleApproveSubstitute(petugasId: string) {
    setActionError(null);
    try {
      await approveSubstitute.mutateAsync({
        petugasId,
        body: { markConfirmed: false },
      });
    } catch (err) {
      setActionError(resolveLiturgiError(err));
    }
  }

  async function handleConfirmDialog() {
    if (!confirmTarget) return;
    setActionError(null);
    try {
      if (confirmTarget.type === "cancel-toggle") {
        await updateSchedule.mutateAsync({ status: confirmTarget.nextStatus });
      } else if (confirmTarget.type === "delete-schedule") {
        await deleteSchedule.mutateAsync(eventId);
        navigate("/liturgi", { replace: true });
        return;
      } else if (confirmTarget.type === "delete-petugas") {
        if (confirmTarget.petugas.id) {
          await deletePetugas.mutateAsync(confirmTarget.petugas.id);
        }
      } else if (confirmTarget.type === "delete-lagu") {
        if (confirmTarget.lagu.id) {
          await deleteLagu.mutateAsync(confirmTarget.lagu.id);
        }
      }
      setConfirmTarget(null);
    } catch (err) {
      setActionError(resolveLiturgiError(err));
      setConfirmTarget(null);
    }
  }

  const confirmPending =
    updateSchedule.isPending ||
    deleteSchedule.isPending ||
    deletePetugas.isPending ||
    deleteLagu.isPending;

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi"), to: "/liturgi" },
          { label: schedule.title ?? t("liturgi.detailTitle") },
        ]}
        title={schedule.title ?? t("liturgi.detailTitle")}
        subtitle={`${statusMeta.label} · ${formatLiturgiRange(schedule.startAt, schedule.endAt, parokiTimezone)}`}
        action={
          <Link
            to={`/event/${schedule.id}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            <ExternalLink aria-hidden className="size-4" />
            {t("liturgi.viewInEventCalendar")}
          </Link>
        }
      />

      {actionError && (
        <Card className="border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
          {actionError}
        </Card>
      )}

      {/* Card 1: Informasi & Waktu Perayaan Misa */}
      <Card className="space-y-4 p-4 sm:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base font-semibold text-fg">
              {t("liturgi.scheduleInfoTitle")}
            </h2>
            <Badge tone={statusMeta.tone}>{statusMeta.label}</Badge>
          </div>
          {schedule.canManageSchedule && !editingInfo && (
            <Button
              variant="outline"
              onClick={() => setEditingInfo(true)}
              className="h-11 w-full sm:w-auto"
            >
              <Pencil aria-hidden className="size-4" />
              {t("liturgi.scheduleEditCta")}
            </Button>
          )}
        </div>

        {!editingInfo ? (
          <div className="space-y-3 text-sm text-fg">
            <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:gap-x-6">
              <span className="inline-flex items-center gap-1.5 text-muted">
                <CalendarDays aria-hidden className="size-4 shrink-0" />
                {formatLiturgiRange(
                  schedule.startAt,
                  schedule.endAt,
                  parokiTimezone,
                )}
              </span>
              <span className="inline-flex items-center gap-1.5 text-muted">
                <MapPin aria-hidden className="size-4 shrink-0" />
                {schedule.location || t("liturgi.scheduleLocationEmpty")}
              </span>
            </div>
            {schedule.description && (
              <p className="whitespace-pre-line text-sm text-fg">
                {schedule.description}
              </p>
            )}
            <div className="flex flex-wrap items-center gap-2 pt-1 text-xs text-muted">
              <Badge tone={schedule.rsvpEnabled ? "success" : "neutral"}>
                {schedule.rsvpEnabled
                  ? t("liturgi.scheduleRsvpEnabledLabel")
                  : t("liturgi.scheduleRsvpDisabledLabel")}
              </Badge>
              {schedule.rsvpEnabled && schedule.rsvpDeadline && (
                <span>
                  {t("liturgi.scheduleRsvpDeadlineLabel")}:{" "}
                  {formatRespondedAt(schedule.rsvpDeadline, parokiTimezone)}
                </span>
              )}
            </div>
          </div>
        ) : (
          <form onSubmit={handleSaveInfo} className="space-y-4">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldTitle")}
              <input
                type="text"
                required
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldLocation")}
              <input
                type="text"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
              />
            </label>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {`${t("liturgi.fieldStartAt")} (${tzShort})`}
                <input
                  type="datetime-local"
                  required
                  value={startAt}
                  onChange={(e) => setStartAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {`${t("liturgi.fieldEndAt")} (${tzShort})`}
                <input
                  type="datetime-local"
                  required
                  value={endAt}
                  onChange={(e) => setEndAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldDescription")}
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-base text-fg sm:text-sm"
              />
            </label>

            <label className="flex min-h-11 items-center gap-2.5 text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={rsvpEnabled}
                onChange={(e) => setRsvpEnabled(e.target.checked)}
                className="size-4 rounded border-line"
              />
              {t("liturgi.fieldRsvpEnabled")}
            </label>

            {rsvpEnabled && (
              <div className="space-y-2">
                <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                  {`${t("liturgi.fieldRsvpDeadline")} (${tzShort})`}
                  <input
                    type="datetime-local"
                    disabled={clearRsvpDeadline}
                    value={clearRsvpDeadline ? "" : rsvpDeadline}
                    onChange={(e) => setRsvpDeadline(e.target.value)}
                    className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg disabled:opacity-50 sm:text-sm"
                  />
                </label>
                {schedule.rsvpDeadline && (
                  <label className="flex min-h-11 items-center gap-2 text-xs text-muted">
                    <input
                      type="checkbox"
                      checked={clearRsvpDeadline}
                      onChange={(e) => setClearRsvpDeadline(e.target.checked)}
                      className="size-4 rounded border-line"
                    />
                    {t("liturgi.scheduleClearRsvpDeadline")}
                  </label>
                )}
              </div>
            )}

            <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditingInfo(false)}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.ministriesCancel")}
              </Button>
              <Button
                type="submit"
                disabled={updateSchedule.isPending}
                className="h-11 w-full sm:w-auto"
              >
                {updateSchedule.isPending
                  ? t("liturgi.scheduleSaving")
                  : t("liturgi.scheduleSaveCta")}
              </Button>
            </div>
          </form>
        )}
      </Card>

      {/* Card 2: RSVP Kehadiran Umat */}
      <Card className="space-y-4 p-4 sm:p-6">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-base font-semibold text-fg">
            {t("liturgi.rsvpCardTitle")}
          </h2>
          {myRsvpStatus && (
            <Badge
              tone={
                myRsvpStatus === "going"
                  ? "success"
                  : myRsvpStatus === "not_going"
                    ? "danger"
                    : "warning"
              }
            >
              {t("liturgi.rsvpMyStatus")}:{" "}
              {myRsvpStatus === "going"
                ? t("liturgi.rsvpGoing")
                : myRsvpStatus === "maybe"
                  ? t("liturgi.rsvpMaybe")
                  : t("liturgi.rsvpNotGoing")}
            </Badge>
          )}
        </div>

        <div className="grid grid-cols-3 gap-3">
          <div className="rounded-lg border border-line bg-surface-2/40 p-3 text-center">
            <p className="text-lg font-bold text-fg">
              {schedule.rsvpSummary?.going ?? 0}
            </p>
            <p className="text-xs text-muted">{t("liturgi.rsvpGoing")}</p>
          </div>
          <div className="rounded-lg border border-line bg-surface-2/40 p-3 text-center">
            <p className="text-lg font-bold text-fg">
              {schedule.rsvpSummary?.maybe ?? 0}
            </p>
            <p className="text-xs text-muted">{t("liturgi.rsvpMaybe")}</p>
          </div>
          <div className="rounded-lg border border-line bg-surface-2/40 p-3 text-center">
            <p className="text-lg font-bold text-fg">
              {schedule.rsvpSummary?.notGoing ?? 0}
            </p>
            <p className="text-xs text-muted">{t("liturgi.rsvpNotGoing")}</p>
          </div>
        </div>

        {schedule.canRsvp && schedule.rsvpOpen ? (
          <div className="space-y-3">
            <input
              type="text"
              value={rsvpNote}
              onChange={(e) => setRsvpNote(e.target.value)}
              placeholder={t("liturgi.rsvpNotePlaceholder")}
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
            />
            <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
              <Button
                type="button"
                variant={myRsvpStatus === "going" ? "primary" : "outline"}
                disabled={upsertRSVP.isPending}
                onClick={() => handleSubmitRsvp("going")}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.rsvpGoing")}
              </Button>
              <Button
                type="button"
                variant={myRsvpStatus === "maybe" ? "primary" : "outline"}
                disabled={upsertRSVP.isPending}
                onClick={() => handleSubmitRsvp("maybe")}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.rsvpMaybe")}
              </Button>
              <Button
                type="button"
                variant={myRsvpStatus === "not_going" ? "danger" : "outline"}
                disabled={upsertRSVP.isPending}
                onClick={() => handleSubmitRsvp("not_going")}
                className="h-11 w-full sm:w-auto"
              >
                {t("liturgi.rsvpNotGoing")}
              </Button>
              {schedule.myRsvp && (
                <Button
                  type="button"
                  variant="ghost"
                  disabled={deleteRSVP.isPending}
                  onClick={handleDeleteRsvp}
                  className="h-11 w-full sm:w-auto"
                >
                  {t("liturgi.rsvpRemove")}
                </Button>
              )}
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted">{t("liturgi.rsvpClosedNote")}</p>
        )}
      </Card>

      {/* Card 3: Petugas Liturgi (6 Ministry Groups) */}
      <Card className="space-y-5 p-4 sm:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <Users aria-hidden className="size-5 text-muted" />
            <h2 className="text-base font-semibold text-fg">
              {t("liturgi.petugasCardTitle")}
            </h2>
          </div>
          {manageableRoles.size > 0 && (
            <Link
              to={`/liturgi/${schedule.id}/petugas/new`}
              className={`${buttonClass("primary")} h-11 w-full sm:w-auto`}
            >
              <Plus aria-hidden className="size-4" />
              {t("liturgi.petugasAddCta")}
            </Link>
          )}
        </div>

        <div className="space-y-5">
          {LITURGI_MINISTRIES.map((ministry) => {
            const rows = groupedPetugas[ministry.slug];
            const canAddRole = manageableRoles.has(ministry.slug);
            return (
              <section
                key={ministry.slug}
                className="space-y-3 border-t border-line pt-4 first:border-t-0 first:pt-0"
              >
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold text-fg">
                      {ministry.label}
                    </h3>
                    <Badge tone="neutral">{rows.length}</Badge>
                  </div>
                  {canAddRole && (
                    <Link
                      to={`/liturgi/${schedule.id}/petugas/new?role=${ministry.slug}`}
                      className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
                    >
                      <Plus aria-hidden className="size-4" />
                      {t("liturgi.petugasAddShort")}
                    </Link>
                  )}
                </div>

                {rows.length === 0 ? (
                  <p className="text-sm text-muted">
                    {t("liturgi.petugasEmptyMinistry")}
                  </p>
                ) : (
                  <ul className="space-y-3">
                    {rows.map((p) => {
                      const pMeta = petugasStatusMeta(p.status);
                      const subLabel = formatSubstituteLabel(p);
                      const canApproveSub =
                        Boolean(p.canApproveSubstitute) &&
                        hasSubstituteProposal(p);
                      const respondedLabel = formatRespondedAt(
                        p.respondedAt,
                        parokiTimezone,
                      );

                      return (
                        <li
                          key={p.id}
                          className="flex flex-col gap-3 rounded-xl border border-line bg-surface-2/30 p-3.5"
                        >
                          <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
                            <div className="space-y-1">
                              <div className="flex flex-wrap items-center gap-2">
                                {p.slotLabel && (
                                  <Badge tone="neutral">{p.slotLabel}</Badge>
                                )}
                                <span className="text-sm font-semibold text-fg">
                                  {p.name}
                                </span>
                                {p.orgId && (
                                  <Badge tone="neutral">
                                    {t("liturgi.petugasOrgIndicator")}
                                  </Badge>
                                )}
                                <Badge tone={pMeta.tone}>{pMeta.label}</Badge>
                              </div>
                              {respondedLabel && (
                                <p className="text-xs text-muted">
                                  Direspon: {respondedLabel}
                                </p>
                              )}
                              {p.note && (
                                <p className="text-sm text-muted">{p.note}</p>
                              )}
                              {subLabel && (
                                <p className="rounded-lg border border-line bg-canvas px-3 py-2 text-xs font-medium text-fg">
                                  {subLabel}
                                </p>
                              )}
                            </div>

                            <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:justify-end">
                              {p.canRespond && p.id && (
                                <>
                                  {p.status !== "confirmed" && (
                                    <Button
                                      type="button"
                                      variant="primary"
                                      disabled={respondPetugas.isPending}
                                      onClick={() =>
                                        handleReadyPetugas(p.id as string)
                                      }
                                      className="h-11 w-full sm:w-auto"
                                    >
                                      <Check aria-hidden className="size-4" />
                                      {t("liturgi.petugasReadyCta")}
                                    </Button>
                                  )}
                                  <Link
                                    to={`/liturgi/${schedule.id}/petugas/${p.id}/respond`}
                                    className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
                                  >
                                    {t("liturgi.petugasRespondCta")}
                                  </Link>
                                </>
                              )}

                              {canApproveSub && p.id && (
                                <Button
                                  type="button"
                                  variant="primary"
                                  disabled={approveSubstitute.isPending}
                                  onClick={() =>
                                    handleApproveSubstitute(p.id as string)
                                  }
                                  className="h-11 w-full sm:w-auto"
                                >
                                  {t("liturgi.petugasApproveSubstituteCta")}
                                </Button>
                              )}

                              {p.canEdit && p.id && (
                                <>
                                  <Link
                                    to={`/liturgi/${schedule.id}/petugas/${p.id}/edit`}
                                    className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
                                  >
                                    <Pencil aria-hidden className="size-4" />
                                    {t("liturgi.petugasEditCta")}
                                  </Link>
                                  <Button
                                    type="button"
                                    variant="danger"
                                    onClick={() =>
                                      setConfirmTarget({
                                        type: "delete-petugas",
                                        petugas: p,
                                      })
                                    }
                                    className="h-11 w-full sm:w-auto"
                                  >
                                    <Trash2 aria-hidden className="size-4" />
                                    {t("liturgi.petugasDeleteCta")}
                                  </Button>
                                </>
                              )}
                            </div>
                          </div>
                        </li>
                      );
                    })}
                  </ul>
                )}
              </section>
            );
          })}
        </div>
      </Card>

      {/* Card 4: Daftar Lagu Perayaan */}
      <Card className="space-y-4 p-4 sm:p-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2">
            <Music aria-hidden className="size-5 text-muted" />
            <h2 className="text-base font-semibold text-fg">
              {t("liturgi.laguCardTitle")}
            </h2>
          </div>
          {schedule.canManageLagu && (
            <Link
              to={`/liturgi/${schedule.id}/lagu/new`}
              className={`${buttonClass("primary")} h-11 w-full sm:w-auto`}
            >
              <Plus aria-hidden className="size-4" />
              {t("liturgi.laguAddCta")}
            </Link>
          )}
        </div>

        {sortedLagu.length === 0 ? (
          <p className="text-sm text-muted">{t("liturgi.laguEmpty")}</p>
        ) : (
          <ul className="space-y-3">
            {sortedLagu.map((lagu) => {
              const partLabel = songPartLabel(lagu.bagian, songPartLookups);
              const codeLabel = [lagu.sumber, lagu.nomor]
                .filter(Boolean)
                .join(" · ");

              return (
                <li
                  key={lagu.id}
                  className="flex flex-col gap-3 rounded-xl border border-line bg-surface-2/30 p-3.5 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      {partLabel && <Badge tone="neutral">{partLabel}</Badge>}
                      {codeLabel && <Badge tone="neutral">{codeLabel}</Badge>}
                      {lagu.nadaDasar && (
                        <Badge tone="neutral">{`1=${lagu.nadaDasar}`}</Badge>
                      )}
                      {lagu.cantusId && (
                        <Badge tone="success">
                          {t("liturgi.laguCatalogBadge")}
                        </Badge>
                      )}
                    </div>
                    <div>
                      {lagu.cantusId ? (
                        <Link
                          to={`/cantus/${lagu.cantusId}`}
                          className="inline-flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline"
                        >
                          {lagu.judul}
                          <ExternalLink aria-hidden className="size-3.5" />
                        </Link>
                      ) : (
                        <span className="text-sm font-semibold text-fg">
                          {lagu.judul}
                        </span>
                      )}
                    </div>
                    {lagu.note && (
                      <p className="text-sm text-muted">{lagu.note}</p>
                    )}
                  </div>

                  {schedule.canManageLagu && lagu.id && (
                    <div className="flex flex-col gap-2 sm:flex-row">
                      <Link
                        to={`/liturgi/${schedule.id}/lagu/${lagu.id}/edit`}
                        className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
                      >
                        <Pencil aria-hidden className="size-4" />
                        {t("liturgi.petugasEditCta")}
                      </Link>
                      <Button
                        type="button"
                        variant="danger"
                        onClick={() =>
                          setConfirmTarget({ type: "delete-lagu", lagu })
                        }
                        className="h-11 w-full sm:w-auto"
                      >
                        <Trash2 aria-hidden className="size-4" />
                        {t("liturgi.petugasDeleteCta")}
                      </Button>
                    </div>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </Card>

      {/* Card 5: Zona Bahaya */}
      {schedule.canManageSchedule && (
        <InductiveDangerZone
          title={t("liturgi.dangerZoneTitle")}
          description={t("liturgi.dangerZoneDescription")}
          action={
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  setConfirmTarget({
                    type: "cancel-toggle",
                    nextStatus:
                      schedule.status === "cancelled"
                        ? "scheduled"
                        : "cancelled",
                  })
                }
                className="h-11 w-full sm:w-auto"
              >
                {schedule.status === "cancelled"
                  ? t("liturgi.reopenScheduleCta")
                  : t("liturgi.cancelScheduleCta")}
              </Button>
              <Button
                type="button"
                variant="danger"
                onClick={() => setConfirmTarget({ type: "delete-schedule" })}
                className="h-11 w-full sm:w-auto"
              >
                <Trash2 aria-hidden className="size-4" />
                {t("liturgi.deleteScheduleCta")}
              </Button>
            </div>
          }
        />
      )}

      <ConfirmDialog
        open={confirmTarget !== null}
        title={
          confirmTarget?.type === "cancel-toggle"
            ? confirmTarget.nextStatus === "cancelled"
              ? t("liturgi.cancelScheduleTitle")
              : t("liturgi.reopenScheduleTitle")
            : confirmTarget?.type === "delete-schedule"
              ? t("liturgi.deleteScheduleTitle")
              : confirmTarget?.type === "delete-petugas"
                ? t("liturgi.petugasDeleteTitle")
                : t("liturgi.laguDeleteTitle")
        }
        description={
          confirmTarget?.type === "cancel-toggle"
            ? confirmTarget.nextStatus === "cancelled"
              ? t("liturgi.cancelScheduleDescription")
              : t("liturgi.reopenScheduleDescription")
            : confirmTarget?.type === "delete-schedule"
              ? t("liturgi.deleteScheduleDescription", {
                  title: schedule.title ?? "",
                })
              : confirmTarget?.type === "delete-petugas"
                ? t("liturgi.petugasDeleteDescription", {
                    name: confirmTarget.petugas.name ?? "",
                  })
                : confirmTarget?.type === "delete-lagu"
                  ? t("liturgi.laguDeleteDescription", {
                      title: confirmTarget.lagu.judul ?? "",
                    })
                  : ""
        }
        confirmLabel={
          confirmTarget?.type === "cancel-toggle"
            ? confirmTarget.nextStatus === "cancelled"
              ? t("liturgi.cancelScheduleCta")
              : t("liturgi.reopenScheduleCta")
            : t("liturgi.petugasDeleteCta")
        }
        confirmVariant={
          confirmTarget?.type === "cancel-toggle" &&
          confirmTarget.nextStatus === "scheduled"
            ? "primary"
            : "danger"
        }
        pending={confirmPending}
        onConfirm={handleConfirmDialog}
        onCancel={() => setConfirmTarget(null)}
      />
    </div>
  );
}
