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
import { useLookupsSearch } from "../api/queries";
import {
  getLookupCategoryLabel,
  hasAnyRefAdminRole,
  isSuperAdmin,
  LOOKUP_CATEGORY_META,
} from "../model";

export function meta() {
  return [{ title: "Kamus Referensi — Domus" }];
}

export default function LookupsPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = hasAnyRefAdminRole(me?.roles);
  const superAdmin = isSuperAdmin(me?.roles);

  const [searchParams, setSearchParams] = useSearchParams();
  const category = searchParams.get("category") ?? "";
  const [q, setQ] = useState("");
  const [scope, setScope] = useState<"all" | "global" | "paroki">("all");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const lookupsQuery = useLookupsSearch(
    {
      q: debouncedQ || undefined,
      filter: category
        ? {
            category,
            scope,
            active: activeFilter === "" ? undefined : activeFilter === "true",
          }
        : undefined,
    },
    canManage && Boolean(category),
  );
  const items = lookupsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          ...(category
            ? [
                { label: "Kamus Referensi", to: "/settings/lookups" },
                { label: getLookupCategoryLabel(category) },
              ]
            : [{ label: "Kamus Referensi" }]),
        ]}
        title={
          category
            ? `Kamus: ${getLookupCategoryLabel(category)}`
            : "Kamus Referensi"
        }
        subtitle={
          category
            ? "Kelola daftar item referensi global maupun khusus paroki aktif."
            : "Pilih kategori kamus referensi untuk melihat atau menambahkan item."
        }
        action={
          canManage && category ? (
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
              <Link
                to={`/settings/lookups/new?category=${encodeURIComponent(category)}`}
                className={buttonClass("primary")}
              >
                <Plus aria-hidden className="size-4" />
                Tambah Item
              </Link>
            </div>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki atau Super Admin yang dapat mengelola kamus
          referensi.
        </div>
      ) : null}

      {canManage && !category ? (
        <InductiveHubGrid>
          {LOOKUP_CATEGORY_META.map((cat, idx) => (
            <InductiveStepCard
              key={cat.slug}
              to={`/settings/lookups?category=${encodeURIComponent(cat.slug)}`}
              title={cat.label}
              description={cat.description}
              icon={BookOpen}
              plate={((idx % 6) + 1) as 1 | 2 | 3 | 4 | 5 | 6}
            />
          ))}
        </InductiveHubGrid>
      ) : null}

      {canManage && category ? (
        <>
          <InductiveFilterBar>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5 sm:min-w-56">
              <label
                htmlFor="lookup-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Item
              </label>
              <input
                id="lookup-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau label…"
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              />
            </div>

            <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
              <label
                htmlFor="lookup-scope-filter"
                className="text-xs font-semibold text-muted"
              >
                Cakupan
              </label>
              <select
                id="lookup-scope-filter"
                value={scope}
                onChange={(e) =>
                  setScope(e.target.value as "all" | "global" | "paroki")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              >
                <option value="all">Semua cakupan</option>
                <option value="global">Global</option>
                <option value="paroki">Paroki</option>
              </select>
            </div>

            <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
              <label
                htmlFor="lookup-status-filter"
                className="text-xs font-semibold text-muted"
              >
                Status
              </label>
              <select
                id="lookup-status-filter"
                value={activeFilter}
                onChange={(e) =>
                  setActiveFilter(e.target.value as "" | "true" | "false")
                }
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              >
                <option value="">Semua</option>
                <option value="true">Aktif</option>
                <option value="false">Nonaktif</option>
              </select>
            </div>
          </InductiveFilterBar>

          {lookupsQuery.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={BookOpen}
              title="Belum ada item referensi"
              description="Tambahkan item baru pada kategori ini."
              action={
                <Link
                  to={`/settings/lookups/new?category=${encodeURIComponent(category)}`}
                  className={buttonClass("primary")}
                >
                  Tambah Item
                </Link>
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
                  {item.parokiId || superAdmin ? (
                    <Link
                      to={`/settings/lookups/${item.id}`}
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
      ) : null}
    </div>
  );
}
