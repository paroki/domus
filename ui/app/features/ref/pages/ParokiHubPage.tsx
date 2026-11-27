import { useMe } from "~/shared/api/session";
import {
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { isParokiAdmin, PAROKI_STEPS } from "../model";

export function meta() {
  return [{ title: "Administrasi Paroki — Domus" }];
}

export default function ParokiHubPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isParokiAdmin(me?.roles);

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Administrasi Paroki" },
        ]}
        title="Administrasi Paroki"
        subtitle="Kelola identitas paroki aktif, wilayah/stasi, dan lingkungan."
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Admin Paroki yang dapat mengelola administrasi paroki.
        </div>
      ) : null}

      {canManage ? (
        <nav aria-label="Administrasi Paroki">
          <InductiveHubGrid>
            {PAROKI_STEPS.map((step, idx) => (
              <InductiveStepCard
                key={step.id}
                to={step.path}
                step={idx + 1}
                title={step.title}
                description={step.description}
                icon={step.icon}
                plate={step.plate}
              />
            ))}
          </InductiveHubGrid>
        </nav>
      ) : null}
    </div>
  );
}
