import { Given, Then, When } from "@cucumber/cucumber";
import { authDB } from "@domus/better-auth/auth";
import { expect } from "@playwright/test";
import { sql } from "drizzle-orm";
import type { DomusWorld } from "../../support/world";

Given(
  "I have finance access to workspace {string} in parish {string}",
  async function (this: DomusWorld, workspaceName: string, parishName: string) {
    const userId = this.userIds[0];
    if (!userId) {
      throw new Error("No active user to associate with created workspace");
    }

    const dioceseRes = await authDB.execute(
      sql`SELECT id FROM public.dioceses WHERE name = 'E2E Keuskupan Finance' LIMIT 1`,
    );
    let dioceseId: number;
    if (dioceseRes.rows.length === 0) {
      const insertDiocese = await authDB.execute(
        sql`INSERT INTO public.dioceses (name, active, created_at, updated_at, created_by, updated_by)
            VALUES ('E2E Keuskupan Finance', true, NOW(), NOW(), ${userId}, ${userId})
            RETURNING id`,
      );
      dioceseId = (insertDiocese.rows[0] as { id: number }).id;
    } else {
      dioceseId = (dioceseRes.rows[0] as { id: number }).id;
    }

    const parishRes = await authDB.execute(
      sql`SELECT id FROM public.parishes WHERE name = ${parishName} LIMIT 1`,
    );
    let parishId: number;
    if (parishRes.rows.length === 0) {
      const insertParish = await authDB.execute(
        sql`INSERT INTO public.parishes (name, active, diocese_parishes)
            VALUES (${parishName}, true, ${dioceseId})
            RETURNING id`,
      );
      parishId = (insertParish.rows[0] as { id: number }).id;
    } else {
      parishId = (parishRes.rows[0] as { id: number }).id;
    }

    const workspaceRes = await authDB.execute(
      sql`SELECT id FROM public.workspaces WHERE name = ${workspaceName} LIMIT 1`,
    );
    let workspaceId: number;
    if (workspaceRes.rows.length === 0) {
      const insertWorkspace = await authDB.execute(
        sql`INSERT INTO public.workspaces (name, status, parish_workspaces, created_at, updated_at, created_by, updated_by)
            VALUES (${workspaceName}, 'active', ${parishId}, NOW(), NOW(), ${userId}, ${userId})
            RETURNING id`,
      );
      workspaceId = (insertWorkspace.rows[0] as { id: number }).id;
    } else {
      workspaceId = (workspaceRes.rows[0] as { id: number }).id;
    }

    await authDB.execute(
      sql`INSERT INTO public.memberships (user_id, scope_type, scope_id, role, created_at)
          VALUES (${userId}, 'workspace', ${workspaceId}, 'finance-writer', NOW())
          ON CONFLICT DO NOTHING`,
    );
  },
);

When("I clear finance local storage", async function (this: DomusWorld) {
  await this.page.addInitScript(() => {
    localStorage.removeItem("activeFinanceWorkspaceId");
    localStorage.removeItem("financeWorkspaces");
  });
  if (this.page.url() !== "about:blank") {
    await this.page.evaluate(() => {
      localStorage.removeItem("activeFinanceWorkspaceId");
      localStorage.removeItem("financeWorkspaces");
    });
  }
});

Then(
  "I should see active finance workspace {string}",
  async function (this: DomusWorld, name: string) {
    await this.financePage.expectSelectedWorkspace(name);
  },
);

When(
  "I switch active finance workspace to {string}",
  async function (this: DomusWorld, name: string) {
    await this.financePage.selectWorkspace(name);
  },
);

Then(
  "localStorage activeFinanceWorkspaceId should be set for workspace {string}",
  async function (this: DomusWorld, workspaceName: string) {
    const workspaceRes = await authDB.execute(
      sql`SELECT id FROM public.workspaces WHERE name = ${workspaceName} LIMIT 1`,
    );
    const expectedId = String((workspaceRes.rows[0] as { id: number }).id);
    const storedId = await this.page.evaluate(() =>
      localStorage.getItem("activeFinanceWorkspaceId"),
    );
    expect(storedId).toBe(expectedId);
  },
);

Then(
  "I should see finance no access warning",
  async function (this: DomusWorld) {
    await this.financePage.expectNoAccessVisible();
  },
);
