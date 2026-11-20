import * as Sentry from "@sentry/react-router";
import { StrictMode, startTransition } from "react";
import { hydrateRoot } from "react-dom/client";
import { HydratedRouter } from "react-router/dom";

// No-op when the DSN is unset (e.g. local dev without Sentry).
Sentry.init({
  dsn: import.meta.env.APP_SENTRY_UI_DSN,
  environment: import.meta.env.APP_SENTRY_ENVIRONMENT || import.meta.env.MODE,
  release: import.meta.env.APP_SENTRY_RELEASE,
  integrations: [
    Sentry.reactRouterTracingIntegration(),
    Sentry.replayIntegration(),
  ],
  // ponytail: fixed 0.1 trace sample rate; read APP_SENTRY_TRACES_SAMPLE_RATE when tuning needed.
  tracesSampleRate: 0.1,
  tracePropagationTargets: [/^\//],
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
});

startTransition(() => {
  hydrateRoot(
    document,
    <StrictMode>
      <HydratedRouter onError={Sentry.sentryOnError} />
    </StrictMode>,
  );
});
