import assert from "node:assert/strict";
import { BasePage } from "../base.page";

export class OrgsPage extends BasePage {
  private cardFor(name: string) {
    return this.page
      .locator("div.grid > div")
      .filter({ has: this.page.getByRole("link", { name, exact: true }) });
  }

  async expectLoaded(): Promise<void> {
    await this.page
      .getByRole("heading", { name: "Direktori Organisasi" })
      .waitFor({ state: "visible", timeout: 10_000 });
  }

  async expectCreateButtonVisible(visible: boolean): Promise<void> {
    const btn = this.page.getByRole("button", { name: "Tambah Organisasi" });
    if (visible) {
      await btn.waitFor({ state: "visible", timeout: 10_000 });
    } else {
      assert.equal(await btn.count(), 0);
    }
  }

  async createOrg(input: {
    kindLabel?: string;
    name: string;
    code?: string;
    parentLabel?: string;
    zoneLabel?: string;
    wardLabel?: string;
    description?: string;
  }): Promise<void> {
    const submitBtn = this.page.getByRole("button", {
      name: "Simpan Organisasi",
    });
    if ((await submitBtn.count()) === 0) {
      await this.page
        .getByRole("button", { name: "Tambah Organisasi" })
        .click();
      await submitBtn.waitFor({ state: "visible", timeout: 10_000 });
    }

    assert.equal(
      await this.page.getByText("Jadikan saya Admin Organisasi ini").count(),
      0,
    );
    assert.equal(await this.page.getByText("Email Admin Awal").count(), 0);
    assert.equal(await this.page.getByText("Nama Admin Awal").count(), 0);

    if (input.kindLabel) {
      await this.page
        .getByLabel("Jenis Organisasi")
        .selectOption({ label: input.kindLabel });
    }
    await this.page.getByLabel("Nama Organisasi").fill(input.name);
    if (input.code !== undefined) {
      await this.page.getByLabel("Kode").fill(input.code);
    }
    if (input.parentLabel) {
      const parentSelect = this.page.getByLabel("Induk Organisasi (opsional)");
      await parentSelect
        .locator("option", { hasText: input.parentLabel })
        .waitFor({ state: "attached", timeout: 10_000 });
      await parentSelect.selectOption({ label: input.parentLabel });
    }
    if (input.zoneLabel) {
      const zoneSelect = this.page.getByLabel("Wilayah / Stasi (opsional)");
      await zoneSelect
        .locator("option", { hasText: input.zoneLabel })
        .waitFor({ state: "attached", timeout: 10_000 });
      await zoneSelect.selectOption({ label: input.zoneLabel });
    }
    if (input.wardLabel) {
      const wardSelect = this.page.getByLabel("Lingkungan (opsional)");
      await wardSelect
        .locator("option", { hasText: input.wardLabel })
        .waitFor({ state: "attached", timeout: 10_000 });
      await wardSelect.selectOption({ label: input.wardLabel });
    }
    if (input.description) {
      await this.page
        .getByLabel("Deskripsi (opsional)")
        .fill(input.description);
    }

    await submitBtn.click();
    await submitBtn.waitFor({ state: "detached", timeout: 10_000 });
  }

  async filterByKind(kindLabel: string): Promise<void> {
    await this.page
      .getByLabel("Semua Jenis")
      .selectOption({ label: kindLabel });
  }

  async search(query: string): Promise<void> {
    await this.page
      .getByPlaceholder("Cari nama atau kode organisasi…")
      .fill(query);
  }

  async toggleMyOrgs(myOrgsOnly: boolean): Promise<void> {
    await this.page
      .getByRole("button", {
        name: myOrgsOnly ? "Organisasi Saya" : "Semua Organisasi",
      })
      .click();
  }

  async expectOrgCardVisible(
    name: string,
    options?: {
      code?: string;
      kindBadge?: string;
      roleBadge?: "Admin Organisasi" | "Anggota";
    },
  ): Promise<void> {
    const card = this.cardFor(name);
    await card.waitFor({ state: "visible", timeout: 10_000 });
    assert.equal(
      await card.getByRole("button", { name: "Masuk sebagai Admin" }).count(),
      0,
    );
    if (options?.code) {
      await card
        .getByText(options.code, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.kindBadge) {
      await card
        .getByText(options.kindBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
    if (options?.roleBadge) {
      await card
        .getByText(options.roleBadge, { exact: true })
        .waitFor({ state: "visible", timeout: 10_000 });
    }
  }

  async openOrgDetail(name: string): Promise<void> {
    await this.page.getByRole("link", { name, exact: true }).click();
  }
}
