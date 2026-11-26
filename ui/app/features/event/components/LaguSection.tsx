import { Music, Pencil, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useCreateEventLagu,
  useDeleteEventLagu,
  useUpdateEventLagu,
} from "../api/actions";
import { resolveEventError } from "../api/errors";
import { useEventSongPartLookups } from "../api/queries";
import type { EventLaguResponse } from "../api/types";
import { DEFAULT_EVENT_SONG_PARTS } from "../model";

interface LaguSectionProps {
  eventId: string;
  canManage: boolean;
  lagu: EventLaguResponse[];
}

export function LaguSection({ eventId, canManage, lagu }: LaguSectionProps) {
  const [showAdd, setShowAdd] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [urutan, setUrutan] = useState("1");
  const [bagian, setBagian] = useState("");
  const [sumber, setSumber] = useState("");
  const [nomor, setNomor] = useState("");
  const [judul, setJudul] = useState("");
  const [nadaDasar, setNadaDasar] = useState("");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);

  const songPartLookupsQuery = useEventSongPartLookups(canManage);
  const createLagu = useCreateEventLagu(eventId);
  const updateLagu = useUpdateEventLagu(eventId);
  const deleteLagu = useDeleteEventLagu(eventId);

  const partSuggestions = Array.from(
    new Set([
      ...(songPartLookupsQuery.data ?? [])
        .map((r) => r.label?.trim())
        .filter((v): v is string => Boolean(v)),
      ...DEFAULT_EVENT_SONG_PARTS,
    ]),
  );

  function resetForm(nextUrutan = lagu.length + 1) {
    setUrutan(String(nextUrutan));
    setBagian("");
    setSumber("");
    setNomor("");
    setJudul("");
    setNadaDasar("");
    setNote("");
  }

  function startEdit(item: EventLaguResponse) {
    setShowAdd(false);
    setEditingId(item.id ?? null);
    setUrutan(String(item.urutan ?? 1));
    setBagian(item.bagian ?? "");
    setSumber(item.sumber ?? "");
    setNomor(item.nomor ?? "");
    setJudul(item.judul ?? "");
    setNadaDasar(item.nadaDasar ?? "");
    setNote(item.note ?? "");
    setError(null);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const parsedUrutan = Number.parseInt(urutan, 10);
    const numUrutan = Number.isNaN(parsedUrutan) ? 0 : parsedUrutan;

    try {
      if (editingId) {
        await updateLagu.mutateAsync({
          laguId: editingId,
          body: {
            urutan: numUrutan,
            bagian: bagian.trim(),
            sumber: sumber.trim(),
            nomor: nomor.trim(),
            judul: judul.trim(),
            nadaDasar: nadaDasar.trim(),
            note: note.trim(),
          },
        });
        setEditingId(null);
      } else {
        await createLagu.mutateAsync({
          urutan: numUrutan,
          bagian: bagian.trim(),
          sumber: sumber.trim() || undefined,
          nomor: nomor.trim() || undefined,
          judul: judul.trim(),
          nadaDasar: nadaDasar.trim() || undefined,
          note: note.trim() || undefined,
        });
        setShowAdd(false);
      }
      resetForm(lagu.length + (editingId ? 1 : 2));
    } catch (err) {
      setError(resolveEventError(err));
    }
  }

  async function handleDelete(laguId: string) {
    if (!window.confirm("Hapus lagu ini dari tata perayaan?")) return;
    setError(null);
    try {
      await deleteLagu.mutateAsync(laguId);
    } catch (err) {
      setError(resolveEventError(err));
    }
  }

  const formOpen = showAdd || Boolean(editingId);

  return (
    <Card className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="flex items-center gap-2 text-lg font-semibold text-ink">
          <Music className="h-5 w-5 text-accent" />
          Daftar Lagu / Tata Perayaan
        </h2>
        {canManage && !formOpen && (
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              resetForm(lagu.length + 1);
              setEditingId(null);
              setShowAdd(true);
              setError(null);
            }}
          >
            <Plus className="h-4 w-4" />
            Tambah Lagu
          </Button>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-danger-fg/30 bg-danger-bg px-3 py-2 text-sm text-danger-fg">
          {error}
        </div>
      )}

      {formOpen && canManage && (
        <form
          onSubmit={handleSubmit}
          className="space-y-4 rounded-lg border border-line bg-surface-muted/40 p-4"
        >
          <div className="grid gap-4 sm:grid-cols-3">
            <label className="block text-sm font-medium text-ink">
              Bagian Lagu
              <input
                type="text"
                list="event-song-part-suggestions"
                value={bagian}
                onChange={(e) => setBagian(e.target.value)}
                placeholder="Contoh: Pembuka, Kemuliaan"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                required
              />
              <datalist id="event-song-part-suggestions">
                {partSuggestions.map((item) => (
                  <option key={item} value={item} />
                ))}
              </datalist>
            </label>

            <label className="block text-sm font-medium text-ink">
              Judul Lagu
              <input
                type="text"
                value={judul}
                onChange={(e) => setJudul(e.target.value)}
                placeholder="Contoh: Mari Kita Bersukacita"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                required
              />
            </label>

            <label className="block text-sm font-medium text-ink">
              Urutan Lagu
              <input
                type="number"
                value={urutan}
                onChange={(e) => setUrutan(e.target.value)}
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>
          </div>

          <div className="grid gap-4 sm:grid-cols-3">
            <label className="block text-sm font-medium text-ink">
              Sumber (opsional)
              <input
                type="text"
                value={sumber}
                onChange={(e) => setSumber(e.target.value)}
                placeholder="Puji Syukur / Madah Bakti"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>
            <label className="block text-sm font-medium text-ink">
              Nomor (opsional)
              <input
                type="text"
                value={nomor}
                onChange={(e) => setNomor(e.target.value)}
                placeholder="Contoh: 319"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>
            <label className="block text-sm font-medium text-ink">
              Nada Dasar (opsional)
              <input
                type="text"
                value={nadaDasar}
                onChange={(e) => setNadaDasar(e.target.value)}
                placeholder="Contoh: G, Do=F"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>
          </div>

          <label className="block text-sm font-medium text-ink">
            Catatan Lagu (opsional)
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Contoh: Bait 1-3, ulangan 2x"
              className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
            />
          </label>

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setShowAdd(false);
                setEditingId(null);
              }}
            >
              Batal
            </Button>
            <Button
              type="submit"
              disabled={createLagu.isPending || updateLagu.isPending}
            >
              {createLagu.isPending || updateLagu.isPending
                ? "Menyimpan…"
                : editingId
                  ? "Simpan Lagu"
                  : "Simpan Lagu"}
            </Button>
          </div>
        </form>
      )}

      {lagu.length === 0 ? (
        <p className="text-sm text-ink-muted">
          Belum ada daftar lagu yang ditambahkan untuk kegiatan ini.
        </p>
      ) : (
        <div className="divide-y divide-line">
          {lagu.map((item) => (
            <div
              key={item.id}
              className="flex flex-wrap items-center justify-between gap-3 py-3"
            >
              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-xs font-semibold text-ink-muted">
                    #{item.urutan}
                  </span>
                  <Badge tone="neutral">{item.bagian}</Badge>
                  <span className="text-sm font-semibold text-ink">
                    {item.judul}
                  </span>
                  {(item.sumber || item.nomor) && (
                    <span className="text-xs text-ink-muted">
                      ({[item.sumber, item.nomor].filter(Boolean).join(" ")})
                    </span>
                  )}
                  {item.nadaDasar && (
                    <Badge tone="success">{item.nadaDasar}</Badge>
                  )}
                </div>
                {item.note && (
                  <p className="text-xs text-ink-muted">{item.note}</p>
                )}
              </div>

              {canManage && (
                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => startEdit(item)}
                  >
                    <Pencil className="h-3.5 w-3.5" />
                    Ubah
                  </Button>
                  <Button
                    type="button"
                    variant="danger"
                    disabled={deleteLagu.isPending}
                    onClick={() => handleDelete(item.id ?? "")}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                    Hapus
                  </Button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}
