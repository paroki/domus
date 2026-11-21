import assert from "node:assert/strict";
import type { Locator } from "playwright";
import { BasePage } from "../base.page";

export class AgentWorkspacePage extends BasePage {
  private sidebar(): Locator {
    return this.page.locator("aside").first();
  }

  private memoryDrawer(): Locator {
    return this.page
      .locator("aside")
      .filter({
        has: this.page.getByRole("heading", { name: "Memori AI" }),
      })
      .first();
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Asisten AI" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectNoParishGuard(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Pilih Paroki Aktif" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async createNewSession(): Promise<void> {
    await this.page.getByTestId("agent-new-session").click();
  }

  async searchSessions(q: string): Promise<void> {
    await this.sidebar().getByPlaceholder("Cari percakapan…").fill(q);
  }

  async expectSessionInSidebar(title: string): Promise<void> {
    await this.sidebar()
      .getByText(title, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectSessionNotInSidebar(title: string): Promise<void> {
    await this.page.waitForFunction(
      ([t]) => {
        const aside = document.querySelector("aside");
        return !aside?.textContent?.includes(t);
      },
      [title] as const,
      { timeout: 10_000 },
    );
  }

  async selectSession(title: string): Promise<void> {
    await this.sidebar()
      .getByRole("button", { name: new RegExp(title) })
      .first()
      .click();
  }

  async renameSession(oldTitle: string, newTitle: string): Promise<void> {
    const row = this.sidebar()
      .locator("div.group")
      .filter({ hasText: oldTitle })
      .first();
    await row.scrollIntoViewIfNeeded();
    await row.hover();
    await row
      .getByRole("button", { name: "Ubah Judul" })
      .click({ force: true });
    const input = this.sidebar().locator('input[type="text"]').first();
    await input.fill(newTitle);
    await input.press("Enter");
    await this.expectSessionInSidebar(newTitle);
  }

  async deleteSession(title: string): Promise<void> {
    const row = this.sidebar()
      .locator("div.group")
      .filter({ hasText: title })
      .first();
    await row.scrollIntoViewIfNeeded();
    await row.hover();
    await row.getByRole("button", { name: "Hapus" }).click({ force: true });
    await this.expectSessionNotInSidebar(title);
  }

  async openMemoryDrawer(): Promise<void> {
    await this.page.getByRole("button", { name: "Memori AI" }).click();
    await this.memoryDrawer().waitFor({ state: "visible", timeout: 10_000 });
  }

  async closeMemoryDrawer(): Promise<void> {
    await this.memoryDrawer().getByRole("button", { name: "Tutup" }).click();
  }

  async switchMemoryScopeTab(
    tabName: "Memori Pribadi" | "FAQ & Pengetahuan Paroki",
  ): Promise<void> {
    await this.memoryDrawer().getByRole("tab", { name: tabName }).click();
  }

  async createMemory(input: {
    category: string;
    content: string;
    isPinned?: boolean;
  }): Promise<void> {
    const drawer = this.memoryDrawer();
    await drawer.getByRole("button", { name: "Tambah Memori" }).click();
    const form = drawer.locator("form");
    await form.locator("select").selectOption(input.category);
    await form.locator("textarea").fill(input.content);
    if (input.isPinned) {
      await form.locator('input[type="checkbox"]').check();
    }
    await form.getByRole("button", { name: "Simpan" }).click();
    await this.expectMemoryVisible(input.content);
  }

  async editMemory(
    oldContent: string,
    input: { content: string; category?: string },
  ): Promise<void> {
    const drawer = this.memoryDrawer();
    const card = drawer
      .locator("div.rounded-lg")
      .filter({ hasText: oldContent })
      .first();
    await card.getByRole("button", { name: "Ubah" }).click();
    const form = drawer.locator("form");
    if (input.category) {
      await form.locator("select").selectOption(input.category);
    }
    await form.locator("textarea").fill(input.content);
    await form.getByRole("button", { name: "Simpan" }).click();
    await this.expectMemoryVisible(input.content);
  }

  async togglePinMemory(content: string): Promise<void> {
    const card = this.memoryDrawer()
      .locator("div.rounded-lg")
      .filter({ hasText: content })
      .first();
    await card.getByRole("button", { name: "Pin" }).click();
  }

  async deleteMemory(content: string): Promise<void> {
    const card = this.memoryDrawer()
      .locator("div.rounded-lg")
      .filter({ hasText: content })
      .first();
    await card.getByRole("button", { name: "Hapus" }).click();
    await this.expectMemoryNotVisible(content);
  }

  async expectMemoryVisible(content: string): Promise<void> {
    await this.memoryDrawer()
      .getByText(content, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectMemoryNotVisible(content: string): Promise<void> {
    await this.page.waitForFunction(
      ([c]) => {
        const asides = Array.from(document.querySelectorAll("aside"));
        const drawer = asides.find((a) => a.textContent?.includes("Memori AI"));
        return !drawer?.textContent?.includes(c);
      },
      [content] as const,
      { timeout: 10_000 },
    );
  }

  async expectParishMemoryReadOnlyNotice(): Promise<void> {
    const drawer = this.memoryDrawer();
    await drawer
      .getByText(
        "Hanya Admin Paroki yang dapat menambah atau mengubah memori pengetahuan paroki.",
      )
      .waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await drawer.getByRole("button", { name: "Tambah Memori" }).count(),
      0,
    );
  }

  async sendMessage(prompt: string): Promise<void> {
    const composer = this.page.getByPlaceholder(
      "Tanyakan jadwal kegiatan, organisasi, atau informasi paroki…",
    );
    await composer.fill(prompt);
    await this.page.getByRole("button", { name: "Kirim" }).click();
  }

  async expectUserMessageVisible(text: string): Promise<void> {
    await this.page
      .locator("main")
      .getByText(text, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectAssistantMessageContaining(text: string): Promise<void> {
    await this.page
      .locator("main")
      .getByText(text, { exact: false })
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async expectThinkingBlockVisible(agentLabel?: string): Promise<void> {
    const pattern = agentLabel
      ? new RegExp(`Penalaran.*${agentLabel}`)
      : /Penalaran/;
    await this.page
      .locator("main")
      .getByText(pattern)
      .first()
      .waitFor({ state: "visible", timeout: 15_000 });
  }

  async expectToolExecutionBadge(toolName: string): Promise<void> {
    const summary = this.page
      .locator("main details summary")
      .filter({ hasText: toolName })
      .first();
    await summary.waitFor({ state: "visible", timeout: 15_000 });
    await summary.click();
  }
}
