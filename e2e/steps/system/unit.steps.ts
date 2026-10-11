import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "unit {string} exists in parish {string} with kind {string}",
  async function (
    this: DomusWorld,
    name: string,
    parishName: string,
    kind: string,
  ) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created unit");
    }

    const parishRes = await authDB.execute(
      sql`SELECT id FROM public.parishes WHERE name = ${parishName} LIMIT 1`,
    );
    let parishId: number;
    if (parishRes.rows.length === 0) {
      const insertParish = await authDB.execute(
        sql`INSERT INTO public.parishes (name, active, created_at, updated_at, created_by, updated_by)
            VALUES (${parishName}, true, NOW(), NOW(), ${userId}, ${userId})
            RETURNING id`,
      );
      parishId = (insertParish.rows[0] as { id: number }).id;
    } else {
      parishId = (parishRes.rows[0] as { id: number }).id;
    }

    await authDB.execute(
      sql`INSERT INTO public.units (name, kind, status, parish_units, created_at, updated_at, created_by, updated_by)
          VALUES (${name}, ${kind}, 'active', ${parishId}, NOW(), NOW(), ${userId}, ${userId})`,
    );
  },
);

When("I click the add unit button", async function (this: DomusWorld) {
  await this.unitPage.openCreatePage();
});

When(
  "I fill the unit name with {string}",
  async function (this: DomusWorld, name: string) {
    await this.unitPage.fillName(name);
  },
);

When(
  "I fill the unit code with {string}",
  async function (this: DomusWorld, code: string) {
    await this.unitPage.fillCode(code);
  },
);

When(
  "I select unit kind {string} in unit form",
  async function (this: DomusWorld, kindLabel: string) {
    await this.unitPage.selectKind(kindLabel);
  },
);

When(
  "I select parish {string} in unit form",
  async function (this: DomusWorld, parishName: string) {
    await this.unitPage.selectParish(parishName);
  },
);

When(
  "I select territory {string} in unit form",
  async function (this: DomusWorld, territoryName: string) {
    await this.unitPage.selectTerritory(territoryName);
  },
);

When(
  "I select parent unit {string} in unit form",
  async function (this: DomusWorld, parentName: string) {
    await this.unitPage.selectParent(parentName);
  },
);

When(
  "I set the unit status to {string}",
  async function (this: DomusWorld, status: string) {
    const checked = status === "Aktif" || status === "active";
    await this.unitPage.setActive(checked);
  },
);

When(
  "I fill the unit description with {string}",
  async function (this: DomusWorld, desc: string) {
    await this.unitPage.fillDescription(desc);
  },
);

When("I submit the unit form", async function (this: DomusWorld) {
  await this.unitPage.submitForm();
});

When("I cancel the unit form", async function (this: DomusWorld) {
  await this.unitPage.cancelForm();
});

When(
  "I click edit for unit {string}",
  async function (this: DomusWorld, name: string) {
    await this.unitPage.openEditPage(name);
  },
);

When(
  "I click delete for unit {string}",
  async function (this: DomusWorld, name: string) {
    await this.unitPage.clickDelete(name);
  },
);

When(
  "I search unit for {string}",
  async function (this: DomusWorld, query: string) {
    await this.unitPage.search(query);
  },
);

Then(
  "I should see unit {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.unitPage.expectUnitVisible(name);
  },
);

Then(
  "I should see unit {string} with kind {string} in the list",
  async function (this: DomusWorld, name: string, kindText: string) {
    await this.unitPage.expectUnitWithKindVisible(name, kindText);
  },
);

Then(
  "I should see unit {string} with status {string} in the list",
  async function (this: DomusWorld, name: string, statusText: string) {
    await this.unitPage.expectUnitWithStatusVisible(name, statusText);
  },
);

Then(
  "I should not see unit {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.unitPage.expectUnitNotVisible(name);
  },
);

Then(
  "I should see a unit validation error {string}",
  async function (this: DomusWorld, text: string) {
    await this.unitPage.expectValidationError(text);
  },
);
