import { ArrowLeft, Pencil, Plus, Power, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useCreateWard,
  useCreateZone,
  useDeleteWard,
  useDeleteZone,
  useUpdateWard,
  useUpdateZone,
} from "../api/actions";
import {
  useCurrentIdentity,
  useWardsSearch,
  useZonesSearch,
} from "../api/queries";
import type { Ward, Zone } from "../api/types";
import { isParishAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Wilayah & Lingkungan — Domus" }];
}

export default function TerritoryPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const allowed = isParishAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);
  const enabled = allowed && hasTenant;

  const [zoneTypeFilter, setZoneTypeFilter] = useState<
    "" | "wilayah" | "stasi"
  >("");
  const [selectedZoneId, setSelectedZoneId] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Zone form state
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [zType, setZType] = useState<"wilayah" | "stasi">("wilayah");
  const [zCode, setZCode] = useState("");
  const [zName, setZName] = useState("");
  const [zCoordName, setZCoordName] = useState("");
  const [zCoordPhone, setZCoordPhone] = useState("");

  // Ward form state
  const [editingWard, setEditingWard] = useState<Ward | null>(null);
  const [wZoneId, setWZoneId] = useState("");
  const [wCode, setWCode] = useState("");
  const [wName, setWName] = useState("");
  const [wPatron, setWPatron] = useState("");
  const [wHeadName, setWHeadName] = useState("");
  const [wHeadPhone, setWHeadPhone] = useState("");

  const zonesQuery = useZonesSearch(
    { filter: zoneTypeFilter ? { jenis: zoneTypeFilter } : undefined },
    enabled,
  );
  const wardsQuery = useWardsSearch(
    { filter: selectedZoneId ? { wilayahId: selectedZoneId } : undefined },
    enabled,
  );
  const zones = zonesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const wards = wardsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createZone = useCreateZone();
  const updateZone = useUpdateZone();
  const deleteZone = useDeleteZone();
  const createWard = useCreateWard();
  const updateWard = useUpdateWard();
  const deleteWard = useDeleteWard();

  function resetZoneForm() {
    setEditingZone(null);
    setZType("wilayah");
    setZCode("");
    setZName("");
    setZCoordName("");
    setZCoordPhone("");
  }

  function startEditZone(z: Zone) {
    setEditingZone(z);
    setZType((z.jenis as "wilayah" | "stasi") || "wilayah");
    setZCode(z.kode ?? "");
    setZName(z.nama ?? "");
    setZCoordName(z.koordinator ?? "");
    setZCoordPhone(z.teleponKoordinator ?? "");
  }

  function resetWardForm() {
    setEditingWard(null);
    setWZoneId(selectedZoneId || zones[0]?.id || "");
    setWCode("");
    setWName("");
    setWPatron("");
    setWHeadName("");
    setWHeadPhone("");
  }

  function startEditWard(w: Ward) {
    setEditingWard(w);
    setWZoneId(w.wilayahId ?? "");
    setWCode(w.kode ?? "");
    setWName(w.nama ?? "");
    setWPatron(w.pelindung ?? "");
    setWHeadName(w.ketua ?? "");
    setWHeadPhone(w.teleponKetua ?? "");
  }

  async function handleSubmitZone(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const code = slugifyCode(zCode || zName);
      if (editingZone?.id) {
        await updateZone.mutateAsync({
          id: editingZone.id,
          body: {
            jenis: zType,
            kode: code,
            nama: zName.trim(),
            koordinator: zCoordName.trim(),
            teleponKoordinator: zCoordPhone.trim(),
          },
        });
      } else {
        const created = await createZone.mutateAsync({
          jenis: zType,
          kode: code,
          nama: zName.trim(),
          koordinator: zCoordName.trim() || undefined,
          teleponKoordinator: zCoordPhone.trim() || undefined,
        });
        if (!wZoneId && created.id) setWZoneId(created.id);
      }
      resetZoneForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan wilayah.",
      );
    }
  }

  async function handleSubmitWard(e: FormEvent) {
    e.preventDefault();
    const targetZoneId = wZoneId || selectedZoneId || zones[0]?.id;
    if (!targetZoneId) {
      setErrorMsg(t("ref.selectZoneFirst"));
      return;
    }
    setErrorMsg(null);
    try {
      const code = slugifyCode(wCode || wName);
      if (editingWard?.id) {
        await updateWard.mutateAsync({
          id: editingWard.id,
          body: {
            wilayahId: targetZoneId,
            kode: code,
            nama: wName.trim(),
            pelindung: wPatron.trim(),
            ketua: wHeadName.trim(),
            teleponKetua: wHeadPhone.trim(),
          },
        });
      } else {
        await createWard.mutateAsync({
          wilayahId: targetZoneId,
          kode: code,
          nama: wName.trim(),
          pelindung: wPatron.trim() || undefined,
          ketua: wHeadName.trim() || undefined,
          teleponKetua: wHeadPhone.trim() || undefined,
        });
      }
      resetWardForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan lingkungan.",
      );
    }
  }

  const zoneNameById = new Map(zones.map((z) => [z.id ?? "", z.nama ?? ""]));

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/ref"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("ref.backToSettings")}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">
          {t("ref.territoryTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.territorySubtitle")}</p>
      </div>

      {!hasTenant ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.noActiveParish")}
        </div>
      ) : !allowed ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenParishAdmin")}
        </div>
      ) : (
        <>
          {errorMsg && (
            <div
              role="alert"
              className="rounded-xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger"
            >
              {errorMsg}
            </div>
          )}

          <div className="grid gap-6 lg:grid-cols-2">
            {/* Zones / Stations */}
            <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold text-fg">
                  {t("ref.zonesSection")}
                </h2>
                <select
                  aria-label={t("ref.zoneTypeLabel")}
                  value={zoneTypeFilter}
                  onChange={(e) =>
                    setZoneTypeFilter(
                      e.target.value as "" | "wilayah" | "stasi",
                    )
                  }
                  className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs text-fg"
                >
                  <option value="">{t("ref.allTypes")}</option>
                  <option value="wilayah">{t("ref.typeZone")}</option>
                  <option value="stasi">{t("ref.typeStation")}</option>
                </select>
              </div>

              <form
                onSubmit={handleSubmitZone}
                className="grid gap-2.5 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-2"
              >
                <select
                  value={zType}
                  onChange={(e) =>
                    setZType(e.target.value as "wilayah" | "stasi")
                  }
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                >
                  <option value="wilayah">{t("ref.typeZone")}</option>
                  <option value="stasi">{t("ref.typeStation")}</option>
                </select>
                <input
                  required
                  value={zCode}
                  onChange={(e) => setZCode(e.target.value)}
                  placeholder={t("ref.zoneCodePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  required
                  value={zName}
                  onChange={(e) => setZName(e.target.value)}
                  placeholder={t("ref.zoneNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg sm:col-span-2"
                />
                <input
                  value={zCoordName}
                  onChange={(e) => setZCoordName(e.target.value)}
                  placeholder={t("ref.coordinatorNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={zCoordPhone}
                  onChange={(e) => setZCoordPhone(e.target.value)}
                  placeholder={t("ref.phoneLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <div className="flex justify-end gap-2 sm:col-span-2">
                  {editingZone && (
                    <button
                      type="button"
                      onClick={resetZoneForm}
                      className="h-9 rounded-lg border border-line px-3 text-xs font-medium text-fg"
                    >
                      {t("users.cancel")}
                    </button>
                  )}
                  <button
                    type="submit"
                    disabled={createZone.isPending || updateZone.isPending}
                    className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                  >
                    <Plus aria-hidden className="size-4" />
                    {editingZone ? t("ref.updateItem") : t("ref.addZone")}
                  </button>
                </div>
              </form>

              <div className="divide-y divide-line">
                {zones.map((z) => (
                  <div
                    key={z.id}
                    className={`flex items-center justify-between gap-2 py-2.5 ${
                      selectedZoneId === z.id
                        ? "bg-primary/5 px-2 rounded-lg"
                        : ""
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() =>
                        setSelectedZoneId((prev) =>
                          prev === z.id ? "" : (z.id ?? ""),
                        )
                      }
                      className="flex-1 text-left"
                    >
                      <p className="text-sm font-semibold text-fg">
                        {z.nama}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({z.kode} ·{" "}
                          {z.jenis === "stasi"
                            ? t("ref.typeStation")
                            : t("ref.typeZone")}
                          )
                        </span>
                      </p>
                      {(z.koordinator || z.teleponKoordinator) && (
                        <p className="text-xs text-muted">
                          {z.koordinator}
                          {z.teleponKoordinator
                            ? ` · ${z.teleponKoordinator}`
                            : ""}
                        </p>
                      )}
                    </button>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        title={z.active ? t("ref.active") : t("ref.inactive")}
                        onClick={async () => {
                          if (!z.id) return;
                          setErrorMsg(null);
                          try {
                            await updateZone.mutateAsync({
                              id: z.id,
                              body: { active: !z.active },
                            });
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal mengubah status.",
                            );
                          }
                        }}
                        className={`inline-flex size-8 items-center justify-center rounded-lg hover:bg-fg/8 ${
                          z.active ? "text-primary" : "text-muted"
                        }`}
                      >
                        <Power aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("ref.editItem")} ${z.nama}`}
                        onClick={() => startEditZone(z)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${z.nama}`}
                        onClick={async () => {
                          if (!z.id) return;
                          setErrorMsg(null);
                          try {
                            await deleteZone.mutateAsync(z.id);
                            if (selectedZoneId === z.id) setSelectedZoneId("");
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal menghapus wilayah.",
                            );
                          }
                        }}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                      >
                        <Trash2 aria-hidden className="size-4" />
                      </button>
                    </div>
                  </div>
                ))}
                {zones.length === 0 && (
                  <p className="py-3 text-sm text-muted">
                    {t("ref.emptyList")}
                  </p>
                )}
              </div>
            </section>

            {/* Wards */}
            <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold text-fg">
                  {t("ref.wardsSection")}
                </h2>
                <select
                  aria-label={t("ref.filterByZone")}
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                  className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs text-fg"
                >
                  <option value="">{t("ref.allZones")}</option>
                  {zones.map((z) => (
                    <option key={z.id} value={z.id}>
                      {z.nama}
                    </option>
                  ))}
                </select>
              </div>

              <form
                onSubmit={handleSubmitWard}
                className="grid gap-2.5 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-2"
              >
                <select
                  value={wZoneId || selectedZoneId || zones[0]?.id || ""}
                  onChange={(e) => setWZoneId(e.target.value)}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                >
                  <option value="">{t("ref.selectZone")}</option>
                  {zones.map((z) => (
                    <option key={z.id} value={z.id}>
                      {z.nama}
                    </option>
                  ))}
                </select>
                <input
                  required
                  value={wCode}
                  onChange={(e) => setWCode(e.target.value)}
                  placeholder={t("ref.wardCodePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  required
                  value={wName}
                  onChange={(e) => setWName(e.target.value)}
                  placeholder={t("ref.wardNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={wPatron}
                  onChange={(e) => setWPatron(e.target.value)}
                  placeholder={t("ref.patronSaintLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={wHeadName}
                  onChange={(e) => setWHeadName(e.target.value)}
                  placeholder={t("ref.headNamePlaceholder")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <input
                  value={wHeadPhone}
                  onChange={(e) => setWHeadPhone(e.target.value)}
                  placeholder={t("ref.phoneLabel")}
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                />
                <div className="flex justify-end gap-2 sm:col-span-2">
                  {editingWard && (
                    <button
                      type="button"
                      onClick={resetWardForm}
                      className="h-9 rounded-lg border border-line px-3 text-xs font-medium text-fg"
                    >
                      {t("users.cancel")}
                    </button>
                  )}
                  <button
                    type="submit"
                    disabled={createWard.isPending || updateWard.isPending}
                    className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                  >
                    <Plus aria-hidden className="size-4" />
                    {editingWard ? t("ref.updateItem") : t("ref.addWard")}
                  </button>
                </div>
              </form>

              <div className="divide-y divide-line">
                {wards.map((w) => (
                  <div
                    key={w.id}
                    className="flex items-center justify-between gap-2 py-2.5"
                  >
                    <div>
                      <p className="text-sm font-semibold text-fg">
                        {w.nama}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({w.kode}
                          {w.wilayahId && zoneNameById.get(w.wilayahId)
                            ? ` · ${zoneNameById.get(w.wilayahId)}`
                            : ""}
                          )
                        </span>
                      </p>
                      {(w.ketua || w.teleponKetua) && (
                        <p className="text-xs text-muted">
                          {w.ketua}
                          {w.teleponKetua ? ` · ${w.teleponKetua}` : ""}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        title={w.active ? t("ref.active") : t("ref.inactive")}
                        onClick={async () => {
                          if (!w.id) return;
                          setErrorMsg(null);
                          try {
                            await updateWard.mutateAsync({
                              id: w.id,
                              body: { active: !w.active },
                            });
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal mengubah status.",
                            );
                          }
                        }}
                        className={`inline-flex size-8 items-center justify-center rounded-lg hover:bg-fg/8 ${
                          w.active ? "text-primary" : "text-muted"
                        }`}
                      >
                        <Power aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("ref.editItem")} ${w.nama}`}
                        onClick={() => startEditWard(w)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${w.nama}`}
                        onClick={async () => {
                          if (!w.id) return;
                          setErrorMsg(null);
                          try {
                            await deleteWard.mutateAsync(w.id);
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal menghapus lingkungan.",
                            );
                          }
                        }}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                      >
                        <Trash2 aria-hidden className="size-4" />
                      </button>
                    </div>
                  </div>
                ))}
                {wards.length === 0 && (
                  <p className="py-3 text-sm text-muted">
                    {t("ref.emptyList")}
                  </p>
                )}
              </div>
            </section>
          </div>
        </>
      )}
    </div>
  );
}
