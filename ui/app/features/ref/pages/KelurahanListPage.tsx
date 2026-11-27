import { Compass, Plus } from "lucide-react";
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
import {
  useKecamatanSearch,
  useKelurahanSearch,
  useKotaSearch,
} from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Kelurahan / Desa — Domus" }];
}

export default function KelurahanListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [searchParams, setSearchParams] = useSearchParams();
  const kotaId = searchParams.get("kotaId") ?? "";
  const kecamatanId = searchParams.get("kecamatanId") ?? "";
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const kotaQuery = useKotaSearch({}, canManage);
  const kotaList = kotaQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const kecQuery = useKecamatanSearch(
    { filter: kotaId ? { kotaId } : undefined },
    canManage,
  );
  const kecList = kecQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const kecMap = new Map(kecList.map((k) => [k.id, k.nama]));

  const kelQuery = useKelurahanSearch(
    {
      q: debouncedQ || undefined,
      filter:
        kecamatanId || activeFilter
          ? {
              kecamatanId: kecamatanId || undefined,
              active: activeFilter === "" ? undefined : activeFilter === "true",
            }
          : undefined,
    },
    canManage && kecList.length > 0,
  );
  const items = kelQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const newParams = new URLSearchParams();
  if (kotaId) newParams.set("kotaId", kotaId);
  if (kecamatanId) newParams.set("kecamatanId", kecamatanId);
  const newHref = newParams.toString()
    ? `/settings/sipil/kelurahan/new?${newParams.toString()}`
    : "/settings/sipil/kelurahan/new";

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Wilayah Sipil", to: "/settings/sipil" },
          { label: "Kelurahan / Desa" },
        ]}
        title="Kelurahan / Desa"
        subtitle="Daftar kelurahan atau desa di bawah masing-masing kecamatan."
        action={
          canManage && kecList.length > 0 ? (
            <Link to={newHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              Tambah Kelurahan
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
        !kecQuery.isLoading && kecList.length === 0 && !kotaId ? (
          <InductiveEmptyState
            icon={Compass}
            title="Buat kecamatan terlebih dahulu"
            description="Kelurahan harus terhubung ke kecamatan. Buat kecamatan pertama terlebih dahulu."
            action={
              <Link
                to="/settings/sipil/kecamatan/new"
                className={buttonClass("primary")}
              >
                Tambah Kecamatan
              </Link>
            }
          />
        ) : (
          <>
            <InductiveFilterBar>
              <div className="flex min-w-0 flex-1 flex-col gap-1.5 sm:min-w-48">
                <label
                  htmlFor="kelurahan-search"
                  className="text-xs font-semibold text-muted"
                >
                  Cari Kelurahan / Desa
                </label>
                <input
                  id="kelurahan-search"
                  type="search"
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Cari kode atau nama kelurahan…"
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
                <label
                  htmlFor="kelurahan-kota-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Filter Kota
                </label>
                <select
                  id="kelurahan-kota-filter"
                  value={kotaId}
                  onChange={(e) => {
                    const next = new URLSearchParams(searchParams);
                    if (e.target.value) {
                      next.set("kotaId", e.target.value);
                    } else {
                      next.delete("kotaId");
                    }
                    next.delete("kecamatanId");
                    setSearchParams(next);
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                >
                  <option value="">Semua Kota</option>
                  {kotaList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
                <label
                  htmlFor="kelurahan-kecamatan-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Filter Kecamatan
                </label>
                <select
                  id="kelurahan-kecamatan-filter"
                  value={kecamatanId}
                  onChange={(e) => {
                    const next = new URLSearchParams(searchParams);
                    if (e.target.value) {
                      next.set("kecamatanId", e.target.value);
                    } else {
                      next.delete("kecamatanId");
                    }
                    setSearchParams(next);
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                >
                  <option value="">Semua Kecamatan</option>
                  {kecList.map((k) => (
                    <option key={k.id} value={k.id}>
                      {k.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
                <label
                  htmlFor="kelurahan-status-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Status
                </label>
                <select
                  id="kelurahan-status-filter"
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

            {kelQuery.isLoading ? (
              <p className="text-sm text-muted">Memuat data…</p>
            ) : items.length === 0 ? (
              <InductiveEmptyState
                icon={Compass}
                title="Belum ada kelurahan/desa"
                description="Tambahkan kelurahan atau desa baru di bawah kecamatan yang dipilih."
                action={
                  <Link to={newHref} className={buttonClass("primary")}>
                    Tambah Kelurahan
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
                          {item.nama}
                        </span>
                        <Badge tone="neutral">{item.kode}</Badge>
                        {item.kodePos ? (
                          <Badge tone="neutral">{item.kodePos}</Badge>
                        ) : null}
                        <Badge tone={item.active ? "success" : "neutral"}>
                          {item.active ? "Aktif" : "Nonaktif"}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted">
                        Kecamatan:{" "}
                        {kecMap.get(item.kecamatanId ?? "") ?? item.kecamatanId}
                      </p>
                    </div>
                    <Link
                      to={`/settings/sipil/kelurahan/${item.id}`}
                      className={buttonClass("outline")}
                    >
                      Ubah
                    </Link>
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
