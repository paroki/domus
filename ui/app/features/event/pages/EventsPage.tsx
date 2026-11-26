import { CalendarDays, MapPin, Plus, Search } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useCreateEvent, useUpsertEventRSVP } from "../api/actions";
import { resolveEventError } from "../api/errors";
import {
  useActiveParoki,
  useEventOrgs,
  useEvents,
  useEventTypeLookups,
} from "../api/queries";
import {
  computeEventLifecycle,
  DEFAULT_PAROKI_TIMEZONE,
  type EventStatus,
  type EventVisibility,
  eventTypeLabel,
  formatEventRange,
  fromLocalDatetimeInputValue,
  lifecycleMeta,
  mergeEventTypeOptions,
  type RSVPStatus,
  rsvpStatusMeta,
  visibilityMeta,
} from "../model";
import type { Route } from "./+types/EventsPage";

const PAROKI_OPTION = "__PAROKI__";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("event.title")} · Domus` }];
}

export default function EventsPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const [myOrgsOnly, setMyOrgsOnly] = useState(false);
  const [orgFilter, setOrgFilter] = useState("");
  const [jenisFilter, setJenisFilter] = useState("");
  const [visibilityFilter, setVisibilityFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [q, setQ] = useState("");

  const [showCreate, setShowCreate] = useState(false);
  const [orgId, setOrgId] = useState("");
  const [title, setTitle] = useState("");
  const [jenis, setJenis] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [location, setLocation] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<EventVisibility>("public");
  const [rsvpEnabled, setRsvpEnabled] = useState(true);
  const [rsvpDeadline, setRsvpDeadline] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const activeParokiQuery = useActiveParoki(hasParish);
  const activeParoki = activeParokiQuery.data;
  const parokiTimezone = activeParoki?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const isParokiAdmin = Boolean(me?.roles?.includes("ref:paroki-admin"));

  const orgsQuery = useEventOrgs(hasParish);
  const orgs = orgsQuery.data ?? [];
  const adminOrgs = orgs.filter((o) => o.myRole === "admin");
  const canCreateEvent = isParokiAdmin || adminOrgs.length > 0;

  const typeLookupsQuery = useEventTypeLookups(hasParish);
  const typeLookups = typeLookupsQuery.data ?? [];
  const eventTypeOptions = mergeEventTypeOptions(typeLookups);

  const eventsQuery = useEvents(
    {
      q: q.trim() || undefined,
      filter: {
        orgId: orgFilter && orgFilter !== PAROKI_OPTION ? orgFilter : undefined,
        parokiWideOnly: orgFilter === PAROKI_OPTION ? true : undefined,
        jenis: jenisFilter || undefined,
        visibility: (visibilityFilter as EventVisibility) || undefined,
        status: (statusFilter as EventStatus) || undefined,
        myOrgsOnly: myOrgsOnly || undefined,
      },
    },
    hasParish,
  );

  const createEvent = useCreateEvent();
  const upsertRSVP = useUpsertEventRSVP();

  const items = eventsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const selectedOrgId =
    orgId || (isParokiAdmin ? PAROKI_OPTION : adminOrgs[0]?.id || "");
  const isCreatingParokiEvent = selectedOrgId === PAROKI_OPTION;

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    if (!selectedOrgId) {
      setFormError("Pilih penyelenggara kegiatan.");
      return;
    }
    try {
      await createEvent.mutateAsync({
        orgId: isCreatingParokiEvent ? undefined : selectedOrgId,
        title: title.trim(),
        jenis: jenis || undefined,
        startAt: fromLocalDatetimeInputValue(startAt, parokiTimezone),
        endAt: fromLocalDatetimeInputValue(endAt, parokiTimezone),
        location: location.trim() || undefined,
        description: description.trim() || undefined,
        visibility: isCreatingParokiEvent ? "public" : visibility,
        rsvpEnabled,
        rsvpDeadline: rsvpDeadline
          ? fromLocalDatetimeInputValue(rsvpDeadline, parokiTimezone)
          : undefined,
      });
      setTitle("");
      setJenis("");
      setStartAt("");
      setEndAt("");
      setLocation("");
      setDescription("");
      setVisibility("public");
      setRsvpEnabled(true);
      setRsvpDeadline("");
      setShowCreate(false);
    } catch (err) {
      setFormError(resolveEventError(err));
    }
  }

  async function handleQuickRSVP(eventId: string, status: RSVPStatus) {
    setActionError(null);
    try {
      await upsertRSVP.mutateAsync({ eventId, body: { status } });
    } catch (err) {
      setActionError(resolveEventError(err));
    }
  }

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

  return (
    <div className="mx-auto max-w-5xl space-y-6 px-4 py-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold text-ink">
            <CalendarDays className="h-6 w-6 text-accent" />
            {t("event.title")}
          </h1>
          <p className="mt-1 text-sm text-ink-muted">{t("event.subtitle")}</p>
        </div>
        {canCreateEvent && (
          <Button
            type="button"
            onClick={() => {
              setShowCreate((v) => !v);
              setFormError(null);
            }}
          >
            <Plus className="h-4 w-4" />
            {t("event.createEvent")}
          </Button>
        )}
      </div>

      {showCreate && canCreateEvent && (
        <Card>
          <h2 className="text-lg font-semibold text-ink">
            {t("event.createEvent")}
          </h2>
          <form onSubmit={handleCreate} className="mt-4 space-y-4">
            {formError && (
              <div className="rounded-lg border border-danger-fg/30 bg-danger-bg px-3 py-2 text-sm text-danger-fg">
                {formError}
              </div>
            )}
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block text-sm font-medium text-ink">
                Penyelenggara
                <select
                  value={selectedOrgId}
                  onChange={(e) => {
                    const next = e.target.value;
                    setOrgId(next);
                    if (next === PAROKI_OPTION) {
                      setVisibility("public");
                    }
                  }}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                >
                  {isParokiAdmin && (
                    <option value={PAROKI_OPTION}>
                      Paroki (Tingkat Paroki)
                    </option>
                  )}
                  {adminOrgs.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name} ({o.code})
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-medium text-ink">
                Visibilitas
                <select
                  value={isCreatingParokiEvent ? "public" : visibility}
                  onChange={(e) =>
                    setVisibility(e.target.value as EventVisibility)
                  }
                  disabled={isCreatingParokiEvent}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink disabled:opacity-60"
                >
                  <option value="public">Publik Paroki</option>
                  <option value="private" disabled={isCreatingParokiEvent}>
                    Internal Organisasi
                  </option>
                </select>
              </label>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <label className="block text-sm font-medium text-ink">
                Judul Kegiatan
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Contoh: Rapat Pleno DPP"
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
                Lokasi (opsional)
                <input
                  type="text"
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  placeholder="Aula Paroki / Gereja"
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                />
              </label>
              <label className="block text-sm font-medium text-ink">
                Batas Waktu RSVP (opsional)
                <input
                  type="datetime-local"
                  value={rsvpDeadline}
                  onChange={(e) => setRsvpDeadline(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                />
              </label>
            </div>

            <label className="block text-sm font-medium text-ink">
              Deskripsi (opsional)
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
              Aktifkan konfirmasi kehadiran (RSVP) anggota organisasi
            </label>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setShowCreate(false)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={createEvent.isPending}>
                {createEvent.isPending ? "Menyimpan…" : "Simpan Kegiatan"}
              </Button>
            </div>
          </form>
        </Card>
      )}

      <Card>
        <div className="flex flex-wrap items-center gap-3">
          <div className="inline-flex rounded-lg border border-line p-0.5">
            <button
              type="button"
              onClick={() => setMyOrgsOnly(false)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition ${
                !myOrgsOnly
                  ? "bg-accent text-white"
                  : "text-ink-muted hover:text-ink"
              }`}
            >
              {t("event.allEvents")}
            </button>
            <button
              type="button"
              onClick={() => setMyOrgsOnly(true)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition ${
                myOrgsOnly
                  ? "bg-accent text-white"
                  : "text-ink-muted hover:text-ink"
              }`}
            >
              {t("event.myOrgsOnly")}
            </button>
          </div>

          <select
            value={orgFilter}
            onChange={(e) => setOrgFilter(e.target.value)}
            className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
          >
            <option value="">{t("event.allOrgs")}</option>
            <option value={PAROKI_OPTION}>Paroki (Tingkat Paroki)</option>
            {orgs.map((o) => (
              <option key={o.id} value={o.id}>
                {o.name}
              </option>
            ))}
          </select>

          <select
            value={jenisFilter}
            onChange={(e) => setJenisFilter(e.target.value)}
            className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
          >
            <option value="">Semua Jenis</option>
            {eventTypeOptions.map((opt) => (
              <option key={opt.code} value={opt.code}>
                {opt.label}
              </option>
            ))}
          </select>

          <select
            value={visibilityFilter}
            onChange={(e) => setVisibilityFilter(e.target.value)}
            className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
          >
            <option value="">{t("event.allVisibility")}</option>
            <option value="public">Publik Paroki</option>
            <option value="private">Internal Organisasi</option>
          </select>

          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
          >
            <option value="">{t("event.allStatus")}</option>
            <option value="scheduled">Terjadwal</option>
            <option value="cancelled">Dibatalkan</option>
          </select>

          <div className="relative min-w-[200px] flex-1">
            <Search className="pointer-events-none absolute top-2.5 left-3 h-4 w-4 text-ink-muted" />
            <input
              type="search"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder={t("event.searchPlaceholder")}
              className="w-full rounded-lg border border-line bg-surface py-2 pr-3 pl-9 text-sm text-ink"
            />
          </div>
        </div>
      </Card>

      {(eventsQuery.isError || actionError) && (
        <div className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
          {actionError ?? resolveEventError(eventsQuery.error)}
        </div>
      )}

      {eventsQuery.isLoading ? (
        <Card>
          <p className="text-sm text-ink-muted">Memuat jadwal kegiatan…</p>
        </Card>
      ) : items.length === 0 ? (
        <Card>
          <p className="text-sm text-ink-muted">
            Belum ada kegiatan yang sesuai dengan filter.
          </p>
        </Card>
      ) : (
        <div className="space-y-4">
          {items.map((ev) => {
            const lifecycle = computeEventLifecycle(ev);
            const lcMeta = lifecycleMeta(lifecycle);
            const visMeta = visibilityMeta(ev.visibility ?? "public");
            const myRsvpMeta = ev.myRsvp
              ? rsvpStatusMeta(ev.myRsvp.status ?? "")
              : null;

            return (
              <Card key={ev.id} className="space-y-3">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge tone="neutral">
                        {ev.orgId
                          ? ev.orgName
                          : (activeParoki?.nama ?? "Paroki")}
                      </Badge>
                      {ev.jenis && (
                        <Badge tone="neutral">
                          {eventTypeLabel(ev.jenis, typeLookups)}
                        </Badge>
                      )}
                      <Badge tone={visMeta.tone}>{visMeta.label}</Badge>
                      <Badge tone={lcMeta.tone}>{lcMeta.label}</Badge>
                      {myRsvpMeta && (
                        <Badge tone={myRsvpMeta.tone}>
                          RSVP Anda: {myRsvpMeta.label}
                        </Badge>
                      )}
                    </div>
                    <h2 className="text-lg font-semibold text-ink">
                      <Link
                        to={`/event/${ev.id}`}
                        className="hover:text-accent hover:underline"
                      >
                        {ev.title}
                      </Link>
                    </h2>
                    <p className="text-sm text-ink-muted">
                      {formatEventRange(ev.startAt, ev.endAt, parokiTimezone)}
                    </p>
                    {ev.location && (
                      <p className="flex items-center gap-1.5 text-sm text-ink-muted">
                        <MapPin className="h-3.5 w-3.5" />
                        {ev.location}
                      </p>
                    )}
                  </div>
                </div>

                {ev.description && (
                  <p className="line-clamp-2 text-sm text-ink-muted">
                    {ev.description}
                  </p>
                )}

                {ev.rsvpSummary && (
                  <div className="flex flex-wrap items-center gap-3 text-xs text-ink-muted">
                    <span>Hadir: {ev.rsvpSummary.going}</span>
                    <span>·</span>
                    <span>Mungkin: {ev.rsvpSummary.maybe}</span>
                    <span>·</span>
                    <span>Tidak Hadir: {ev.rsvpSummary.notGoing}</span>
                  </div>
                )}

                {Boolean(ev.canRsvp) && ev.rsvpOpen && (
                  <div className="flex flex-wrap items-center gap-2 border-t border-line pt-3">
                    <span className="text-xs font-medium text-ink-muted">
                      Konfirmasi Cepat:
                    </span>
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
                        onClick={() => handleQuickRSVP(ev.id ?? "", opt.status)}
                      >
                        {opt.label}
                      </Button>
                    ))}
                  </div>
                )}
              </Card>
            );
          })}

          {eventsQuery.hasNextPage && (
            <div className="flex justify-center">
              <Button
                type="button"
                variant="outline"
                disabled={eventsQuery.isFetchingNextPage}
                onClick={() => eventsQuery.fetchNextPage()}
              >
                {eventsQuery.isFetchingNextPage
                  ? "Memuat…"
                  : t("event.loadMore")}
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
