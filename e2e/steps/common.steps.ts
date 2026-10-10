import { Given, Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../support/world";

Given("I am logged in as an admin", async function (this: DomusWorld) {
  await this.loginAs({ role: "admin" });
});

When("I open {string}", async function (this: DomusWorld, path: string) {
  await this.page.goto(path);
});

Then("I should be on {string}", async function (this: DomusWorld, path: string) {
  await this.loginPage.expectPath(path);
});
