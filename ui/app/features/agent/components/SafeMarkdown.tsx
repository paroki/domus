import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { safeUrlTransform } from "../model";

interface SafeMarkdownProps {
  content: string;
}

export function SafeMarkdown({ content }: SafeMarkdownProps) {
  return (
    <div className="prose prose-sm max-w-none break-words text-fg dark:prose-invert">
      <Markdown
        remarkPlugins={[remarkGfm]}
        urlTransform={safeUrlTransform}
        components={{
          a: ({ href, children }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary underline underline-offset-2 hover:opacity-80"
            >
              {children}
            </a>
          ),
        }}
      >
        {content}
      </Markdown>
    </div>
  );
}
