import { expect, type Locator } from "@playwright/test";
import { BasePage } from "../BasePage";

export class FinancePage extends BasePage {
  readonly path = "/fin";

  get unitSelector(): Locator {
    return this.page.locator("[data-testid='finance-unit-selector']");
  }

  get unitSelectInput(): Locator {
    return this.unitSelector.locator(".ant-select");
  }

  get noAccessAlert(): Locator {
    return this.page.locator("[data-testid='finance-no-access']");
  }

  async expectUnitSelectorVisible() {
    await expect(this.unitSelector).toBeVisible();
  }

  async expectSelectedUnit(name: string) {
    await expect(this.unitSelectInput).toContainText(name);
  }

  async selectUnit(name: string) {
    await this.unitSelectInput.click();
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
