import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import { Button } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductivePageHeader } from "~/shared/ui/iui";
import { resolveCantusErrorMessage } from "../api/errors";
import { useCantusDetail } from "../api/queries";
import { CantusSectionCards } from "../components/CantusSectionCards";
import { isCantusAdmin } from "../model";

export function meta() {
  return [{ title: "Detail Nyanyian · Nyanyian · Domus" }];
}

export default function CantusDetailPage() {
  const { cantusId = "" } = useParams();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const [feedback, setFeedback] = useState<{
    tone: "success" | "error";
    message: string;
  } | null>(null);

  const detailQuery = useCantusDetail(cantusId, Boolean(me && cantusId));
  const cantus = detailQuery.data;
  const canManage = Boolean(cantus?.canManage || isCantusAdmin(me));

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: "Beranda", to: "/" },
          { label: "Nyanyian", to: "/cantus" },
          { label: cantus?.judul ?? "Detail Nyanyian" },
        ]}
        title={cantus?.judul ?? "Detail Nyanyian"}
        subtitle="Lihat lirik, bait mazmur & BPI, kredit pencipta, partitur, dan audio latihan."
      />

      {feedback && (
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
      )}

      {detailQuery.isLoading ? (
        <Card className="p-6 text-sm text-muted">Memuat detail nyanyian…</Card>
      ) : detailQuery.isError || !cantus ? (
        <Card className="flex flex-col items-start gap-3 p-6">
          <p className="text-sm text-danger-fg">
            {resolveCantusErrorMessage(
              detailQuery.error,
              "Data nyanyian tidak ditemukan.",
            )}
          </p>
          <Link to="/cantus" className="no-underline">
            <Button variant="outline">Kembali ke Katalog Nyanyian</Button>
          </Link>
        </Card>
      ) : (
        <CantusSectionCards
          cantus={cantus}
          canManage={canManage}
          onFeedback={setFeedback}
          onDeleted={() => navigate("/cantus", { replace: true })}
        />
      )}
    </div>
  );
}
