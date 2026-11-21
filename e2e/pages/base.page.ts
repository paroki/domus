import assert from "node:assert/strict";
import type { Page } from "playwright";

export class BasePage {
  constructor(protected readonly page: Page) {}

  async goto(path: string): Promise<void> {
    await this.page.goto(path, { waitUntil: "domcontentloaded" });
  }

  async expectPath(expectedPath: string): Promise<void> {
    await this.page.waitForURL((url) => url.pathname === expectedPath, {
      timeout: 10_000,
    });
    const current = new URL(this.page.url());
    assert.equal(current.pathname, expectedPath);
  }
}
