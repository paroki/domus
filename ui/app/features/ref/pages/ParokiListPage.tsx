import { Church, Plus } from "lucide-react";
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
import { useKeuskupanSearch, useParokiSearch } from "../api/queries";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Paroki — Domus" }];
}

export default function ParokiListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const [searchParams, setSearchParams] = useSearchParams();
  const keuskupanId = searchParams.get("keuskupanId") ?? "";
  const [q, setQ] = useState("");
  const debouncedQ = useDebouncedValue(q);

  const keuskupanQuery = useKeuskupanSearch({}, canManage);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const parokiQuery = useParokiSearch(
    {
      q: debouncedQ || undefined,
      filter: keuskupanId ? { keuskupanId } : undefined,
    },
    canManage,
  );
  const items = parokiQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const newHref = keuskupanId
    ? `/settings/gerejawi/paroki/new?keuskupanId=${keuskupanId}`
    : "/settings/gerejawi/paroki/new";

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
          { label: "Paroki" },
        ]}
        title="Paroki"
        subtitle="Daftar tenant paroki yang terdaftar di Domus."
        action={
          canManage ? (
            <Link to={newHref} className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              Tambah Paroki
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola daftar paroki.
        </div>
      ) : null}

      {canManage ? (
        <>
          <InductiveFilterBar>
            <div className="flex min-w-56 flex-1 flex-col gap-1.5">
              <label
                htmlFor="paroki-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Paroki
              </label>
              <input
                id="paroki-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau nama paroki…"
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
              />
            </div>

            <div className="flex min-w-56 flex-col gap-1.5">
              <label
                htmlFor="paroki-keuskupan-filter"
                className="text-xs font-semibold text-muted"
              >
                Filter Keuskupan
              </label>
              <select
                id="paroki-keuskupan-filter"
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
          </InductiveFilterBar>

          {parokiQuery.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={Church}
              title="Belum ada paroki"
              description="Tambahkan paroki baru untuk memulai administrasi tenant."
              action={
                <Link to={newHref} className={buttonClass("primary")}>
                  Tambah Paroki
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
                      <Badge tone={item.active ? "success" : "neutral"}>
                        {item.active ? "Aktif" : "Nonaktif"}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted">
                      Kode: <span className="font-mono">{item.kode}</span>
                      {item.timezone ? ` · Zona waktu: ${item.timezone}` : ""}
                    </p>
                  </div>
                  <Link
                    to={`/settings/gerejawi/paroki/${item.id}`}
                    className={buttonClass("outline")}
                  >
                    Ubah &amp; Kelola Admin
                  </Link>
                </Card>
              ))}
            </div>
          )}
        </>
      ) : null}
    </div>
  );
}
