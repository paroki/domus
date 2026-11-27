import { ChevronRight, Music, Plus, Search } from "lucide-react";
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
} from "~/shared/ui/iui";
import { useDebouncedValue } from "~/shared/util/use-debounced-value";
import { useCantusSearch } from "../api/queries";
import {
  CANTUS_KATEGORI_SUGGESTIONS,
  CANTUS_SUMBER_SUGGESTIONS,
  cantusJenisLabel,
  isCantusAdmin,
  parseCantusDirectoryFilters,
  parseMazmurMeta,
} from "../model";

export function meta() {
  return [{ title: "Katalog Nyanyian & Mazmur · Domus" }];
}

const JENIS_CHIPS = [
  { value: "", label: "Semua Jenis" },
  { value: "lagu", label: "Lagu Liturgi" },
  { value: "mazmur", label: "Mazmur & BPI" },
  { value: "profan", label: "Profan" },
] as const;

export default function CantusDirectoryPage() {
  const { data: me } = useMe();
  const [searchParams, setSearchParams] = useSearchParams();
  const filters = parseCantusDirectoryFilters(searchParams);
  const [qInput, setQInput] = useState(filters.q);
  const debouncedQ = useDebouncedValue(qInput.trim(), 250);

  const canManage = isCantusAdmin(me);

  const activeBool =
    filters.active === "all" ? undefined : filters.active === "active";

  const searchQuery = useCantusSearch(
    {
      q: debouncedQ || undefined,
      filter: {
        ...(filters.jenis
          ? { jenis: filters.jenis as "lagu" | "mazmur" | "profan" }
          : {}),
        ...(filters.sumber.trim() ? { sumber: filters.sumber.trim() } : {}),
        ...(filters.kategori.trim()
          ? { kategori: filters.kategori.trim() }
          : {}),
        ...(filters.tag.trim() ? { tag: filters.tag.trim() } : {}),
        ...(filters.nadaDasar.trim()
          ? { nadaDasar: filters.nadaDasar.trim() }
          : {}),
        ...(activeBool !== undefined ? { active: activeBool } : {}),
      },
      limit: 20,
    },
    Boolean(me),
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
    debouncedQ ||
      filters.jenis ||
      filters.sumber ||
      filters.kategori ||
      filters.tag ||
      filters.nadaDasar ||
      filters.active !== "active",
  );

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[{ label: "Beranda", to: "/" }, { label: "Nyanyian" }]}
        title="Katalog Nyanyian & Mazmur"
        subtitle="Inventaris global lagu liturgi resmi (Puji Syukur, Madah Bakti, Kidung Adi), Mazmur Tanggapan & BPI, serta lagu rohani lintas paroki."
        action={
          canManage ? (
            <Link to="/cantus/new" className="w-full no-underline sm:w-auto">
              <Button className="w-full sm:w-auto">
                <Plus aria-hidden className="size-4" />
                Tambah Nyanyian
              </Button>
            </Link>
          ) : undefined
        }
      />

      <InductiveFilterBar>
        <div className="flex flex-col gap-3">
          <div className="relative w-full">
            <Search
              aria-hidden
              className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted"
            />
            <input
              type="search"
              value={qInput}
              onChange={(e) => {
                setQInput(e.target.value);
                updateParam("q", e.target.value.trim());
              }}
              placeholder="Cari judul, nomor lagu, lirik, atau pengarang…"
              aria-label="Cari judul, nomor lagu, lirik, atau pengarang"
              className="h-11 w-full rounded-lg border border-line bg-canvas pr-3.5 pl-10 text-sm text-fg placeholder:text-muted"
            />
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {JENIS_CHIPS.map((chip) => {
              const selected = filters.jenis === chip.value;
              return (
                <button
                  key={chip.value || "all"}
                  type="button"
                  onClick={() => updateParam("jenis", chip.value)}
                  className={`min-h-10 rounded-full px-3.5 text-xs font-semibold transition-colors ${
                    selected
                      ? "bg-primary text-on-primary"
                      : "border border-line bg-surface text-muted hover:text-fg"
                  }`}
                >
                  {chip.label}
                </button>
              );
            })}
          </div>

          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-5">
            <select
              value={filters.sumber}
              onChange={(e) => updateParam("sumber", e.target.value)}
              aria-label="Filter sumber"
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="">Semua Sumber</option>
              {CANTUS_SUMBER_SUGGESTIONS.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>

            <div>
              <input
                type="text"
                list="cantus-kategori-suggestions"
                value={filters.kategori}
                onChange={(e) => updateParam("kategori", e.target.value)}
                placeholder="Kategori (mis. Pembuka)"
                aria-label="Filter kategori"
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted"
              />
              <datalist id="cantus-kategori-suggestions">
                {CANTUS_KATEGORI_SUGGESTIONS.map((k) => (
                  <option key={k} value={k} />
                ))}
              </datalist>
            </div>

            <input
              type="text"
              value={filters.tag}
              onChange={(e) => updateParam("tag", e.target.value)}
              placeholder="Filter tag (mis. Adven, Tahun A)"
              aria-label="Filter tag"
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted"
            />

            <input
              type="text"
              value={filters.nadaDasar}
              onChange={(e) => updateParam("nadaDasar", e.target.value)}
              placeholder="Nada dasar (mis. D, C)"
              aria-label="Filter nada dasar"
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg placeholder:text-muted"
            />

            <select
              value={filters.active}
              onChange={(e) =>
                updateParam(
                  "active",
                  e.target.value === "active"
                    ? ""
                    : e.target.value === "inactive"
                      ? "false"
                      : "all",
                )
              }
              aria-label="Filter status aktif"
              className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
            >
              <option value="active">Aktif saja</option>
              <option value="all">Semua status</option>
              <option value="inactive">Nonaktif</option>
            </select>
          </div>
        </div>
      </InductiveFilterBar>

      {searchQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">Memuat katalog nyanyian…</Card>
      ) : rows.length === 0 ? (
        <InductiveEmptyState
          icon={Music}
          title={
            hasFilters
              ? "Tidak ada nyanyian yang cocok"
              : "Katalog nyanyian masih kosong"
          }
          description={
            hasFilters
              ? "Coba ubah kata kunci pencarian atau reset filter untuk melihat nyanyian lainnya."
              : "Belum ada entri lagu, mazmur, atau BPI yang terdaftar di katalog global."
          }
          action={
            hasFilters ? (
              <Button variant="outline" onClick={resetFilters}>
                Reset Filter
              </Button>
            ) : canManage ? (
              <Link to="/cantus/new" className="no-underline">
                <Button>
                  <Plus aria-hidden className="size-4" />
                  Tambah Nyanyian
                </Button>
              </Link>
            ) : undefined
          }
        />
      ) : (
        <Card className="divide-y divide-line">
          {rows.map((item) => {
            const mazmurMeta = parseMazmurMeta(
              item.meta as Record<string, unknown> | undefined,
            );
            const previewText =
              mazmurMeta.ulangan?.trim() || item.lirik?.trim() || "";
            const codeBadge = item.nomor?.trim() || item.sumber?.trim();
            const tags = (item.tag ?? []).slice(0, 4);

            return (
              <Link
                key={item.id}
                to={`/cantus/${item.id}`}
                className="flex items-center justify-between gap-3 p-4 no-underline transition-colors hover:bg-surface-2/50 sm:px-6"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-1.5">
                    {codeBadge && (
                      <span className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-xs font-semibold text-fg">
                        {codeBadge}
                      </span>
                    )}
                    <span className="truncate text-sm font-bold text-fg">
                      {item.judul}
                    </span>
                    <Badge tone="neutral">{cantusJenisLabel(item.jenis)}</Badge>
                    {item.kategori && (
                      <Badge tone="neutral">{item.kategori}</Badge>
                    )}
                    {item.nadaDasar && (
                      <Badge tone="neutral">{item.nadaDasar}</Badge>
                    )}
                    {item.active === false && (
                      <Badge tone="danger">Nonaktif</Badge>
                    )}
                  </div>

                  {(tags.length > 0 || previewText) && (
                    <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
                      {tags.map((tg) => (
                        <span
                          key={tg}
                          className="rounded-full bg-surface-2 px-2 py-0.5 text-[11px] font-medium text-muted"
                        >
                          #{tg}
                        </span>
                      ))}
                      {previewText && (
                        <p className="line-clamp-1 text-xs text-muted">
                          {previewText}
                        </p>
                      )}
                    </div>
                  )}
                </div>

                <div className="flex shrink-0 items-center gap-1.5 text-muted">
                  <span className="hidden text-xs font-semibold sm:inline-flex">
                    Lihat Detail
                  </span>
                  <ChevronRight aria-hidden className="size-4 shrink-0" />
                </div>
              </Link>
            );
          })}
        </Card>
      )}

      {searchQuery.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => searchQuery.fetchNextPage()}
            disabled={searchQuery.isFetchingNextPage}
          >
            {searchQuery.isFetchingNextPage ? "Memuat…" : "Muat lebih banyak"}
          </Button>
        </div>
      )}
    </div>
  );
}
