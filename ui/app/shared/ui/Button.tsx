import type { ButtonHTMLAttributes } from "react";

export type ButtonVariant = "primary" | "outline" | "ghost" | "danger";

const base =
  "inline-flex h-11 items-center justify-center gap-2 rounded-lg px-5 text-sm font-semibold no-underline transition-colors disabled:cursor-not-allowed disabled:opacity-50";

const variants: Record<ButtonVariant, string> = {
  primary: "bg-primary text-on-primary hover:opacity-90",
  outline:
    "border border-primary bg-transparent text-primary hover:bg-primary/10",
  ghost: "bg-transparent text-primary hover:bg-primary/10",
  danger: "bg-danger text-on-danger hover:opacity-90",
};

/** Class names for button styling; use on <Link> or <a> too. */
export function buttonClass(variant: ButtonVariant = "primary") {
  return `${base} ${variants[variant]}`;
}

export function Button({
  variant = "primary",
  type = "button",
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant }) {
  return (
    <button
      type={type}
      className={`${buttonClass(variant)} ${className}`}
      {...props}
    />
  );
}
