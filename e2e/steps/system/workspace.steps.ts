import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "workspace {string} exists in parish {string}",
  async function (this: DomusWorld, name: string, parishName: string) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created workspace");
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
      sql`INSERT INTO public.workspaces (name, status, parish_workspaces, created_at, updated_at, created_by, updated_by)
          VALUES (${name}, 'active', ${parishId}, NOW(), NOW(), ${userId}, ${userId})`,
    );
  },
);

When("I click the add workspace button", async function (this: DomusWorld) {
  await this.workspacePage.openCreatePage();
});

When(
  "I fill the workspace name with {string}",
  async function (this: DomusWorld, name: string) {
    await this.workspacePage.fillName(name);
  },
);

When(
  "I select parish {string} in workspace form",
  async function (this: DomusWorld, parishName: string) {
    await this.workspacePage.selectParish(parishName);
  },
);

When(
  "I set the workspace status to {string}",
  async function (this: DomusWorld, status: string) {
    const checked = status === "Aktif" || status === "active";
    await this.workspacePage.setActive(checked);
  },
);

When(
  "I fill the workspace description with {string}",
  async function (this: DomusWorld, desc: string) {
    await this.workspacePage.fillDescription(desc);
  },
);

When("I submit the workspace form", async function (this: DomusWorld) {
  await this.workspacePage.submitForm();
});

When("I cancel the workspace form", async function (this: DomusWorld) {
  await this.workspacePage.cancelForm();
});

When(
  "I click edit for workspace {string}",
  async function (this: DomusWorld, name: string) {
    await this.workspacePage.openEditPage(name);
  },
);

When(
  "I click delete for workspace {string}",
  async function (this: DomusWorld, name: string) {
    await this.workspacePage.clickDelete(name);
  },
);

When(
  "I search workspace for {string}",
  async function (this: DomusWorld, query: string) {
    await this.workspacePage.search(query);
  },
);

Then(
  "I should see workspace {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.workspacePage.expectWorkspaceVisible(name);
  },
);

Then(
  "I should see workspace {string} with status {string} in the list",
  async function (this: DomusWorld, name: string, statusText: string) {
    await this.workspacePage.expectWorkspaceWithStatusVisible(name, statusText);
  },
);

Then(
  "I should not see workspace {string} in the list",
  async function (this: DomusWorld, name: string) {
    await this.workspacePage.expectWorkspaceNotVisible(name);
  },
);

Then(
  "I should see a workspace validation error {string}",
  async function (this: DomusWorld, text: string) {
    await this.workspacePage.expectValidationError(text);
  },
);
