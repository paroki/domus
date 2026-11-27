import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { resolveUmatErrorMessage } from "../api/errors";
import { useUmatDetail } from "../api/queries";
import { UmatForbiddenCard } from "../components/UmatForbiddenCard";
import { UmatSectionCards } from "../components/UmatSectionCards";
import { isUmatAdmin } from "../model";

export function meta() {
  return [{ title: "Profil Umat · Data Umat · Domus" }];
}

export default function UmatDetailPage() {
  const { umatId = "" } = useParams();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const [feedback, setFeedback] = useState<{
    tone: "success" | "error";
    message: string;
  } | null>(null);

  const activeParokiId = me?.activeParokiId ?? "";
  const allowed = isUmatAdmin(me);

  const detailQuery = useUmatDetail(
    activeParokiId,
    umatId,
    Boolean(me && allowed && activeParokiId && umatId),
  );

  if (me && (!allowed || !activeParokiId)) {
    return <UmatForbiddenCard />;
  }

  const umat = detailQuery.data;
  const breadcrumbs = [
    { label: "Beranda", to: "/" },
    { label: "Data Umat", to: "/umat" },
    ...(umat?.keluarga?.id
      ? [
          {
            label: `KK ${umat.keluarga.namaKeluarga}`,
            to: `/umat/keluarga/${umat.keluarga.id}`,
          },
        ]
      : []),
    { label: umat?.nama ?? "Profil Umat" },
  ];

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-6 sm:px-6 sm:py-8">
      <InductivePageHeader
        breadcrumbs={breadcrumbs}
        title={umat?.nama ?? "Profil Umat"}
        subtitle={
          umat?.nomorInduk
            ? `Nomor Induk: ${umat.nomorInduk} · Kelola identitas pribadi, Kartu Keluarga, kontak, pekerjaan, tautan akun, dan sakramen.`
            : "Kelola identitas pribadi, Kartu Keluarga, kontak, pekerjaan, tautan akun, dan sakramen."
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
        <Card className="p-4 text-sm text-muted sm:p-6">
          Memuat profil umat…
        </Card>
      ) : detailQuery.isError || !umat ? (
        <Card className="flex flex-col items-start gap-3 p-4 sm:p-6">
          <p className="text-sm text-danger-fg">
            {resolveUmatErrorMessage(
              detailQuery.error,
              "Data umat tidak ditemukan.",
            )}
          </p>
          <Link to="/umat" className="no-underline">
            <Button variant="outline">Kembali ke direktori umat</Button>
          </Link>
        </Card>
      ) : (
        <UmatSectionCards
          parokiId={activeParokiId}
          umat={umat}
          onFeedback={setFeedback}
          onDeleted={() => navigate("/umat", { replace: true })}
        />
      )}
    </div>
  );
}
