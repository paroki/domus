import { BookOpen, Home, Plus, Search, Users } from "lucide-react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { useDebouncedValue } from "~/shared/util/use-debounced-value";
import {
  useKeluargaSearch,
  useUmatLingkunganList,
  useUmatWilayahList,
} from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import {
  isUmatAdmin,
  type KeluargaAktifFilter,
  parseKeluargaDirectoryFilters,
} from "../model";

export function meta() {
  return [{ title: "Kartu Keluarga · Data Umat · Domus" }];
}

export default function KeluargaListPage() {
  const { data: me } = useMe();
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = parseKeluargaDirectoryFilters(searchParams);
  const [qInput, setQInput] = useState(filters.q);
  const debouncedQ = useDebouncedValue(qInput.trim(), 250);

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const wilayahQuery = useUmatWilayahList(
    activeParokiId,
    Boolean(me && allowed && activeParokiId),
  );
  const lingkunganQuery = useUmatLingkunganList(
    activeParokiId,
    filters.wilayahId || undefined,
    Boolean(me && allowed && activeParokiId),
  );

  const searchQuery = useKeluargaSearch(
    activeParokiId,
    {
      q: debouncedQ || undefined,
      filter: {
        ...(filters.nomorKk.trim() ? { nomorKk: filters.nomorKk.trim() } : {}),
        ...(filters.wilayahId ? { wilayahId: filters.wilayahId } : {}),
        ...(filters.lingkunganId ? { lingkunganId: filters.lingkunganId } : {}),
        ...(filters.aktif !== "all" ? { aktif: filters.aktif === "true" } : {}),
      },
      limit: 25,
    },
    Boolean(me && allowed && activeParokiId),
  );

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  const wilayahMap = new Map(
    (wilayahQuery.data ?? []).map((w) => [w.id ?? "", w.nama ?? ""]),
  );
  const lingkunganMap = new Map(
    (lingkunganQuery.data ?? []).map((l) => [l.id ?? "", l.nama ?? ""]),
  );

  const rows = searchQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  function updateParam(key: string, value: string) {
    const next = new URLSearchParams(searchParams);
    if (value) next.set(key, value);
    else next.delete(key);
    setSearchParams(next, { replace: true });
  }

  function resetFilters() {
    setQInput("");
    setSearchParams(new URLSearchParams(), { replace: true });
  }

  const hasFilters = Boolean(
    filters.q ||
      filters.nomorKk ||
      filters.wilayahId ||
      filters.lingkunganId ||
      filters.aktif !== "all",
  );

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          { label: "Kartu Keluarga" },
        ]}
        title="Kartu Keluarga"
        subtitle="Daftar Kartu Keluarga (KK) paroki, alamat domisili, dan anggota keluarga."
        action={
          <Link to="/umat/keluarga/new" className="no-underline">
            <Button>
              <Plus aria-hidden className="size-4" />
              Tambah Kartu Keluarga
            </Button>
          </Link>
        }
      />

      <InductiveHubGrid>
        <InductiveStepCard
          to="/umat/referensi"
          icon={BookOpen}
          plate={4}
          title="Kamus Referensi Umat"
          description="Kelola pilihan status tempat tinggal, status ekonomi, hubungan keluarga, dan referensi umat lainnya"
        />
      </InductiveHubGrid>

      <InductiveFilterBar>
        <label className="flex min-w-52 flex-1 flex-col gap-1 text-xs font-medium text-muted">
          Cari Nama Keluarga / Alamat
          <div className="relative">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted"
            />
            <input
              type="search"
              value={qInput}
              onChange={(e) => {
                setQInput(e.target.value);
                updateParam("q", e.target.value);
              }}
              placeholder="Nama keluarga atau alamat…"
              className="h-10 w-full rounded-lg border border-line bg-canvas pr-3 pl-9 text-sm text-fg"
            />
          </div>
        </label>

        <label className="flex w-44 flex-col gap-1 text-xs font-medium text-muted">
          Nomor KK (Tepat)
          <input
            type="text"
            value={filters.nomorKk}
            onChange={(e) => updateParam("nomorKk", e.target.value)}
            placeholder="Cari Nomor KK…"
            className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg"
          />
        </label>

        <label className="flex w-44 flex-col gap-1 text-xs font-medium text-muted">
          Wilayah
          <select
            value={filters.wilayahId}
            onChange={(e) => {
              const next = new URLSearchParams(searchParams);
              if (e.target.value) next.set("wilayahId", e.target.value);
              else next.delete("wilayahId");
              next.delete("lingkunganId");
              setSearchParams(next, { replace: true });
            }}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">Semua Wilayah</option>
            {(wilayahQuery.data ?? []).map((w) => (
              <option key={w.id} value={w.id}>
                {w.nama}
              </option>
            ))}
          </select>
        </label>

        <label className="flex w-44 flex-col gap-1 text-xs font-medium text-muted">
          Lingkungan
          <select
            value={filters.lingkunganId}
            onChange={(e) => updateParam("lingkunganId", e.target.value)}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">Semua Lingkungan</option>
            {(lingkunganQuery.data ?? []).map((l) => (
              <option key={l.id} value={l.id}>
                {l.nama}
              </option>
            ))}
          </select>
        </label>

        <label className="flex w-36 flex-col gap-1 text-xs font-medium text-muted">
          Status
          <select
            value={filters.aktif}
            onChange={(e) =>
              updateParam("aktif", e.target.value as KeluargaAktifFilter)
            }
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="all">Semua</option>
            <option value="true">Aktif</option>
            <option value="false">Nonaktif</option>
          </select>
        </label>
      </InductiveFilterBar>

      {searchQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">Memuat Kartu Keluarga…</Card>
      ) : rows.length === 0 ? (
        <InductiveEmptyState
          icon={Home}
          title="Belum ada Kartu Keluarga"
          description={
            hasFilters
              ? "Tidak ada Kartu Keluarga yang cocok dengan filter pencarian Anda."
              : "Mulai dengan menambahkan Kartu Keluarga pertama di paroki ini."
          }
          action={
            hasFilters ? (
              <Button variant="outline" onClick={resetFilters}>
                Atur ulang filter
              </Button>
            ) : (
              <Link to="/umat/keluarga/new" className="no-underline">
                <Button>
                  <Plus aria-hidden className="size-4" />
                  Tambah Kartu Keluarga
                </Button>
              </Link>
            )
          }
        />
      ) : (
        <Card className="divide-y divide-line">
          {rows.map((k) => {
            const wilName = k.wilayahId ? wilayahMap.get(k.wilayahId) : "";
            const lingName = k.lingkunganId
              ? lingkunganMap.get(k.lingkunganId)
              : "";
            const addParams = new URLSearchParams();
            if (k.id) addParams.set("keluargaId", k.id);
            if (k.wilayahId) addParams.set("wilayahId", k.wilayahId);
            if (k.lingkunganId) addParams.set("lingkunganId", k.lingkunganId);

            return (
              <div
                key={k.id}
                className="flex flex-wrap items-center justify-between gap-4 p-4 sm:px-6"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-bold text-fg">
                      {k.namaKeluarga}
                    </span>
                    {k.nomorKkMasked ? (
                      <Badge tone="neutral">
                        <span className="font-mono">{k.nomorKkMasked}</span>
                      </Badge>
                    ) : null}
                    <Badge tone={k.aktif ? "success" : "neutral"}>
                      {k.aktif ? "Aktif" : "Nonaktif"}
                    </Badge>
                  </div>
                  <p className="mt-1 truncate text-xs text-muted">
                    {[wilName, lingName, k.alamat]
                      .filter(Boolean)
                      .join(" · ") || "-"}
                  </p>
                </div>

                <div className="flex flex-wrap items-center gap-3 text-xs font-medium">
                  <Link
                    to={`/umat/new?${addParams.toString()}`}
                    className="inline-flex items-center gap-1 text-muted no-underline hover:text-primary"
                  >
                    <Users aria-hidden className="size-3.5" />+ Tambah Anggota →
                  </Link>
                  <Link
                    to={`/umat/keluarga/${k.id}`}
                    className="text-primary no-underline hover:underline"
                  >
                    Kelola KK →
                  </Link>
                </div>
              </div>
            );
          })}
        </Card>
      )}

      {searchQuery.hasNextPage ? (
        <div className="flex justify-center">
          <Button
            variant="outline"
            disabled={searchQuery.isFetchingNextPage}
            onClick={() => searchQuery.fetchNextPage()}
          >
            {searchQuery.isFetchingNextPage ? "Memuat…" : "Muat lebih banyak"}
          </Button>
        </div>
      ) : null}
    </div>
  );
}
