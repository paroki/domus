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
import { useKeuskupanSearch } from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Keuskupan — Domus" }];
}

export default function KeuskupanListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [q, setQ] = useState("");
  const [activeFilter, setActiveFilter] = useState<"" | "true" | "false">("");
  const debouncedQ = useDebouncedValue(q);

  const search = useKeuskupanSearch(
    {
      q: debouncedQ || undefined,
      filter: activeFilter ? { active: activeFilter === "true" } : undefined,
    },
    canManage,
  );
  const items = search.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Keuskupan" },
        ]}
        title="Keuskupan"
        subtitle="Daftar keuskupan yang menaungi kevikepan dan tenant paroki."
        action={
          canManage ? (
            <Link
              to="/settings/gerejawi/keuskupan/new"
              className={buttonClass("primary")}
            >
              <Plus aria-hidden className="size-4" />
              Tambah Keuskupan
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola keuskupan.
        </div>
      ) : null}

      {canManage ? (
        <>
          <InductiveFilterBar>
            <div className="flex min-w-0 flex-1 flex-col gap-1.5 sm:min-w-60">
              <label
                htmlFor="keuskupan-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Keuskupan
              </label>
              <input
                id="keuskupan-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau nama keuskupan…"
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-base text-fg focus:border-primary focus:outline-none sm:text-sm"
              />
            </div>

            <div className="flex w-full flex-col gap-1.5 sm:w-auto sm:min-w-36">
              <label
                htmlFor="keuskupan-status-filter"
                className="text-xs font-semibold text-muted"
              >
                Status
              </label>
              <select
                id="keuskupan-status-filter"
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

          {search.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={Building2}
              title="Belum ada keuskupan"
              description="Tambahkan keuskupan pertama sebelum membuat kevikepan atau paroki."
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
                      {item.uskup ? ` · Uskup: ${item.uskup}` : ""}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      to={`/settings/gerejawi/kevikepan?keuskupanId=${item.id}`}
                      className={buttonClass("ghost")}
                    >
                      Kelola Kevikepan →
                    </Link>
                    <Link
                      to={`/settings/gerejawi/paroki?keuskupanId=${item.id}`}
                      className={buttonClass("ghost")}
                    >
                      Kelola Paroki →
                    </Link>
                    <Link
                      to={`/settings/gerejawi/keuskupan/${item.id}`}
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
