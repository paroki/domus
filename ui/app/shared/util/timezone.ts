export const DEFAULT_TIMEZONE = "Asia/Makassar";

export const INDONESIA_TIMEZONES = [
  { value: "Asia/Jakarta", label: "Asia/Jakarta (WIB)", short: "WIB" },
  { value: "Asia/Makassar", label: "Asia/Makassar (WITA)", short: "WITA" },
  { value: "Asia/Jayapura", label: "Asia/Jayapura (WIT)", short: "WIT" },
] as const;

export interface TimezoneOption {
  value: string;
  label: string;
  short: string;
}

function isValidIANATimezone(tz: string): boolean {
  const trimmed = tz.trim();
  if (!trimmed) return false;
  try {
    Intl.DateTimeFormat(undefined, { timeZone: trimmed });
    return true;
  } catch {
    return false;
  }
}

export function detectBrowserTimezone(): string {
  try {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone?.trim();
    if (tz && tz !== "UTC" && tz !== "Etc/UTC" && isValidIANATimezone(tz)) {
      return tz;
    }
  } catch {
    // Fallback to default timezone if Intl is unavailable.
  }
  return DEFAULT_TIMEZONE;
}

export function getTimezoneOptions(currentValue?: string): TimezoneOption[] {
  const options: TimezoneOption[] = [...INDONESIA_TIMEZONES];
  const candidate = (currentValue?.trim() || detectBrowserTimezone()).trim();
  if (
    candidate &&
    !options.some((opt) => opt.value === candidate) &&
    isValidIANATimezone(candidate)
  ) {
    options.push({ value: candidate, label: candidate, short: candidate });
  }
  return options;
}
