import { Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../support/world";

When("I open {string}", async function (this: DomusWorld, path: string) {
  await this.pages.base.goto(path);
});

Then("I am on {string}", async function (this: DomusWorld, path: string) {
  await this.pages.base.expectPath(path);
});
