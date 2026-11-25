import { MapPin } from "lucide-react";
import { useLocation } from "react-router";
import { useMe } from "~/shared/api/session";
import { InductivePageHeader } from "~/shared/ui/iui";
import { UnderConstruction } from "~/shared/ui/UnderConstruction";
import { isSuperAdmin } from "../model";

export function meta() {
  return [{ title: "Segera Hadir — Domus" }];
}

export default function RefPlaceholderPage() {
  const { pathname } = useLocation();
  const { data: me, isLoading: meLoading } = useMe();
  const canManage = isSuperAdmin(me?.roles);
  const isSipil = pathname.startsWith("/settings/sipil");

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-6 py-8">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Pengaturan", to: "/settings" },
          ...(isSipil
            ? [{ label: "Wilayah Sipil" }]
            : [
                { label: "Struktur Gerejawi", to: "/settings/gerejawi" },
                { label: "Kevikepan" },
              ]),
        ]}
        title={isSipil ? "Wilayah Sipil" : "Kevikepan"}
        subtitle={
          isSipil
            ? "Pengelolaan data kota/kabupaten, kecamatan, dan kelurahan."
            : "Pengelolaan data kevikepan di bawah keuskupan."
        }
      />

      {!meLoading && !canManage ? (
        <div
          role="alert"
          className="rounded-xl border border-warning-fg/30 bg-warning-bg px-4 py-3 text-sm text-warning-fg"
        >
          Hanya Super Admin yang dapat mengakses halaman ini.
        </div>
      ) : null}

      {canManage ? (
        <UnderConstruction
          app={{
            id: "settings",
            path: "/settings",
            order: 5,
            plate: 2,
            icon: MapPin,
          }}
        />
      ) : null}
    </div>
  );
}
