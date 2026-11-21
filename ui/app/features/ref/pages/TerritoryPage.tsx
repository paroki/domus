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
  const hasTenant = Boolean(me?.activeParishId);
  const enabled = allowed && hasTenant;

  const [zoneTypeFilter, setZoneTypeFilter] = useState<"" | "zone" | "station">(
    "",
  );
  const [selectedZoneId, setSelectedZoneId] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Zone form state
  const [editingZone, setEditingZone] = useState<Zone | null>(null);
  const [zType, setZType] = useState<"zone" | "station">("zone");
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
    { filter: zoneTypeFilter ? { type: zoneTypeFilter } : undefined },
    enabled,
  );
  const wardsQuery = useWardsSearch(
    { filter: selectedZoneId ? { zoneId: selectedZoneId } : undefined },
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
    setZType("zone");
    setZCode("");
    setZName("");
    setZCoordName("");
    setZCoordPhone("");
  }

  function startEditZone(z: Zone) {
    setEditingZone(z);
    setZType((z.type as "zone" | "station") || "zone");
    setZCode(z.code ?? "");
    setZName(z.name ?? "");
    setZCoordName(z.coordinatorName ?? "");
    setZCoordPhone(z.coordinatorPhone ?? "");
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
    setWZoneId(w.zoneId ?? "");
    setWCode(w.code ?? "");
    setWName(w.name ?? "");
    setWPatron(w.patronSaint ?? "");
    setWHeadName(w.headName ?? "");
    setWHeadPhone(w.headPhone ?? "");
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
            type: zType,
            code,
            name: zName.trim(),
            coordinatorName: zCoordName.trim(),
            coordinatorPhone: zCoordPhone.trim(),
          },
        });
      } else {
        const created = await createZone.mutateAsync({
          type: zType,
          code,
          name: zName.trim(),
          coordinatorName: zCoordName.trim() || undefined,
          coordinatorPhone: zCoordPhone.trim() || undefined,
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
            zoneId: targetZoneId,
            code,
            name: wName.trim(),
            patronSaint: wPatron.trim(),
            headName: wHeadName.trim(),
            headPhone: wHeadPhone.trim(),
          },
        });
      } else {
        await createWard.mutateAsync({
          zoneId: targetZoneId,
          code,
          name: wName.trim(),
          patronSaint: wPatron.trim() || undefined,
          headName: wHeadName.trim() || undefined,
          headPhone: wHeadPhone.trim() || undefined,
        });
      }
      resetWardForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan lingkungan.",
      );
    }
  }

  const zoneNameById = new Map(zones.map((z) => [z.id ?? "", z.name ?? ""]));

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
                    setZoneTypeFilter(e.target.value as "" | "zone" | "station")
                  }
                  className="h-8 rounded-lg border border-line bg-canvas px-2 text-xs text-fg"
                >
                  <option value="">{t("ref.allTypes")}</option>
                  <option value="zone">{t("ref.typeZone")}</option>
                  <option value="station">{t("ref.typeStation")}</option>
                </select>
              </div>

              <form
                onSubmit={handleSubmitZone}
                className="grid gap-2.5 rounded-xl border border-line bg-canvas p-4 sm:grid-cols-2"
              >
                <select
                  value={zType}
                  onChange={(e) =>
                    setZType(e.target.value as "zone" | "station")
                  }
                  className="h-9 rounded-lg border border-line bg-surface px-3 text-sm text-fg"
                >
                  <option value="zone">{t("ref.typeZone")}</option>
                  <option value="station">{t("ref.typeStation")}</option>
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
                        {z.name}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({z.code} ·{" "}
                          {z.type === "station"
                            ? t("ref.typeStation")
                            : t("ref.typeZone")}
                          )
                        </span>
                      </p>
                      {(z.coordinatorName || z.coordinatorPhone) && (
                        <p className="text-xs text-muted">
                          {z.coordinatorName}
                          {z.coordinatorPhone ? ` · ${z.coordinatorPhone}` : ""}
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
                        aria-label={`${t("ref.editItem")} ${z.name}`}
                        onClick={() => startEditZone(z)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${z.name}`}
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
                      {z.name}
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
                      {z.name}
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
                        {w.name}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({w.code}
                          {w.zoneId && zoneNameById.get(w.zoneId)
                            ? ` · ${zoneNameById.get(w.zoneId)}`
                            : ""}
                          )
                        </span>
                      </p>
                      {(w.headName || w.headPhone) && (
                        <p className="text-xs text-muted">
                          {w.headName}
                          {w.headPhone ? ` · ${w.headPhone}` : ""}
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
                        aria-label={`${t("ref.editItem")} ${w.name}`}
                        onClick={() => startEditWard(w)}
                        className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                      >
                        <Pencil aria-hidden className="size-4" />
                      </button>
                      <button
                        type="button"
                        aria-label={`${t("users.deleteUser")} ${w.name}`}
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
