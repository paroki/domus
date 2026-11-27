import { ArrowLeft, Building2, Church, Pencil, UserCheck } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveEmptyState, InductivePageHeader } from "~/shared/ui/iui";
import { useUpdateLiturgiMinistry } from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import {
  useLiturgiMinistries,
  useLiturgiOrgs,
  useLiturgiParokiMembers,
} from "../api/queries";
import type { LiturgiMinistry } from "../api/types";
import {
  canManageLiturgiGlobal,
  getLiturgiMinistryMeta,
  LITURGI_MINISTRIES,
  type LiturgiRoleSlug,
} from "../model";
import type { Route } from "./+types/LiturgiMinistriesPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.ministriesTitle")} · Domus` }];
}

export default function LiturgiMinistriesPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const activeParokiId = me?.activeParokiId ?? "";
  const canEdit = canManageLiturgiGlobal(me?.roles);

  const ministriesQuery = useLiturgiMinistries(Boolean(activeParokiId));
  const [editingRole, setEditingRole] = useState<LiturgiRoleSlug | null>(null);
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [selectedLeaderUserId, setSelectedLeaderUserId] = useState("");
  const [memberSearch, setMemberSearch] = useState("");
  const [feedback, setFeedback] = useState<{
    tone: "success" | "error";
    message: string;
  } | null>(null);

  const orgsQuery = useLiturgiOrgs(editingRole !== null);
  const membersQuery = useLiturgiParokiMembers(
    activeParokiId,
    memberSearch,
    Boolean(activeParokiId) && editingRole !== null,
  );
  const updateMinistry = useUpdateLiturgiMinistry();

  if (me && !activeParokiId) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8">
        <InductiveEmptyState
          icon={Church}
          title={t("liturgi.noParishTitle")}
          description={t("liturgi.noParishDescription")}
        />
      </div>
    );
  }

  const bySlug = new Map<string, LiturgiMinistry>();
  for (const item of ministriesQuery.data ?? []) {
    if (item.roleSlug) bySlug.set(item.roleSlug, item);
  }

  const orgs = orgsQuery.data ?? [];
  const members = membersQuery.data ?? [];

  function startEdit(slug: LiturgiRoleSlug, row?: LiturgiMinistry) {
    setFeedback(null);
    setMemberSearch("");
    setSelectedOrgId(row?.orgId ?? "");
    setSelectedLeaderUserId(row?.leaderUserId ?? "");
    setEditingRole(slug);
  }

  async function handleSave(e: FormEvent, slug: LiturgiRoleSlug) {
    e.preventDefault();
    setFeedback(null);
    const meta = getLiturgiMinistryMeta(slug);
    try {
      await updateMinistry.mutateAsync({
        roleSlug: slug,
        body: {
          ...(selectedOrgId ? { orgId: selectedOrgId } : { clearOrgId: true }),
          ...(selectedLeaderUserId
            ? { leaderUserId: selectedLeaderUserId }
            : { clearLeaderUserId: true }),
        },
      });
      setEditingRole(null);
      setFeedback({
        tone: "success",
        message: t("liturgi.ministriesSaved", { role: meta?.label ?? slug }),
      });
    } catch (err) {
      setFeedback({ tone: "error", message: resolveLiturgiError(err) });
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi"), to: "/liturgi" },
          { label: t("liturgi.ministriesTitle") },
        ]}
        title={t("liturgi.ministriesTitle")}
        subtitle={t("liturgi.ministriesSubtitle")}
        action={
          <Link to="/liturgi" className={buttonClass("outline")}>
            <ArrowLeft aria-hidden className="size-4" />
            {t("liturgi.ministriesBack")}
          </Link>
        }
      />

      {feedback && (
        <Card
          className={`px-4 py-3 text-sm ${
            feedback.tone === "error"
              ? "border-danger/30 bg-danger-bg text-danger-fg"
              : "border-primary/30 bg-primary/10 text-fg"
          }`}
        >
          {feedback.message}
        </Card>
      )}

      {ministriesQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">{t("auth.loading")}</Card>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {LITURGI_MINISTRIES.map((meta) => {
            const row = bySlug.get(meta.slug);
            const isEditing = editingRole === meta.slug;

            return (
              <Card key={meta.slug} className="flex flex-col gap-4 p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h2 className="text-base font-bold text-fg">
                      {meta.label}
                    </h2>
                    <p className="mt-0.5 text-sm text-muted">
                      {meta.description}
                    </p>
                  </div>
                  {canEdit && !isEditing && (
                    <Button
                      variant="outline"
                      className="h-9 shrink-0 px-3 text-xs"
                      onClick={() => startEdit(meta.slug, row)}
                    >
                      <Pencil aria-hidden className="size-3.5" />
                      {t("liturgi.ministriesEdit")}
                    </Button>
                  )}
                </div>

                {!isEditing ? (
                  <div className="flex flex-col gap-2 border-t border-line pt-3 text-sm">
                    <div className="flex flex-wrap items-center gap-2">
                      <Building2 aria-hidden className="size-4 text-muted" />
                      <Badge tone={row?.orgName ? "neutral" : "warning"}>
                        {row?.orgName || t("liturgi.ministriesNoOrg")}
                      </Badge>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                      <UserCheck aria-hidden className="size-4 text-muted" />
                      <Badge tone={row?.leaderName ? "neutral" : "warning"}>
                        {row?.leaderName || t("liturgi.ministriesNoLeader")}
                      </Badge>
                    </div>
                    {row?.canManage && (
                      <div className="pt-1">
                        <Badge tone="success">
                          {t("liturgi.ministriesCanManage")}
                        </Badge>
                      </div>
                    )}
                  </div>
                ) : (
                  <form
                    onSubmit={(e) => handleSave(e, meta.slug)}
                    className="flex flex-col gap-3 border-t border-line pt-4"
                  >
                    <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                      {t("liturgi.ministriesOrgLabel")}
                      <select
                        value={selectedOrgId}
                        onChange={(e) => setSelectedOrgId(e.target.value)}
                        className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                      >
                        <option value="">
                          {t("liturgi.ministriesOrgNone")}
                        </option>
                        {orgs.map((org) => (
                          <option key={org.id} value={org.id}>
                            {org.name}
                          </option>
                        ))}
                      </select>
                    </label>

                    <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                      {t("liturgi.ministriesLeaderLabel")}
                      <input
                        type="search"
                        value={memberSearch}
                        onChange={(e) => setMemberSearch(e.target.value)}
                        placeholder={t("users.searchPlaceholder")}
                        className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg placeholder:text-muted sm:text-sm"
                      />
                      <select
                        value={selectedLeaderUserId}
                        onChange={(e) =>
                          setSelectedLeaderUserId(e.target.value)
                        }
                        className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-base text-fg sm:text-sm"
                      >
                        <option value="">
                          {t("liturgi.ministriesLeaderNone")}
                        </option>
                        {row?.leaderUserId &&
                          !members.some(
                            (m) => m.userId === row.leaderUserId,
                          ) && (
                            <option value={row.leaderUserId}>
                              {row.leaderName || row.leaderEmail}
                            </option>
                          )}
                        {members.map((m) => (
                          <option key={m.userId} value={m.userId}>
                            {m.name || m.email}
                          </option>
                        ))}
                      </select>
                    </label>

                    <div className="flex flex-col-reverse gap-2 pt-1 sm:flex-row sm:justify-end">
                      <Button
                        variant="outline"
                        className="h-11 w-full sm:w-auto"
                        onClick={() => setEditingRole(null)}
                      >
                        {t("liturgi.ministriesCancel")}
                      </Button>
                      <Button
                        type="submit"
                        disabled={updateMinistry.isPending}
                        className="h-11 w-full sm:w-auto"
                      >
                        {t("liturgi.ministriesSave")}
                      </Button>
                    </div>
                  </form>
                )}
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
