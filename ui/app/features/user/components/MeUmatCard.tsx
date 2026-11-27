import { Eye, EyeOff, FileImage, ShieldCheck, UserCheck } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";
import { getActiveParokiId } from "~/shared/api/client";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { useMyUmat } from "../api/queries";
import type { SelfDokumenMeta } from "../api/types";

function maskDigits(val?: string | null): string {
  if (!val) return "-";
  if (val.length <= 4) return "*".repeat(val.length);
  return `${"*".repeat(val.length - 4)}${val.slice(-4)}`;
}

function dokumenLabel(jenis?: string): string {
  switch (jenis) {
    case "ktp":
      return "KTP";
    case "kk":
      return "Kartu Keluarga";
    case "kartu_pelajar":
    case "kartuPelajar":
      return "Kartu Pelajar";
    case "kia":
      return "KIA";
    default:
      return jenis || "Dokumen";
  }
}

function DokumenPreview({ doc }: { doc: SelfDokumenMeta }) {
  const [objectUrl, setObjectUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!doc.id) return;
    let active = true;
    let createdUrl: string | null = null;
    const controller = new AbortController();

    async function load() {
      try {
        const headers: Record<string, string> = {};
        const parokiId = getActiveParokiId();
        if (parokiId) {
          headers["X-Paroki-Id"] = parokiId;
        }
        const res = await fetch(`/api/umat/dokumen/${doc.id}`, {
          credentials: "include",
          headers,
          signal: controller.signal,
        });
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const blob = await res.blob();
        createdUrl = URL.createObjectURL(blob);
        if (active) {
          setObjectUrl(createdUrl);
        } else {
          URL.revokeObjectURL(createdUrl);
        }
      } catch {
        if (active && !controller.signal.aborted) {
          setFailed(true);
        }
      }
    }

    void load();

    return () => {
      active = false;
      controller.abort();
      if (createdUrl) {
        URL.revokeObjectURL(createdUrl);
      }
    };
  }, [doc.id]);

  return (
    <div className="flex flex-col gap-2 rounded-xl border border-line bg-canvas p-3">
      <div className="flex items-center gap-2 text-xs font-semibold text-muted">
        <FileImage aria-hidden className="size-4" />
        <span>{dokumenLabel(doc.jenis)}</span>
      </div>
      {objectUrl ? (
        <img
          src={objectUrl}
          alt={dokumenLabel(doc.jenis)}
          className="max-h-40 w-full rounded-lg border border-line object-contain bg-surface"
        />
      ) : failed ? (
        <p className="text-xs text-muted">Pratinjau dokumen tidak tersedia.</p>
      ) : (
        <p className="text-xs text-muted">Memuat pratinjau…</p>
      )}
    </div>
  );
}

export function MeUmatCard({ verified }: { verified?: boolean }) {
  const { data: profile, isLoading, isError } = useMyUmat(Boolean(verified));
  const [showSensitive, setShowSensitive] = useState(false);

  if (!verified) {
    return (
      <Card className="flex flex-col gap-4 p-4 sm:p-6">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-fg">Verifikasi Data Umat</h2>
            <p className="mt-1 text-sm text-muted">
              Hubungkan akun Anda dengan data umat dan Kartu Keluarga paroki
              untuk mendapatkan Nomor Induk Umat.
            </p>
          </div>
          <Badge tone="neutral">Belum Terverifikasi</Badge>
        </div>
        <div>
          <Link
            to="/me/verify"
            className={`${buttonClass("primary")} w-full sm:w-auto`}
          >
            <UserCheck aria-hidden className="size-4" />
            Verifikasi Data Umat
          </Link>
        </div>
      </Card>
    );
  }

  const umat = profile?.umat;
  const keluarga = profile?.keluarga;
  const members = profile?.members ?? [];
  const dokumen = profile?.dokumen ?? [];

  return (
    <Card className="flex flex-col gap-5 p-4 sm:p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-bold text-fg">
            Data Umat &amp; Keluarga Saya
          </h2>
          <p className="mt-0.5 text-sm text-muted">
            Identitas umat paroki dan anggota Kartu Keluarga Anda.
          </p>
        </div>
        <Badge tone="success">
          <ShieldCheck aria-hidden className="mr-1 size-4" />
          Terverifikasi
        </Badge>
      </div>

      {isLoading ? (
        <p className="text-sm text-muted">Memuat data umat…</p>
      ) : isError || !profile ? (
        <p className="text-sm text-danger-fg">
          Gagal memuat data umat &amp; keluarga.
        </p>
      ) : (
        <>
          <div className="flex items-center justify-end">
            <Button
              variant="outline"
              onClick={() => setShowSensitive((v) => !v)}
              className="w-full sm:w-auto"
            >
              {showSensitive ? (
                <>
                  <EyeOff aria-hidden className="size-4" />
                  Sembunyikan NIK &amp; KK
                </>
              ) : (
                <>
                  <Eye aria-hidden className="size-4" />
                  Tampilkan
                </>
              )}
            </Button>
          </div>

          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs font-medium text-muted">
                Nomor Induk Umat
              </dt>
              <dd className="mt-0.5 font-mono text-sm font-semibold text-fg">
                {umat?.nomorInduk || "-"}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">
                Nomor Registrasi Keluarga
              </dt>
              <dd className="mt-0.5 font-mono text-sm font-semibold text-fg">
                {keluarga?.nomorRegistrasi || "-"}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">NIK</dt>
              <dd className="mt-0.5 font-mono text-sm text-fg">
                {showSensitive ? umat?.nik || "-" : maskDigits(umat?.nik)}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Nomor KK</dt>
              <dd className="mt-0.5 font-mono text-sm text-fg">
                {showSensitive
                  ? keluarga?.nomorKk || "-"
                  : maskDigits(keluarga?.nomorKk)}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">Nama Lengkap</dt>
              <dd className="mt-0.5 text-sm font-medium text-fg">
                {umat?.nama || "-"}
              </dd>
            </div>
            <div>
              <dt className="text-xs font-medium text-muted">
                Kepala Keluarga
              </dt>
              <dd className="mt-0.5 text-sm font-medium text-fg">
                {keluarga?.namaKeluarga || "-"}
              </dd>
            </div>
            {(keluarga?.alamat || umat?.alamat) && (
              <div className="sm:col-span-2">
                <dt className="text-xs font-medium text-muted">Alamat</dt>
                <dd className="mt-0.5 text-sm text-fg">
                  {keluarga?.alamat || umat?.alamat}
                </dd>
              </div>
            )}
          </dl>

          {members.length > 0 && (
            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold text-fg">
                Anggota Keluarga ({members.length})
              </h3>
              <ul className="divide-y divide-line rounded-xl border border-line bg-canvas">
                {members.map((m, idx) => (
                  <li
                    key={m.id ?? `${m.nama ?? "member"}-${idx}`}
                    className="flex flex-wrap items-center justify-between gap-2 px-3.5 py-2.5 text-sm"
                  >
                    <div>
                      <p className="font-medium text-fg">{m.nama || "-"}</p>
                      <p className="text-xs text-muted">
                        {m.hubunganKeluarga || "Anggota"}
                      </p>
                    </div>
                    {m.nomorInduk && (
                      <span className="font-mono text-xs text-muted">
                        {m.nomorInduk}
                      </span>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {dokumen.length > 0 && (
            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold text-fg">
                Dokumen Tersimpan ({dokumen.length})
              </h3>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                {dokumen.map((doc, idx) => (
                  <DokumenPreview key={doc.id ?? idx} doc={doc} />
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </Card>
  );
}
