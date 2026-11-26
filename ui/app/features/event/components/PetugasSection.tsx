import { Check, Plus, Trash2, Users, X } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  useCreateEventPetugas,
  useDeleteEventPetugas,
  useUpdateEventPetugas,
} from "../api/actions";
import { resolveEventError } from "../api/errors";
import {
  useEventOrgs,
  useEventRoleLookups,
  useOrgMembersForEvent,
  useParokiMembersForEvent,
} from "../api/queries";
import type { EventPetugasResponse } from "../api/types";
import { DEFAULT_EVENT_ROLES, petugasStatusMeta } from "../model";

type AssigneeMode = "manual" | "org" | "user";

interface PetugasSectionProps {
  eventId: string;
  parokiId: string;
  orgId?: string;
  canManage: boolean;
  petugas: EventPetugasResponse[];
}

export function PetugasSection({
  eventId,
  parokiId,
  orgId,
  canManage,
  petugas,
}: PetugasSectionProps) {
  const [showAdd, setShowAdd] = useState(false);
  const [mode, setMode] = useState<AssigneeMode>("manual");
  const [role, setRole] = useState("");
  const [manualName, setManualName] = useState("");
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [selectedUserId, setSelectedUserId] = useState("");
  const [urutan, setUrutan] = useState("1");
  const [note, setNote] = useState("");
  const [error, setError] = useState<string | null>(null);

  const roleLookupsQuery = useEventRoleLookups(canManage);
  const orgsQuery = useEventOrgs(canManage);
  const orgMembersQuery = useOrgMembersForEvent(
    orgId ?? "",
    canManage && Boolean(orgId),
  );
  const parokiMembersQuery = useParokiMembersForEvent(
    parokiId,
    canManage && !orgId,
  );

  const createPetugas = useCreateEventPetugas(eventId);
  const updatePetugas = useUpdateEventPetugas(eventId);
  const deletePetugas = useDeleteEventPetugas(eventId);

  const roleSuggestions = Array.from(
    new Set([
      ...(roleLookupsQuery.data ?? [])
        .map((r) => r.label?.trim())
        .filter((v): v is string => Boolean(v)),
      ...DEFAULT_EVENT_ROLES,
    ]),
  );

  const orgOptions = orgsQuery.data ?? [];
  const memberOptions = orgId
    ? (orgMembersQuery.data ?? [])
    : (parokiMembersQuery.data ?? []);

  async function handleAdd(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const parsedUrutan = Number.parseInt(urutan, 10);
    try {
      await createPetugas.mutateAsync({
        role: role.trim(),
        userId: mode === "user" && selectedUserId ? selectedUserId : undefined,
        orgId: mode === "org" && selectedOrgId ? selectedOrgId : undefined,
        name: mode === "manual" ? manualName.trim() : undefined,
        urutan: Number.isNaN(parsedUrutan) ? 0 : parsedUrutan,
        note: note.trim() || undefined,
      });
      setRole("");
      setManualName("");
      setSelectedOrgId("");
      setSelectedUserId("");
      setUrutan(String(petugas.length + 2));
      setNote("");
      setShowAdd(false);
    } catch (err) {
      setError(resolveEventError(err));
    }
  }

  async function handleConfirm(
    petugasId: string,
    status: "confirmed" | "declined",
  ) {
    setError(null);
    try {
      await updatePetugas.mutateAsync({
        petugasId,
        body: { status },
      });
    } catch (err) {
      setError(resolveEventError(err));
    }
  }

  async function handleDelete(petugasId: string) {
    if (!window.confirm("Hapus penugasan petugas ini?")) return;
    setError(null);
    try {
      await deletePetugas.mutateAsync(petugasId);
    } catch (err) {
      setError(resolveEventError(err));
    }
  }

  return (
    <Card className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="flex items-center gap-2 text-lg font-semibold text-ink">
          <Users className="h-5 w-5 text-accent" />
          Petugas Pelayanan
        </h2>
        {canManage && !showAdd && (
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              setShowAdd(true);
              setError(null);
              setUrutan(String(petugas.length + 1));
            }}
          >
            <Plus className="h-4 w-4" />
            Tambah Petugas
          </Button>
        )}
      </div>

      {error && (
        <div className="rounded-lg border border-danger-fg/30 bg-danger-bg px-3 py-2 text-sm text-danger-fg">
          {error}
        </div>
      )}

      {showAdd && canManage && (
        <form
          onSubmit={handleAdd}
          className="space-y-4 rounded-lg border border-line bg-surface-muted/40 p-4"
        >
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="block text-sm font-medium text-ink">
              Peran Petugas
              <input
                type="text"
                list="event-role-suggestions"
                value={role}
                onChange={(e) => setRole(e.target.value)}
                placeholder="Contoh: Organis, Lektor 1, Selebran"
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                required
              />
              <datalist id="event-role-suggestions">
                {roleSuggestions.map((item) => (
                  <option key={item} value={item} />
                ))}
              </datalist>
            </label>

            <label className="block text-sm font-medium text-ink">
              Tipe Penugasan
              <select
                value={mode}
                onChange={(e) => setMode(e.target.value as AssigneeMode)}
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              >
                <option value="manual">Nama Manual</option>
                <option value="org">Organisasi / Lingkungan</option>
                <option value="user">Anggota Organisasi / Paroki</option>
              </select>
            </label>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            {mode === "manual" && (
              <label className="block text-sm font-medium text-ink">
                Nama Petugas
                <input
                  type="text"
                  value={manualName}
                  onChange={(e) => setManualName(e.target.value)}
                  placeholder="Contoh: RD. Yohanes"
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                />
              </label>
            )}

            {mode === "org" && (
              <label className="block text-sm font-medium text-ink">
                Pilih Organisasi / Lingkungan
                <select
                  value={selectedOrgId}
                  onChange={(e) => setSelectedOrgId(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                >
                  <option value="">-- Pilih Organisasi --</option>
                  {orgOptions.map((o) => (
                    <option key={o.id} value={o.id}>
                      {o.name} ({o.code})
                    </option>
                  ))}
                </select>
              </label>
            )}

            {mode === "user" && (
              <label className="block text-sm font-medium text-ink">
                Pilih Anggota
                <select
                  value={selectedUserId}
                  onChange={(e) => setSelectedUserId(e.target.value)}
                  className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
                  required
                >
                  <option value="">-- Pilih Anggota --</option>
                  {memberOptions.map((m) => (
                    <option key={m.userId} value={m.userId}>
                      {m.name} ({m.email})
                    </option>
                  ))}
                </select>
              </label>
            )}

            <label className="block text-sm font-medium text-ink">
              Urutan
              <input
                type="number"
                value={urutan}
                onChange={(e) => setUrutan(e.target.value)}
                className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
              />
            </label>
          </div>

          <label className="block text-sm font-medium text-ink">
            Catatan (opsional)
            <input
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Catatan tambahan untuk petugas"
              className="mt-1 w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink"
            />
          </label>

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowAdd(false)}
            >
              Batal
            </Button>
            <Button type="submit" disabled={createPetugas.isPending}>
              {createPetugas.isPending ? "Menyimpan…" : "Simpan Petugas"}
            </Button>
          </div>
        </form>
      )}

      {petugas.length === 0 ? (
        <p className="text-sm text-ink-muted">
          Belum ada petugas pelayanan yang ditugaskan untuk kegiatan ini.
        </p>
      ) : (
        <div className="divide-y divide-line">
          {petugas.map((item) => {
            const st = petugasStatusMeta(item.status ?? "assigned");
            return (
              <div
                key={item.id}
                className="flex flex-wrap items-center justify-between gap-3 py-3"
              >
                <div className="space-y-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-ink">
                      {item.role}
                    </span>
                    <span className="text-sm text-ink-muted">·</span>
                    <span className="text-sm font-medium text-ink">
                      {item.name}
                    </span>
                    <Badge tone={st.tone}>{st.label}</Badge>
                  </div>
                  {item.note && (
                    <p className="text-xs text-ink-muted">{item.note}</p>
                  )}
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {item.canConfirm && (
                    <div
                      role="group"
                      className="flex items-center gap-1.5"
                      aria-label="Konfirmasi Tugas"
                    >
                      <Button
                        type="button"
                        variant={
                          item.status === "confirmed" ? "primary" : "outline"
                        }
                        disabled={updatePetugas.isPending}
                        onClick={() =>
                          handleConfirm(item.id ?? "", "confirmed")
                        }
                      >
                        <Check className="h-3.5 w-3.5" />
                        Siap Bertugas
                      </Button>
                      <Button
                        type="button"
                        variant={
                          item.status === "declined" ? "danger" : "outline"
                        }
                        disabled={updatePetugas.isPending}
                        onClick={() => handleConfirm(item.id ?? "", "declined")}
                      >
                        <X className="h-3.5 w-3.5" />
                        Berhalangan
                      </Button>
                    </div>
                  )}
                  {canManage && (
                    <Button
                      type="button"
                      variant="danger"
                      disabled={deletePetugas.isPending}
                      onClick={() => handleDelete(item.id ?? "")}
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Hapus
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </Card>
  );
}
