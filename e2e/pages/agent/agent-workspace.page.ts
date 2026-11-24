import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class AgentWorkspacePage extends BasePage {
  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("button", { name: "Percakapan Baru" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page
      .getByPlaceholder(
        "Tanyakan jadwal kegiatan, organisasi, atau informasi paroki…",
      )
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async sendMessage(text: string): Promise<void> {
    const composer = this.page.getByPlaceholder(
      "Tanyakan jadwal kegiatan, organisasi, atau informasi paroki…",
    );
    await composer.fill(text);
    await this.page.getByRole("button", { name: "Kirim" }).click();
  }

  async expectHandoff(from: string, to: string): Promise<void> {
    await this.page
      .locator("span")
      .filter({ hasText: from })
      .filter({ hasText: to })
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async expectToolBadge(toolName: string): Promise<void> {
    await this.page
      .getByText(toolName, { exact: true })
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async expectAssistantReply(textIncludes: string): Promise<void> {
    await this.page
      .getByText(textIncludes)
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async expectSessionUrlAndReload(expectedReply: string): Promise<void> {
    await this.page.waitForURL(/\/agent\/sessions\/[0-9a-f-]+$/, {
      timeout: 15_000,
    });
    await this.page
      .getByRole("button", { name: "Kirim" })
      .waitFor({ state: "visible", timeout: 15_000 });
    await this.page.reload({ waitUntil: "domcontentloaded" });
    await this.expectAssistantReply(expectedReply);
  }

  async sendSlowStreamAndAbort(text: string): Promise<void> {
    await this.sendMessage(text);
    const stopBtn = this.page.getByRole("button", { name: "Hentikan" });
    await stopBtn.waitFor({ state: "visible", timeout: 10_000 });
    await stopBtn.click();
    await stopBtn.waitFor({ state: "detached", timeout: 10_000 });
    await this.page
      .getByRole("button", { name: "Kirim" })
      .waitFor({ state: "visible", timeout: 10_000 });
    await this.page.waitForTimeout(300);
    const alertCount = await this.page
      .locator("main")
      .getByRole("alert")
      .count();
    assert.equal(
      alertCount,
      0,
      "Expected no error banner after clicking Hentikan",
    );
  }

  async renameActiveSession(newTitle: string): Promise<void> {
    const row = this.page.locator("aside li").first();
    await row.hover();
    await row.getByRole("button", { name: "Ubah Judul" }).click();
    const input = this.page.locator("aside li input[type='text']").first();
    await input.fill(newTitle);
    await input.press("Enter");
    await this.page
      .locator("aside")
      .getByRole("link", { name: newTitle })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteActiveSession(title: string): Promise<void> {
    const row = this.page
      .locator("aside li")
      .filter({ hasText: title })
      .first();
    await row.hover();
    await row.getByRole("button", { name: "Hapus" }).click();
    await this.page.waitForURL(/\/agent$/, { timeout: 10_000 });
  }

  async openMemoryDrawer(): Promise<void> {
    await this.page.getByRole("button", { name: "Memori AI" }).click();
    await this.page
      .getByRole("heading", { name: "Memori AI" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async addMemory(content: string): Promise<void> {
    const drawer = this.page.getByRole("region", { name: "Memori AI" });
    await drawer
      .getByPlaceholder("Tulis fakta atau preferensi (maks. 2000 karakter)…")
      .fill(content);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/agent/memories") &&
          r.request().method() === "POST" &&
          r.status() === 201,
      ),
      drawer.getByRole("button", { name: "Tambah Memori" }).click(),
    ]);
    await drawer
      .getByText(content)
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async searchMemory(query: string, expectedText: string): Promise<void> {
    const drawer = this.page.getByRole("region", { name: "Memori AI" });
    await drawer.getByPlaceholder("Cari memori…").fill(query);
    await drawer
      .getByText(expectedText)
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async editMemory(oldText: string, newContent: string): Promise<void> {
    const drawer = this.page.getByRole("region", { name: "Memori AI" });
    const item = drawer.locator("li").filter({ hasText: oldText }).first();
    await item.getByRole("button", { name: "Ubah Memori" }).click();
    await drawer
      .getByPlaceholder("Tulis fakta atau preferensi (maks. 2000 karakter)…")
      .fill(newContent);
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/agent/memories/") &&
          r.request().method() === "PATCH" &&
          r.status() === 200,
      ),
      drawer.getByRole("button", { name: "Simpan Memori" }).click(),
    ]);
    await drawer
      .getByText(newContent)
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async deleteMemory(text: string): Promise<void> {
    const drawer = this.page.getByRole("region", { name: "Memori AI" });
    const item = drawer.locator("li").filter({ hasText: text }).first();
    await Promise.all([
      this.page.waitForResponse(
        (r) =>
          r.url().includes("/api/agent/memories/") &&
          r.request().method() === "DELETE" &&
          r.status() === 204,
      ),
      item.getByRole("button", { name: "Hapus" }).click(),
    ]);
    await item.waitFor({ state: "detached", timeout: 10_000 });
  }
}
