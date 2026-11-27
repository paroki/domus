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
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { useDebouncedValue } from "~/shared/util/use-debounced-value";
import {
  useUmatLingkunganList,
  useUmatLookups,
  useUmatSearch,
  useUmatWilayahList,
} from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import {
  genderLabel,
  isUmatAdmin,
  parseUmatDirectoryFilters,
  type UmatAlmarhumFilter,
} from "../model";

export function meta() {
  return [{ title: "Data Umat · Domus" }];
}

export default function UmatDirectoryPage() {
  const { data: me } = useMe();
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = parseUmatDirectoryFilters(searchParams);
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
  const agamaQuery = useUmatLookups(
    activeParokiId,
    "agama",
    Boolean(me && allowed && activeParokiId),
  );
  const statusPernikahanQuery = useUmatLookups(
    activeParokiId,
    "status_pernikahan",
    Boolean(me && allowed && activeParokiId),
  );

  const searchQuery = useUmatSearch(
    activeParokiId,
    {
      q: debouncedQ || undefined,
      filter: {
        ...(filters.nik.trim() ? { nik: filters.nik.trim() } : {}),
        ...(filters.wilayahId ? { wilayahId: filters.wilayahId } : {}),
        ...(filters.lingkunganId ? { lingkunganId: filters.lingkunganId } : {}),
        ...(filters.jenisKelamin ? { jenisKelamin: filters.jenisKelamin } : {}),
        ...(filters.agama ? { agama: filters.agama } : {}),
        ...(filters.statusPernikahan
          ? { statusPernikahan: filters.statusPernikahan }
          : {}),
        ...(filters.almarhum !== "all"
          ? { almarhum: filters.almarhum === "deceased" }
          : {}),
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
      filters.nik ||
      filters.wilayahId ||
      filters.lingkunganId ||
      filters.jenisKelamin ||
      filters.agama ||
      filters.statusPernikahan ||
      filters.almarhum !== "all",
  );

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[{ label: "Beranda", to: "/" }, { label: "Data Umat" }]}
        title="Data Umat"
        subtitle="Direktori umat paroki, pencarian NIK terenkripsi, dan pengelolaan Kartu Keluarga."
        action={
          <Link to="/umat/new" className="no-underline">
            <Button>
              <Plus aria-hidden className="size-4" />
              Tambah Umat
            </Button>
          </Link>
        }
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <InductiveStepCard
          to="/umat/keluarga"
          icon={Home}
          plate={3}
          title="Kartu Keluarga"
          description="Kelola Kartu Keluarga, alamat domisili, dan daftar anggota per Wilayah & Lingkungan"
        />
        <InductiveStepCard
          to="/umat/referensi"
          icon={BookOpen}
          plate={4}
          title="Kamus Referensi Umat"
          description="Kelola daftar pilihan agama, pendidikan, pekerjaan, golongan darah, dan referensi umat lainnya"
        />
      </div>

      <InductiveFilterBar>
        <label className="flex min-w-52 flex-1 flex-col gap-1 text-xs font-medium text-muted">
          Cari Nama / Nama Baptis
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
              placeholder="Nama atau nama baptis…"
              className="h-10 w-full rounded-lg border border-line bg-canvas pr-3 pl-9 text-sm text-fg"
            />
          </div>
        </label>

        <label className="flex w-44 flex-col gap-1 text-xs font-medium text-muted">
          NIK (Tepat)
          <input
            type="text"
            value={filters.nik}
            onChange={(e) => updateParam("nik", e.target.value)}
            placeholder="Cari NIK…"
            className="h-10 rounded-lg border border-line bg-canvas px-3 font-mono text-sm text-fg"
          />
        </label>

        <label className="flex w-40 flex-col gap-1 text-xs font-medium text-muted">
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

        <label className="flex w-40 flex-col gap-1 text-xs font-medium text-muted">
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
          Jenis Kelamin
          <select
            value={filters.jenisKelamin}
            onChange={(e) => updateParam("jenisKelamin", e.target.value)}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">Semua</option>
            <option value="pria">Pria</option>
            <option value="wanita">Wanita</option>
          </select>
        </label>

        <label className="flex w-36 flex-col gap-1 text-xs font-medium text-muted">
          Agama
          <select
            value={filters.agama}
            onChange={(e) => updateParam("agama", e.target.value)}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">Semua</option>
            {(agamaQuery.data ?? []).map((l) => (
              <option key={l.id} value={l.code}>
                {l.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex w-40 flex-col gap-1 text-xs font-medium text-muted">
          Status Pernikahan
          <select
            value={filters.statusPernikahan}
            onChange={(e) => updateParam("statusPernikahan", e.target.value)}
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">Semua</option>
            {(statusPernikahanQuery.data ?? []).map((l) => (
              <option key={l.id} value={l.code}>
                {l.label}
              </option>
            ))}
          </select>
        </label>

        <label className="flex w-36 flex-col gap-1 text-xs font-medium text-muted">
          Status Hidup
          <select
            value={filters.almarhum}
            onChange={(e) =>
              updateParam("almarhum", e.target.value as UmatAlmarhumFilter)
            }
            className="h-10 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="all">Semua</option>
            <option value="alive">Hidup</option>
            <option value="deceased">Almarhum</option>
          </select>
        </label>
      </InductiveFilterBar>

      {searchQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">Memuat data umat…</Card>
      ) : rows.length === 0 ? (
        <InductiveEmptyState
          icon={Users}
          title="Belum ada data umat"
          description={
            hasFilters
              ? "Tidak ada data umat yang cocok dengan filter pencarian Anda."
              : "Mulai dengan menambahkan data umat pertama di paroki ini."
          }
          action={
            hasFilters ? (
              <Button variant="outline" onClick={resetFilters}>
                Atur ulang filter
              </Button>
            ) : (
              <Link to="/umat/new" className="no-underline">
                <Button>
                  <Plus aria-hidden className="size-4" />
                  Tambah Umat
                </Button>
              </Link>
            )
          }
        />
      ) : (
        <Card className="divide-y divide-line">
          {rows.map((u) => {
            const wilName = u.wilayahId ? wilayahMap.get(u.wilayahId) : "";
            const lingName = u.lingkunganId
              ? lingkunganMap.get(u.lingkunganId)
              : "";
            return (
              <Link
                key={u.id}
                to={`/umat/${u.id}`}
                className="flex flex-wrap items-center justify-between gap-4 p-4 no-underline transition-colors hover:bg-surface-2/50 sm:px-6"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-bold text-fg">
                      {u.nama}
                      {u.namaBaptis ? ` (${u.namaBaptis})` : ""}
                    </span>
                    {u.nikMasked ? (
                      <Badge tone="neutral">
                        <span className="font-mono">{u.nikMasked}</span>
                      </Badge>
                    ) : null}
                    {u.almarhum ? <Badge tone="danger">Almarhum</Badge> : null}
                  </div>
                  <p className="mt-1 truncate text-xs text-muted">
                    {[genderLabel(u.jenisKelamin), wilName, lingName]
                      .filter(Boolean)
                      .join(" · ")}
                  </p>
                </div>
                <span className="shrink-0 text-xs font-medium text-primary">
                  Ubah / Detail →
                </span>
              </Link>
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
