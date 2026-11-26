import { Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../../support/world";

function splitCsv(raw: string): string[] {
  return raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

Then("I see the events page", async function (this: DomusWorld) {
  await this.pages.event.events.expectLoaded();
});

When(
  "I submit an event with title {string} start {string} end {string} visibility {string} location {string}",
  async function (
    this: DomusWorld,
    title: string,
    startAt: string,
    endAt: string,
    visibilityLabel: string,
    location: string,
  ) {
    await this.pages.event.events.createEvent({
      title,
      startAt,
      endAt,
      visibilityLabel: visibilityLabel as
        | "Publik Paroki"
        | "Internal Organisasi",
      location,
    });
  },
);

Then(
  "I see event form error {string}",
  async function (this: DomusWorld, message: string) {
    await this.pages.event.events.expectFormError(message);
  },
);

When(
  "I create an event with title {string} start {string} end {string} visibility {string} location {string}",
  async function (
    this: DomusWorld,
    title: string,
    startAt: string,
    endAt: string,
    visibilityLabel: string,
    location: string,
  ) {
    await this.pages.event.events.createEvent({
      title,
      startAt,
      endAt,
      visibilityLabel: visibilityLabel as
        | "Publik Paroki"
        | "Internal Organisasi",
      location,
    });
    await this.pages.event.events.expectFormClosed();
  },
);

Then(
  "I see event card {string} with visibility {string}",
  async function (this: DomusWorld, title: string, visibilityBadge: string) {
    await this.pages.event.events.expectEventCardVisible(title, {
      visibilityBadge,
    });
  },
);

Then(
  "I do not see event card {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.event.events.expectEventCardNotVisible(title);
  },
);

When(
  "I filter events by visibility {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.event.events.filterByVisibility(label);
  },
);

When(
  "I search events for {string}",
  async function (this: DomusWorld, query: string) {
    await this.pages.event.events.search(query);
  },
);

When(
  "I open event detail {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.event.events.openEventDetail(title);
  },
);

Then(
  "I see event detail {string}",
  async function (this: DomusWorld, title: string) {
    await this.pages.event.detail.expectLoaded(title);
  },
);

When(
  "I edit the event to title {string} location {string} visibility {string}",
  async function (
    this: DomusWorld,
    title: string,
    location: string,
    visibilityLabel: string,
  ) {
    await this.pages.event.detail.editEvent({
      title,
      location,
      visibilityLabel: visibilityLabel as
        | "Publik Paroki"
        | "Internal Organisasi",
    });
  },
);

Then(
  "I see event detail location {string}",
  async function (this: DomusWorld, location: string) {
    await this.pages.event.detail.expectLocationVisible(location);
  },
);

Then(
  "I see event badges {string}",
  async function (this: DomusWorld, badges: string) {
    await this.pages.event.detail.expectBadges(splitCsv(badges));
  },
);

When("I cancel the event", async function (this: DomusWorld) {
  await this.pages.event.detail.cancelEvent();
});

Then("I see event RSVP closed message", async function (this: DomusWorld) {
  await this.pages.event.detail.expectRsvpClosed();
});

When("I reschedule the event", async function (this: DomusWorld) {
  await this.pages.event.detail.rescheduleEvent();
});

Then("I see event RSVP open buttons", async function (this: DomusWorld) {
  await this.pages.event.detail.expectRsvpOpen();
});

When(
  "I create a parish-level event with title {string} start {string} end {string} location {string}",
  async function (
    this: DomusWorld,
    title: string,
    startAt: string,
    endAt: string,
    location: string,
  ) {
    await this.pages.event.events.createEvent({
      orgLabel: "Paroki (Tingkat Paroki)",
      title,
      startAt,
      endAt,
      location,
    });
    await this.pages.event.events.expectFormClosed();
  },
);

