import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "diocese {string} exists",
  async function (this: DomusWorld, name: string) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created diocese");
    }
    await authDB.execute(
      sql`INSERT INTO public.dioceses (name, created_at, updated_at, created_by, updated_by)
          VALUES (${name}, NOW(), NOW(), ${userId}, ${userId})`,
    );
  },
);

When("I click the add diocese button", async function (this: DomusWorld) {
  await this.diocesePage.openCreatePage();
});

When(
  "I fill the diocese name with {string}",
  async function (this: DomusWorld, name: string) {
    await this.diocesePage.fillName(name);
  },
);

When("I submit the diocese form", async function (this: DomusWorld) {
  await this.diocesePage.submitForm();
});

When("I cancel the diocese form", async function (this: DomusWorld) {
  await this.diocesePage.cancelForm();
});

When(
  "I click edit for diocese {string}",
  async function (this: DomusWorld, name: string) {
    await this.diocesePage.openEditPage(name);
  },
);

When(
  "I click delete for diocese {string}",
  async function (this: DomusWorld, name: string) {
    await this.diocesePage.clickDelete(name);
  },
);

When(
  "I search diocese for {string}",
  async function (this: DomusWorld, query: string) {
    await this.diocesePage.search(query);
  },
);

Then(
  "I should see diocese {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.diocesePage.expectDioceseVisible(name);
  },
);

Then(
  "I should not see diocese {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.diocesePage.expectDioceseNotVisible(name);
  },
);

Then(
  "I should see a diocese validation error {string}",
  async function (this: DomusWorld, text: string) {
    await this.diocesePage.expectValidationError(text);
  },
);
