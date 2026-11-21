import { Then } from "@cucumber/cucumber";
import type { DomusWorld } from "../../support/world";

Then("I see the login page", async function (this: DomusWorld) {
  await this.pages.user.login.expectLoaded();
});

Then(
  "I see the home page with admin launcher menus",
  async function (this: DomusWorld) {
    await this.pages.shell.home.expectRendered();
  },
);
