import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Navigate, Outlet, useLocation } from "react-router";
import { ApiError, setActiveParish } from "~/shared/api/client";
import { useMe } from "../api/queries";

export default function RequireAuth() {
  const { t } = useTranslation();
  const location = useLocation();
  const { data: me, isLoading, error } = useMe();

  useEffect(() => {
    if (me) {
      setActiveParish(me.activeParishId ?? null);
    } else if (error instanceof ApiError && error.status === 401) {
      setActiveParish(null);
    }
  }, [me, error]);

  if (isLoading) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="flex min-h-dvh items-center justify-center bg-canvas text-sm font-medium text-muted"
      >
        {t("auth.loading")}
      </div>
    );
  }

  if (error || !me) {
    const redirectTarget = encodeURIComponent(
      `${location.pathname}${location.search}`,
    );
    return <Navigate to={`/login?redirect=${redirectTarget}`} replace />;
  }

  return <Outlet />;
}
