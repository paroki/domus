import { Link2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import {
  InductiveEmptyState,
  InductiveFormSection,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useCreateOrgInvite } from "../api/actions";
import { resolveOrgError } from "../api/errors";
import { useOrg } from "../api/queries";
import { INVITE_TTL_OPTIONS } from "../model";
import type { Route } from "./+types/OrgInviteCreatePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.createInvite")} · Domus` }];
}

export default function OrgInviteCreatePage() {
  const { t } = useTranslation();
  const { orgId = "" } = useParams();
  const navigate = useNavigate();

  const orgQuery = useOrg(orgId);
  const org = orgQuery.data;
  const isOrgAdmin = org?.myRole === "admin";

  const createInvite = useCreateOrgInvite(orgId);

  const [role, setRole] = useState<"member" | "admin">("member");
  const [ttlSeconds, setTtlSeconds] = useState(86400);
  const [maxUses, setMaxUses] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const parsedMax = maxUses.trim()
        ? Number.parseInt(maxUses.trim(), 10)
        : undefined;
      await createInvite.mutateAsync({
        role,
        ttlSeconds,
        maxUses: parsedMax && parsedMax > 0 ? parsedMax : undefined,
      });
      navigate(`/org/${orgId}#invites`, { replace: true });
    } catch (err) {
      setErrorMsg(resolveOrgError(err));
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Organisasi", to: "/org" },
          { label: org?.name ?? "Detail Organisasi", to: `/org/${orgId}` },
          { label: t("org.createInvite") },
        ]}
        title={t("org.createInvite")}
        subtitle="Bagikan tautan undangan agar umat atau pengurus dapat bergabung ke organisasi ini."
      />

      {orgQuery.isLoading ? (
        <p className="text-sm text-muted">Memuat detail organisasi…</p>
      ) : !isOrgAdmin ? (
        <InductiveEmptyState
          icon={Link2}
          title="Akses Terbatas"
          description="Hanya Admin Organisasi yang dapat membuat tautan undangan."
        />
      ) : (
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          {errorMsg && (
            <div
              role="alert"
              className="rounded-lg border border-danger-fg/30 bg-danger-bg px-4 py-3 text-sm font-medium text-danger-fg"
            >
              {errorMsg}
            </div>
          )}

          <InductiveFormSection
            title="Pengaturan Tautan Undangan"
            description="Tentukan peran bawaan, masa berlaku, dan batas pemakaian tautan."
          >
            <div className="grid gap-4 sm:grid-cols-3">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Peran Anggota
                <select
                  value={role}
                  onChange={(e) =>
                    setRole(e.target.value as "member" | "admin")
                  }
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  <option value="member">Anggota (member)</option>
                  <option value="admin">Admin Organisasi (admin)</option>
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Masa Berlaku
                <select
                  value={ttlSeconds}
                  onChange={(e) => setTtlSeconds(Number(e.target.value))}
                  className="h-11 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
                >
                  {INVITE_TTL_OPTIONS.map((opt) => (
                    <option key={opt.seconds} value={opt.seconds}>
                      Berlaku {opt.label}
                    </option>
                  ))}
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                Maks. Pemakaian (opsional)
                <input
                  type="number"
                  min={1}
                  value={maxUses}
                  onChange={(e) => setMaxUses(e.target.value)}
                  placeholder="Maks. pakai (opsional)"
                  className="h-11 rounded-lg border border-line bg-canvas px-3.5 text-sm text-fg"
                />
              </label>
            </div>
          </InductiveFormSection>

          <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <Button
              variant="outline"
              onClick={() => navigate(`/org/${orgId}#invites`)}
              className="w-full sm:w-auto"
            >
              Batal
            </Button>
            <Button
              type="submit"
              disabled={createInvite.isPending}
              className="w-full sm:w-auto"
            >
              {t("org.createInvite")}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
