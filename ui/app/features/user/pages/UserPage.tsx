import { ShieldAlert } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import i18n from "~/shared/i18n";
import { Button } from "~/shared/ui/Button";
import { useAuthRoles, useMe } from "../api/queries";
import { GlobalUsersScope } from "../components/GlobalUsersScope";
import { OrgMembersScope } from "../components/OrgMembersScope";
import { ParishMembersScope } from "../components/ParishMembersScope";
import { getAvailableUserScopes, type UserDirectoryScope } from "../model";
import type { Route } from "./+types/UserPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("apps.user.name")} · Domus` }];
}

export default function UserPage() {
  const { t } = useTranslation();
  const { data: me } = useMe();
  useAuthRoles();

  const scopes = getAvailableUserScopes(me?.roles);
  const [activeScope, setActiveScope] = useState<UserDirectoryScope | null>(
    null,
  );
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (!scopes.length) {
      setActiveScope(null);
      return;
    }
    if (!activeScope || !scopes.includes(activeScope)) {
      setActiveScope(scopes[0] ?? null);
    }
  }, [scopes, activeScope]);

  if (me && scopes.length === 0) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-3 px-4 py-16 text-center">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">
          {t("users.forbiddenTitle")}
        </h1>
        <p className="text-sm text-muted">{t("users.forbiddenDescription")}</p>
      </div>
    );
  }

  const scopeLabels: Record<UserDirectoryScope, string> = {
    global: t("users.scopeGlobal"),
    parish: t("users.scopeParish"),
    org: t("users.scopeOrg"),
  };

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <div>
        <h1 className="text-2xl font-bold text-fg">{t("users.title")}</h1>
        <p className="text-sm text-muted">{t("users.subtitle")}</p>
      </div>

      {scopes.length > 1 && (
        <div
          role="tablist"
          aria-label={t("users.title")}
          className="flex flex-wrap gap-2"
        >
          {scopes.map((scope) => (
            <Button
              key={scope}
              role="tab"
              aria-selected={activeScope === scope}
              variant={activeScope === scope ? "primary" : "outline"}
              onClick={() => {
                setErrorMsg(null);
                setActiveScope(scope);
              }}
            >
              {scopeLabels[scope]}
            </Button>
          ))}
        </div>
      )}

      {errorMsg && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {me && activeScope === "global" && (
        <GlobalUsersScope onError={setErrorMsg} />
      )}
      {me && activeScope === "parish" && (
        <ParishMembersScope me={me} onError={setErrorMsg} />
      )}
      {me && activeScope === "org" && (
        <OrgMembersScope me={me} onError={setErrorMsg} />
      )}
    </div>
  );
}
