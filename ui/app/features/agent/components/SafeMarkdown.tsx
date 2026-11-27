import { Check, Copy } from "lucide-react";
import { isValidElement, type ReactNode, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import Markdown from "react-markdown";
import { Link } from "react-router";
import remarkGfm from "remark-gfm";
import { safeUrlTransform } from "../model";

interface SafeMarkdownProps {
  content: string;
}

function extractLanguage(children: ReactNode): string | null {
  if (!isValidElement<{ className?: string }>(children)) return null;
  const cls = children.props.className ?? "";
  const match = /language-([a-zA-Z0-9_-]+)/.exec(cls);
  return match?.[1] ?? null;
}

function CodeBlock({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  const preRef = useRef<HTMLPreElement>(null);
  const [copied, setCopied] = useState(false);
  const lang = extractLanguage(children);

  function handleCopy() {
    const text = preRef.current?.textContent ?? "";
    if (!text || typeof navigator === "undefined" || !navigator.clipboard) {
      return;
    }
    void navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  return (
    <div className="group/code my-2.5 overflow-hidden rounded-xl border border-line bg-canvas">
      <div className="flex items-center justify-between border-b border-line/70 bg-surface/60 px-3 py-1.5">
        <span className="font-mono text-[11px] font-medium uppercase tracking-wide text-muted">
          {lang ?? "text"}
        </span>
        <button
          type="button"
          onClick={handleCopy}
          aria-label={t("agent.copyCode")}
          className="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[11px] font-medium text-muted transition-colors hover:bg-fg/8 hover:text-fg"
        >
          {copied ? (
            <>
              <Check
                aria-hidden
                className="size-3 text-emerald-600 dark:text-emerald-400"
              />
              {t("agent.copiedMessage")}
            </>
          ) : (
            <>
              <Copy aria-hidden className="size-3" />
              {t("agent.copyCode")}
            </>
          )}
        </button>
      </div>
      <pre
        ref={preRef}
        className="m-0 overflow-x-auto bg-transparent px-3.5 py-3 font-mono text-xs leading-relaxed text-fg"
      >
        {children}
      </pre>
    </div>
  );
}

export function SafeMarkdown({ content }: SafeMarkdownProps) {
  return (
    <div className="max-w-none break-words text-sm leading-relaxed text-fg">
      <Markdown
        remarkPlugins={[remarkGfm]}
        urlTransform={safeUrlTransform}
        components={{
          p: ({ children }) => (
            <p className="my-2 first:mt-0 last:mb-0">{children}</p>
          ),
          h1: ({ children }) => (
            <h1 className="mt-4 mb-2 text-base font-bold text-fg first:mt-0">
              {children}
            </h1>
          ),
          h2: ({ children }) => (
            <h2 className="mt-3.5 mb-1.5 text-[15px] font-semibold text-fg first:mt-0">
              {children}
            </h2>
          ),
          h3: ({ children }) => (
            <h3 className="mt-3 mb-1 text-sm font-semibold text-fg first:mt-0">
              {children}
            </h3>
          ),
          h4: ({ children }) => (
            <h4 className="mt-2.5 mb-1 text-xs font-semibold uppercase tracking-wide text-muted first:mt-0">
              {children}
            </h4>
          ),
          ul: ({ children }) => (
            <ul className="my-2 list-disc space-y-1 pl-5 first:mt-0 last:mb-0">
              {children}
            </ul>
          ),
          ol: ({ children }) => (
            <ol className="my-2 list-decimal space-y-1 pl-5 first:mt-0 last:mb-0">
              {children}
            </ol>
          ),
          li: ({ children }) => <li className="leading-relaxed">{children}</li>,
          blockquote: ({ children }) => (
            <blockquote className="my-2.5 border-l-2 border-primary/50 pl-3 text-muted italic">
              {children}
            </blockquote>
          ),
          hr: () => <hr className="my-3 border-line" />,
          strong: ({ children }) => (
            <strong className="font-semibold text-fg">{children}</strong>
          ),
          a: ({ href, children }) =>
            href?.startsWith("/") && !href.startsWith("//") ? (
              <Link
                to={href}
                className="font-medium text-primary underline underline-offset-2 hover:opacity-80"
              >
                {children}
              </Link>
            ) : (
              <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-primary underline underline-offset-2 hover:opacity-80"
              >
                {children}
              </a>
            ),
          code: ({ className, children }) => {
            const isBlock = Boolean(className?.includes("language-"));
            if (isBlock) {
              return <code className={className}>{children}</code>;
            }
            return (
              <code className="rounded-md border border-line/70 bg-canvas px-1.5 py-0.5 font-mono text-[12px] text-fg">
                {children}
              </code>
            );
          },
          pre: ({ children }) => <CodeBlock>{children}</CodeBlock>,
          table: ({ children }) => (
            <div className="my-2.5 overflow-x-auto rounded-xl border border-line">
              <table className="m-0 w-full border-collapse text-xs">
                {children}
              </table>
            </div>
          ),
          thead: ({ children }) => (
            <thead className="bg-canvas/80 text-fg">{children}</thead>
          ),
          tr: ({ children }) => (
            <tr className="border-b border-line last:border-b-0">{children}</tr>
          ),
          th: ({ children }) => (
            <th className="px-3 py-2 text-left font-semibold text-fg">
              {children}
            </th>
          ),
          td: ({ children }) => (
            <td className="px-3 py-2 align-top text-fg">{children}</td>
          ),
        }}
      >
        {content}
      </Markdown>
    </div>
  );
}
