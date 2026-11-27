import { Building2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import {
  InductiveEmptyState,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useCreateOrg } from "../api/actions";
import { resolveOrgError } from "../api/errors";
import { useOrgLingkungan, useOrgs, useOrgWilayah } from "../api/queries";
import { isParokiAdmin, ORG_KINDS, type OrgKind, slugifyCode } from "../model";
import type { Route } from "./+types/OrgCreatePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.createOrgTitle")} · Domus` }];
}

export default function OrgCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me, isLoading: meLoading } = useMe();
  const hasParoki = Boolean(me?.activeParokiId);
  const parokiAdmin = isParokiAdmin(me?.roles);

  const [kind, setKind] = useState<OrgKind>("section");
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [codeEdited, setCodeEdited] = useState(false);
  const [description, setDescription] = useState("");
  const [parentId, setParentId] = useState("");
  const [wilayahId, setWilayahId] = useState("");
  const [lingkunganId, setLingkunganId] = useState("");

  const [selfAssign, setSelfAssign] = useState(false);
  const [initialAdminEmail, setInitialAdminEmail] = useState("");
  const [initialAdminName, setInitialAdminName] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const allOrgsForParent = useOrgs({ limit: 100 }, hasParoki && parokiAdmin);
  const wilayahQuery = useOrgWilayah(
    hasParoki && parokiAdmin && kind === "wilayah",
  );
  const lingkunganQuery = useOrgLingkungan(
    hasParoki && parokiAdmin && kind === "lingkungan",
  );
  const createOrg = useCreateOrg();

  const parentOptions =
    allOrgsForParent.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  function handleNameChange(val: string) {
    setName(val);
    if (!codeEdited) {
      setCode(slugifyCode(val));
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(null);
    if (!name.trim() || !code.trim()) {
      setFormError("Nama dan kode organisasi wajib diisi.");
      return;
    }

    try {
      await createOrg.mutateAsync({
        kind,
        code: code.trim(),
        name: name.trim(),
        description: description.trim() || undefined,
        parentId: parentId || undefined,
        wilayahId: kind === "wilayah" && wilayahId ? wilayahId : undefined,
        lingkunganId:
          kind === "lingkungan" && lingkunganId ? lingkunganId : undefined,
        initialAdminUserId: selfAssign && me?.id ? me.id : undefined,
        initialAdminEmail:
          !selfAssign && initialAdminEmail.trim()
            ? initialAdminEmail.trim()
            : undefined,
        initialAdminName:
          !selfAssign && initialAdminName.trim()
            ? initialAdminName.trim()
            : undefined,
      });
      navigate("/org", { replace: true });
    } catch (err) {
      setFormError(resolveOrgError(err));
    }
  }

  if (!meLoading && !hasParoki) {
    return (
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={Building2}
          title={t("org.noParishTitle")}
          description={t("org.noParishDescription")}
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Organisasi", to: "/org" },
          { label: t("org.addOrg") },
        ]}
        title={t("org.createOrgTitle")}
        subtitle="Tambahkan organisasi, seksi, wilayah, atau komunitas baru di paroki."
      />

      {!meLoading && !parokiAdmin ? (
        <InductiveEmptyState
          icon={Building2}
          title="Akses Terbatas"
          description="Hanya Admin Paroki yang dapat menambahkan organisasi baru."
        />
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          {formError && (
            <div
              role="alert"
              className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm font-medium text-danger-fg"
            >
              {formError}
            </div>
          )}

          <InductiveFormSection
            title="Informasi Organisasi"
            description="Lengkapi identitas, jenis, serta induk organisasi bila ada."
          >
            <div className="grid gap-4 sm:grid-cols-2">
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
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 font-mono text-sm text-fg"
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
                    value={wilayahId}
                    onChange={(e) => setWilayahId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                  >
                    <option value="">Pilih wilayah</option>
                    {(wilayahQuery.data ?? []).map((z) => (
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
                    value={lingkunganId}
                    onChange={(e) => setLingkunganId(e.target.value)}
                    className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                  >
                    <option value="">Pilih lingkungan</option>
                    {(lingkunganQuery.data ?? []).map((w) => (
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
            </div>
          </InductiveFormSection>

          <InductiveFormSection
            title="Admin Organisasi Awal (Opsional)"
            description="Anda dapat langsung menunjuk Admin Organisasi saat pembuatan, atau menambahkannya nanti melalui menu Pengguna."
          >
            <div className="grid gap-4 sm:grid-cols-2">
              <label className="flex min-h-11 cursor-pointer items-center gap-2.5 text-sm font-medium text-fg sm:col-span-2">
                <input
                  type="checkbox"
                  checked={selfAssign}
                  onChange={(e) => setSelfAssign(e.target.checked)}
                  className="size-4 accent-primary"
                />
                Jadikan saya Admin Organisasi ini
              </label>

              {!selfAssign && (
                <>
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    Email Admin Awal (opsional)
                    <input
                      type="email"
                      value={initialAdminEmail}
                      onChange={(e) => setInitialAdminEmail(e.target.value)}
                      placeholder="admin@paroki.id"
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                    />
                  </label>
                  <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                    Nama Admin Awal (opsional)
                    <input
                      type="text"
                      value={initialAdminName}
                      onChange={(e) => setInitialAdminName(e.target.value)}
                      placeholder="Nama lengkap admin"
                      className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                    />
                  </label>
                </>
              )}
            </div>
          </InductiveFormSection>

          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Button
              variant="outline"
              onClick={() => navigate("/org")}
              className="w-full sm:w-auto"
            >
              Batal
            </Button>
            <Button
              type="submit"
              disabled={createOrg.isPending}
              className="w-full sm:w-auto"
            >
              {createOrg.isPending ? "Menyimpan…" : "Simpan Organisasi"}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
