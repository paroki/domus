import { useMe } from "~/shared/api/session";
import {
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { GEREJAWI_STEPS, isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Struktur Gerejawi — Domus" }];
}

export default function GerejawiHubPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Struktur Gerejawi" },
        ]}
        title="Struktur Gerejawi"
        subtitle="Kelola keuskupan, kevikepan, dan tenant paroki secara berurutan."
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola struktur gerejawi.
        </div>
      ) : null}

      {canManage ? (
        <nav aria-label="Struktur Gerejawi">
          <InductiveHubGrid>
            {GEREJAWI_STEPS.map((step, idx) => (
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
