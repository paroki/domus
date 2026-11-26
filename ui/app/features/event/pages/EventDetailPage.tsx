import { ArrowLeft, CalendarDays, MapPin, Pencil, Trash2 } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
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
import { LaguSection } from "../components/LaguSection";
import { PetugasSection } from "../components/PetugasSection";
import {
  computeEventLifecycle,
  DEFAULT_PAROKI_TIMEZONE,
  type EventVisibility,
  eventTypeLabel,
  formatEventRange,
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
      <div className="mx-auto max-w-4xl px-4 py-8">
        <Card>
          <h1 className="text-xl font-bold text-ink">
            {t("event.noParishTitle")}
          </h1>
          <p className="mt-2 text-sm text-ink-muted">
            {t("event.noParishDescription")}
          </p>
        </Card>
      </div>
    );
  }

  if (eventQuery.isLoading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <Card>
          <p className="text-sm text-ink-muted">Memuat detail kegiatan…</p>
        </Card>
      </div>
    );
  }

  if (eventQuery.isError || !ev) {
    return (
      <div className="mx-auto max-w-4xl space-y-4 px-4 py-8">
        <Link
          to="/event"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:underline"
        >
          <ArrowLeft className="h-4 w-4" />
          {t("event.backToList")}
        </Link>
        <Card>
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

  async function handleToggleCancel() {
    if (!ev) return;
    const nextStatus = ev.status === "cancelled" ? "scheduled" : "cancelled";
    const promptMsg =
      nextStatus === "cancelled"
        ? "Batalkan kegiatan ini? Anggota tidak lagi dapat mengirim RSVP."
        : "Jadwalkan kembali kegiatan ini?";
    if (!window.confirm(promptMsg)) return;
    setActionError(null);
    try {
      await updateEvent.mutateAsync({ status: nextStatus });
    } catch (err) {
      setActionError(resolveEventError(err));
    }
  }

  async function handleDeleteEvent() {
    if (!window.confirm("Hapus kegiatan ini secara permanen?")) return;
    setActionError(null);
    try {
      await deleteEvent.mutateAsync(eventId);
      navigate("/event", { replace: true });
    } catch (err) {
      setActionError(resolveEventError(err));
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

  async function handleDeleteRSVP() {
    if (!window.confirm("Hapus konfirmasi kehadiran (RSVP) Anda?")) return;
    setActionError(null);
    try {
      await deleteRSVP.mutateAsync();
      setRsvpNote("");
    } catch (err) {
      setActionError(resolveEventError(err));
    }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6 px-4 py-8">
      <Link
        to="/event"
        className="inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:underline"
      >
        <ArrowLeft className="h-4 w-4" />
        {t("event.backToList")}
      </Link>

      {actionError && (
        <div className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
          {actionError}
        </div>
      )}

      <Card className="space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-2">
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
            <h1 className="text-2xl font-bold text-ink">{ev.title}</h1>
            <p className="flex items-center gap-2 text-sm text-ink-muted">
              <CalendarDays className="h-4 w-4" />
              {formatEventRange(ev.startAt, ev.endAt, parokiTimezone)}
            </p>
            {ev.location && (
              <p className="flex items-center gap-2 text-sm text-ink-muted">
                <MapPin className="h-4 w-4" />
                {ev.location}
              </p>
            )}
            <p className="text-xs text-ink-muted">
              {ev.rsvpEnabled
                ? ev.rsvpDeadline
                  ? `Batas RSVP: ${formatEventRange(ev.rsvpDeadline, undefined, parokiTimezone)}`
                  : "RSVP dibuka hingga 30 menit setelah acara dimulai"
                : "RSVP dinonaktifkan untuk kegiatan ini"}
            </p>
          </div>

          {canManage && (
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing((v) => !v)}
              >
                <Pencil className="h-4 w-4" />
                {t("event.editEvent")}
              </Button>
              <Button
                type="button"
                variant="outline"
                disabled={updateEvent.isPending}
                onClick={handleToggleCancel}
              >
                {ev.status === "cancelled"
                  ? t("event.rescheduleEvent")
                  : t("event.cancelEvent")}
              </Button>
              <Button
                type="button"
                variant="danger"
                disabled={deleteEvent.isPending}
                onClick={handleDeleteEvent}
              >
                <Trash2 className="h-4 w-4" />
                {t("event.deleteEvent")}
              </Button>
            </div>
          )}
        </div>

        {ev.description && (
          <p className="whitespace-pre-line border-t border-line pt-4 text-sm text-ink">
            {ev.description}
          </p>
        )}
      </Card>

      {editing && canManage && (
        <Card>
          <h2 className="text-lg font-semibold text-ink">
            {t("event.editEvent")}
          </h2>
          <form onSubmit={handleSaveEdit} className="mt-4 space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block text-sm font-medium text-ink">
                Judul Kegiatan
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                />
              </label>
              <label className="block text-sm font-medium text-ink">
                Jenis Kegiatan
                <select
                  value={jenis}
                  onChange={(e) => setJenis(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
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
              <label className="block text-sm font-medium text-ink">
                Waktu Mulai
                <input
                  type="datetime-local"
                  value={startAt}
                  onChange={(e) => setStartAt(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                />
              </label>
              <label className="block text-sm font-medium text-ink">
                Waktu Selesai
                <input
                  type="datetime-local"
                  value={endAt}
                  onChange={(e) => setEndAt(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                />
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block text-sm font-medium text-ink">
                Lokasi
                <input
                  type="text"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                />
              </label>
              <label className="block text-sm font-medium text-ink">
                Visibilitas
                <select
                  value={ev.orgId ? visibility : "public"}
                  onChange={(e) =>
                    setVisibility(e.target.value as EventVisibility)
                  }
                  disabled={!ev.orgId}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink disabled:opacity-60"
                >
                  <option value="public">Publik Paroki</option>
                  <option value="private" disabled={!ev.orgId}>
                    Internal Organisasi
                  </option>
                </select>
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block text-sm font-medium text-ink">
                Batas Waktu RSVP
                <input
                  type="datetime-local"
                  value={rsvpDeadline}
                  disabled={clearRsvpDeadline}
                  onChange={(e) => setRsvpDeadline(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink disabled:opacity-50"
                />
              </label>
              <div className="flex items-end pb-2">
                <label className="flex items-center gap-2 text-sm text-ink">
                  <input
                    type="checkbox"
                    checked={clearRsvpDeadline}
                    onChange={(e) => setClearRsvpDeadline(e.target.checked)}
                    className="h-4 w-4 rounded border-line"
                  />
                  Hapus batas waktu RSVP
                </label>
              </div>
            </div>

            <label className="block text-sm font-medium text-ink">
              Deskripsi
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>

            <label className="flex items-center gap-2 text-sm text-ink">
              <input
                type="checkbox"
                checked={rsvpEnabled}
                onChange={(e) => setRsvpEnabled(e.target.checked)}
                className="h-4 w-4 rounded border-line"
              />
              Aktifkan konfirmasi kehadiran (RSVP)
            </label>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setEditing(false)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={updateEvent.isPending}>
                {updateEvent.isPending ? "Menyimpan…" : "Simpan Perubahan"}
              </Button>
            </div>
          </form>
        </Card>
      )}

      <PetugasSection
        eventId={eventId}
        parokiId={me?.activeParokiId ?? ""}
        orgId={ev.orgId}
        canManage={canManage}
        petugas={ev.petugas ?? []}
      />

      <LaguSection
        eventId={eventId}
        canManage={canManage}
        lagu={ev.lagu ?? []}
      />

      {!canRsvp ? (
        <Card>
          <h2 className="text-base font-semibold text-ink">
            {t("event.rsvpRestrictedTitle")}
          </h2>
          {ev.orgId ? (
            <p className="mt-1 text-sm text-ink-muted">
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
            <p className="mt-1 text-sm text-ink-muted">
              Konfirmasi kehadiran (RSVP) dan daftar peserta hanya tersedia bagi
              umat terdaftar di paroki ini.
            </p>
          )}
        </Card>
      ) : (
        <>
          <Card className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="text-lg font-semibold text-ink">
                Konfirmasi Kehadiran (RSVP)
              </h2>
              {myRsvpMeta && (
                <Badge tone={myRsvpMeta.tone}>Status: {myRsvpMeta.label}</Badge>
              )}
            </div>

            {ev.myRsvp?.note && (
              <p className="text-sm text-ink-muted">
                Catatan Anda: <span className="text-ink">{ev.myRsvp.note}</span>
              </p>
            )}

            {ev.rsvpOpen ? (
              <div className="space-y-3">
                <input
                  type="text"
                  value={rsvpNote}
                  onChange={(e) => setRsvpNote(e.target.value)}
                  placeholder="Catatan kehadiran (opsional)"
                  className="w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                />
                <div className="flex flex-wrap items-center gap-2">
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
                        ev.myRsvp?.status === opt.status ? "primary" : "outline"
                      }
                      disabled={upsertRSVP.isPending}
                      onClick={() => handleSubmitRSVP(opt.status)}
                    >
                      {opt.label}
                    </Button>
                  ))}
                  {ev.myRsvp && (
                    <Button
                      type="button"
                      variant="danger"
                      disabled={deleteRSVP.isPending}
                      onClick={handleDeleteRSVP}
                    >
                      {t("event.removeRsvp")}
                    </Button>
                  )}
                </div>
              </div>
            ) : (
              <p className="text-sm text-ink-muted">{t("event.rsvpClosed")}</p>
            )}
          </Card>

          {summary && (
            <div className="grid gap-4 sm:grid-cols-4">
              <Card>
                <p className="text-xs font-medium text-ink-muted">Hadir</p>
                <p className="mt-1 text-2xl font-bold text-success-fg">
                  {summary.going}
                </p>
              </Card>
              <Card>
                <p className="text-xs font-medium text-ink-muted">Mungkin</p>
                <p className="mt-1 text-2xl font-bold text-warning-fg">
                  {summary.maybe}
                </p>
              </Card>
              <Card>
                <p className="text-xs font-medium text-ink-muted">
                  Tidak Hadir
                </p>
                <p className="mt-1 text-2xl font-bold text-danger-fg">
                  {summary.notGoing}
                </p>
              </Card>
              <Card>
                <p className="text-xs font-medium text-ink-muted">
                  Belum Konfirmasi
                </p>
                <p className="mt-1 text-2xl font-bold text-ink">
                  {summary.unresponded}
                </p>
                <p className="text-xs text-ink-muted">
                  {ev.orgId
                    ? `dari ${summary.totalActiveMembers} anggota aktif`
                    : `dari ${summary.totalActiveMembers} umat paroki aktif`}
                </p>
              </Card>
            </div>
          )}

          <Card className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-ink">
                {t("event.rsvpRosterTitle")}
              </h2>
              <div className="flex flex-wrap gap-2">
                <select
                  value={rosterStatus}
                  onChange={(e) => setRosterStatus(e.target.value)}
                  className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
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
                  className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                />
              </div>
            </div>

            {rsvpsQuery.isLoading ? (
              <p className="text-sm text-ink-muted">Memuat daftar peserta…</p>
            ) : rsvpRows.length === 0 ? (
              <p className="text-sm text-ink-muted">
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
                      <div className="space-y-0.5">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-medium text-ink">
                            {row.name}
                          </span>
                          {row.orgRole && (
                            <Badge tone="neutral">{row.orgRole}</Badge>
                          )}
                        </div>
                        <p className="text-xs text-ink-muted">{row.email}</p>
                        {row.note && (
                          <p className="text-sm text-ink">{row.note}</p>
                        )}
                      </div>
                      <div className="text-right">
                        <Badge tone={st.tone}>{st.label}</Badge>
                        <p className="mt-1 text-xs text-ink-muted">
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
    </div>
  );
}
