import assert from "node:assert/strict";
import { Given, Then, When } from "@cucumber/cucumber";
import { TestClient } from "../../support/test-client";
import type { DomusWorld } from "../../support/world";

const lastSeededEventIds = new WeakMap<DomusWorld, string>();

async function seedEvent(
  world: DomusWorld,
  token: string,
  orgId: string,
  visibility: "public" | "private",
  title: string,
  rsvpEnabled: boolean,
): Promise<string> {
  assert.ok(world.activeParishId, "activeParishId must be set");
  const adminClient = new TestClient();
  adminClient.setToken(token);
  adminClient.setParishId(world.activeParishId);
  const start = new Date(Date.now() + 24 * 60 * 60 * 1000);
  const end = new Date(start.getTime() + 2 * 60 * 60 * 1000);
  const res = await adminClient.post<{ data?: { id?: string } }>("/events", {
    orgId,
    title,
    visibility,
    startAt: start.toISOString(),
    endAt: end.toISOString(),
    location: "Aula Paroki",
    rsvpEnabled,
  });
  assert.equal(
    res.status,
    201,
    `Expected 201 seeding event "${title}", got ${res.status}: ${JSON.stringify(res.data)}`,
  );
  const eventId = res.data?.data?.id;
  assert.ok(eventId, `Missing event id for "${title}"`);
  return eventId;
}

When("I open my active organization detail", async function (this: DomusWorld) {
  assert.ok(this.activeOrgId, "activeOrgId must be set");
  await this.pages.org.detail.goto(`/org/${this.activeOrgId}`);
});

Then(
  "I see organization detail without admin controls or invites",
  async function (this: DomusWorld) {
    assert.ok(this.activeOrgName, "activeOrgName must be set");
    await this.pages.org.detail.expectLoaded(this.activeOrgName);
    await this.pages.org.detail.expectEditAndDeleteVisibility({
      canEdit: false,
      canDelete: false,
    });
    await this.pages.org.detail.expectInvitesSectionVisible(false);
  },
);

When(
  "I open the sibling organization detail",
  async function (this: DomusWorld) {
    assert.ok(this.otherOrgId, "otherOrgId must be set");
    await this.pages.org.detail.goto(`/org/${this.otherOrgId}`);
  },
);

Then(
  "I see the organization roster restricted banner",
  async function (this: DomusWorld) {
    await this.pages.org.detail.expectRosterRestrictedBanner();
  },
);

When("I leave my active organization", async function (this: DomusWorld) {
  await this.pages.org.detail.leaveOrganization();
});

Given(
  "my organization has a {string} event {string} with RSVP enabled",
  async function (this: DomusWorld, visibility: string, title: string) {
    assert.ok(this.orgAdminToken, "orgAdminToken must be set");
    assert.ok(this.activeOrgId, "activeOrgId must be set");
    const eventId = await seedEvent(
      this,
      this.orgAdminToken,
      this.activeOrgId,
      visibility === "private" ? "private" : "public",
      title,
      true,
    );
    lastSeededEventIds.set(this, eventId);
  },
);

Given(
  "the sibling organization has a {string} event {string}",
  async function (this: DomusWorld, visibility: string, title: string) {
    assert.ok(this.orgAdminToken, "orgAdminToken must be set");
    assert.ok(this.otherOrgId, "otherOrgId must be set");
    await seedEvent(
      this,
      this.orgAdminToken,
      this.otherOrgId,
      visibility === "private" ? "private" : "public",
      title,
      true,
    );
  },
);

Then("I do not see the create event button", async function (this: DomusWorld) {
  await this.pages.event.events.expectCreateButtonVisible(false);
});

Then("I see the RSVP restricted banner", async function (this: DomusWorld) {
  await this.pages.event.detail.expectRsvpRestrictedBanner();
});

When(
  "I submit RSVP {string} with note {string}",
  async function (this: DomusWorld, label: string, note: string) {
    await this.pages.event.detail.submitRsvp(
      label as "Hadir" | "Mungkin" | "Tidak Hadir",
      note || undefined,
    );
  },
);

Then(
  "I see my RSVP status {string} with note {string}",
  async function (this: DomusWorld, statusLabel: string, note: string) {
    await this.pages.event.detail.expectRsvpStatus(
      statusLabel,
      note || undefined,
    );
  },
);

Then(
  "I see my RSVP status {string}",
  async function (this: DomusWorld, statusLabel: string) {
    await this.pages.event.detail.expectRsvpStatus(statusLabel);
  },
);

Then(
  "I see RSVP summary counts going {int} maybe {int} notGoing {int}",
  async function (
    this: DomusWorld,
    going: number,
    maybe: number,
    notGoing: number,
  ) {
    await this.pages.event.detail.expectSummaryCounts({
      going,
      maybe,
      notGoing,
    });
  },
);

Then(
  "I see RSVP attendee {string} with status {string}",
  async function (this: DomusWorld, name: string, statusLabel: string) {
    await this.pages.event.detail.expectAttendeeVisible(name, statusLabel);
  },
);

When("I remove my RSVP", async function (this: DomusWorld) {
  await this.pages.event.detail.removeRsvp();
});

Then(
  "I see my RSVP removed for {string}",
  async function (this: DomusWorld, memberName: string) {
    await this.pages.event.detail.expectRsvpRemoved(memberName);
  },
);

When(
  "the organization admin cancels the seeded event via API",
  async function (this: DomusWorld) {
    assert.ok(this.orgAdminToken, "orgAdminToken must be set");
    assert.ok(this.activeParishId, "activeParishId must be set");
    const eventId = lastSeededEventIds.get(this);
    assert.ok(eventId, "seeded event id must be set");
    const adminClient = new TestClient();
    adminClient.setToken(this.orgAdminToken);
    adminClient.setParishId(this.activeParishId);
    const res = await adminClient.patch(`/events/${eventId}`, {
      status: "cancelled",
    });
    assert.equal(
      res.status,
      200,
      `Expected 200 cancelling event, got ${res.status}: ${JSON.stringify(res.data)}`,
    );
    await this.pages.event.detail.goto(`/event/${eventId}`);
  },
);
