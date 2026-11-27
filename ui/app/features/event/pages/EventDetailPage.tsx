import { ArrowLeft, CalendarDays, Church, Pencil, Trash2 } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { ConfirmDialog } from "~/shared/ui/ConfirmDialog";
import {
  InductiveDangerZone,
  InductiveEmptyState,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import {
  useDeleteEvent,
  useDeleteEventRSVP,
  useUpdateEvent,
  useUpsertEventRSVP,
} from "../api/actions";
import { resolveEventError } from "../api/errors";
import {
  useActiveParoki,
  useEvent,
  useEventRSVPSummary,
  useEventRSVPs,
  useEventTypeLookups,
} from "../api/queries";
import {
  computeEventLifecycle,
  DEFAULT_PAROKI_TIMEZONE,
  type EventVisibility,
  eventTypeLabel,
  formatEventRange,
  formatParokiTimezoneShort,
  fromLocalDatetimeInputValue,
  lifecycleMeta,
  mergeEventTypeOptions,
  type RSVPStatus,
  rsvpStatusMeta,
  toLocalDatetimeInputValue,
  visibilityMeta,
} from "../model";
import type { Route } from "./+types/EventDetailPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("event.title")} · Domus` }];
}

export default function EventDetailPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { eventId = "" } = useParams();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const activeParokiQuery = useActiveParoki(hasParish);
  const activeParoki = activeParokiQuery.data;
  const parokiTimezone = activeParoki?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const tzShort = formatParokiTimezoneShort(parokiTimezone);

  const typeLookupsQuery = useEventTypeLookups(hasParish);
  const typeLookups = typeLookupsQuery.data ?? [];
  const eventTypeOptions = mergeEventTypeOptions(typeLookups);

  const eventQuery = useEvent(eventId, hasParish);
  const ev = eventQuery.data;
  const canRsvp = Boolean(ev?.canRsvp);
  const canManage = Boolean(ev?.canManage);

  const summaryQuery = useEventRSVPSummary(eventId, hasParish && canRsvp);
  const [rosterQ, setRosterQ] = useState("");
  const [rosterStatus, setRosterStatus] = useState("");
  const rsvpsQuery = useEventRSVPs(
    eventId,
    {
      q: rosterQ.trim() || undefined,
      filter: rosterStatus ? { status: rosterStatus as RSVPStatus } : undefined,
    },
    hasParish && canRsvp,
  );

  const updateEvent = useUpdateEvent(eventId);
  const deleteEvent = useDeleteEvent();
  const upsertRSVP = useUpsertEventRSVP();
  const deleteRSVP = useDeleteEventRSVP(eventId);

  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState("");
  const [jenis, setJenis] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [location, setLocation] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<EventVisibility>("public");
  const [rsvpEnabled, setRsvpEnabled] = useState(true);
  const [rsvpDeadline, setRsvpDeadline] = useState("");
  const [clearRsvpDeadline, setClearRsvpDeadline] = useState(false);

  const [rsvpNote, setRsvpNote] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [pendingCancelToggle, setPendingCancelToggle] = useState(false);
  const [pendingDeleteEvent, setPendingDeleteEvent] = useState(false);
  const [pendingDeleteRsvp, setPendingDeleteRsvp] = useState(false);

  useEffect(() => {
    if (!ev) return;
    setTitle(ev.title ?? "");
    setJenis(ev.jenis ?? "");
    setStartAt(toLocalDatetimeInputValue(ev.startAt, parokiTimezone));
    setEndAt(toLocalDatetimeInputValue(ev.endAt, parokiTimezone));
    setLocation(ev.location ?? "");
    setDescription(ev.description ?? "");
    setVisibility((ev.visibility ?? "public") as EventVisibility);
    setRsvpEnabled(ev.rsvpEnabled ?? true);
    setRsvpDeadline(toLocalDatetimeInputValue(ev.rsvpDeadline, parokiTimezone));
    setClearRsvpDeadline(false);
    setRsvpNote(ev.myRsvp?.note ?? "");
  }, [ev, parokiTimezone]);

  if (!hasParish) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={CalendarDays}
          title={t("event.noParishTitle")}
          description={t("event.noParishDescription")}
        />
      </div>
    );
  }

  if (eventQuery.isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <Card className="p-4 sm:p-6">
          <p className="text-sm text-muted">Memuat detail kegiatan…</p>
        </Card>
      </div>
    );
  }

  if (eventQuery.isError || !ev) {
    return (
      <div className="mx-auto max-w-4xl space-y-4 px-4 py-6 sm:px-6 sm:py-8">
        <Link
          to="/event"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:underline"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("event.backToList")}
        </Link>
        <Card className="p-4 sm:p-6">
          <p className="text-sm text-danger-fg">
            {resolveEventError(eventQuery.error)}
          </p>
        </Card>
      </div>
    );
  }

  const lifecycle = computeEventLifecycle(ev);
  const lcMeta = lifecycleMeta(lifecycle);
  const visMeta = visibilityMeta(ev.visibility ?? "public");
  const summary = summaryQuery.data ?? ev.rsvpSummary;
  const rsvpRows = rsvpsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const myRsvpMeta = ev.myRsvp ? rsvpStatusMeta(ev.myRsvp.status ?? "") : null;
  const isCancelled = ev.status === "cancelled";

  async function handleSaveEdit(e: FormEvent) {
    e.preventDefault();
    setActionError(null);
    try {
      await updateEvent.mutateAsync({
        title: title.trim(),
        jenis,
        startAt: fromLocalDatetimeInputValue(startAt, parokiTimezone),
        endAt: fromLocalDatetimeInputValue(endAt, parokiTimezone),
        location,
        description,
        visibility: ev?.orgId ? visibility : "public",
        rsvpEnabled,
        clearRsvpDeadline: clearRsvpDeadline || undefined,
        rsvpDeadline:
          !clearRsvpDeadline && rsvpDeadline
            ? fromLocalDatetimeInputValue(rsvpDeadline, parokiTimezone)
            : undefined,
      });
      setEditing(false);
    } catch (err) {
      setActionError(resolveEventError(err));
    }
  }

  async function confirmToggleCancel() {
    if (!ev) return;
    const nextStatus = ev.status === "cancelled" ? "scheduled" : "cancelled";
    setActionError(null);
    try {
      await updateEvent.mutateAsync({ status: nextStatus });
      setPendingCancelToggle(false);
    } catch (err) {
      setActionError(resolveEventError(err));
      setPendingCancelToggle(false);
    }
  }

  async function confirmDeleteEvent() {
    setActionError(null);
    try {
      await deleteEvent.mutateAsync(eventId);
      navigate("/event", { replace: true });
    } catch (err) {
      setActionError(resolveEventError(err));
      setPendingDeleteEvent(false);
    }
  }

  async function handleSubmitRSVP(status: RSVPStatus) {
    setActionError(null);
    try {
      await upsertRSVP.mutateAsync({
        eventId,
        body: { status, note: rsvpNote.trim() || undefined },
      });
    } catch (err) {
      setActionError(resolveEventError(err));
    }
  }

  async function confirmDeleteRSVP() {
    setActionError(null);
    try {
      await deleteRSVP.mutateAsync();
      setRsvpNote("");
      setPendingDeleteRsvp(false);
    } catch (err) {
      setActionError(resolveEventError(err));
      setPendingDeleteRsvp(false);
    }
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("event.title"), to: "/event" },
          { label: ev.title ?? "Detail Kegiatan" },
        ]}
        title={ev.title ?? ""}
        subtitle={formatEventRange(ev.startAt, ev.endAt, parokiTimezone)}
      />

      {actionError && (
        <div
          role="alert"
          className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {actionError}
        </div>
      )}

      <Card className="p-4 sm:p-6">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-line pb-4">
          <div className="flex flex-wrap items-center gap-2">
            {ev.orgId ? (
              <Link
                to={`/org/${ev.orgId}`}
                className="inline-flex items-center rounded-full bg-neutral-bg px-3 py-1 text-[13px] font-semibold text-neutral-fg hover:underline"
              >
                {ev.orgName}
              </Link>
            ) : (
              <Badge tone="neutral">{activeParoki?.nama ?? "Paroki"}</Badge>
            )}
            {ev.jenis && (
              <Badge tone="neutral">
                {eventTypeLabel(ev.jenis, typeLookups)}
              </Badge>
            )}
            <Badge tone={visMeta.tone}>{visMeta.label}</Badge>
            <Badge tone={lcMeta.tone}>{lcMeta.label}</Badge>
          </div>

          {canManage && !editing && (
            <Button
              type="button"
              variant="outline"
              onClick={() => setEditing(true)}
              className="min-h-10 w-full sm:w-auto"
            >
              <Pencil aria-hidden className="size-4" />
              Ubah
            </Button>
          )}
        </div>

        {!editing ? (
          <dl className="mt-4 grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs font-medium text-muted">Penyelenggara</dt>
              <dd className="mt-1 text-sm font-medium text-fg">
                {ev.orgId ? ev.orgName : (activeParoki?.nama ?? "Paroki")}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Jenis Kegiatan</dt>
              <dd className="mt-1 text-sm text-fg">
                {ev.jenis ? eventTypeLabel(ev.jenis, typeLookups) : "Umum"}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Waktu</dt>
              <dd className="mt-1 text-sm text-fg">
                {formatEventRange(ev.startAt, ev.endAt, parokiTimezone)}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Lokasi</dt>
              <dd className="mt-1 text-sm text-fg">{ev.location || "—"}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Visibilitas</dt>
              <dd className="mt-1 text-sm text-fg">{visMeta.label}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">
                Konfirmasi Kehadiran (RSVP)
              </dt>
              <dd className="mt-1 text-sm text-fg">
                {ev.rsvpEnabled
                  ? ev.rsvpDeadline
                    ? `Batas RSVP: ${formatEventRange(ev.rsvpDeadline, undefined, parokiTimezone)}`
                    : "Dibuka hingga 30 menit setelah mulai"
                  : "Dinonaktifkan"}
              </dd>
            </div>
            <div className="sm:col-span-2">
              <dt className="text-xs font-medium text-muted">Deskripsi</dt>
              <dd className="mt-1 whitespace-pre-line text-sm text-fg">
                {ev.description || "—"}
              </dd>
            </div>
          </dl>
        ) : (
          <form onSubmit={handleSaveEdit} className="mt-4 space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Judul Kegiatan
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  required
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Jenis Kegiatan
                <select
                  value={jenis}
                  onChange={(e) => setJenis(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="">Tanpa Jenis / Umum</option>
                  {eventTypeOptions.map((opt) => (
                    <option key={opt.code} value={opt.code}>
                      {opt.label}
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Waktu Mulai ({tzShort})
                <input
                  type="datetime-local"
                  value={startAt}
                  onChange={(e) => setStartAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  required
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Waktu Selesai ({tzShort})
                <input
                  type="datetime-local"
                  value={endAt}
                  onChange={(e) => setEndAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                  required
                />
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Lokasi
                <input
                  type="text"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Visibilitas
                <select
                  value={ev.orgId ? visibility : "public"}
                  onChange={(e) =>
                    setVisibility(e.target.value as EventVisibility)
                  }
                  disabled={!ev.orgId}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg disabled:opacity-60"
                >
                  <option value="public">Publik Paroki</option>
                  <option value="private" disabled={!ev.orgId}>
                    Internal Organisasi
                  </option>
                </select>
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Batas Waktu RSVP ({tzShort})
                <input
                  type="datetime-local"
                  value={rsvpDeadline}
                  disabled={clearRsvpDeadline}
                  onChange={(e) => setRsvpDeadline(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg disabled:opacity-50"
                />
              </label>
              <div className="flex items-end">
                <label className="flex min-h-11 cursor-pointer items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={clearRsvpDeadline}
                    onChange={(e) => setClearRsvpDeadline(e.target.checked)}
                    className="size-4 rounded border-line"
                  />
                  Hapus batas waktu RSVP
                </label>
              </div>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Deskripsi
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-sm text-fg"
              />
            </label>

            <label className="flex min-h-11 cursor-pointer items-center gap-2 text-sm text-fg">
              <input
                type="checkbox"
                checked={rsvpEnabled}
                onChange={(e) => setRsvpEnabled(e.target.checked)}
                className="size-4 rounded border-line"
              />
              Aktifkan konfirmasi kehadiran (RSVP)
            </label>

            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(false)}
                className="w-full sm:w-auto"
              >
                Batal
              </Button>
              <Button
                type="submit"
                disabled={updateEvent.isPending}
                className="w-full sm:w-auto"
              >
                {updateEvent.isPending ? "Menyimpan…" : "Simpan Perubahan"}
              </Button>
            </div>
          </form>
        )}
      </Card>

      {ev.jenis === "misa" && (
        <InductiveStepCard
          to={`/liturgi/${ev.id}`}
          icon={Church}
          plate={1}
          title={t("event.liturgiShortcutCta")}
          description={t("event.liturgiShortcutDescription")}
          badge={
            <Badge tone="neutral">{t("event.liturgiShortcutTitle")}</Badge>
          }
        />
      )}

      {!canRsvp ? (
        <Card className="p-4 sm:p-6">
          <h2 className="text-base font-semibold text-fg">
            {t("event.rsvpRestrictedTitle")}
          </h2>
          {ev.orgId ? (
            <p className="mt-1 text-sm text-muted">
              Konfirmasi kehadiran (RSVP) dan daftar peserta hanya tersedia bagi
              anggota aktif{" "}
              <Link
                to={`/org/${ev.orgId}`}
                className="font-medium text-accent hover:underline"
              >
                {ev.orgName}
              </Link>
              .
            </p>
          ) : (
            <p className="mt-1 text-sm text-muted">
              Konfirmasi kehadiran (RSVP) dan daftar peserta hanya tersedia bagi
              umat terdaftar di paroki ini.
            </p>
          )}
        </Card>
      ) : (
        <>
          <Card className="space-y-4 p-4 sm:p-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg font-semibold text-fg">
                Konfirmasi Kehadiran (RSVP)
              </h2>
              {myRsvpMeta && (
                <Badge tone={myRsvpMeta.tone}>Status: {myRsvpMeta.label}</Badge>
              )}
            </div>

            {ev.myRsvp?.note && (
              <p className="text-sm text-muted">
                Catatan Anda: <span className="text-fg">{ev.myRsvp.note}</span>
              </p>
            )}

            {ev.rsvpOpen ? (
              <div className="space-y-3">
                <input
                  type="text"
                  value={rsvpNote}
                  onChange={(e) => setRsvpNote(e.target.value)}
                  placeholder="Catatan kehadiran (opsional)"
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
                />
                <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
                  <div className="grid grid-cols-3 gap-2 sm:flex sm:items-center">
                    {(
                      [
                        { status: "going", label: "Hadir" },
                        { status: "maybe", label: "Mungkin" },
                        { status: "not_going", label: "Tidak Hadir" },
                      ] as const
                    ).map((opt) => (
                      <Button
                        key={opt.status}
                        type="button"
                        variant={
                          ev.myRsvp?.status === opt.status
                            ? "primary"
                            : "outline"
                        }
                        disabled={upsertRSVP.isPending}
                        onClick={() => handleSubmitRSVP(opt.status)}
                        className="w-full sm:w-auto"
                      >
                        {opt.label}
                      </Button>
                    ))}
                  </div>
                  {ev.myRsvp && (
                    <Button
                      type="button"
                      variant="danger"
                      disabled={deleteRSVP.isPending}
                      onClick={() => setPendingDeleteRsvp(true)}
                      className="w-full sm:w-auto"
                    >
                      {t("event.removeRsvp")}
                    </Button>
                  )}
                </div>
              </div>
            ) : (
              <p className="text-sm text-muted">{t("event.rsvpClosed")}</p>
            )}
          </Card>

          {summary && (
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
              <Card className="p-4 sm:p-6">
                <p className="text-xs font-medium text-muted">Hadir</p>
                <p className="mt-1 text-2xl font-bold text-success-fg">
                  {summary.going}
                </p>
              </Card>
              <Card className="p-4 sm:p-6">
                <p className="text-xs font-medium text-muted">Mungkin</p>
                <p className="mt-1 text-2xl font-bold text-warning-fg">
                  {summary.maybe}
                </p>
              </Card>
              <Card className="p-4 sm:p-6">
                <p className="text-xs font-medium text-muted">Tidak Hadir</p>
                <p className="mt-1 text-2xl font-bold text-danger-fg">
                  {summary.notGoing}
                </p>
              </Card>
              <Card className="p-4 sm:p-6">
                <p className="text-xs font-medium text-muted">
                  Belum Konfirmasi
                </p>
                <p className="mt-1 text-2xl font-bold text-fg">
                  {summary.unresponded}
                </p>
                <p className="text-xs text-muted">
                  {ev.orgId
                    ? `dari ${summary.totalActiveMembers} anggota aktif`
                    : `dari ${summary.totalActiveMembers} umat paroki aktif`}
                </p>
              </Card>
            </div>
          )}

          <Card className="space-y-4 p-4 sm:p-6">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <h2 className="text-lg font-semibold text-fg">
                {t("event.rsvpRosterTitle")}
              </h2>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <select
                  aria-label="Filter status kehadiran"
                  value={rosterStatus}
                  onChange={(e) => setRosterStatus(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
                >
                  <option value="">Semua Status</option>
                  <option value="going">Hadir</option>
                  <option value="maybe">Mungkin</option>
                  <option value="not_going">Tidak Hadir</option>
                </select>
                <input
                  type="search"
                  value={rosterQ}
                  onChange={(e) => setRosterQ(e.target.value)}
                  placeholder="Cari nama, email, catatan…"
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted sm:w-auto"
                />
              </div>
            </div>

            {rsvpsQuery.isLoading ? (
              <p className="text-sm text-muted">Memuat daftar peserta…</p>
            ) : rsvpRows.length === 0 ? (
              <p className="text-sm text-muted">
                Belum ada anggota yang mengirimkan konfirmasi kehadiran.
              </p>
            ) : (
              <div className="divide-y divide-line">
                {rsvpRows.map((row) => {
                  const st = rsvpStatusMeta(row.status ?? "");
                  return (
                    <div
                      key={row.id}
                      className="flex flex-wrap items-center justify-between gap-3 py-3"
                    >
                      <div className="min-w-0 flex-1 space-y-0.5">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="truncate font-medium text-fg">
                            {row.name}
                          </span>
                          {row.orgRole && (
                            <Badge tone="neutral">{row.orgRole}</Badge>
                          )}
                        </div>
                        <p className="truncate text-xs text-muted">
                          {row.email}
                        </p>
                        {row.note && (
                          <p className="text-sm text-fg">{row.note}</p>
                        )}
                      </div>
                      <div className="shrink-0 text-right">
                        <Badge tone={st.tone}>{st.label}</Badge>
                        <p className="mt-1 text-xs text-muted">
                          {formatEventRange(
                            row.updatedAt,
                            undefined,
                            parokiTimezone,
                          )}
                        </p>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {rsvpsQuery.hasNextPage && (
              <div className="flex justify-center pt-2">
                <Button
                  type="button"
                  variant="outline"
                  disabled={rsvpsQuery.isFetchingNextPage}
                  onClick={() => rsvpsQuery.fetchNextPage()}
                  className="w-full sm:w-auto"
                >
                  {rsvpsQuery.isFetchingNextPage
                    ? "Memuat…"
                    : t("event.loadMore")}
                </Button>
              </div>
            )}
          </Card>
        </>
      )}

      {canManage && (
        <InductiveDangerZone
          title="Zona Berbahaya"
          description="Tindakan di bawah ini membatalkan, menjadwalkan kembali, atau menghapus kegiatan secara permanen."
          action={
            <div className="flex flex-col gap-2 sm:flex-row">
              <Button
                type="button"
                variant="outline"
                disabled={updateEvent.isPending}
                onClick={() => setPendingCancelToggle(true)}
                className="w-full sm:w-auto"
              >
                {isCancelled
                  ? t("event.rescheduleEvent")
                  : t("event.cancelEvent")}
              </Button>
              <Button
                type="button"
                variant="danger"
                disabled={deleteEvent.isPending}
                onClick={() => setPendingDeleteEvent(true)}
                className="w-full sm:w-auto"
              >
                <Trash2 aria-hidden className="size-4" />
                {t("event.deleteEvent")}
              </Button>
            </div>
          }
        />
      )}

      <ConfirmDialog
        open={pendingCancelToggle}
        title={
          isCancelled ? t("event.rescheduleEvent") : t("event.cancelEvent")
        }
        description={
          isCancelled
            ? "Jadwalkan kembali kegiatan ini?"
            : "Batalkan kegiatan ini? Anggota tidak lagi dapat mengirim RSVP."
        }
        confirmLabel={
          isCancelled ? t("event.rescheduleEvent") : t("event.cancelEvent")
        }
        confirmVariant={isCancelled ? "primary" : "danger"}
        pending={updateEvent.isPending}
        onConfirm={confirmToggleCancel}
        onCancel={() => setPendingCancelToggle(false)}
      />

      <ConfirmDialog
        open={pendingDeleteEvent}
        title={t("event.deleteEvent")}
        description="Hapus kegiatan ini secara permanen?"
        confirmLabel={t("event.deleteEvent")}
        confirmVariant="danger"
        pending={deleteEvent.isPending}
        onConfirm={confirmDeleteEvent}
        onCancel={() => setPendingDeleteEvent(false)}
      />

      <ConfirmDialog
        open={pendingDeleteRsvp}
        title={t("event.removeRsvp")}
        description="Hapus konfirmasi kehadiran (RSVP) Anda?"
        confirmLabel={t("event.removeRsvp")}
        confirmVariant="danger"
        pending={deleteRSVP.isPending}
        onConfirm={confirmDeleteRSVP}
        onCancel={() => setPendingDeleteRsvp(false)}
      />
    </div>
  );
}
