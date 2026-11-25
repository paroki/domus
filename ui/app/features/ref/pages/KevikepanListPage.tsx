import { Network, Plus } from "lucide-react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useDebouncedValue } from "~/shared/util/use-debounced-value";
import { useKeuskupanSearch, useKevikepanSearch } from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Kevikepan — Domus" }];
}

export default function KevikepanListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [searchParams, setSearchParams] = useSearchParams();
  const keuskupanId = searchParams.get("keuskupanId") ?? "";
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const keuskupanQuery = useKeuskupanSearch({}, canManage);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const keuskupanMap = new Map(keuskupanList.map((k) => [k.id, k.nama]));

  const kevikepanQuery = useKevikepanSearch(
    {
      q: debouncedQ || undefined,
      filter:
        keuskupanId || activeFilter
          ? {
              keuskupanId: keuskupanId || undefined,
              active: activeFilter === "" ? undefined : activeFilter === "true",
            }
          : undefined,
    },
    canManage && keuskupanList.length > 0,
  );
  const items = kevikepanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const newHref = keuskupanId
    ? `/settings/gerejawi/kevikepan/new?keuskupanId=${keuskupanId}`
    : "/settings/gerejawi/kevikepan/new";

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Kevikepan" },
        ]}
        title="Kevikepan"
        subtitle="Daftar kevikepan di bawah masing-masing keuskupan."
        action={
          canManage && keuskupanList.length > 0 ? (
            <Link to={newHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              Tambah Kevikepan
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola kevikepan.
        </div>
      ) : null}

      {canManage ? (
        !keuskupanQuery.isLoading && keuskupanList.length === 0 ? (
          <InductiveEmptyState
            icon={Network}
            title="Buat keuskupan terlebih dahulu"
            description="Kevikepan harus terhubung ke keuskupan induk. Buat keuskupan pertama terlebih dahulu."
            action={
              <Link
                to="/settings/gerejawi/keuskupan/new"
                className={buttonClass("primary")}
              >
                Tambah Keuskupan
              </Link>
            }
          />
        ) : (
          <>
            <InductiveFilterBar>
              <div className="flex min-w-52 flex-1 flex-col gap-1.5">
                <label
                  htmlFor="kevikepan-search"
                  className="text-xs font-semibold text-muted"
                >
                  Cari Kevikepan
                </label>
                <input
                  id="kevikepan-search"
                  type="search"
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Cari kode atau nama kevikepan…"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                />
              </div>

              <div className="flex min-w-52 flex-col gap-1.5">
                <label
                  htmlFor="kevikepan-keuskupan-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Filter Keuskupan
                </label>
                <select
                  id="kevikepan-keuskupan-filter"
                  value={keuskupanId}
                  onChange={(e) => {
                    const next = new URLSearchParams(searchParams);
                    if (e.target.value) {
                      next.set("keuskupanId", e.target.value);
                    } else {
                      next.delete("keuskupanId");
                    }
                    setSearchParams(next);
                  }}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                >
                  <option value="">Semua Keuskupan</option>
                  {keuskupanList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex min-w-36 flex-col gap-1.5">
                <label
                  htmlFor="kevikepan-status-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Status
                </label>
                <select
                  id="kevikepan-status-filter"
                  value={activeFilter}
                  onChange={(e) =>
                    setActiveFilter(e.target.value as "" | "true" | "false")
                  }
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                >
                  <option value="">Semua</option>
                  <option value="true">Aktif</option>
                  <option value="false">Nonaktif</option>
                </select>
              </div>
            </InductiveFilterBar>

            {kevikepanQuery.isLoading ? (
              <p className="text-sm text-muted">Memuat data…</p>
            ) : items.length === 0 ? (
              <InductiveEmptyState
                icon={Network}
                title="Belum ada kevikepan"
                description="Tambahkan kevikepan baru di bawah keuskupan yang dipilih."
                action={
                  <Link to={newHref} className={buttonClass("primary")}>
                    Tambah Kevikepan
                  </Link>
                }
              />
            ) : (
              <div className="flex flex-col gap-3">
                {items.map((item) => (
                  <Card
                    key={item.id}
                    className="flex flex-col justify-between gap-4 p-5 sm:flex-row sm:items-center"
                  >
                    <div className="flex flex-col gap-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-base font-semibold text-fg">
                          {item.nama}
                        </span>
                        <Badge tone="neutral">{item.kode}</Badge>
                        <Badge tone={item.active ? "success" : "neutral"}>
                          {item.active ? "Aktif" : "Nonaktif"}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted">
                        Keuskupan:{" "}
                        {keuskupanMap.get(item.keuskupanId ?? "") ??
                          item.keuskupanId}
                        {item.vikep ? ` · Vikep: ${item.vikep}` : ""}
                      </p>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <Link
                        to={`/settings/gerejawi/paroki?keuskupanId=${item.keuskupanId}&kevikepanId=${item.id}`}
                        className={buttonClass("ghost")}
                      >
                        Kelola Paroki →
                      </Link>
                      <Link
                        to={`/settings/gerejawi/kevikepan/${item.id}`}
                        className={buttonClass("outline")}
                      >
                        Ubah
                      </Link>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </>
        )
      ) : null}
    </div>
  );
}
