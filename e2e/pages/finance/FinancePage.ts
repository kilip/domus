import { expect, type Locator } from "@playwright/test";
import { BasePage } from "../BasePage";

export class FinancePage extends BasePage {
  readonly path = "/fin";

  get workspaceSelector(): Locator {
    return this.page.locator("[data-testid='finance-workspace-selector']");
  }

  get workspaceSelectInput(): Locator {
    return this.workspaceSelector.locator(".ant-select");
  }

  get noAccessAlert(): Locator {
    return this.page.locator("[data-testid='finance-no-access']");
  }

  async expectWorkspaceSelectorVisible() {
    await expect(this.workspaceSelector).toBeVisible();
  }

  async expectSelectedWorkspace(name: string) {
    await expect(this.workspaceSelectInput).toContainText(name);
  }

  async selectWorkspace(name: string) {
    await this.workspaceSelectInput.click();
    const dropdown = this.page.locator(
      ".ant-select-dropdown:not(.ant-select-dropdown-hidden)",
    );
    const option = dropdown
      .locator(".ant-select-item-option-content")
      .filter({ hasText: name });
    await option.click();
  }

  async expectNoAccessVisible() {
    await expect(this.noAccessAlert).toBeVisible();
    await expect(this.noAccessAlert).toContainText(
      "Anda tidak memiliki akses untuk modul finance",
    );
  }
}
