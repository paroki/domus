import {
  CalendarDays,
  Church,
  MapPin,
  Plus,
  Search,
  Users,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useSearchParams } from "react-router";
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
import { useActiveParoki, useLiturgiSchedules } from "../api/queries";
import type { LiturgiSchedule } from "../api/types";
import {
  canManageLiturgiGlobal,
  DEFAULT_PAROKI_TIMEZONE,
  formatLiturgiRange,
  fromLocalDatetimeInputValue,
  groupPetugasByRole,
  isUserAssignedToSchedule,
  LITURGI_MINISTRIES,
  type LiturgiRoleSlug,
  petugasStatusMeta,
} from "../model";
import type { Route } from "./+types/LiturgiSchedulesPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.title")} · Domus` }];
}

export default function LiturgiSchedulesPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const { data: activeParoki } = useActiveParoki();
  const [searchParams, setSearchParams] = useSearchParams();

  const qParam = searchParams.get("q") ?? "";
  const fromParam = searchParams.get("from") ?? "";
  const toParam = searchParams.get("to") ?? "";
  const assignedToMe = searchParams.get("assignedToMe") === "true";
  const hasDeclined = searchParams.get("hasDeclined") === "true";
  const roleParam = (searchParams.get("role") ?? "") as LiturgiRoleSlug | "";

  const [qInput, setQInput] = useState(qParam);
  useEffect(() => {
    setQInput(qParam);
  }, [qParam]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const trimmed = qInput.trim();
      if (trimmed === qParam) return;
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          if (trimmed) next.set("q", trimmed);
          else next.delete("q");
          return next;
        },
        { replace: true },
      );
    }, 250);
    return () => window.clearTimeout(timer);
  }, [qInput, qParam, setSearchParams]);

  const activeParokiId = me?.activeParokiId ?? "";
  const parokiTimezone = activeParoki?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const canManage = canManageLiturgiGlobal(me?.roles);

  const fromIso = fromParam
    ? fromLocalDatetimeInputValue(`${fromParam}T00:00:00`, parokiTimezone)
    : undefined;
  const toIso = toParam
    ? fromLocalDatetimeInputValue(`${toParam}T23:59:59`, parokiTimezone)
    : undefined;

  const hasFilter =
    Boolean(fromIso) ||
    Boolean(toIso) ||
    assignedToMe ||
    hasDeclined ||
    Boolean(roleParam);

  const schedulesQuery = useLiturgiSchedules(
    {
      q: qParam || undefined,
      filter: hasFilter
        ? {
            from: fromIso,
            to: toIso,
            assignedToMe: assignedToMe || undefined,
            hasDeclined: hasDeclined || undefined,
            role: roleParam || undefined,
          }
        : undefined,
    },
    Boolean(activeParokiId),
  );

  const schedules: LiturgiSchedule[] =
    schedulesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  function updateParam(key: string, value?: string) {
    setSearchParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (value) next.set(key, value);
        else next.delete(key);
        return next;
      },
      { replace: true },
    );
  }

  function resetFilters() {
    setQInput("");
    setSearchParams(new URLSearchParams(), { replace: true });
  }

  if (me && !activeParokiId) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8">
        <InductiveEmptyState
          icon={Church}
          title={t("liturgi.noParishTitle")}
          description={t("liturgi.noParishDescription")}
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi") },
        ]}
        title={t("liturgi.title")}
        subtitle={t("liturgi.subtitle")}
        action={
          <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
            <Link
              to="/liturgi/ministries"
              className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
            >
              {t("liturgi.ministriesCta")}
            </Link>
            {canManage && (
              <Link
                to="/liturgi/new"
                className={`${buttonClass("primary")} h-11 w-full sm:w-auto`}
              >
                <Plus aria-hidden className="size-4" />
                {t("liturgi.createScheduleCta").replace(/^\+\s*/, "")}
              </Link>
            )}
          </div>
        }
      />

      <InductiveFilterBar className="flex-col items-stretch gap-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <div className="relative md:col-span-2">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted"
            />
            <input
              type="search"
              value={qInput}
              onChange={(e) => setQInput(e.target.value)}
              placeholder={t("liturgi.searchPlaceholder")}
              aria-label={t("liturgi.searchPlaceholder")}
              className="h-11 w-full rounded-lg border border-line bg-canvas pr-3.5 pl-10 text-base text-fg placeholder:text-muted sm:text-sm"
            />
          </div>

          <label className="flex flex-col gap-1 text-xs font-medium text-muted">
            {t("liturgi.filterFromLabel")}
            <input
              type="date"
              value={fromParam}
              onChange={(e) => updateParam("from", e.target.value)}
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
            />
          </label>

          <label className="flex flex-col gap-1 text-xs font-medium text-muted">
            {t("liturgi.filterToLabel")}
            <input
              type="date"
              value={toParam}
              onChange={(e) => updateParam("to", e.target.value)}
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
            />
          </label>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={resetFilters}
            className={`min-h-10 rounded-full px-3.5 text-xs font-semibold transition-colors ${
              !hasFilter && !qParam
                ? "bg-primary text-on-primary"
                : "border border-line bg-surface text-muted hover:text-fg"
            }`}
          >
            {t("liturgi.filterAll")}
          </button>

          <button
            type="button"
            onClick={() =>
              updateParam("assignedToMe", assignedToMe ? undefined : "true")
            }
            className={`min-h-10 rounded-full px-3.5 text-xs font-semibold transition-colors ${
              assignedToMe
                ? "bg-primary text-on-primary"
                : "border border-line bg-surface text-muted hover:text-fg"
            }`}
          >
            {t("liturgi.filterAssignedToMe")}
          </button>

          <button
            type="button"
            onClick={() =>
              updateParam("hasDeclined", hasDeclined ? undefined : "true")
            }
            className={`min-h-10 rounded-full px-3.5 text-xs font-semibold transition-colors ${
              hasDeclined
                ? "bg-danger text-on-danger"
                : "border border-line bg-surface text-muted hover:text-fg"
            }`}
          >
            {t("liturgi.filterHasDeclined")}
          </button>

          {LITURGI_MINISTRIES.map((m) => {
            const active = roleParam === m.slug;
            return (
              <button
                key={m.slug}
                type="button"
                onClick={() => updateParam("role", active ? undefined : m.slug)}
                className={`min-h-10 rounded-full px-3.5 text-xs font-semibold transition-colors ${
                  active
                    ? "bg-primary text-on-primary"
                    : "border border-line bg-surface text-muted hover:text-fg"
                }`}
              >
                {m.label}
              </button>
            );
          })}
        </div>
      </InductiveFilterBar>

      {schedulesQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">{t("auth.loading")}</Card>
      ) : schedules.length === 0 ? (
        <InductiveEmptyState
          icon={Church}
          title={t("liturgi.emptyTitle")}
          description={t("liturgi.emptyDescription")}
          action={
            hasFilter || qParam ? (
              <Button variant="outline" onClick={resetFilters}>
                {t("liturgi.filterReset")}
              </Button>
            ) : canManage ? (
              <Link
                to="/liturgi/new"
                className={`${buttonClass("primary")} h-11 w-full sm:w-auto`}
              >
                <Plus aria-hidden className="size-4" />
                {t("liturgi.createScheduleCta").replace(/^\+\s*/, "")}
              </Link>
            ) : undefined
          }
        />
      ) : (
        <>
          {/* Mobile Card Stack (< lg) */}
          <div className="space-y-3 lg:hidden">
            {schedules.map((s) => {
              const grouped = groupPetugasByRole(s.petugas);
              const myDuty = isUserAssignedToSchedule(s.petugas, me?.id);
              return (
                <Link
                  key={s.id}
                  to={`/liturgi/${s.id}`}
                  className="block rounded-xl border border-line bg-surface p-4 text-fg no-underline transition-colors hover:border-primary/50"
                >
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div className="min-w-0 flex-1">
                      <h2 className="text-base font-bold text-fg">{s.title}</h2>
                      <p className="mt-1 flex items-center gap-1.5 text-xs text-muted">
                        <CalendarDays
                          aria-hidden
                          className="size-3.5 shrink-0"
                        />
                        {formatLiturgiRange(s.startAt, s.endAt, parokiTimezone)}
                      </p>
                      {s.location && (
                        <p className="mt-0.5 flex items-center gap-1.5 text-xs text-muted">
                          <MapPin aria-hidden className="size-3.5 shrink-0" />
                          {s.location}
                        </p>
                      )}
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5">
                      {myDuty && (
                        <Badge tone="success">
                          {t("liturgi.yourDutyBadge")}
                        </Badge>
                      )}
                      {s.rsvpEnabled && s.rsvpSummary && (
                        <Badge tone="neutral">
                          <Users aria-hidden className="mr-1 inline size-3" />
                          {s.rsvpSummary.going ?? 0}
                        </Badge>
                      )}
                    </div>
                  </div>

                  <div className="mt-3 grid grid-cols-2 gap-2 border-t border-line pt-3 sm:grid-cols-3">
                    {LITURGI_MINISTRIES.map((m) => {
                      const slots = grouped[m.slug];
                      const hasDeclinedSlot = slots.some(
                        (p) => p.status === "declined",
                      );
                      const isEmpty = slots.length === 0;
                      const names = slots
                        .map((p) => p.name || p.slotLabel)
                        .filter(Boolean)
                        .join(", ");
                      return (
                        <div
                          key={m.slug}
                          className={`rounded-lg border px-2.5 py-2 text-xs ${
                            hasDeclinedSlot
                              ? "border-danger/40 bg-danger-bg/50"
                              : isEmpty
                                ? "border-line bg-canvas text-muted"
                                : "border-line bg-surface-2/40"
                          }`}
                        >
                          <div className="flex items-center justify-between gap-1 font-semibold">
                            <span>{m.shortLabel}</span>
                            <span
                              className={
                                hasDeclinedSlot
                                  ? "text-danger-fg"
                                  : "text-muted"
                              }
                            >
                              {slots.length}
                            </span>
                          </div>
                          <p className="mt-0.5 truncate text-[11px] text-muted">
                            {names || "—"}
                          </p>
                        </div>
                      );
                    })}
                  </div>
                </Link>
              );
            })}
          </div>

          {/* Desktop 7-Column Matrix (lg+) */}
          <div className="hidden overflow-x-auto rounded-xl border border-line bg-surface lg:block">
            <table className="w-full border-collapse text-left text-sm">
              <thead>
                <tr className="border-b border-line bg-surface-2/60 text-xs font-semibold text-muted">
                  <th className="min-w-56 px-4 py-3">
                    {t("liturgi.matrixScheduleCol")}
                  </th>
                  {LITURGI_MINISTRIES.map((m) => (
                    <th key={m.slug} className="min-w-40 px-3 py-3">
                      {m.label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {schedules.map((s) => {
                  const grouped = groupPetugasByRole(s.petugas);
                  const myDuty = isUserAssignedToSchedule(s.petugas, me?.id);
                  const manageable = new Set(s.manageableRoles ?? []);
                  return (
                    <tr key={s.id} className="align-top hover:bg-surface-2/25">
                      <td className="px-4 py-3">
                        <Link
                          to={`/liturgi/${s.id}`}
                          className="font-semibold text-fg no-underline hover:text-primary"
                        >
                          {s.title}
                        </Link>
                        <p className="mt-1 text-xs text-muted">
                          {formatLiturgiRange(
                            s.startAt,
                            s.endAt,
                            parokiTimezone,
                          )}
                        </p>
                        {s.location && (
                          <p className="mt-0.5 text-xs text-muted">
                            {s.location}
                          </p>
                        )}
                        <div className="mt-2 flex flex-wrap gap-1">
                          {myDuty && (
                            <Badge tone="success">
                              {t("liturgi.yourDutyBadge")}
                            </Badge>
                          )}
                          {s.rsvpEnabled && s.rsvpSummary && (
                            <Badge tone="neutral">
                              Hadir: {s.rsvpSummary.going ?? 0}
                            </Badge>
                          )}
                        </div>
                      </td>

                      {LITURGI_MINISTRIES.map((m) => {
                        const slots = grouped[m.slug];
                        const cellHref = manageable.has(m.slug)
                          ? `/liturgi/${s.id}/petugas/new?role=${m.slug}`
                          : `/liturgi/${s.id}`;
                        return (
                          <td key={m.slug} className="px-3 py-3">
                            {slots.length === 0 ? (
                              <Link
                                to={cellHref}
                                className="inline-block text-xs text-muted no-underline hover:text-primary"
                              >
                                —
                              </Link>
                            ) : (
                              <div className="flex flex-col gap-1.5">
                                {slots.map((slot) => {
                                  const st = petugasStatusMeta(slot.status);
                                  return (
                                    <Link
                                      key={slot.id}
                                      to={`/liturgi/${s.id}`}
                                      className="block rounded-lg border border-line bg-canvas p-2 text-xs no-underline transition-colors hover:border-primary/40"
                                    >
                                      {slot.slotLabel && (
                                        <span className="block text-[11px] font-medium text-muted">
                                          {slot.slotLabel}
                                        </span>
                                      )}
                                      <span className="block font-semibold text-fg">
                                        {slot.name || "—"}
                                      </span>
                                      <div className="mt-1 flex flex-wrap items-center gap-1">
                                        <Badge tone={st.tone}>{st.label}</Badge>
                                        {slot.substituteName && (
                                          <Badge tone="warning">
                                            Pengganti: {slot.substituteName}
                                          </Badge>
                                        )}
                                      </div>
                                    </Link>
                                  );
                                })}
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {schedulesQuery.hasNextPage && (
            <div className="flex justify-center">
              <Button
                variant="outline"
                disabled={schedulesQuery.isFetchingNextPage}
                onClick={() => schedulesQuery.fetchNextPage()}
              >
                {t("liturgi.loadMore")}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
