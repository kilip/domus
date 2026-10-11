import { Given, Then, When } from "@cucumber/cucumber";
import { expect } from "@playwright/test";
import type { DomusWorld } from "../support/world";

Given("I am logged in as an admin", async function (this: DomusWorld) {
  await this.loginAs({ role: "admin" });
});

Given("I am logged in as a regular user", async function (this: DomusWorld) {
  await this.loginAs({ role: "user" });
});

When("I open {string}", async function (this: DomusWorld, path: string) {
  await this.page.goto(path);
});

When("I refresh the page", async function (this: DomusWorld) {
  await this.page.reload();
});

When("I confirm the deletion", async function (this: DomusWorld) {
  await this.page
    .locator(".ant-popconfirm")
    .getByRole("button", { name: "Hapus" })
    .click();
});

When("I cancel the deletion", async function (this: DomusWorld) {
  await this.page
    .locator(".ant-popconfirm")
    .getByRole("button", { name: "Batal" })
    .click();
});

Then(
  "I should be on {string}",
  async function (this: DomusWorld, path: string) {
    await this.loginPage.expectPath(path);
  },
);

Then(
  "I should see a success notification {string}",
  async function (this: DomusWorld, text: string) {
    await expect(
      this.page.locator(".ant-message-success").filter({ hasText: text }),
    ).toBeVisible();
  },
);

Then(
  "I should see an error or forbidden message",
  async function (this: DomusWorld) {
    await expect(this.page.locator("[role='alert']")).toBeVisible();
  },
);
