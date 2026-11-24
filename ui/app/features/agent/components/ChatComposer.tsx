import { Send, Square } from "lucide-react";
import { type FormEvent, type KeyboardEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/shared/ui/Button";
import { MAX_CHAT_MESSAGE_CHARS, validateChatMessage } from "../model";

interface ChatComposerProps {
  streaming: boolean;
  disabled?: boolean;
  onSend: (message: string) => void;
  onStop: () => void;
}

export function ChatComposer({
  streaming,
  disabled,
  onSend,
  onStop,
}: ChatComposerProps) {
  const { t } = useTranslation();
  const [text, setText] = useState("");
  const [err, setErr] = useState<string | null>(null);

  function submit() {
    if (streaming || disabled) return;
    const validationErr = validateChatMessage(text);
    if (validationErr) {
      setErr(validationErr);
      return;
    }
    setErr(null);
    const trimmed = text.trim();
    setText("");
    onSend(trimmed);
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    submit();
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="border-t border-line bg-surface px-4 py-3 sm:px-6"
    >
      <div className="mx-auto flex max-w-3xl flex-col gap-1.5">
        {err ? <p className="text-xs text-red-600">{err}</p> : null}
        <div className="flex items-end gap-2">
          <textarea
            rows={2}
            value={text}
            maxLength={MAX_CHAT_MESSAGE_CHARS}
            disabled={streaming || disabled}
            onChange={(e) => {
              setText(e.target.value);
              if (err) setErr(null);
            }}
            onKeyDown={handleKeyDown}
            placeholder={t("agent.composerPlaceholder")}
            className="min-h-11 max-h-40 flex-1 resize-none rounded-xl border border-line bg-bg px-3.5 py-2.5 text-sm text-fg placeholder:text-muted/70 focus:border-accent focus:outline-none disabled:opacity-60"
          />

          {streaming ? (
            <Button
              type="button"
              variant="outline"
              onClick={onStop}
              className="h-11 shrink-0 gap-1.5"
            >
              <Square aria-hidden className="size-4" />
              {t("agent.stop")}
            </Button>
          ) : (
            <Button
              type="submit"
              disabled={disabled || !text.trim()}
              className="h-11 shrink-0 gap-1.5"
            >
              <Send aria-hidden className="size-4" />
              {t("agent.send")}
            </Button>
          )}
        </div>
      </div>
    </form>
  );
}
