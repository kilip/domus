import { expect } from "@playwright/test";
import { BasePage } from "../BasePage";

export class ParishPage extends BasePage {
  readonly path = "/sys/parish";
  readonly createPath = "/sys/parish/create";

  get addButton() {
    return this.page.getByRole("button", { name: "Tambah paroki" });
  }

  get searchInput() {
    return this.page.getByPlaceholder(/cari/i);
  }

  get nameInput() {
    return this.page.locator("#name");
  }

  get activeSwitch() {
    return this.page.locator("#active");
  }

  get churchInput() {
    return this.page.locator("#church");
  }

  get dioceseSelect() {
    return this.page.locator("#dioceseId");
  }

  get addressInput() {
    return this.page.locator("#address");
  }

  get phoneInput() {
    return this.page.locator("#phone");
  }

  get emailInput() {
    return this.page.locator("#email");
  }

  get saveButton() {
    return this.page.getByRole("button", { name: "Simpan" });
  }

  get cancelButton() {
    return this.page.getByRole("button", { name: "Batal" });
  }

  get validationError() {
    return this.page.locator(".ant-form-item-explain-error");
  }

  get popconfirm() {
    return this.page.locator(".ant-popconfirm");
  }

  get popconfirmOkButton() {
    return this.popconfirm.getByRole("button", { name: "Hapus" });
  }

  get popconfirmCancelButton() {
    return this.popconfirm.getByRole("button", { name: "Batal" });
  }

  get toastSuccess() {
    return this.page.locator(".ant-message-success");
  }

  get alertError() {
    return this.page.locator("[role='alert']");
  }

  row(name: string) {
    return this.page.locator(".ant-table-row").filter({
      has: this.page.getByRole("cell", { name, exact: true }),
    });
  }

  editButton(name: string) {
    return this.row(name).locator("button[aria-label='Ubah']");
  }

  deleteButton(name: string) {
    return this.row(name).locator("button[aria-label='Hapus']");
  }

  async openCreatePage() {
    await this.addButton.click();
    await expect(this.page).toHaveURL(/\/sys\/parish\/create$/);
  }

  async fillName(name: string) {
    await this.nameInput.fill(name);
  }

  async setActive(checked: boolean) {
    const isChecked =
      (await this.activeSwitch.getAttribute("aria-checked")) === "true";
    if (isChecked !== checked) {
      await this.activeSwitch.click();
    }
  }

  async fillChurch(church: string) {
    await this.churchInput.fill(church);
  }

  async fillAddress(address: string) {
    await this.addressInput.fill(address);
  }

  async fillPhone(phone: string) {
    await this.phoneInput.fill(phone);
  }

  async fillEmail(email: string) {
    await this.emailInput.fill(email);
  }

  async submitForm() {
    await this.saveButton.click();
  }

  async cancelForm() {
    await this.cancelButton.click();
  }

  async openEditPage(name: string) {
    await this.editButton(name).click();
    await expect(this.page).toHaveURL(/\/sys\/parish\/update\/\d+$/);
  }

  async clickDelete(name: string) {
    await this.deleteButton(name).click();
    await expect(this.popconfirm).toBeVisible();
  }

  async confirmDelete() {
    await this.popconfirmOkButton.click();
  }

  async cancelDelete() {
    await this.popconfirmCancelButton.click();
  }

  async search(query: string) {
    await this.searchInput.fill(query);
    await this.searchInput.press("Enter");
  }

  async expectParishVisible(name: string) {
    await expect(this.row(name)).toBeVisible();
  }

  async expectParishWithStatusVisible(name: string, statusText: string) {
    const rowLocator = this.row(name);
    await expect(rowLocator).toBeVisible();
    await expect(
      rowLocator.getByText(statusText, { exact: true }),
    ).toBeVisible();
  }

  async expectParishWithChurchVisible(name: string, church: string) {
    const rowLocator = this.row(name);
    await expect(rowLocator).toBeVisible();
    await expect(
      rowLocator.getByRole("cell", { name: church, exact: true }),
    ).toBeVisible();
  }

  async expectParishNotVisible(name: string) {
    await expect(this.row(name)).not.toBeVisible();
  }

  async expectSuccessMessage(text: string) {
    await expect(this.toastSuccess.filter({ hasText: text })).toBeVisible();
  }

  async expectValidationError(text: string) {
    await expect(this.validationError).toHaveText(text);
  }

  async expectForbiddenOrError() {
    await expect(this.alertError).toBeVisible();
  }
}
