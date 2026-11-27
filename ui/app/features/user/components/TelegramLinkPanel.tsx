import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Copy, ExternalLink, RefreshCw, Send } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { ApiError } from "~/shared/api/client";
import { Button, buttonClass } from "~/shared/ui/Button";
import { useCreateTelegramLinkToken } from "../api/actions";
import { userKeys } from "../api/keys";
import { meQueryOptions, useMyAccounts } from "../api/queries";
import type { TelegramLinkTokenResponse, UserAccount } from "../api/types";
import { QrCodeSvg } from "./QrCodeSvg";

interface TelegramLinkPanelProps {
  autoStart?: boolean;
  isAlreadyLinked?: boolean;
  onLinked?: () => void;
}

function telegramAccountKey(accounts: UserAccount[]): string | null {
  const acc = accounts.find((a) => a.provider === "telegram");
  if (!acc) return null;
  return `${acc.id ?? ""}:${acc.subject ?? ""}:${acc.createdAt ?? ""}`;
}

export function TelegramLinkPanel({
  autoStart = false,
  isAlreadyLinked = false,
  onLinked,
}: TelegramLinkPanelProps) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const createToken = useCreateTelegramLinkToken();
  const { data: accountsData } = useMyAccounts();

  const [token, setToken] = useState<TelegramLinkTokenResponse | null>(null);
  const [awaitingLink, setAwaitingLink] = useState(false);
  const [copied, setCopied] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const baselineRef = useRef<{
    hadTelegram: boolean;
    accountKey: string | null;
  } | null>(null);
  const autoStartedRef = useRef(false);

  const meQuery = useQuery({
    ...meQueryOptions(),
    refetchInterval: awaitingLink ? 3000 : false,
  });

  const generateToken = useCallback(async () => {
    setErrorMsg(null);
    setCopied(false);
    const currentAccounts = accountsData?.data ?? [];
    const currentMe = meQuery.data;
    baselineRef.current = {
      hadTelegram:
        isAlreadyLinked ||
        Boolean(currentMe?.linkedProviders?.includes("telegram")) ||
        currentAccounts.some((a) => a.provider === "telegram"),
      accountKey: telegramAccountKey(currentAccounts),
    };

    try {
      const res = await createToken.mutateAsync();
      setToken(res);
      setAwaitingLink(true);
    } catch (err) {
      setAwaitingLink(false);
      if (err instanceof ApiError) {
        setErrorMsg(err.message);
      }
    }
  }, [accountsData?.data, createToken, isAlreadyLinked, meQuery.data]);

  useEffect(() => {
    if (autoStart && !autoStartedRef.current) {
      autoStartedRef.current = true;
      void generateToken();
    }
  }, [autoStart, generateToken]);

  useEffect(() => {
    if (!awaitingLink || !baselineRef.current?.hadTelegram) return;
    if (meQuery.dataUpdatedAt === 0) return;
    void qc.invalidateQueries({ queryKey: userKeys.accounts() });
  }, [awaitingLink, meQuery.dataUpdatedAt, qc]);

  useEffect(() => {
    if (!awaitingLink || !baselineRef.current) return;
    const baseline = baselineRef.current;

    const freshLinked =
      !baseline.hadTelegram &&
      Boolean(meQuery.data?.linkedProviders?.includes("telegram"));

    const currentKey = telegramAccountKey(accountsData?.data ?? []);
    const relinked =
      baseline.hadTelegram &&
      currentKey !== null &&
      currentKey !== baseline.accountKey;

    if (freshLinked || relinked) {
      setAwaitingLink(false);
      void Promise.all([
        qc.invalidateQueries({ queryKey: userKeys.accounts() }),
        qc.invalidateQueries({ queryKey: userKeys.me() }),
      ]);
      onLinked?.();
    }
  }, [
    accountsData?.data,
    awaitingLink,
    meQuery.data?.linkedProviders,
    onLinked,
    qc,
  ]);

  async function handleCopy() {
    if (!token?.deepLinkUrl) return;
    try {
      await navigator.clipboard.writeText(token.deepLinkUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API unavailable; user can still click the link or scan QR.
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {errorMsg && (
        <div
          role="alert"
          className="rounded-lg bg-danger-bg px-3.5 py-2.5 text-sm font-medium text-danger-fg"
        >
          {errorMsg}
        </div>
      )}

      {!token?.deepLinkUrl ? (
        <div className="flex items-center gap-3">
          <Button
            type="button"
            disabled={createToken.isPending}
            onClick={() => void generateToken()}
          >
            <Send aria-hidden className="size-4" />
            {createToken.isPending
              ? t("me.telegramGenerating")
              : t("me.telegramGenerate")}
          </Button>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          <p className="text-sm text-muted">
            {token.botUsername
              ? t("me.telegramHint", { bot: `@${token.botUsername}` })
              : t("me.telegramHintNoBot")}
          </p>

          <a
            href={token.deepLinkUrl}
            target="_blank"
            rel="noopener noreferrer"
            className={`${buttonClass("primary")} w-full`}
          >
            <ExternalLink aria-hidden className="size-4" />
            {t("me.telegramOpenBot")}
          </a>

          <div className="self-center overflow-hidden rounded-xl border border-line bg-white p-2">
            <QrCodeSvg
              value={token.deepLinkUrl}
              size={176}
              label={t("me.telegramQrAlt")}
            />
          </div>

          <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-center">
            <Button
              type="button"
              variant="outline"
              onClick={handleCopy}
              className="w-full sm:w-auto"
            >
              {copied ? (
                <Check aria-hidden className="size-4" />
              ) : (
                <Copy aria-hidden className="size-4" />
              )}
              {copied ? t("me.telegramCopied") : t("me.telegramCopyLink")}
            </Button>

            <Button
              type="button"
              variant="ghost"
              disabled={createToken.isPending}
              onClick={() => void generateToken()}
              className="w-full sm:w-auto"
            >
              <RefreshCw aria-hidden className="size-4" />
              {t("me.telegramRegenerate")}
            </Button>
          </div>

          {awaitingLink && (
            <p className="text-center text-xs font-medium text-muted">
              {t("me.telegramWaiting")}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
