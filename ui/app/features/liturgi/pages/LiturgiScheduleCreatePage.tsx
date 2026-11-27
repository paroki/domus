import { ArrowLeft, ShieldAlert } from "lucide-react";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import { useMe } from "~/shared/api/session";
import i18n from "~/shared/i18n";
import { Button, buttonClass } from "~/shared/ui/Button";
import { Card } from "~/shared/ui/Card";
import { InductiveFormSection, InductivePageHeader } from "~/shared/ui/iui";
import { useCreateLiturgiSchedule } from "../api/actions";
import { resolveLiturgiError } from "../api/errors";
import { useActiveParoki } from "../api/queries";
import {
  canManageLiturgiGlobal,
  DEFAULT_PAROKI_TIMEZONE,
  formatParokiTimezoneShort,
  fromLocalDatetimeInputValue,
} from "../model";
import type { Route } from "./+types/LiturgiScheduleCreatePage";

export function meta({}: Route.MetaArgs) {
  return [{ title: `${i18n.t("liturgi.createTitle")} · Domus` }];
}

export default function LiturgiScheduleCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const { data: activeParoki } = useActiveParoki();
  const createSchedule = useCreateLiturgiSchedule();

  const canManage = canManageLiturgiGlobal(me?.roles);
  const parokiTimezone = activeParoki?.timezone || DEFAULT_PAROKI_TIMEZONE;
  const tzShort = formatParokiTimezoneShort(parokiTimezone);

  const [title, setTitle] = useState("");
  const [location, setLocation] = useState("");
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [description, setDescription] = useState("");
  const [rsvpEnabled, setRsvpEnabled] = useState(true);
  const [rsvpDeadline, setRsvpDeadline] = useState("");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  if (me && !canManage) {
    return (
      <div className="mx-auto flex w-full max-w-2xl flex-col items-center gap-4 px-4 py-16 text-center">
        <span className="flex size-12 items-center justify-center rounded-xl bg-danger-bg text-danger-fg">
          <ShieldAlert aria-hidden className="size-6" />
        </span>
        <h1 className="text-xl font-bold text-fg">
          {t("liturgi.createForbiddenTitle")}
        </h1>
        <p className="text-sm text-muted">
          {t("liturgi.createForbiddenDescription")}
        </p>
        <Link to="/liturgi" className={buttonClass("outline")}>
          <ArrowLeft aria-hidden className="size-4" />
          {t("liturgi.backToSchedules")}
        </Link>
      </div>
    );
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setErrorMsg(null);

    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      setErrorMsg(t("liturgi.validationTitleRequired"));
      return;
    }
    if (!startAt || !endAt) {
      setErrorMsg(t("liturgi.validationStartEndRequired"));
      return;
    }

    const startIso = fromLocalDatetimeInputValue(startAt, parokiTimezone);
    const endIso = fromLocalDatetimeInputValue(endAt, parokiTimezone);
    if (!startIso || !endIso || new Date(endIso) <= new Date(startIso)) {
      setErrorMsg(t("liturgi.validationEndAfterStart"));
      return;
    }

    const deadlineIso =
      rsvpEnabled && rsvpDeadline
        ? fromLocalDatetimeInputValue(rsvpDeadline, parokiTimezone)
        : undefined;

    try {
      const created = await createSchedule.mutateAsync({
        title: trimmedTitle,
        location: location.trim() || undefined,
        startAt: startIso,
        endAt: endIso,
        description: description.trim() || undefined,
        rsvpEnabled,
        rsvpDeadline: deadlineIso || undefined,
      });
      if (created?.id) {
        navigate(`/liturgi/${created.id}`);
      } else {
        navigate("/liturgi");
      }
    } catch (err) {
      setErrorMsg(resolveLiturgiError(err));
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 sm:px-6">
      <InductivePageHeader
        breadcrumbs={[
          { label: t("liturgi.breadcrumbHome"), to: "/" },
          { label: t("liturgi.breadcrumbLiturgi"), to: "/liturgi" },
          { label: t("liturgi.createTitle") },
        ]}
        title={t("liturgi.createTitle")}
        subtitle={`${t("liturgi.createSubtitle")} (${tzShort})`}
        action={
          <Link to="/liturgi" className={buttonClass("outline")}>
            <ArrowLeft aria-hidden className="size-4" />
            {t("liturgi.backToSchedules")}
          </Link>
        }
      />

      <form onSubmit={handleSubmit} className="max-w-2xl space-y-5">
        {errorMsg && (
          <Card className="border-danger/30 bg-danger-bg px-4 py-3 text-sm text-danger-fg">
            {errorMsg}
          </Card>
        )}

        <InductiveFormSection title={t("liturgi.createTitle")}>
          <div className="space-y-4">
            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldTitle")}
              <input
                type="text"
                required
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={t("liturgi.fieldTitlePlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldLocation")}
              <input
                type="text"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder={t("liturgi.fieldLocationPlaceholder")}
                className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {`${t("liturgi.fieldStartAt")} (${tzShort})`}
                <input
                  type="datetime-local"
                  required
                  value={startAt}
                  onChange={(e) => setStartAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>

              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {`${t("liturgi.fieldEndAt")} (${tzShort})`}
                <input
                  type="datetime-local"
                  required
                  value={endAt}
                  onChange={(e) => setEndAt(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>
            </div>

            <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
              {t("liturgi.fieldDescription")}
              <textarea
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t("liturgi.fieldDescriptionPlaceholder")}
                className="w-full rounded-lg border border-line bg-canvas px-3.5 py-2.5 text-base text-fg placeholder:text-muted sm:text-sm"
              />
            </label>

            <label className="flex min-h-11 items-center gap-2.5 text-sm font-medium text-fg">
              <input
                type="checkbox"
                checked={rsvpEnabled}
                onChange={(e) => setRsvpEnabled(e.target.checked)}
                className="size-4 rounded border-line"
              />
              {t("liturgi.fieldRsvpEnabled")}
            </label>

            {rsvpEnabled && (
              <label className="flex flex-col gap-1.5 text-sm font-medium text-fg">
                {`${t("liturgi.fieldRsvpDeadline")} (${tzShort})`}
                <input
                  type="datetime-local"
                  value={rsvpDeadline}
                  onChange={(e) => setRsvpDeadline(e.target.value)}
                  className="h-11 w-full rounded-lg border border-line bg-canvas px-3.5 text-base text-fg sm:text-sm"
                />
              </label>
            )}
          </div>
        </InductiveFormSection>

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Link
            to="/liturgi"
            className={`${buttonClass("outline")} h-11 w-full sm:w-auto`}
          >
            {t("liturgi.ministriesCancel")}
          </Link>
          <Button
            type="submit"
            disabled={createSchedule.isPending}
            className="h-11 w-full sm:w-auto"
          >
            {createSchedule.isPending
              ? t("liturgi.creating")
              : t("liturgi.createSubmit")}
          </Button>
        </div>
      </form>
    </div>
  );
}
