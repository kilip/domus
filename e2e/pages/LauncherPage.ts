import { expect } from "@playwright/test";
import { BasePage } from "./BasePage";

export class LauncherPage extends BasePage {
  readonly path = "/";

  get heading() {
    return this.page.getByRole("heading", { name: "Semua modul" });
  }

  moduleLink(name: string) {
    return this.page.getByRole("link", { name, exact: true });
  }

  async expectVisible() {
    await expect(this.heading).toBeVisible();
  }

  async expectModules(names: string[]) {
    for (const name of names) await expect(this.moduleLink(name)).toBeVisible();
  }

  async openModule(name: string) {
    await this.moduleLink(name).click();
  }
}
