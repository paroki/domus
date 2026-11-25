import { ArrowLeft, CheckCircle2, Pencil, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useCreatePeriod,
  useDeletePeriod,
  useUpdatePeriod,
} from "../api/actions";
import { useCurrentIdentity, usePeriodsSearch } from "../api/queries";
import type { Period } from "../api/types";
import { isParokiAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Periode — Domus" }];
}

export default function PeriodsPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const allowed = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);
  const enabled = allowed && hasTenant;

  const [typeFilter, setTypeFilter] = useState<"" | "fiscal" | "liturgical">(
    "",
  );
  const [statusFilter, setStatusFilter] = useState<"" | "open" | "closed">("");
  const [editing, setEditing] = useState<Period | null>(null);
  const [pType, setPType] = useState<"fiscal" | "liturgical">("fiscal");
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [isCurrent, setIsCurrent] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const periodsQuery = usePeriodsSearch(
    {
      filter: {
        type: typeFilter || undefined,
        status: statusFilter || undefined,
      },
    },
    enabled,
  );
  const periods = periodsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createPeriod = useCreatePeriod();
  const updatePeriod = useUpdatePeriod();
  const deletePeriod = useDeletePeriod();

  function resetForm() {
    setEditing(null);
    setPType("fiscal");
    setCode("");
    setName("");
    setStartDate("");
    setEndDate("");
    setIsCurrent(false);
  }

  function startEdit(p: Period) {
    setEditing(p);
    setPType((p.type as "fiscal" | "liturgical") || "fiscal");
    setCode(p.code ?? "");
    setName(p.name ?? "");
    setStartDate(p.startDate ?? "");
    setEndDate(p.endDate ?? "");
    setIsCurrent(Boolean(p.isCurrent));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const normCode = slugifyCode(code || name);
      if (editing?.id) {
        await updatePeriod.mutateAsync({
          id: editing.id,
          body: {
            code: normCode,
            name: name.trim(),
            startDate,
            endDate,
            isCurrent,
          },
        });
      } else {
        await createPeriod.mutateAsync({
          type: pType,
          code: normCode,
          name: name.trim(),
          startDate,
          endDate,
          isCurrent,
        });
      }
      resetForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan periode.",
      );
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/ref"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("ref.backToSettings")}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">
          {t("ref.periodsTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.periodsSubtitle")}</p>
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

          <form
            onSubmit={handleSubmit}
            className="flex flex-col gap-3 rounded-2xl border border-line bg-surface p-6"
          >
            <div className="grid gap-3 sm:grid-cols-5">
              <select
                disabled={Boolean(editing)}
                value={pType}
                onChange={(e) =>
                  setPType(e.target.value as "fiscal" | "liturgical")
                }
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg disabled:opacity-60"
              >
                <option value="fiscal">{t("ref.periodFiscal")}</option>
                <option value="liturgical">{t("ref.periodLiturgical")}</option>
              </select>
              <input
                required
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder={t("ref.periodNamePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                type="date"
                aria-label={t("ref.startDateLabel")}
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                type="date"
                aria-label={t("ref.endDateLabel")}
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              <label className="inline-flex items-center gap-2 text-sm text-fg">
                <input
                  type="checkbox"
                  checked={isCurrent}
                  onChange={(e) => setIsCurrent(e.target.checked)}
                  className="size-4 rounded border-line"
                />
                {t("ref.setAsCurrent")}
              </label>

              <div className="flex gap-2">
                {editing && (
                  <button
                    type="button"
                    onClick={resetForm}
                    className="h-9 rounded-lg border border-line px-3 text-xs font-medium text-fg"
                  >
                    {t("users.cancel")}
                  </button>
                )}
                <button
                  type="submit"
                  disabled={createPeriod.isPending || updatePeriod.isPending}
                  className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                >
                  <Plus aria-hidden className="size-4" />
                  {editing ? t("ref.updateItem") : t("ref.addPeriod")}
                </button>
              </div>
            </div>
          </form>

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <div className="flex flex-wrap gap-2">
              <select
                aria-label={t("ref.periodTypeLabel")}
                value={typeFilter}
                onChange={(e) =>
                  setTypeFilter(e.target.value as "" | "fiscal" | "liturgical")
                }
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                <option value="">{t("ref.allTypes")}</option>
                <option value="fiscal">{t("ref.periodFiscal")}</option>
                <option value="liturgical">{t("ref.periodLiturgical")}</option>
              </select>
              <select
                aria-label={t("ref.periodStatusLabel")}
                value={statusFilter}
                onChange={(e) =>
                  setStatusFilter(e.target.value as "" | "open" | "closed")
                }
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                <option value="">{t("users.statusAll")}</option>
                <option value="open">{t("ref.statusOpen")}</option>
                <option value="closed">{t("ref.statusClosed")}</option>
              </select>
            </div>

            <div className="divide-y divide-line">
              {periods.map((p) => (
                <div
                  key={p.id}
                  className="flex flex-wrap items-center justify-between gap-3 py-3"
                >
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="text-sm font-semibold text-fg">
                        {p.name}{" "}
                        <span className="text-xs font-normal text-muted">
                          ({p.code})
                        </span>
                      </p>
                      <span className="rounded-full bg-fg/10 px-2 py-0.5 text-[11px] font-medium text-fg">
                        {p.type === "liturgical"
                          ? t("ref.periodLiturgical")
                          : t("ref.periodFiscal")}
                      </span>
                      <span
                        className={`rounded-full px-2 py-0.5 text-[11px] font-medium ${
                          p.status === "open"
                            ? "bg-primary/15 text-primary"
                            : "bg-fg/10 text-muted"
                        }`}
                      >
                        {p.status === "open"
                          ? t("ref.statusOpen")
                          : t("ref.statusClosed")}
                      </span>
                      {p.isCurrent && (
                        <span className="inline-flex items-center gap-1 rounded-full bg-primary px-2 py-0.5 text-[11px] font-semibold text-on-primary">
                          <CheckCircle2 aria-hidden className="size-3" />
                          {t("ref.currentPeriodBadge")}
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-muted">
                      {p.startDate} – {p.endDate}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-1.5">
                    {!p.isCurrent && (
                      <button
                        type="button"
                        onClick={async () => {
                          if (!p.id) return;
                          setErrorMsg(null);
                          try {
                            await updatePeriod.mutateAsync({
                              id: p.id,
                              body: { isCurrent: true },
                            });
                          } catch (err) {
                            setErrorMsg(
                              err instanceof Error
                                ? err.message
                                : "Gagal menetapkan periode aktif.",
                            );
                          }
                        }}
                        className="h-8 rounded-lg border border-line px-2.5 text-xs font-medium text-fg hover:bg-fg/8"
                      >
                        {t("ref.makeCurrent")}
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={async () => {
                        if (!p.id) return;
                        setErrorMsg(null);
                        try {
                          await updatePeriod.mutateAsync({
                            id: p.id,
                            body: {
                              status: p.status === "open" ? "closed" : "open",
                            },
                          });
                        } catch (err) {
                          setErrorMsg(
                            err instanceof Error
                              ? err.message
                              : "Gagal mengubah status.",
                          );
                        }
                      }}
                      className="h-8 rounded-lg border border-line px-2.5 text-xs font-medium text-fg hover:bg-fg/8"
                    >
                      {p.status === "open"
                        ? t("ref.closePeriod")
                        : t("ref.reopenPeriod")}
                    </button>
                    <button
                      type="button"
                      aria-label={`${t("ref.editItem")} ${p.name}`}
                      onClick={() => startEdit(p)}
                      className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                    >
                      <Pencil aria-hidden className="size-4" />
                    </button>
                    <button
                      type="button"
                      aria-label={`${t("users.deleteUser")} ${p.name}`}
                      onClick={async () => {
                        if (!p.id) return;
                        setErrorMsg(null);
                        try {
                          await deletePeriod.mutateAsync(p.id);
                        } catch (err) {
                          setErrorMsg(
                            err instanceof Error
                              ? err.message
                              : "Gagal menghapus periode.",
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
              {periods.length === 0 && (
                <p className="py-4 text-center text-sm text-muted">
                  {t("ref.emptyList")}
                </p>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  );
}
