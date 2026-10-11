import { expect, type Page } from "@playwright/test";

export abstract class BasePage {
  abstract readonly path: string;

  constructor(protected readonly page: Page) {}

  async open(path?: string) {
    await this.page.goto(path ?? this.path);
  }

  async expectPath(path: string) {
    await expect(this.page).toHaveURL(
      new RegExp(`${path.replace(/\//g, "\\/")}(\\?.*)?$`),
      { timeout: 20_000 },
    );
  }
}
