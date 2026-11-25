import { MapPin, Plus } from "lucide-react";
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
import { useKecamatanSearch, useKotaSearch } from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Kecamatan — Domus" }];
}

export default function KecamatanListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [searchParams, setSearchParams] = useSearchParams();
  const kotaId = searchParams.get("kotaId") ?? "";
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const kotaQuery = useKotaSearch({}, canManage);
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kotaMap = new Map(kotaList.map((k) => [k.id, k.nama]));

  const kecQuery = useKecamatanSearch(
    {
      q: debouncedQ || undefined,
      filter:
        kotaId || activeFilter
          ? {
              kotaId: kotaId || undefined,
              active: activeFilter === "" ? undefined : activeFilter === "true",
            }
          : undefined,
    },
    canManage && kotaList.length > 0,
  );
  const items = kecQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const newHref = kotaId
    ? `/settings/sipil/kecamatan/new?kotaId=${kotaId}`
    : "/settings/sipil/kecamatan/new";

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Wilayah Sipil", to: "/settings/sipil" },
          { label: "Kecamatan" },
        ]}
        title="Kecamatan"
        subtitle="Daftar kecamatan di bawah masing-masing kota/kabupaten."
        action={
          canManage && kotaList.length > 0 ? (
            <Link to={newHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              Tambah Kecamatan
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola wilayah sipil.
        </div>
      ) : null}

      {canManage ? (
        !kotaQuery.isLoading && kotaList.length === 0 ? (
          <InductiveEmptyState
            icon={MapPin}
            title="Buat kota terlebih dahulu"
            description="Kecamatan harus terhubung ke kota/kabupaten. Buat kota pertama terlebih dahulu."
            action={
              <Link
                to="/settings/sipil/kota/new"
                className={buttonClass("primary")}
              >
                Tambah Kota
              </Link>
            }
          />
        ) : (
          <>
            <InductiveFilterBar>
              <div className="flex min-w-52 flex-1 flex-col gap-1.5">
                <label
                  htmlFor="kecamatan-search"
                  className="text-xs font-semibold text-muted"
                >
                  Cari Kecamatan
                </label>
                <input
                  id="kecamatan-search"
                  type="search"
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Cari kode atau nama kecamatan…"
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                />
              </div>

              <div className="flex min-w-52 flex-col gap-1.5">
                <label
                  htmlFor="kecamatan-kota-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Filter Kota / Kabupaten
                </label>
                <select
                  id="kecamatan-kota-filter"
                  value={kotaId}
                  onChange={(e) => {
                    const next = new URLSearchParams(searchParams);
                    if (e.target.value) {
                      next.set("kotaId", e.target.value);
                    } else {
                      next.delete("kotaId");
                    }
                    setSearchParams(next);
                  }}
                  className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
                >
                  <option value="">Semua Kota</option>
                  {kotaList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex min-w-36 flex-col gap-1.5">
                <label
                  htmlFor="kecamatan-status-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Status
                </label>
                <select
                  id="kecamatan-status-filter"
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

            {kecQuery.isLoading ? (
              <p className="text-sm text-muted">Memuat data…</p>
            ) : items.length === 0 ? (
              <InductiveEmptyState
                icon={MapPin}
                title="Belum ada kecamatan"
                description="Tambahkan kecamatan baru di bawah kota/kabupaten yang dipilih."
                action={
                  <Link to={newHref} className={buttonClass("primary")}>
                    Tambah Kecamatan
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
                        Kota: {kotaMap.get(item.kotaId ?? "") ?? item.kotaId}
                      </p>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <Link
                        to={`/settings/sipil/kelurahan?kotaId=${item.kotaId}&kecamatanId=${item.id}`}
                        className={buttonClass("ghost")}
                      >
                        Kelola Kelurahan →
                      </Link>
                      <Link
                        to={`/settings/sipil/kecamatan/${item.id}`}
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
