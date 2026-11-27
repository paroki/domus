import { AlertTriangle, ChevronRight, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router";
import { Card } from "./Card";
import { plateBg } from "./plate";

export interface BreadcrumbItem {
  label: string;
  to?: string;
}

export function InductivePageHeader({
  breadcrumbs,
  title,
  subtitle,
  action,
}: {
  breadcrumbs: BreadcrumbItem[];
  title: string;
  subtitle: string;
  action?: ReactNode;
}) {
  return (
    <header className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div className="flex flex-col gap-1.5">
        <nav
          aria-label="Breadcrumb"
          className="flex flex-wrap items-center gap-1 text-xs text-muted"
        >
          {breadcrumbs.map((crumb, idx) => (
            <span
              key={`${crumb.label}-${idx}`}
              className="inline-flex items-center gap-1"
            >
              {idx > 0 ? (
                <ChevronRight aria-hidden className="size-3.5" />
              ) : null}
              {crumb.to ? (
                <Link
                  to={crumb.to}
                  className="text-muted no-underline hover:text-fg"
                >
                  {crumb.label}
                </Link>
              ) : (
                <span className="font-medium text-fg">{crumb.label}</span>
              )}
            </span>
          ))}
        </nav>
        <h1 className="text-2xl font-bold text-fg">{title}</h1>
        <p className="text-sm text-muted">{subtitle}</p>
      </div>
      {action ? (
        <div className="w-full shrink-0 sm:w-auto [&>*]:w-full sm:[&>*]:w-auto">
          {action}
        </div>
      ) : null}
    </header>
  );
}

export function InductiveStepCard({
  to,
  title,
  description,
  icon: Icon,
  plate = 1,
  step,
  badge,
}: {
  to: string;
  title: string;
  description: string;
  icon: LucideIcon;
  plate?: 1 | 2 | 3 | 4 | 5 | 6;
  step?: number;
  badge?: ReactNode;
}) {
  return (
    <Link
      to={to}
      className="group flex items-start gap-4 rounded-xl border border-line bg-surface p-4 text-left text-fg no-underline transition-colors hover:border-primary/50 hover:bg-surface-2/40 sm:p-5"
    >
      <span
        className={`flex size-12 shrink-0 items-center justify-center rounded-xl text-white ${plateBg[plate]}`}
      >
        <Icon aria-hidden className="size-6" strokeWidth={1.75} />
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex items-center justify-between gap-2">
          <span className="text-base font-semibold text-fg group-hover:text-primary">
            {step != null ? `${step}. ` : ""}
            {title}
          </span>
          {badge}
        </div>
        <p className="text-sm text-muted">{description}</p>
      </div>
    </Link>
  );
}

export function InductiveHubGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {children}
    </div>
  );
}

export function InductiveFilterBar({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <Card className={`flex flex-wrap items-end gap-3 p-4 ${className}`}>
      {children}
    </Card>
  );
}

export function InductiveFormSection({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <Card className="flex flex-col gap-4 p-4 sm:p-6">
      <div>
        <h2 className="text-base font-semibold text-fg">{title}</h2>
        {description ? (
          <p className="mt-0.5 text-sm text-muted">{description}</p>
        ) : null}
      </div>
      {children}
    </Card>
  );
}

export function InductiveDangerZone({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action: ReactNode;
}) {
  return (
    <Card className="flex flex-col justify-between gap-4 border-danger/30 bg-danger-bg/40 p-4 sm:flex-row sm:items-center sm:p-6">
      <div className="flex items-start gap-3">
        <AlertTriangle
          aria-hidden
          className="mt-0.5 size-5 shrink-0 text-danger-fg"
        />
        <div>
          <h2 className="text-base font-semibold text-danger-fg">{title}</h2>
          <p className="mt-0.5 text-sm text-muted">{description}</p>
        </div>
      </div>
      <div className="w-full shrink-0 sm:w-auto [&>*]:w-full sm:[&>*]:w-auto">
        {action}
      </div>
    </Card>
  );
}

export function InductiveEmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon?: LucideIcon;
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <Card className="flex flex-col items-center gap-3 p-6 text-center sm:p-10">
      {Icon ? (
        <span className="flex size-12 items-center justify-center rounded-full bg-surface-2 text-muted">
          <Icon aria-hidden className="size-6" />
        </span>
      ) : null}
      <div className="flex flex-col gap-1">
        <h2 className="text-base font-semibold text-fg">{title}</h2>
        <p className="max-w-md text-sm text-muted">{description}</p>
      </div>
      {action ? (
        <div className="mt-1 w-full sm:w-auto [&>*]:w-full sm:[&>*]:w-auto">
          {action}
        </div>
      ) : null}
    </Card>
  );
}
