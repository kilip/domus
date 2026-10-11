import { Then, When } from "@cucumber/cucumber";
import type { DomusWorld } from "../support/world";

Then("I should see the launcher", async function (this: DomusWorld) {
  await this.launcherPage.expectVisible();
});

Then(
  "I should see modules {string}",
  async function (this: DomusWorld, csv: string) {
    await this.launcherPage.expectModules(csv.split(",").map((s) => s.trim()));
  },
);

When(
  "I click module {string}",
  async function (this: DomusWorld, name: string) {
    await this.launcherPage.openModule(name);
  },
);
