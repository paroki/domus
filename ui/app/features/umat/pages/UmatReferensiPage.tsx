import { BookOpen, Plus } from "lucide-react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { useDebouncedValue } from "~/shared/util/use-debounced-value";
import { useUmatReferensiSearch } from "../api/queries";
import {
  canEditUmatReferensi,
  getUmatReferensiCategoryLabel,
  isSuperAdmin,
  isUmatAdmin,
  UMAT_REFERENSI_CATEGORIES,
} from "../model";

export function meta() {
  return [{ title: "Kamus Referensi Umat — Domus" }];
}

export default function UmatReferensiPage() {
  const { data: me } = useMe();
  const parokiId = me?.activeParokiId ?? "";
  const canCreate = isUmatAdmin(me) || isSuperAdmin(me);

  const [searchParams, setSearchParams] = useSearchParams();
  const category = searchParams.get("category") ?? "";
  const [q, setQ] = useState("");
  const [scope, setScope] = useState<"all" | "global" | "paroki">("all");
  const [status, setStatus] = useState<"all" | "active" | "inactive">("all");
  const debouncedQ = useDebouncedValue(q);

  const activeFilter =
    status === "active" ? true : status === "inactive" ? false : undefined;

  const referensiQuery = useUmatReferensiSearch(
    parokiId,
    {
      q: debouncedQ || undefined,
      filter: category
        ? {
            category,
            scope,
            ...(activeFilter !== undefined ? { active: activeFilter } : {}),
          }
        : undefined,
    },
    Boolean(me) && Boolean(category),
  );
  const items = referensiQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          ...(category
            ? [
                { label: "Kamus Referensi Umat", to: "/umat/referensi" },
                { label: getUmatReferensiCategoryLabel(category) },
              ]
            : [{ label: "Kamus Referensi Umat" }]),
        ]}
        title={
          category
            ? `Kamus: ${getUmatReferensiCategoryLabel(category)}`
            : "Kamus Referensi Umat"
        }
        subtitle={
          category
            ? "Daftar nilai referensi Data Umat global maupun khusus paroki aktif."
            : "Pilih kategori kamus referensi Data Umat untuk melihat atau mengelola item."
        }
        action={
          category ? (
            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={() => {
                  const next = new URLSearchParams(searchParams);
                  next.delete("category");
                  setSearchParams(next);
                }}
                className={buttonClass("outline")}
              >
                Ganti Kategori
              </button>
              {canCreate ? (
                <Link
                  to={`/umat/referensi/new?category=${encodeURIComponent(category)}`}
                  className={buttonClass("primary")}
                >
                  <Plus aria-hidden className="size-4" />
                  Tambah Item
                </Link>
              ) : null}
            </div>
          ) : undefined
        }
      />

      {!category ? (
        <InductiveHubGrid>
          {UMAT_REFERENSI_CATEGORIES.map((cat, idx) => (
            <InductiveStepCard
              key={cat.slug}
              to={`/umat/referensi?category=${encodeURIComponent(cat.slug)}`}
              title={cat.label}
              description={cat.description}
              icon={BookOpen}
              plate={((idx % 6) + 1) as 1 | 2 | 3 | 4 | 5 | 6}
            />
          ))}
        </InductiveHubGrid>
      ) : (
        <>
          <InductiveFilterBar>
            <div className="flex min-w-56 flex-1 flex-col gap-1.5">
              <label
                htmlFor="referensi-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Item
              </label>
              <input
                id="referensi-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau label…"
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              />
            </div>

            <div className="flex min-w-40 flex-col gap-1.5">
              <label
                htmlFor="referensi-scope-filter"
                className="text-xs font-semibold text-muted"
              >
                Cakupan
              </label>
              <select
                id="referensi-scope-filter"
                value={scope}
                onChange={(e) =>
                  setScope(e.target.value as "all" | "global" | "paroki")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              >
                <option value="all">Semua</option>
                <option value="global">Global</option>
                <option value="paroki">Paroki</option>
              </select>
            </div>

            <div className="flex min-w-40 flex-col gap-1.5">
              <label
                htmlFor="referensi-status-filter"
                className="text-xs font-semibold text-muted"
              >
                Status
              </label>
              <select
                id="referensi-status-filter"
                value={status}
                onChange={(e) =>
                  setStatus(e.target.value as "all" | "active" | "inactive")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              >
                <option value="all">Semua</option>
                <option value="active">Aktif</option>
                <option value="inactive">Nonaktif</option>
              </select>
            </div>
          </InductiveFilterBar>

          {referensiQuery.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={BookOpen}
              title="Belum ada item referensi"
              description="Belum ada item referensi yang sesuai pada kategori ini."
              action={
                canCreate ? (
                  <Link
                    to={`/umat/referensi/new?category=${encodeURIComponent(category)}`}
                    className={buttonClass("primary")}
                  >
                    Tambah Item
                  </Link>
                ) : undefined
              }
            />
          ) : (
            <div className="flex flex-col gap-3">
              {items.map((item) => (
                <Card
                  key={item.id}
                  className="flex flex-col justify-between gap-4 p-4 sm:flex-row sm:items-center sm:p-6"
                >
                  <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-base font-semibold text-fg">
                        {item.label}
                      </span>
                      <Badge tone={item.parokiId ? "neutral" : "warning"}>
                        {item.parokiId ? "Paroki" : "Global"}
                      </Badge>
                      <Badge tone={item.active ? "success" : "neutral"}>
                        {item.active ? "Aktif" : "Nonaktif"}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted">
                      Kode: <span className="font-mono">{item.code}</span> ·
                      Urutan: {item.sortOrder}
                    </p>
                  </div>
                  {canEditUmatReferensi(me, item.parokiId) ? (
                    <Link
                      to={`/umat/referensi/${item.id}`}
                      className={buttonClass("outline")}
                    >
                      Ubah
                    </Link>
                  ) : null}
                </Card>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
