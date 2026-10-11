import { expect } from "@playwright/test";
import { BasePage } from "./BasePage";

export class LoginPage extends BasePage {
  readonly path = "/login";

  get heading() {
    return this.page.getByRole("heading", { name: "Masuk ke Domus" });
  }
  get googleButton() {
    return this.page.getByRole("button", { name: "Lanjutkan dengan Google" });
  }
  get githubButton() {
    return this.page.getByRole("button", { name: "Lanjutkan dengan GitHub" });
  }

  async expectVisible() {
    await expect(this.heading).toBeVisible();
    await expect(this.googleButton).toBeVisible();
    await expect(this.githubButton).toBeVisible();
  }
}
