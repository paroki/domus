import { Building2, Plus, Search } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import {
  InductiveEmptyState,
  InductiveFilterBar,
  InductivePageHeader,
} from "~/shared/ui/iui";
import { useOrgs } from "../api/queries";
import { isParokiAdmin, kindLabel, ORG_KINDS, type OrgKind } from "../model";
import type { Route } from "./+types/OrgsPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("org.title")} · Domus` }];
}

export default function OrgsPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  const hasParoki = Boolean(me?.activeParokiId);
  const parokiAdmin = isParokiAdmin(me?.roles);

  const [myOrgsOnly, setMyOrgsOnly] = useState(false);
  const [kindFilter, setKindFilter] = useState("");
  const [q, setQ] = useState("");

  const orgsQuery = useOrgs(
    {
      q: q.trim() || undefined,
      filter: {
        kind: (kindFilter as OrgKind) || undefined,
        myOrgsOnly: myOrgsOnly || undefined,
      },
    },
    hasParoki,
  );

  const items = orgsQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  if (!hasParoki) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
        <InductiveEmptyState
          icon={Building2}
          title={t("org.noParishTitle")}
          description={t("org.noParishDescription")}
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[{ label: "Beranda", to: "/" }, { label: t("org.title") }]}
        title={t("org.title")}
        subtitle={t("org.subtitle")}
        action={
          parokiAdmin ? (
            <Link to="/org/new" className={buttonClass("primary")}>
              <Plus aria-hidden className="size-4" />
              {t("org.addOrg")}
            </Link>
          ) : undefined
        }
      />

      <InductiveFilterBar className="flex-col items-stretch sm:flex-row sm:items-center sm:justify-between">
        <div className="flex flex-wrap gap-2">
          <Button
            variant={myOrgsOnly ? "outline" : "primary"}
            onClick={() => setMyOrgsOnly(false)}
            className="min-h-10"
          >
            {t("org.allOrgs")}
          </Button>
          <Button
            variant={myOrgsOnly ? "primary" : "outline"}
            onClick={() => setMyOrgsOnly(true)}
            className="min-h-10"
          >
            {t("org.myOrgs")}
          </Button>
        </div>

        <div className="flex flex-1 flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
          <select
            aria-label={t("org.allKinds")}
            value={kindFilter}
            onChange={(e) => setKindFilter(e.target.value)}
            className="h-11 w-full rounded-lg border border-line bg-canvas px-3 text-sm text-fg sm:w-auto"
          >
            <option value="">{t("org.allKinds")}</option>
            {ORG_KINDS.map((k) => (
              <option key={k.value} value={k.value}>
                {k.label}
              </option>
            ))}
          </select>

          <div className="relative w-full sm:max-w-xs">
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
      </InductiveFilterBar>

      {orgsQuery.isLoading ? (
        <p className="text-sm text-muted">Memuat organisasi…</p>
      ) : items.length === 0 ? (
        <InductiveEmptyState
          icon={Building2}
          title="Organisasi tidak ditemukan"
          description="Belum ada organisasi yang sesuai dengan filter."
          action={
            parokiAdmin ? (
              <Link to="/org/new" className={buttonClass("primary")}>
                <Plus aria-hidden className="size-4" />
                {t("org.addOrg")}
              </Link>
            ) : undefined
          }
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {items.map((org) => (
            <Card
              key={org.id}
              className="flex flex-col justify-between gap-4 p-4 sm:p-5"
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
            className="w-full sm:w-auto"
          >
            {orgsQuery.isFetchingNextPage ? "Memuat…" : t("org.loadMore")}
          </Button>
        </div>
      )}
    </div>
  );
}
