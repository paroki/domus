import { Compass, Plus } from "lucide-react";
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
import { useWilayahSearch } from "../api/queries";
import { isParokiAdmin } from "../model";

export function meta() {
  return [{ title: "Wilayah & Stasi — Domus" }];
}

export default function WilayahListPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isParokiAdmin(me?.roles);
  const hasTenant = Boolean(me?.activeParokiId);

  const [q, setQ] = useState("");
  const [jenis, setJenis] = useState<"" | "wilayah" | "stasi">("");
  const debouncedQ = useDebouncedValue(q);

  const search = useWilayahSearch(
    {
      q: debouncedQ || undefined,
      filter: jenis ? { jenis } : undefined,
    },
    canManage && hasTenant,
  );
  const items = search.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Administrasi Paroki", to: "/settings/paroki" },
          { label: "Wilayah & Stasi" },
        ]}
        title="Wilayah &amp; Stasi"
        subtitle="Kelola wilayah teritorial dan stasi pada paroki aktif."
        action={
          canManage && hasTenant ? (
            <Link
              to="/settings/paroki/wilayah/new"
              className={buttonClass("primary")}
            >
              <Plus aria-hidden className="size-4" />
              Tambah Wilayah
            </Link>
          ) : undefined
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki yang dapat mengelola wilayah dan stasi.
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
        <>
          <InductiveFilterBar>
            <div className="flex min-w-56 flex-1 flex-col gap-1.5">
              <label
                htmlFor="wilayah-search"
                className="text-xs font-semibold text-muted"
              >
                Cari Wilayah / Stasi
              </label>
              <input
                id="wilayah-search"
                type="search"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Cari kode atau nama…"
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
              />
            </div>

            <div className="flex min-w-44 flex-col gap-1.5">
              <label
                htmlFor="wilayah-jenis-filter"
                className="text-xs font-semibold text-muted"
              >
                Jenis
              </label>
              <select
                id="wilayah-jenis-filter"
                value={jenis}
                onChange={(e) =>
                  setJenis(e.target.value as "" | "wilayah" | "stasi")
                }
                className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg focus:border-primary focus:outline-none"
              >
                <option value="">Semua jenis</option>
                <option value="wilayah">Wilayah</option>
                <option value="stasi">Stasi</option>
              </select>
            </div>
          </InductiveFilterBar>

          {search.isLoading ? (
            <p className="text-sm text-muted">Memuat data…</p>
          ) : items.length === 0 ? (
            <InductiveEmptyState
              icon={Compass}
              title="Belum ada wilayah atau stasi"
              description="Tambahkan wilayah atau stasi pertama sebelum membuat lingkungan."
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
                      <Badge tone="neutral">
                        {item.jenis === "stasi" ? "Stasi" : "Wilayah"}
                      </Badge>
                      <Badge tone={item.active ? "success" : "neutral"}>
                        {item.active ? "Aktif" : "Nonaktif"}
                      </Badge>
                    </div>
                    <p className="text-xs text-muted">
                      Kode: <span className="font-mono">{item.kode}</span>
                      {item.koordinator
                        ? ` · Koordinator: ${item.koordinator}`
                        : ""}
                    </p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <Link
                      to={`/settings/paroki/lingkungan?wilayahId=${item.id}`}
                      className={buttonClass("ghost")}
                    >
                      Kelola Lingkungan →
                    </Link>
                    <Link
                      to={`/settings/paroki/wilayah/${item.id}`}
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
