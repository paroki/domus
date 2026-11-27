import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { resolveUmatErrorMessage } from "../api/errors";
import { useKeluargaDetail } from "../api/queries";
import { KeluargaSectionCards } from "../components/KeluargaSectionCards";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import { isUmatAdmin } from "../model";

export function meta() {
  return [{ title: "Detail Kartu Keluarga · Data Umat · Domus" }];
}

export default function KeluargaDetailPage() {
  const { keluargaId = "" } = useParams();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const [feedback, setFeedback] = useState<{
    tone: "success" | "error";
    message: string;
  } | null>(null);

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const detailQuery = useKeluargaDetail(
    activeParokiId,
    keluargaId,
    Boolean(me && allowed && activeParokiId && keluargaId),
  );

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Data Umat", to: "/umat" },
          { label: "Kartu Keluarga", to: "/umat/keluarga" },
          { label: detailQuery.data?.namaKeluarga ?? "Detail" },
        ]}
        title={detailQuery.data?.namaKeluarga ?? "Detail Kartu Keluarga"}
        subtitle={
          detailQuery.data?.nomorRegistrasi
            ? `Nomor Registrasi: ${detailQuery.data.nomorRegistrasi} · Kelola identitas Kartu Keluarga, alamat domisili, kondisi rumah tangga, dan anggota.`
            : "Kelola identitas Kartu Keluarga, alamat domisili, kondisi rumah tangga, dan anggota."
        }
      />

      {feedback ? (
        <div
          role="alert"
          className={`rounded-lg border px-4 py-3 text-sm ${
            feedback.tone === "error"
              ? "border-danger/30 bg-danger-bg text-danger-fg"
              : "border-success/30 bg-success-bg text-success-fg"
          }`}
        >
          {feedback.message}
        </div>
      ) : null}

      {detailQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">
          Memuat detail Kartu Keluarga…
        </Card>
      ) : detailQuery.isError || !detailQuery.data ? (
        <Card className="flex flex-col items-start gap-3 p-6">
          <p className="text-sm text-danger-fg">
            {resolveUmatErrorMessage(
              detailQuery.error,
              "Kartu Keluarga tidak ditemukan.",
            )}
          </p>
          <Link to="/umat/keluarga" className="no-underline">
            <Button variant="outline">Kembali ke daftar Kartu Keluarga</Button>
          </Link>
        </Card>
      ) : (
        <KeluargaSectionCards
          parokiId={activeParokiId}
          keluarga={detailQuery.data}
          onFeedback={setFeedback}
          onDeleted={() => navigate("/umat/keluarga", { replace: true })}
        />
      )}
    </div>
  );
}
