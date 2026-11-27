import { Send, Square } from "lucide-react";
import {
  type FormEvent,
  type KeyboardEvent,
  useEffect,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/shared/ui/Button";
import { MAX_CHAT_MESSAGE_CHARS, validateChatMessage } from "../model";

interface ChatComposerProps {
  streaming: boolean;
  disabled?: boolean;
  focusKey?: string;
  onSend: (message: string) => void;
  onStop: () => void;
}

export function ChatComposer({
  streaming,
  disabled,
  focusKey,
  onSend,
  onStop,
}: ChatComposerProps) {
  const { t } = useTranslation();
  const [text, setText] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (disabled) return;
    if (
      typeof window !== "undefined" &&
      window.matchMedia?.("(pointer: coarse)").matches
    ) {
      return;
    }
    textareaRef.current?.focus();
  }, [focusKey, disabled]);

  function resizeTextarea() {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${Math.min(el.scrollHeight, 176)}px`;
  }

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
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
    onSend(trimmed);
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    submit();
  }

  function handleKeyDown(e: KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey && !e.nativeEvent.isComposing) {
      e.preventDefault();
      if (!streaming) {
        submit();
      }
    }
  }

  const showCharCount = text.length > MAX_CHAT_MESSAGE_CHARS * 0.8;

  return (
    <form
      onSubmit={handleSubmit}
      className="border-t border-line bg-surface px-4 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:px-6"
    >
      <div className="mx-auto flex max-w-3xl flex-col gap-1.5">
        {err ? <p className="text-xs text-danger">{err}</p> : null}

        <div className="flex flex-col rounded-2xl border border-line bg-bg shadow-2xs transition-colors focus-within:border-primary/60 focus-within:ring-2 focus-within:ring-primary/15">
          <textarea
            ref={textareaRef}
            rows={1}
            value={text}
            maxLength={MAX_CHAT_MESSAGE_CHARS}
            disabled={disabled}
            onChange={(e) => {
              setText(e.target.value);
              if (err) setErr(null);
              resizeTextarea();
            }}
            onKeyDown={handleKeyDown}
            placeholder={t("agent.composerPlaceholder")}
            className="min-h-11 max-h-44 w-full resize-none bg-transparent px-3.5 pt-3 pb-1 text-base text-fg placeholder:text-muted/70 focus:outline-none disabled:opacity-60 sm:text-sm"
          />

          <div className="flex items-center justify-between gap-2 px-3 pb-2">
            <div className="flex items-center gap-2 text-[11px] text-muted/80">
              <span className="hidden sm:inline">
                {t("agent.composerHint")}
              </span>
              {showCharCount ? (
                <span className="tabular-nums">
                  {text.length} / {MAX_CHAT_MESSAGE_CHARS}
                </span>
              ) : null}
            </div>

            {streaming ? (
              <Button
                type="button"
                variant="outline"
                onClick={onStop}
                className="h-11 shrink-0 gap-1.5 rounded-lg px-4 text-xs sm:text-sm"
              >
                <Square aria-hidden className="size-3.5" />
                {t("agent.stop")}
              </Button>
            ) : (
              <Button
                type="submit"
                disabled={disabled || !text.trim()}
                className="h-11 shrink-0 gap-1.5 rounded-lg px-4 text-xs sm:text-sm"
              >
                <Send aria-hidden className="size-3.5" />
                {t("agent.send")}
              </Button>
            )}
          </div>
        </div>
      </div>
    </form>
  );
}
