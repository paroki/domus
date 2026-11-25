import { ArrowLeft, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useActivateSessionParoki,
  useCreateKeuskupan,
  useCreateParoki,
  useDeleteKeuskupan,
  useDeleteParoki,
} from "../api/actions";
import {
  useCurrentIdentity,
  useKeuskupanSearch,
  useParokiSearch,
} from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Keuskupan & Paroki — Domus" }];
}

export default function ParokiPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const allowed = isSuperAdmin(me?.roles);

  const keuskupanQuery = useKeuskupanSearch({}, allowed);
  const parokiQuery = useParokiSearch({}, allowed);
  const keuskupanList =
    keuskupanQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const parokiList = parokiQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createKeuskupan = useCreateKeuskupan();
  const deleteKeuskupan = useDeleteKeuskupan();
  const createParoki = useCreateParoki();
  const deleteParoki = useDeleteParoki();
  const activateSessionParoki = useActivateSessionParoki();

  const [kKode, setKKode] = useState("");
  const [kNama, setKNama] = useState("");
  const [kUskup, setKUskup] = useState("");

  const [pKeuskupanId, setPKeuskupanId] = useState("");
  const [pKode, setPKode] = useState("");
  const [pNama, setPNama] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  async function handleCreateKeuskupan(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const created = await createKeuskupan.mutateAsync({
        kode: slugifyCode(kKode || kNama),
        nama: kNama.trim(),
        uskup: kUskup.trim() || undefined,
      });
      setKKode("");
      setKNama("");
      setKUskup("");
      if (!pKeuskupanId && created.id) setPKeuskupanId(created.id);
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal membuat keuskupan.",
      );
    }
  }

  async function handleCreateParoki(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    const keuskupanId = pKeuskupanId || keuskupanList[0]?.id;
    if (!keuskupanId) {
      setErrorMsg(t("ref.selectDioceseFirst"));
      return;
    }
    try {
      const created = await createParoki.mutateAsync({
        keuskupanId,
        kode: slugifyCode(pKode || pNama),
        nama: pNama.trim(),
      });
      setPKode("");
      setPNama("");
      if (!me?.activeParokiId && created.id) {
        await activateSessionParoki.mutateAsync(created.id);
      }
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal membuat paroki.");
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-8 px-4 py-8 sm:px-6">
      <div>
        <Link
          to="/ref"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted no-underline hover:text-fg"
        >
          <ArrowLeft aria-hidden className="size-4" />
          {t("ref.backToSettings")}
        </Link>
        <h1 className="mt-2 text-2xl font-bold text-fg">
          {t("ref.parishesTitle")}
        </h1>
        <p className="text-sm text-muted">{t("ref.parishesSubtitle")}</p>
      </div>

      {!allowed ? (
        <div className="rounded-xl border border-line bg-surface p-6 text-sm text-muted">
          {t("ref.forbiddenSuperAdmin")}
        </div>
      ) : (
        <>
          {errorMsg && (
            <div
              role="alert"
              className="rounded-xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger"
            >
              {errorMsg}
            </div>
          )}

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <h2 className="text-lg font-semibold text-fg">
              {t("ref.diocesesSection")}
            </h2>
            <form
              onSubmit={handleCreateKeuskupan}
              className="grid gap-3 sm:grid-cols-4"
            >
              <input
                required
                value={kKode}
                onChange={(e) => setKKode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={kNama}
                onChange={(e) => setKNama(e.target.value)}
                placeholder={t("ref.dioceseNamePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                value={kUskup}
                onChange={(e) => setKUskup(e.target.value)}
                placeholder={t("ref.bishopPlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <button
                type="submit"
                disabled={createKeuskupan.isPending}
                className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                <Plus aria-hidden className="size-4" />
                {t("ref.addDiocese")}
              </button>
            </form>

            <div className="divide-y divide-line">
              {keuskupanList.map((d) => (
                <div
                  key={d.id}
                  className="flex items-center justify-between py-2.5"
                >
                  <div>
                    <p className="text-sm font-semibold text-fg">
                      {d.nama}{" "}
                      <span className="text-xs font-normal text-muted">
                        ({d.kode})
                      </span>
                    </p>
                    {d.uskup && <p className="text-xs text-muted">{d.uskup}</p>}
                  </div>
                  <button
                    type="button"
                    aria-label={`${t("users.deleteUser")} ${d.nama}`}
                    onClick={async () => {
                      if (!d.id) return;
                      setErrorMsg(null);
                      try {
                        await deleteKeuskupan.mutateAsync(d.id);
                      } catch (err) {
                        setErrorMsg(
                          err instanceof Error
                            ? err.message
                            : "Gagal menghapus keuskupan.",
                        );
                      }
                    }}
                    className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                  >
                    <Trash2 aria-hidden className="size-4" />
                  </button>
                </div>
              ))}
              {keuskupanList.length === 0 && (
                <p className="py-3 text-sm text-muted">{t("ref.emptyList")}</p>
              )}
            </div>
          </section>

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <h2 className="text-lg font-semibold text-fg">
              {t("ref.parishesSection")}
            </h2>
            <form
              onSubmit={handleCreateParoki}
              className="grid gap-3 sm:grid-cols-4"
            >
              <select
                value={pKeuskupanId || keuskupanList[0]?.id || ""}
                onChange={(e) => setPKeuskupanId(e.target.value)}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                <option value="">{t("ref.selectDiocese")}</option>
                {keuskupanList.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.nama}
                  </option>
                ))}
              </select>
              <input
                required
                value={pKode}
                onChange={(e) => setPKode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={pNama}
                onChange={(e) => setPNama(e.target.value)}
                placeholder={t("ref.parishNamePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <button
                type="submit"
                disabled={createParoki.isPending}
                className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                <Plus aria-hidden className="size-4" />
                {t("ref.addParish")}
              </button>
            </form>

            <div className="divide-y divide-line">
              {parokiList.map((p) => (
                <div
                  key={p.id}
                  className="flex items-center justify-between py-2.5"
                >
                  <div>
                    <p className="text-sm font-semibold text-fg">
                      {p.nama}{" "}
                      <span className="text-xs font-normal text-muted">
                        ({p.kode})
                      </span>
                    </p>
                  </div>
                  <button
                    type="button"
                    aria-label={`${t("users.deleteUser")} ${p.nama}`}
                    onClick={async () => {
                      if (!p.id) return;
                      setErrorMsg(null);
                      try {
                        await deleteParoki.mutateAsync(p.id);
                      } catch (err) {
                        setErrorMsg(
                          err instanceof Error
                            ? err.message
                            : "Gagal menghapus paroki.",
                        );
                      }
                    }}
                    className="inline-flex size-8 items-center justify-center rounded-lg text-danger hover:bg-danger/10"
                  >
                    <Trash2 aria-hidden className="size-4" />
                  </button>
                </div>
              ))}
              {parokiList.length === 0 && (
                <p className="py-3 text-sm text-muted">{t("ref.emptyList")}</p>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  );
}
