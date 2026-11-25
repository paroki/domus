import { Building2, Plus, Search } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useCreateOrg } from "../api/actions";
import { resolveOrgError } from "../api/errors";
import { useOrgs, useOrgWards, useOrgZones } from "../api/queries";
import {
  isParishAdmin,
  kindLabel,
  ORG_KINDS,
  type OrgKind,
  slugifyCode,
} from "../model";
import type { Route } from "./+types/OrgsPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.title")} · Domus` }];
}

export default function OrgsPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);
  const parishAdmin = isParishAdmin(me?.roles);

  const [myOrgsOnly, setMyOrgsOnly] = useState(false);
  const [kindFilter, setKindFilter] = useState("");
  const [q, setQ] = useState("");

  const [showCreate, setShowCreate] = useState(false);
  const [kind, setKind] = useState<OrgKind>("section");
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [codeEdited, setCodeEdited] = useState(false);
  const [description, setDescription] = useState("");
  const [parentId, setParentId] = useState("");
  const [zoneId, setZoneId] = useState("");
  const [wardId, setWardId] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const orgsQuery = useOrgs(
    {
      q: q.trim() || undefined,
      filter: {
        kind: (kindFilter as OrgKind) || undefined,
        myOrgsOnly: myOrgsOnly || undefined,
      },
    },
    hasParish,
  );
  const allOrgsForParent = useOrgs({ limit: 100 }, hasParish && showCreate);
  const zonesQuery = useOrgZones(hasParish && showCreate && kind === "wilayah");
  const wardsQuery = useOrgWards(
    hasParish && showCreate && kind === "lingkungan",
  );

  const createOrg = useCreateOrg();

  const items = orgsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const parentOptions =
    allOrgsForParent.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  function handleNameChange(val: string) {
    setName(val);
    if (!codeEdited) {
      setCode(slugifyCode(val));
    }
  }

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    try {
      await createOrg.mutateAsync({
        kind,
        code: code.trim(),
        name: name.trim(),
        description: description.trim() || undefined,
        parentId: parentId || undefined,
        wilayahId: kind === "wilayah" && zoneId ? zoneId : undefined,
        lingkunganId: kind === "lingkungan" && wardId ? wardId : undefined,
      });
      setName("");
      setCode("");
      setCodeEdited(false);
      setDescription("");
      setParentId("");
      setZoneId("");
      setWardId("");
      setShowCreate(false);
    } catch (err) {
      setFormError(resolveOrgError(err));
    }
  }

  if (!hasParish) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-8">
        <Card className="p-6 text-center">
          <Building2 aria-hidden className="mx-auto mb-3 size-8 text-muted" />
          <h1 className="text-lg font-semibold text-fg">
            {t("org.noParishTitle")}
          </h1>
          <p className="mt-1 text-sm text-muted">
            {t("org.noParishDescription")}
          </p>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-fg">{t("org.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("org.subtitle")}</p>
        </div>
        {parishAdmin && (
          <Button onClick={() => setShowCreate((v) => !v)}>
            <Plus aria-hidden className="size-4" />
            {t("org.addOrg")}
          </Button>
        )}
      </div>

      {parishAdmin && showCreate && (
        <Card className="p-5">
          <h2 className="mb-4 text-base font-semibold text-fg">
            {t("org.createOrgTitle")}
          </h2>
          {formError && (
            <div
              role="alert"
              className="mb-4 rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
            >
              {formError}
            </div>
          )}
          <form onSubmit={handleCreate} className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Jenis Organisasi
              <select
                value={kind}
                onChange={(e) => setKind(e.target.value as OrgKind)}
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                {ORG_KINDS.map((k) => (
                  <option key={k.value} value={k.value}>
                    {k.label}
                  </option>
                ))}
              </select>
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Nama Organisasi
              <input
                type="text"
                required
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="mis. Seksi Liturgi"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Kode
              <input
                type="text"
                required
                value={code}
                onChange={(e) => {
                  setCodeEdited(true);
                  setCode(e.target.value);
                }}
                placeholder="mis. sek-liturgi"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              Induk Organisasi (opsional)
              <select
                value={parentId}
                onChange={(e) => setParentId(e.target.value)}
                className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                <option value="">Tanpa induk</option>
                {parentOptions.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name} ({o.code})
                  </option>
                ))}
              </select>
            </label>

            {kind === "wilayah" && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Wilayah / Stasi (opsional)
                <select
                  value={zoneId}
                  onChange={(e) => setZoneId(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="">Pilih wilayah</option>
                  {(zonesQuery.data ?? []).map((z) => (
                    <option key={z.id} value={z.id}>
                      {z.nama} ({z.kode})
                    </option>
                  ))}
                </select>
              </label>
            )}

            {kind === "lingkungan" && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Lingkungan (opsional)
                <select
                  value={wardId}
                  onChange={(e) => setWardId(e.target.value)}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="">Pilih lingkungan</option>
                  {(wardsQuery.data ?? []).map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.nama} ({w.kode})
                    </option>
                  ))}
                </select>
              </label>
            )}

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg sm:col-span-2">
              Deskripsi (opsional)
              <input
                type="text"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Keterangan singkat organisasi"
                className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
              />
            </label>

            <div className="flex justify-end gap-2 sm:col-span-2">
              <Button variant="outline" onClick={() => setShowCreate(false)}>
                Batal
              </Button>
              <Button type="submit" disabled={createOrg.isPending}>
                {createOrg.isPending ? "Menyimpan…" : "Simpan Organisasi"}
              </Button>
            </div>
          </form>
        </Card>
      )}

      <Card className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2">
          <Button
            variant={myOrgsOnly ? "outline" : "primary"}
            onClick={() => setMyOrgsOnly(false)}
          >
            {t("org.allOrgs")}
          </Button>
          <Button
            variant={myOrgsOnly ? "primary" : "outline"}
            onClick={() => setMyOrgsOnly(true)}
          >
            {t("org.myOrgs")}
          </Button>
        </div>

        <div className="flex flex-1 flex-wrap items-center justify-end gap-2">
          <select
            aria-label={t("org.allKinds")}
            value={kindFilter}
            onChange={(e) => setKindFilter(e.target.value)}
            className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
          >
            <option value="">{t("org.allKinds")}</option>
            {ORG_KINDS.map((k) => (
              <option key={k.value} value={k.value}>
                {k.label}
              </option>
            ))}
          </select>

          <div className="relative min-w-[220px] flex-1 sm:max-w-xs">
            <Search
              aria-hidden
              className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted"
            />
            <input
              type="search"
              value={q}
              onChange={(e) => setQ(e.target.value)}
              placeholder={t("org.searchPlaceholder")}
              className="h-11 w-full rounded-lg border border-line bg-canvas pl-9 pr-3 text-sm text-fg placeholder:text-muted"
            />
          </div>
        </div>
      </Card>

      {orgsQuery.isLoading ? (
        <p className="text-sm text-muted">Memuat organisasi…</p>
      ) : items.length === 0 ? (
        <Card className="p-6 text-center text-sm text-muted">
          Belum ada organisasi yang sesuai dengan filter.
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {items.map((org) => (
            <Card
              key={org.id}
              className="flex flex-col justify-between gap-4 p-5"
            >
              <div className="flex flex-col gap-2">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <Link
                    to={`/org/${org.id}`}
                    className="text-base font-semibold text-fg hover:text-primary"
                  >
                    {org.name}
                  </Link>
                  <Badge tone={org.active ? "success" : "neutral"}>
                    {org.active ? "Aktif" : "Nonaktif"}
                  </Badge>
                </div>

                <div className="flex flex-wrap items-center gap-2 text-xs text-muted">
                  <span className="font-mono">{org.code}</span>
                  <Badge tone="neutral">{kindLabel(org.kind ?? "")}</Badge>
                  {org.myRole === "admin" && (
                    <Badge tone="warning">{t("org.orgAdminBadge")}</Badge>
                  )}
                  {org.myRole === "member" && (
                    <Badge tone="success">{t("org.orgMemberBadge")}</Badge>
                  )}
                </div>

                {org.description && (
                  <p className="text-sm text-muted">{org.description}</p>
                )}
              </div>

              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-line pt-3">
                <Link
                  to={`/org/${org.id}`}
                  className="text-sm font-semibold text-primary hover:underline"
                >
                  Buka Detail →
                </Link>
              </div>
            </Card>
          ))}
        </div>
      )}

      {orgsQuery.hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            disabled={orgsQuery.isFetchingNextPage}
            onClick={() => orgsQuery.fetchNextPage()}
          >
            {orgsQuery.isFetchingNextPage ? "Memuat…" : "Muat Lebih Banyak"}
          </Button>
        </div>
      )}
    </div>
  );
}
