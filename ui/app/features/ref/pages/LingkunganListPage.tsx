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
import { useLingkunganSearch, useWilayahSearch } from "../api/queries";
import { isParokiAdmin } from "../model";

export function meta() {
  return [{ title: "Lingkungan — Domus" }];
}

export default function LingkunganListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);

  const [searchParams, setSearchParams] = useSearchParams();
  const wilayahId = searchParams.get("wilayahId") ?? "";
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const wilayahQuery = useWilayahSearch({}, canManage && hasTenant);
  const wilayahList =
    wilayahQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const wilayahMap = new Map(wilayahList.map((w) => [w.id, w.nama]));

  const lingkunganQuery = useLingkunganSearch(
    {
      q: debouncedQ || undefined,
      filter:
        wilayahId || activeFilter
          ? {
              wilayahId: wilayahId || undefined,
              active: activeFilter === "" ? undefined : activeFilter === "true",
            }
          : undefined,
    },
    canManage && hasTenant,
  );
  const items = lingkunganQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const newHref = wilayahId
    ? `/settings/paroki/lingkungan/new?wilayahId=${wilayahId}`
    : "/settings/paroki/lingkungan/new";

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Administrasi Paroki", to: "/settings/paroki" },
          { label: "Lingkungan" },
        ]}
        title="Lingkungan"
        subtitle="Kelola lingkungan di bawah setiap wilayah atau stasi paroki aktif."
        action={
          canManage && hasTenant && wilayahList.length > 0 ? (
            <Link to={newHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              Tambah Lingkungan
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki yang dapat mengelola lingkungan.
        </div>
      ) : null}

      {canManage && !hasTenant ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Pilih paroki aktif terlebih dahulu melalui menu pindah paroki di atas.
        </div>
      ) : null}

      {canManage && hasTenant ? (
        !wilayahQuery.isLoading && wilayahList.length === 0 ? (
          <InductiveEmptyState
            icon={MapPin}
            title="Buat wilayah terlebih dahulu"
            description="Lingkungan harus terhubung ke wilayah atau stasi. Buat wilayah pertama terlebih dahulu."
            action={
              <Link
                to="/settings/paroki/wilayah/new"
                className={buttonClass("primary")}
              >
                Tambah Wilayah
              </Link>
            }
          />
        ) : (
          <>
            <InductiveFilterBar>
              <div className="flex min-w-0 flex-1 flex-col gap-1.5 sm:min-w-56">
                <label
                  htmlFor="lingkungan-search"
                  className="text-xs font-semibold text-muted"
                >
                  Cari Lingkungan
                </label>
                <input
                  id="lingkungan-search"
                  type="search"
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Cari kode atau nama lingkungan…"
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                />
              </div>

              <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
                <label
                  htmlFor="lingkungan-wilayah-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Filter Wilayah
                </label>
                <select
                  id="lingkungan-wilayah-filter"
                  value={wilayahId}
                  onChange={(e) => {
                    const next = new URLSearchParams(searchParams);
                    if (e.target.value) {
                      next.set("wilayahId", e.target.value);
                    } else {
                      next.delete("wilayahId");
                    }
                    setSearchParams(next);
                  }}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
                >
                  <option value="">Semua wilayah</option>
                  {wilayahList.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.nama}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-44">
                <label
                  htmlFor="lingkungan-status-filter"
                  className="text-xs font-semibold text-muted"
                >
                  Status
                </label>
                <select
                  id="lingkungan-status-filter"
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

            {lingkunganQuery.isLoading ? (
              <p className="text-sm text-muted">Memuat data…</p>
            ) : items.length === 0 ? (
              <InductiveEmptyState
                icon={MapPin}
                title="Belum ada lingkungan"
                description="Tambahkan lingkungan baru di bawah wilayah atau stasi yang dipilih."
                action={
                  <Link to={newHref} className={buttonClass("primary")}>
                    Tambah Lingkungan
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
                        <Badge tone={item.active ? "success" : "neutral"}>
                          {item.active ? "Aktif" : "Nonaktif"}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted">
                        Kode: <span className="font-mono">{item.kode}</span>
                        {wilayahMap.get(item.wilayahId)
                          ? ` · Wilayah: ${wilayahMap.get(item.wilayahId)}`
                          : ""}
                        {item.ketua ? ` · Ketua: ${item.ketua}` : ""}
                      </p>
                    </div>
                    <Link
                      to={`/settings/paroki/lingkungan/${item.id}`}
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
