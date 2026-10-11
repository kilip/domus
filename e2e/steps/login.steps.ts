import { Then } from "@cucumber/cucumber";
import type { DomusWorld } from "../support/world";

Then("I should see the login form", async function (this: DomusWorld) {
  await this.loginPage.expectVisible();
});
