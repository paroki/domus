import assert from "node:assert/strict";
import test from "node:test";
import {
  type BeforeInstallPromptEvent,
  getDeferredInstallPrompt,
  isIosDevice,
  isStandaloneMode,
  setDeferredInstallPrompt,
} from "./usePwaInstall";

test("isStandaloneMode detects display-mode standalone and iOS navigator.standalone", () => {
  assert.equal(isStandaloneMode(undefined), false);
  assert.equal(
    isStandaloneMode({
      matchMedia: () => ({ matches: true }) as MediaQueryList,
      navigator: {} as Navigator,
    }),
    true,
  );
  assert.equal(
    isStandaloneMode({
      matchMedia: () => ({ matches: false }) as MediaQueryList,
      navigator: { standalone: true } as unknown as Navigator,
    }),
    true,
  );
  assert.equal(
    isStandaloneMode({
      matchMedia: () => ({ matches: false }) as MediaQueryList,
      navigator: { standalone: false } as unknown as Navigator,
    }),
    false,
  );
});

test("isIosDevice detects iPhone, iPad, iPod, and iPadOS MacIntel touch devices", () => {
  assert.equal(isIosDevice(undefined), false);
  assert.equal(
    isIosDevice({
      userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)",
      platform: "iPhone",
      maxTouchPoints: 5,
    }),
    true,
  );
  assert.equal(
    isIosDevice({
      userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
      platform: "MacIntel",
      maxTouchPoints: 5,
    }),
    true,
  );
  assert.equal(
    isIosDevice({
      userAgent: "Mozilla/5.0 (Linux; Android 14)",
      platform: "Linux armv8l",
      maxTouchPoints: 5,
    }),
    false,
  );
});

test("deferred install prompt store sets and clears prompt event", () => {
  const fakeEvent = {
    prompt: async () => {},
    userChoice: Promise.resolve({ outcome: "accepted" as const }),
  } as unknown as BeforeInstallPromptEvent;

  setDeferredInstallPrompt(fakeEvent);
  assert.equal(getDeferredInstallPrompt(), fakeEvent);

  setDeferredInstallPrompt(null);
  assert.equal(getDeferredInstallPrompt(), null);
});
