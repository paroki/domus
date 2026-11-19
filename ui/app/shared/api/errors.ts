export const BASE_ERRORS: Record<string, string> = {
  BAD_REQUEST: "The request could not be processed.",
  INVALID_CURSOR: "Invalid or expired pagination cursor.",
  UNAUTHORIZED: "Authentication is required to continue.",
  FORBIDDEN: "You do not have permission to perform this action.",
  NOT_FOUND: "The requested resource was not found.",
  CONFLICT: "This action conflicts with the current state.",
  GONE: "This resource is no longer available.",
  VALIDATION_ERROR: "One or more fields failed validation.",
  RATE_LIMITED: "Too many requests. Please try again shortly.",
  INTERNAL_ERROR: "An unexpected internal error occurred.",
};

export function resolveErrorMessage(
  code: string,
  fallbackMessage?: string,
  featureErrors?: Record<string, string>,
): string {
  return (
    featureErrors?.[code] ??
    BASE_ERRORS[code] ??
    fallbackMessage ??
    "An unexpected error occurred."
  );
}
