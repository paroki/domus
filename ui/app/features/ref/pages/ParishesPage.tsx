import { ArrowLeft, Plus, Trash2 } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import {
  useActivateSessionParish,
  useCreateDiocese,
  useCreateParish,
  useDeleteDiocese,
  useDeleteParish,
} from "../api/actions";
import {
  useCurrentIdentity,
  useDiocesesSearch,
  useParishesSearch,
} from "../api/queries";
import { isSuperAdmin, slugifyCode } from "../model";

export function meta() {
  return [{ title: "Keuskupan & Paroki — Domus" }];
}

export default function ParishesPage() {
  const { t } = useTranslation();
  const { data: me } = useCurrentIdentity();
  const allowed = isSuperAdmin(me?.roles);

  const diocesesQuery = useDiocesesSearch({}, allowed);
  const parishesQuery = useParishesSearch({}, allowed);
  const dioceses = diocesesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];
  const parishes = parishesQuery.data?.pages.flatMap((p) => p.data ?? []) ?? [];

  const createDiocese = useCreateDiocese();
  const deleteDiocese = useDeleteDiocese();
  const createParish = useCreateParish();
  const deleteParish = useDeleteParish();
  const activateSessionParish = useActivateSessionParish();

  const [dCode, setDCode] = useState("");
  const [dName, setDName] = useState("");
  const [dBishop, setDBishop] = useState("");

  const [pDioceseId, setPDioceseId] = useState("");
  const [pCode, setPCode] = useState("");
  const [pName, setPName] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  async function handleCreateDiocese(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    try {
      const created = await createDiocese.mutateAsync({
        kode: slugifyCode(dCode || dName),
        nama: dName.trim(),
        uskup: dBishop.trim() || undefined,
      });
      setDCode("");
      setDName("");
      setDBishop("");
      if (!pDioceseId && created.id) setPDioceseId(created.id);
    } catch (err) {
      setErrorMsg(
        err instanceof Error ? err.message : "Gagal membuat keuskupan.",
      );
    }
  }

  async function handleCreateParish(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);
    const keuskupanId = pDioceseId || dioceses[0]?.id;
    if (!keuskupanId) {
      setErrorMsg(t("ref.selectDioceseFirst"));
      return;
    }
    try {
      const created = await createParish.mutateAsync({
        keuskupanId,
        kode: slugifyCode(pCode || pName),
        nama: pName.trim(),
      });
      setPCode("");
      setPName("");
      if (!me?.activeParokiId && created.id) {
        await activateSessionParish.mutateAsync(created.id);
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
              onSubmit={handleCreateDiocese}
              className="grid gap-3 sm:grid-cols-4"
            >
              <input
                required
                value={dCode}
                onChange={(e) => setDCode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={dName}
                onChange={(e) => setDName(e.target.value)}
                placeholder={t("ref.dioceseNamePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                value={dBishop}
                onChange={(e) => setDBishop(e.target.value)}
                placeholder={t("ref.bishopPlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <button
                type="submit"
                disabled={createDiocese.isPending}
                className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                <Plus aria-hidden className="size-4" />
                {t("ref.addDiocese")}
              </button>
            </form>

            <div className="divide-y divide-line">
              {dioceses.map((d) => (
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
                        await deleteDiocese.mutateAsync(d.id);
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
              {dioceses.length === 0 && (
                <p className="py-3 text-sm text-muted">{t("ref.emptyList")}</p>
              )}
            </div>
          </section>

          <section className="flex flex-col gap-4 rounded-2xl border border-line bg-surface p-6">
            <h2 className="text-lg font-semibold text-fg">
              {t("ref.parishesSection")}
            </h2>
            <form
              onSubmit={handleCreateParish}
              className="grid gap-3 sm:grid-cols-4"
            >
              <select
                value={pDioceseId || dioceses[0]?.id || ""}
                onChange={(e) => setPDioceseId(e.target.value)}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              >
                <option value="">{t("ref.selectDiocese")}</option>
                {dioceses.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.nama}
                  </option>
                ))}
              </select>
              <input
                required
                value={pCode}
                onChange={(e) => setPCode(e.target.value)}
                placeholder={t("ref.codePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <input
                required
                value={pName}
                onChange={(e) => setPName(e.target.value)}
                placeholder={t("ref.parishNamePlaceholder")}
                className="h-9 rounded-lg border border-line bg-canvas px-3 text-sm text-fg"
              />
              <button
                type="submit"
                disabled={createParish.isPending}
                className="inline-flex h-9 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 text-sm font-semibold text-on-primary disabled:opacity-50"
              >
                <Plus aria-hidden className="size-4" />
                {t("ref.addParish")}
              </button>
            </form>

            <div className="divide-y divide-line">
              {parishes.map((p) => (
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
                        await deleteParish.mutateAsync(p.id);
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
              {parishes.length === 0 && (
                <p className="py-3 text-sm text-muted">{t("ref.emptyList")}</p>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  );
}
