import type { AppDef } from "../app-def";

// Literal class names so Tailwind can see them at build time.
export const plateBg: Record<AppDef["plate"], string> = {
  1: "bg-plate-1",
  2: "bg-plate-2",
  3: "bg-plate-3",
  4: "bg-plate-4",
  5: "bg-plate-5",
  6: "bg-plate-6",
};
