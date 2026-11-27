import { ArrowLeft, Music } from "lucide-react";
import { Link, useParams } from "react-router";
import { InductiveEmptyState, InductivePageHeader } from "~/shared/ui/iui";

export function meta() {
  return [{ title: "Nyanyian · Domus" }];
}

export default function CantusPlaceholderPage() {
  const { cantusId } = useParams();
  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Nyanyian", to: "/cantus" },
          { label: cantusId ? "Detail" : "Tambah Nyanyian" },
        ]}
        title={cantusId ? "Detail Nyanyian" : "Tambah Nyanyian"}
        subtitle="Halaman ini sedang disiapkan."
      />
      <InductiveEmptyState
        icon={Music}
        title="Segera Hadir"
        description="Tampilan detail dan formulir nyanyian sedang dibangun pada tahap berikutnya."
        action={
          <Link
            to="/cantus"
            className="inline-flex items-center gap-1.5 text-sm font-semibold text-primary no-underline"
          >
            <ArrowLeft aria-hidden className="size-4" />
            Kembali ke Katalog Nyanyian
          </Link>
        }
      />
    </div>
  );
}
