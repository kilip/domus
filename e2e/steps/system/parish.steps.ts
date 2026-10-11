import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "parish {string} exists",
  async function (this: DomusWorld, name: string) {
    await authDB.execute(
      sql`INSERT INTO public.parishes (name)
          VALUES (${name})`,
    );
  },
);

Given(
  "parish {string} exists under diocese {string}",
  async function (this: DomusWorld, parishName: string, dioceseName: string) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created diocese");
    }
    const result = await authDB.execute(
      sql`INSERT INTO public.dioceses (name, created_at, updated_at, created_by, updated_by)
          VALUES (${dioceseName}, NOW(), NOW(), ${userId}, ${userId})
          RETURNING id`,
    );
    const dioceseId = (result.rows[0] as { id: number }).id;
    await authDB.execute(
      sql`INSERT INTO public.parishes (name, diocese_parishes)
          VALUES (${parishName}, ${dioceseId})`,
    );
  },
);

Given(
  "parish {string} with church {string} exists",
  async function (this: DomusWorld, name: string, church: string) {
    await authDB.execute(
      sql`INSERT INTO public.parishes (name, church)
          VALUES (${name}, ${church})`,
    );
  },
);

When("I click the add parish button", async function (this: DomusWorld) {
  await this.parishPage.openCreatePage();
});

When(
  "I fill the parish name with {string}",
  async function (this: DomusWorld, name: string) {
    await this.parishPage.fillName(name);
  },
);

When(
  "I fill the parish church with {string}",
  async function (this: DomusWorld, church: string) {
    await this.parishPage.fillChurch(church);
  },
);

When(
  "I fill the parish address with {string}",
  async function (this: DomusWorld, address: string) {
    await this.parishPage.fillAddress(address);
  },
);

When(
  "I fill the parish phone with {string}",
  async function (this: DomusWorld, phone: string) {
    await this.parishPage.fillPhone(phone);
  },
);

When(
  "I fill the parish email with {string}",
  async function (this: DomusWorld, email: string) {
    await this.parishPage.fillEmail(email);
  },
);

When(
  "I select diocese {string} in parish form",
  async function (this: DomusWorld, dioceseName: string) {
    await this.parishPage.dioceseSelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: dioceseName })
      .first()
      .click();
  },
);

When("I submit the parish form", async function (this: DomusWorld) {
  await this.parishPage.submitForm();
});

When("I cancel the parish form", async function (this: DomusWorld) {
  await this.parishPage.cancelForm();
});

When(
  "I click edit for parish {string}",
  async function (this: DomusWorld, name: string) {
    await this.parishPage.openEditPage(name);
  },
);

When(
  "I click delete for parish {string}",
  async function (this: DomusWorld, name: string) {
    await this.parishPage.clickDelete(name);
  },
);

When(
  "I search parish for {string}",
  async function (this: DomusWorld, query: string) {
    await this.parishPage.search(query);
  },
);

Then(
  "I should see parish {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.parishPage.expectParishVisible(name);
  },
);

Then(
  "I should see parish {string} with church {string} in the list",
  async function (this: DomusWorld, name: string, church: string) {
    await this.parishPage.expectParishWithChurchVisible(name, church);
  },
);

Then(
  "I should not see parish {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.parishPage.expectParishNotVisible(name);
  },
);

Then(
  "I should see a parish validation error {string}",
  async function (this: DomusWorld, text: string) {
    await this.parishPage.expectValidationError(text);
  },
);
