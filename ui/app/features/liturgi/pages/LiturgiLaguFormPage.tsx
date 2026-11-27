import { ArrowLeft, Search, Unlink } from "lucide-react";
import { type FormEvent, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate, useParams } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Badge } from "~/shared/ui/Badge";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateLiturgiLagu, useUpdateLiturgiLagu } from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import {
  useLiturgiCantusSearch,
  useLiturgiSchedule,
  useLiturgiSongPartLookups,
} from "../api/queries";
import type { CantusItem } from "../api/types";
import { mergeCantusIntoLaguForm, resolveSongPartOptions } from "../model";
import type { Route } from "./+types/LiturgiLaguFormPage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.laguFormCreateTitle")} · Domus` }];
}

export default function LiturgiLaguFormPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { eventId = "", laguId } = useParams();
  const isEdit = Boolean(laguId);

  const { data: me } = useMe();
  const hasParish = Boolean(me?.activeParokiId);

  const scheduleQuery = useLiturgiSchedule(eventId, hasParish);
  const schedule = scheduleQuery.data;
  const songPartLookupsQuery = useLiturgiSongPartLookups(hasParish);
  const songPartOptions = useMemo(
    () => resolveSongPartOptions(songPartLookupsQuery.data ?? []),
    [songPartLookupsQuery.data],
  );

  const existingLagu = useMemo(
    () => schedule?.lagu?.find((l) => l.id === laguId),
    [schedule?.lagu, laguId],
  );

  const createLagu = useCreateLiturgiLagu(eventId);
  const updateLagu = useUpdateLiturgiLagu(eventId);

  const [bagian, setBagian] = useState("pembuka");
  const [urutan, setUrutan] = useState(0);
  const [formState, setFormState] = useState({
    cantusId: "",
    sumber: "",
    nomor: "",
    judul: "",
    nadaDasar: "",
    note: "",
  });
  const [cantusSearchInput, setCantusSearchInput] = useState("");
  const [debouncedCantusQ, setDebouncedCantusQ] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedCantusQ(cantusSearchInput.trim());
    }, 300);
    return () => window.clearTimeout(timer);
  }, [cantusSearchInput]);

  const cantusSearchQuery = useLiturgiCantusSearch(debouncedCantusQ, hasParish);
  const cantusResults = cantusSearchQuery.data ?? [];

  useEffect(() => {
    if (isEdit && existingLagu) {
      setBagian(existingLagu.bagian ?? songPartOptions[0]?.value ?? "pembuka");
      setUrutan(existingLagu.urutan ?? 0);
      setFormState({
        cantusId: existingLagu.cantusId ?? "",
        sumber: existingLagu.sumber ?? "",
        nomor: existingLagu.nomor ?? "",
        judul: existingLagu.judul ?? "",
        nadaDasar: existingLagu.nadaDasar ?? "",
        note: existingLagu.note ?? "",
      });
    }
  }, [isEdit, existingLagu, songPartOptions]);

  function handleSelectCantus(item: CantusItem) {
    setFormState((prev) => mergeCantusIntoLaguForm(prev, item));
  }

  function handleClearCantus() {
    setFormState((prev) => ({ ...prev, cantusId: "" }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    const trimmedJudul = formState.judul.trim();
    if (!trimmedJudul) {
      setErrorMsg(t("liturgi.validationJudulRequired"));
      return;
    }

    try {
      if (isEdit && laguId) {
        const hadCantus = Boolean(existingLagu?.cantusId);
        const clearedCantus = hadCantus && !formState.cantusId;
        await updateLagu.mutateAsync({
          laguId,
          body: {
            bagian,
            urutan,
            cantusId: formState.cantusId || undefined,
            clearCantusId: clearedCantus ? true : undefined,
            sumber: formState.sumber.trim(),
            nomor: formState.nomor.trim(),
            judul: trimmedJudul,
            nadaDasar: formState.nadaDasar.trim(),
            note: formState.note.trim(),
          },
        });
      } else {
        await createLagu.mutateAsync({
          bagian,
          urutan,
          cantusId: formState.cantusId || undefined,
          sumber: formState.sumber.trim() || undefined,
          nomor: formState.nomor.trim() || undefined,
          judul: trimmedJudul,
          nadaDasar: formState.nadaDasar.trim() || undefined,
          note: formState.note.trim() || undefined,
        });
      }
      navigate(`/liturgi/${eventId}`);
    } catch (err) {
      setErrorMsg(resolveLiturgiError(err));
    }
  }

  const pageTitle = isEdit
    ? t("liturgi.laguFormEditTitle")
    : t("liturgi.laguFormCreateTitle");
  const pending = createLagu.isPending || updateLagu.isPending;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi"), to: "/liturgi" },
          {
            label: schedule?.title ?? t("liturgi.detailTitle"),
            to: `/liturgi/${eventId}`,
          },
          { label: pageTitle },
        ]}
        title={pageTitle}
        subtitle={t("liturgi.laguFormSubtitle")}
        action={
          <Link
            to={`/liturgi/${eventId}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            <ArrowLeft aria-hidden className="size-4" />
            {t("liturgi.backToDetail")}
          </Link>
        }
      />

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-5">
        {errorMsg && (
          <Card className="border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
            {errorMsg}
          </Card>
        )}

        <InductiveFormSection title={t("liturgi.laguCantusSearchTitle")}>
          <div className="space-y-3">
            <div className="relative">
              <Search
                aria-hidden
                className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted"
              />
              <input
                type="search"
                value={cantusSearchInput}
                onChange={(e) => setCantusSearchInput(e.target.value)}
                placeholder={t("liturgi.laguCantusSearchPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas pr-3.5 pl-10 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </div>

            {formState.cantusId && (
              <div className="flex flex-col gap-2 rounded-lg border border-line bg-surface-2/40 p-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="flex items-center gap-2">
                  <Badge tone="success">{t("liturgi.laguCantusLinked")}</Badge>
                  <span className="text-sm font-medium text-fg">
                    {formState.judul}
                  </span>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleClearCantus}
                  className="h-11 w-full sm:w-auto"
                >
                  <Unlink aria-hidden className="size-4" />
                  {t("liturgi.laguCantusUnlink")}
                </Button>
              </div>
            )}

            {cantusResults.length === 0 ? (
              <p className="text-sm text-muted">
                {t("liturgi.laguCantusEmpty")}
              </p>
            ) : (
              <ul className="max-h-56 space-y-2 overflow-y-auto pr-1">
                {cantusResults.map((item) => {
                  const selected = item.id === formState.cantusId;
                  return (
                    <li key={item.id}>
                      <button
                        type="button"
                        onClick={() => handleSelectCantus(item)}
                        className={`flex w-full items-center justify-between gap-3 rounded-lg border px-3 py-2.5 text-left transition-colors ${
                          selected
                            ? "border-primary bg-primary/10"
                            : "border-line bg-canvas hover:bg-surface-2/50"
                        }`}
                      >
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium text-fg">
                            {item.judul}
                          </p>
                          <p className="text-xs text-muted">
                            {[
                              item.sumber,
                              item.nomor,
                              item.nadaDasar ? `1=${item.nadaDasar}` : "",
                            ]
                              .filter(Boolean)
                              .join(" · ")}
                          </p>
                        </div>
                        {selected && <Badge tone="success">Dipilih</Badge>}
                      </button>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </InductiveFormSection>

        <InductiveFormSection title={pageTitle}>
          <div className="space-y-4">
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.laguFieldBagian")}
                <select
                  value={bagian}
                  onChange={(e) => setBagian(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                >
                  {songPartOptions.map((opt) => (
                    <option key={opt.value} value={opt.value}>
                      {opt.label}
                    </option>
                  ))}
                  {bagian &&
                    !songPartOptions.some((opt) => opt.value === bagian) && (
                      <option value={bagian}>{bagian}</option>
                    )}
                </select>
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.laguFieldUrutan")}
                <input
                  type="number"
                  value={urutan}
                  onChange={(e) => setUrutan(Number(e.target.value) || 0)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.laguFieldSumber")}
                <input
                  type="text"
                  value={formState.sumber}
                  onChange={(e) =>
                    setFormState((p) => ({ ...p, sumber: e.target.value }))
                  }
                  placeholder={t("liturgi.laguFieldSumberPlaceholder")}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {t("liturgi.laguFieldNomor")}
                <input
                  type="text"
                  value={formState.nomor}
                  onChange={(e) =>
                    setFormState((p) => ({ ...p, nomor: e.target.value }))
                  }
                  placeholder={t("liturgi.laguFieldNomorPlaceholder")}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
                />
              </label>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.laguFieldJudul")}
              <input
                type="text"
                required
                value={formState.judul}
                onChange={(e) =>
                  setFormState((p) => ({ ...p, judul: e.target.value }))
                }
                placeholder={t("liturgi.laguFieldJudulPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.laguFieldNadaDasar")}
              <input
                type="text"
                value={formState.nadaDasar}
                onChange={(e) =>
                  setFormState((p) => ({ ...p, nadaDasar: e.target.value }))
                }
                placeholder={t("liturgi.laguFieldNadaDasarPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.laguFieldNote")}
              <textarea
                rows={3}
                value={formState.note}
                onChange={(e) =>
                  setFormState((p) => ({ ...p, note: e.target.value }))
                }
                placeholder={t("liturgi.laguFieldNotePlaceholder")}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>
          </div>
        </InductiveFormSection>

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Link
            to={`/liturgi/${eventId}`}
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            {t("liturgi.ministriesCancel")}
          </Link>
          <Button
            type="submit"
            disabled={pending}
            className="h-11 w-full sm:w-auto"
          >
            {pending ? t("liturgi.creating") : t("liturgi.laguFormSave")}
          </Button>
        </div>
      </form>
    </div>
  );
}