Then(
  "I see event card {string} with timezone {string}",
  async function (this: DomusWorld, title: string, tzLabel: string) {
    await this.pages.event.events.expectTimeRangeTimezone(title, tzLabel);
  },
);

When(
  "I filter events by organizer {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.event.events.filterByOrganizer(label);
  },
);

Then(
  "I see event detail timezone {string}",
  async function (this: DomusWorld, tzLabel: string) {
    await this.pages.event.detail.expectTimezoneVisible(tzLabel);
  },
);

When(
  "I create an event with title {string} jenis {string} start {string} end {string} visibility {string} location {string}",
  async function (
    this: DomusWorld,
    title: string,
    jenisLabel: string,
    startAt: string,
    endAt: string,
    visibilityLabel: string,
    location: string,
  ) {
    await this.pages.event.events.createEvent({
      title,
      jenisLabel,
      startAt,
      endAt,
      visibilityLabel: visibilityLabel as
        | "Publik Paroki"
        | "Internal Organisasi",
      location,
    });
    await this.pages.event.events.expectFormClosed();
  },
);

Then(
  "I see event card {string} with jenis {string}",
  async function (this: DomusWorld, title: string, jenisBadge: string) {
    await this.pages.event.events.expectEventCardVisible(title, {
      jenisBadge,
    });
  },
);

When(
  "I filter events by jenis {string}",
  async function (this: DomusWorld, label: string) {
    await this.pages.event.events.filterByJenis(label);
  },
);

When(
  "I edit the event jenis to {string}",
  async function (this: DomusWorld, jenisLabel: string) {
    await this.pages.event.detail.editEvent({ jenisLabel });
  },
);

When("I toggle Jadwal Tugas Saya", async function (this: DomusWorld) {
  await this.pages.event.events.toggleAssignedToMe();
});

When(
  "I add manual petugas role {string} name {string}",
  async function (this: DomusWorld, role: string, name: string) {
    await this.pages.event.detail.addPetugas({
      role,
      mode: "manual",
      name,
    });
  },
);

When(
  "I add org petugas role {string} for my organization",
  async function (this: DomusWorld, role: string) {
    const orgLabel = `${this.activeOrgName} (${this.activeOrgCode})`;
    await this.pages.event.detail.addPetugas({
      role,
      mode: "org",
      orgLabel,
    });
  },
);

When(
  "I confirm petugas {string} as {string}",
  async function (this: DomusWorld, role: string, statusLabel: string) {
    await this.pages.event.detail.confirmPetugas(
      role,
      statusLabel as "Siap Bertugas" | "Berhalangan",
    );
  },
);

Then(
  "I see petugas {string} assigned to {string} with status {string}",
  async function (
    this: DomusWorld,
    role: string,
    name: string,
    statusLabel: string,
  ) {
    await this.pages.event.detail.expectPetugasVisible(role, name, statusLabel);
  },
);

When(
  "I add lagu bagian {string} judul {string} sumber {string} nomor {string} nada {string}",
  async function (
    this: DomusWorld,
    bagian: string,
    judul: string,
    sumber: string,
    nomor: string,
    nadaDasar: string,
  ) {
    await this.pages.event.detail.addLagu({
      bagian,
      judul,
      sumber,
      nomor,
      nadaDasar,
    });
  },
);

Then(
  "I see lagu bagian {string} judul {string}",
  async function (this: DomusWorld, bagian: string, judul: string) {
    await this.pages.event.detail.expectLaguVisible(bagian, judul);
  },
);

When(
  "I edit lagu {string} to judul {string}",
  async function (this: DomusWorld, currentJudul: string, nextJudul: string) {
    await this.pages.event.detail.editLagu(currentJudul, nextJudul);
  },
);

When(
  "I delete lagu {string}",
  async function (this: DomusWorld, judul: string) {
    await this.pages.event.detail.deleteLagu(judul);
  },
);

When("I delete the event", async function (this: DomusWorld) {
  await this.pages.event.detail.deleteEvent();
});
