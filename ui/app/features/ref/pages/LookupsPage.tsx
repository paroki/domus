import { ArrowLeft, Pencil, Plus, Power, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useCreateLookup,
  useDeleteLookup,
  useUpdateLookup,
} from "../api/actions";
import { useCurrentIdentity, useLookupsSearch } from "../api/queries";
import type { Lookup } from "../api/types";
import {
  isParokiAdmin,
  isSuperAdmin,
  LOOKUP_CATEGORIES,
  slugifyCode,
} from "../model";

export function meta() {
  return [{ title: "Kamus Referensi — Domus" }];
}

export default function LookupsPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const superAdmin = isSuperAdmin(me?.roles);
  const parokiAdmin = isParokiAdmin(me?.roles);
  const allowed = superAdmin || parokiAdmin;
  const bothRoles = superAdmin && parokiAdmin;

  const [categoryFilter, setCategoryFilter] = useState("");
  const [scopeFilter, setScopeFilter] = useState<"all" | "global" | "paroki">(
    "all",
  );
  const [editing, setEditing] = useState<Lookup | null>(null);
  const [category, setCategory] = useState<string>(LOOKUP_CATEGORIES[0]);
  const [code, setCode] = useState("");
  const [label, setLabel] = useState("");
  const [sortOrder, setSortOrder] = useState("0");
  const [asGlobal, setAsGlobal] = useState(superAdmin && !parokiAdmin);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const lookupsQuery = useLookupsSearch(
    {
      filter: {
        category: categoryFilter.trim() || undefined,
        scope: scopeFilter,
      },
    },
    allowed,
  );
  const lookups = lookupsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createLookup = useCreateLookup();
  const updateLookup = useUpdateLookup();
  const deleteLookup = useDeleteLookup();

  function resetForm() {
    setEditing(null);
    setCode("");
    setLabel("");
    setSortOrder("0");
    setAsGlobal(superAdmin && !parokiAdmin);
  }

  function startEdit(item: Lookup) {
    setEditing(item);
    setCategory(item.category ?? LOOKUP_CATEGORIES[0]);
    setCode(item.code ?? "");
    setLabel(item.label ?? "");
    setSortOrder(String(item.sortOrder ?? 0));
  }

  function canModify(item: Lookup): boolean {
    if (!item.parokiId) return superAdmin;
    return parokiAdmin && Boolean(me?.activeParokiId);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    const normCategory = slugifyCode(category);
    const normCode = slugifyCode(code || label);
    const orderNum = Number.parseInt(sortOrder, 10) || 0;
    try {
      if (editing?.id) {
        await updateLookup.mutateAsync({
          id: editing.id,
          body: {
            category: normCategory,
            code: normCode,
            label: label.trim(),
            sortOrder: orderNum,
          },
        });
      } else {
        await createLookup.mutateAsync({
          category: normCategory,
          code: normCode,
          label: label.trim(),
          sortOrder: orderNum,
          global: bothRoles ? asGlobal : superAdmin,
        });
      }
      resetForm();
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal menyimpan item referensi.",
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
          {t("ref.lookupsTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.lookupsSubtitle")}</p>
      </div>

      {!allowed ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenMasterData")}
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
            <div className="grid gap-3 sm:grid-cols-4">
              <input
                required
                list="lookup-categories"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                placeholder={t("ref.categoryPlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <datalist id="lookup-categories">
                {LOOKUP_CATEGORIES.map((c) => (
                  <option key={c} value={c} />
                ))}
              </datalist>
              <input
                required
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={label}
                onChange={(e) => setLabel(e.target.value)}
                placeholder={t("ref.labelPlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                type="number"
                value={sortOrder}
                onChange={(e) => setSortOrder(e.target.value)}
                placeholder={t("ref.sortOrderPlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3">
              {!editing && bothRoles ? (
                <label className="inline-flex items-center gap-2 text-sm text-fg">
                  <input
                    type="checkbox"
                    checked={asGlobal}
                    onChange={(e) => setAsGlobal(e.target.checked)}
                    className="size-4 rounded border-line"
                  />
                  {t("ref.saveAsGlobal")}
                </label>
              ) : (
                <span className="text-xs text-muted">
                  {editing
                    ? editing.parokiId
                      ? t("ref.scopeParish")
                      : t("ref.scopeGlobal")
                    : superAdmin
                      ? t("ref.scopeGlobal")
                      : t("ref.scopeParish")}
                </span>
              )}

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
                  disabled={createLookup.isPending || updateLookup.isPending}
                  className="inline-flex h-9 items-center gap-1.5 rounded-lg bg-primary px-4 text-sm font-semibold text-on-primary disabled:opacity-50"
                >
                  <Plus aria-hidden className="size-4" />
                  {editing ? t("ref.updateItem") : t("ref.addLookup")}
                </button>
              </div>
            </div>
          </form>

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex flex-wrap gap-2">
                <select
                  aria-label={t("ref.categoryLabel")}
                  value={categoryFilter}
                  onChange={(e) => setCategoryFilter(e.target.value)}
                  className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="">{t("ref.allCategories")}</option>
                  {LOOKUP_CATEGORIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
                <select
                  aria-label={t("ref.scopeLabel")}
                  value={scopeFilter}
                  onChange={(e) =>
                    setScopeFilter(
                      e.target.value as "all" | "global" | "paroki",
                    )
                  }
                  className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="all">{t("ref.scopeAll")}</option>
                  <option value="global">{t("ref.scopeGlobal")}</option>
                  <option value="paroki">{t("ref.scopeParish")}</option>
                </select>
              </div>
            </div>

            <div className="divide-y divide-line">
              {lookups.map((item) => {
                const editable = canModify(item);
                return (
                  <div
                    key={item.id}
                    className="flex flex-wrap items-center justify-between gap-3 py-3"
                  >
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-semibold text-fg">
                          {item.label}
                        </p>
                        <span
                          className={`rounded-full px-2 py-0.5 text-[11px] font-medium ${
                            item.parokiId
                              ? "bg-primary/15 text-primary"
                              : "bg-fg/10 text-muted"
                          }`}
                        >
                          {item.parokiId
                            ? t("ref.scopeParish")
                            : t("ref.scopeGlobal")}
                        </span>
                      </div>
                      <p className="text-xs text-muted">
                        {item.category} · {item.code} · #{item.sortOrder ?? 0}
                      </p>
                    </div>
                    {editable && (
                      <div className="flex items-center gap-1">
                        <button
                          type="button"
                          title={
                            item.active ? t("ref.active") : t("ref.inactive")
                          }
                          onClick={async () => {
                            if (!item.id) return;
                            setErrorMsg(null);
                            try {
                              await updateLookup.mutateAsync({
                                id: item.id,
                                body: { active: !item.active },
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
                            item.active ? "text-primary" : "text-muted"
                          }`}
                        >
                          <Power aria-hidden className="size-4" />
                        </button>
                        <button
                          type="button"
                          aria-label={`${t("ref.editItem")} ${item.label}`}
                          onClick={() => startEdit(item)}
                          className="inline-flex size-8 items-center justify-center rounded-lg text-muted hover:bg-fg/8 hover:text-fg"
                        >
                          <Pencil aria-hidden className="size-4" />
                        </button>
                        <button
                          type="button"
                          aria-label={`${t("users.deleteUser")} ${item.label}`}
                          onClick={async () => {
                            if (!item.id) return;
                            setErrorMsg(null);
                            try {
                              await deleteLookup.mutateAsync(item.id);
                            } catch (err) {
                              setErrorMsg(
                                err instanceof Error
                                  ? err.message
                                  : "Gagal menghapus item.",
                              );
                            }
                          }}
                          className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                        >
                          <Trash2 aria-hidden className="size-4" />
                        </button>
                      </div>
                    )}
                  </div>
                );
              })}
              {lookups.length === 0 && (
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
