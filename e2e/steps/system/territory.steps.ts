import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "territory {string} exists in parish {string} with kind {string}",
  async function (
    this: DomusWorld,
    name: string,
    parishName: string,
    kind: string,
  ) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created territory");
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
      sql`INSERT INTO public.territories (name, kind, status, parish_territories, created_at, updated_at, created_by, updated_by)
          VALUES (${name}, ${kind}, 'active', ${parishId}, NOW(), NOW(), ${userId}, ${userId})`,
    );
  },
);

When("I click the add territory button", async function (this: DomusWorld) {
  await this.territoryPage.openCreatePage();
});

When(
  "I fill the territory name with {string}",
  async function (this: DomusWorld, name: string) {
    await this.territoryPage.fillName(name);
  },
);

When(
  "I fill the territory code with {string}",
  async function (this: DomusWorld, code: string) {
    await this.territoryPage.fillCode(code);
  },
);

When(
  "I select territory kind {string} in territory form",
  async function (this: DomusWorld, kindLabel: string) {
    await this.territoryPage.selectKind(kindLabel);
  },
);

When(
  "I select parish {string} in territory form",
  async function (this: DomusWorld, parishName: string) {
    await this.territoryPage.selectParish(parishName);
  },
);

When(
  "I select parent territory {string} in territory form",
  async function (this: DomusWorld, parentName: string) {
    await this.territoryPage.selectParent(parentName);
  },
);

When(
  "I set the territory status to {string}",
  async function (this: DomusWorld, status: string) {
    const checked = status === "Aktif" || status === "active";
    await this.territoryPage.setActive(checked);
  },
);

When(
  "I fill the territory description with {string}",
  async function (this: DomusWorld, desc: string) {
    await this.territoryPage.fillDescription(desc);
  },
);

When("I submit the territory form", async function (this: DomusWorld) {
  await this.territoryPage.submitForm();
});

When("I cancel the territory form", async function (this: DomusWorld) {
  await this.territoryPage.cancelForm();
});

When(
  "I click edit for territory {string}",
  async function (this: DomusWorld, name: string) {
    await this.territoryPage.openEditPage(name);
  },
);

When(
  "I click delete for territory {string}",
  async function (this: DomusWorld, name: string) {
    await this.territoryPage.clickDelete(name);
  },
);

When(
  "I search territory for {string}",
  async function (this: DomusWorld, query: string) {
    await this.territoryPage.search(query);
  },
);

Then(
  "I should see territory {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.territoryPage.expectTerritoryVisible(name);
  },
);

Then(
  "I should see territory {string} with kind {string} in the list",
  async function (this: DomusWorld, name: string, kindText: string) {
    await this.territoryPage.expectTerritoryWithKindVisible(name, kindText);
  },
);

Then(
  "I should see territory {string} with status {string} in the list",
  async function (this: DomusWorld, name: string, statusText: string) {
    await this.territoryPage.expectTerritoryWithStatusVisible(name, statusText);
  },
);

Then(
  "I should not see territory {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.territoryPage.expectTerritoryNotVisible(name);
  },
);

Then(
  "I should see a territory validation error {string}",
  async function (this: DomusWorld, text: string) {
    await this.territoryPage.expectValidationError(text);
  },
);
