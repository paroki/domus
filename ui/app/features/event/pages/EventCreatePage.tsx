import { CalendarDays } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import {
  InductiveEmptyState,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useCreateEvent } from "../api/actions";
import { resolveEventError } from "../api/errors";
import {
  useActiveParoki,
  useEventOrgs,
  useEventTypeLookups,
} from "../api/queries";
import {
  DEFAULT_PAROKI_TIMEZONE,
  type EventVisibility,
  formatParokiTimezoneShort,
  fromLocalDatetimeInputValue,
  mergeEventTypeOptions,
} from "../model";
import type { Route } from "./+types/EventCreatePage";

const PAROKI_OPTION = "__PAROKI__";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("event.createEvent")} · Domus` }];
}

export default function EventCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const activeParokiQuery = useActiveParoki(hasParish);
  const activeParoki = activeParokiQuery.data;
  const parokiTimezone = activeParoki?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const tzShort = formatParokiTimezoneShort(parokiTimezone);
  const isParokiAdmin = Boolean(me?.roles?.includes("ref:paroki-admin"));

  const orgsQuery = useEventOrgs(hasParish);
  const orgs = orgsQuery.data ?? [];
  const adminOrgs = orgs.filter((o) => o.myRole === "admin");
  const canCreateEvent = isParokiAdmin || adminOrgs.length > 0;

  const typeLookupsQuery = useEventTypeLookups(hasParish);
  const eventTypeOptions = mergeEventTypeOptions(typeLookupsQuery.data ?? []);

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

  const createEvent = useCreateEvent();

  const selectedOrgId =
    orgId || (isParokiAdmin ? PAROKI_OPTION : adminOrgs[0]?.id || "");
  const isCreatingParokiEvent = selectedOrgId === PAROKI_OPTION;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    if (!selectedOrgId) {
      setFormError("Pilih penyelenggara kegiatan.");
      return;
    }
    try {
      const created = await createEvent.mutateAsync({
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
      navigate(created?.id ? `/event/${created.id}` : "/event", {
        replace: true,
      });
    } catch (err) {
      setFormError(resolveEventError(err));
    }
  }

  if (!meLoading && !hasParish) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={CalendarDays}
          title={t("event.noParishTitle")}
          description={t("event.noParishDescription")}
        />
      </div>
    );
  }

  if (!meLoading && !orgsQuery.isLoading && !canCreateEvent) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={CalendarDays}
          title="Akses Terbatas"
          description="Hanya Admin Paroki atau Admin Organisasi yang dapat membuat kegiatan baru."
          action={
            <Link to="/event" className={buttonClass("outline")}>
              Kembali ke Daftar Kegiatan
            </Link>
          }
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: t("event.title"), to: "/event" },
          { label: t("event.createEvent") },
        ]}
        title={t("event.createEvent")}
        subtitle="Jadwalkan kegiatan paroki atau organisasi dengan waktu lokal paroki dan opsi RSVP."
      />

      {formError && (
        <div
          role="alert"
          className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg"
        >
          {formError}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col gap-6">
        <InductiveFormSection
          title="Informasi Dasar"
          description="Tentukan penyelenggara, visibilitas, judul, dan jenis kegiatan."
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
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
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                required
              >
                {isParokiAdmin && (
                  <option value={PAROKI_OPTION}>Paroki (Tingkat Paroki)</option>
                )}
                {adminOrgs.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name} ({o.code})
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Visibilitas
              <select
                value={isCreatingParokiEvent ? "public" : visibility}
                onChange={(e) =>
                  setVisibility(e.target.value as EventVisibility)
                }
                disabled={isCreatingParokiEvent}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg disabled:opacity-60"
              >
                <option value="public">Publik Paroki</option>
                <option value="private" disabled={isCreatingParokiEvent}>
                  Internal Organisasi
                </option>
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Judul Kegiatan
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Contoh: Rapat Pleno DPP"
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
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

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Deskripsi (opsional)
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-sm text-fg"
              />
            </label>
          </div>
        </InductiveFormSection>

        <InductiveFormSection
          title="Waktu, Lokasi & RSVP"
          description={`Seluruh jadwal menggunakan zona waktu paroki aktif (${tzShort}).`}
        >
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

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Lokasi (opsional)
              <input
                type="text"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="Aula Paroki / Gereja"
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg placeholder:text-muted"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Batas Waktu RSVP (opsional · {tzShort})
              <input
                type="datetime-local"
                value={rsvpDeadline}
                onChange={(e) => setRsvpDeadline(e.target.value)}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>
          </div>

          <label className="flex min-h-11 cursor-pointer items-center gap-3 text-sm text-fg">
            <input
              type="checkbox"
              checked={rsvpEnabled}
              onChange={(e) => setRsvpEnabled(e.target.checked)}
              className="size-4 rounded border-line"
            />
            {isCreatingParokiEvent
              ? "Aktifkan konfirmasi kehadiran (RSVP) umat paroki"
              : "Aktifkan konfirmasi kehadiran (RSVP) anggota organisasi"}
          </label>
        </InductiveFormSection>

        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Link
            to="/event"
            className={`${buttonClass("outline")} w-full sm:w-auto`}
          >
            Batal
          </Link>
          <Button
            type="submit"
            disabled={createEvent.isPending}
            className="w-full sm:w-auto"
          >
            {createEvent.isPending ? "Menyimpan…" : "Simpan Kegiatan"}
          </Button>
        </div>
      </form>
    </div>
  );
}
