import { getInitials } from "../model";

export { useDebouncedValue } from "~/shared/util/use-debounced-value";

export function UserAvatar({
  name,
  email,
  highlight = false,
}: {
  name?: string | null;
  email?: string | null;
  highlight?: boolean;
}) {
  return (
    <span
      aria-hidden
      className={`flex size-10 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
        highlight ? "bg-primary/15 text-primary" : "bg-surface-2 text-fg"
      }`}
    >
      {getInitials(name, email)}
    </span>
  );
}
