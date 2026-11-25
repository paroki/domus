import { useMe } from "~/shared/api/session";
import {
  InductiveHubGrid,
  InductivePageHeader,
  InductiveStepCard,
} from "~/shared/ui/iui";
import { isSuperAdmin, SIPIL_STEPS } from "../model";

export function meta() {
  return [{ title: "Wilayah Sipil — Domus" }];
}

export default function SipilHubPage() {
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          { label: "Wilayah Sipil" },
        ]}
        title="Wilayah Sipil"
        subtitle="Kelola data referensi kota/kabupaten, kecamatan, dan kelurahan/desa secara berurutan."
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengelola wilayah sipil.
        </div>
      ) : null}

      {canManage ? (
        <nav aria-label="Wilayah Sipil">
          <InductiveHubGrid>
            {SIPIL_STEPS.map((step, idx) => (
              <InductiveStepCard
                key={step.id}
                to={step.path}
                step={idx + 1}
                title={`Langkah ${idx + 1}: ${step.title}`}
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
