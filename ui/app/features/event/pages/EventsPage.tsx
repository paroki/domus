import { CalendarDays, MapPin, Plus, Search } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useUpsertEventRSVP } from "../api/actions";
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

  const upsertRSVP = useUpsertEventRSVP();
  const items = eventsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

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
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={CalendarDays}
          title={t("event.noParishTitle")}
          description={t("event.noParishDescription")}
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("event.title") },
        ]}
        title={t("event.title")}
        subtitle={t("event.subtitle")}
        action={
          canCreateEvent ? (
            <Link
              to="/event/new"
              className={`${buttonClass("primary")} w-full sm:w-auto`}
            >
              <Plus aria-hidden className="size-4" />
              {t("event.createEvent")}
            </Link>
          ) : undefined
        }
      />

      <InductiveFilterBar className="flex-col items-stretch sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            variant={!myOrgsOnly ? "primary" : "outline"}
            onClick={() => setMyOrgsOnly(false)}
            className="min-h-10"
          >
            {t("event.allEvents")}
          </Button>
          <Button
            type="button"
            variant={myOrgsOnly ? "primary" : "outline"}
            onClick={() => setMyOrgsOnly(true)}
            className="min-h-10"
          >
            {t("event.myOrgsOnly")}
          </Button>
        </div>

        <div className="flex flex-1 flex-wrap items-center gap-2 sm:justify-end">
          <select
            aria-label={t("event.allOrgs")}
            value={orgFilter}
            onChange={(e) => setOrgFilter(e.target.value)}
            className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
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
            aria-label="Semua Jenis"
            value={jenisFilter}
            onChange={(e) => setJenisFilter(e.target.value)}
            className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
          >
            <option value="">Semua Jenis</option>
            {eventTypeOptions.map((opt) => (
              <option key={opt.code} value={opt.code}>
                {opt.label}
              </option>
            ))}
          </select>

          <select
            aria-label={t("event.allVisibility")}
            value={visibilityFilter}
            onChange={(e) => setVisibilityFilter(e.target.value)}
            className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
          >
            <option value="">{t("event.allVisibility")}</option>
            <option value="public">Publik Paroki</option>
            <option value="private">Internal Organisasi</option>
          </select>

          <select
            aria-label={t("event.allStatus")}
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
          >
            <option value="">{t("event.allStatus")}</option>
            <option value="scheduled">Terjadwal</option>
            <option value="cancelled">Dibatalkan</option>
          </select>

          <div className="relative w-full sm:w-auto sm:min-w-[220px] sm:flex-1">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted"
            />
            <input
              type="search"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder={t("event.searchPlaceholder")}
              className="h-11 w-full rounded-lg border border-line bg-canvas pr-3 pl-9 text-sm text-fg placeholder:text-muted"
            />
          </div>
        </div>
      </InductiveFilterBar>

      {(eventsQuery.isError || actionError) && (
        <div
          role="alert"
          className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {actionError ?? resolveEventError(eventsQuery.error)}
        </div>
      )}

      {eventsQuery.isLoading ? (
        <Card className="p-4 sm:p-6">
          <p className="text-sm text-muted">Memuat jadwal kegiatan…</p>
        </Card>
      ) : items.length === 0 ? (
        <InductiveEmptyState
          icon={CalendarDays}
          title="Kegiatan tidak ditemukan"
          description="Belum ada kegiatan yang sesuai dengan filter."
          action={
            canCreateEvent ? (
              <Link
                to="/event/new"
                className={`${buttonClass("primary")} w-full sm:w-auto`}
              >
                <Plus aria-hidden className="size-4" />
                {t("event.createEvent")}
              </Link>
            ) : undefined
          }
        />
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
              <Card key={ev.id} className="space-y-3 p-4 sm:p-6">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0 flex-1 space-y-1">
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
                    <h2 className="text-lg font-semibold text-fg">
                      <Link
                        to={`/event/${ev.id}`}
                        className="hover:text-accent hover:underline"
                      >
                        {ev.title}
                      </Link>
                    </h2>
                    <p className="text-sm text-muted">
                      {formatEventRange(ev.startAt, ev.endAt, parokiTimezone)}
                    </p>
                    {ev.location && (
                      <p className="flex items-center gap-1.5 text-sm text-muted">
                        <MapPin aria-hidden className="size-3.5 shrink-0" />
                        <span className="truncate">{ev.location}</span>
                      </p>
                    )}
                  </div>
                </div>

                {ev.description && (
                  <p className="line-clamp-2 text-sm text-muted">
                    {ev.description}
                  </p>
                )}

                {ev.rsvpSummary && (
                  <div className="flex flex-wrap items-center gap-3 text-xs text-muted">
                    <span>Hadir: {ev.rsvpSummary.going}</span>
                    <span>·</span>
                    <span>Mungkin: {ev.rsvpSummary.maybe}</span>
                    <span>·</span>
                    <span>Tidak Hadir: {ev.rsvpSummary.notGoing}</span>
                  </div>
                )}

                {Boolean(ev.canRsvp) && ev.rsvpOpen && (
                  <div className="flex flex-col gap-2 border-t border-line pt-3 sm:flex-row sm:items-center">
                    <span className="text-xs font-medium text-muted">
                      Konfirmasi Cepat:
                    </span>
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
                          onClick={() =>
                            handleQuickRSVP(ev.id ?? "", opt.status)
                          }
                          className="w-full sm:w-auto"
                        >
                          {opt.label}
                        </Button>
                      ))}
                    </div>
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
                className="w-full sm:w-auto"
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
