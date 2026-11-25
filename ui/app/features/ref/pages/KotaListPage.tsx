import { Building2, Plus } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router";
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
import { useKotaSearch } from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Kota / Kabupaten — Domus" }];
}

export default function KotaListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const search = useKotaSearch(
    {
      q: debouncedQ || undefined,
      filter:
        activeFilter === "" ? undefined : { active: activeFilter === "true" },
    },
    canManage,
  );
  const items = search.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Wilayah Sipil", to: "/settings/sipil" },
          { label: "Kota / Kabupaten" },
        ]}
        title="Kota / Kabupaten"
        subtitle="Daftar kota dan kabupaten sebagai induk kecamatan."
        action={
          canManage ? (
            <Link
              to="/settings/sipil/kota/new"
              className={buttonClass("primary")}
            >
              <Plus aria-hidden className="size-4" />
              Tambah Kota
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
        <>
          <InductiveFilterBar>
            <div className="flex min-w-56 flex-1 flex-col gap-1.5">
              <label
                htmlFor="kota-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Kota / Kabupaten
              </label>
              <input
                id="kota-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau nama kota…"
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
              />
            </div>

            <div className="flex min-w-36 flex-col gap-1.5">
              <label
                htmlFor="kota-status-filter"
                className="text-xs font-semibold text-muted"
              >
                Status
              </label>
              <select
                id="kota-status-filter"
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

          {search.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={Building2}
              title="Belum ada kota/kabupaten"
              description="Tambahkan kota atau kabupaten pertama sebelum membuat kecamatan."
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
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      to={`/settings/sipil/kecamatan?kotaId=${item.id}`}
                      className={buttonClass("ghost")}
                    >
                      Kelola Kecamatan →
                    </Link>
                    <Link
                      to={`/settings/sipil/kota/${item.id}`}
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
      ) : null}
    </div>
  );
}
