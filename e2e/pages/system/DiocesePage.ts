import { expect } from "@playwright/test";
import { BasePage } from "../BasePage";

export class DiocesePage extends BasePage {
  readonly path = "/sys/diocese";
  readonly createPath = "/sys/diocese/create";

  get addButton() {
    return this.page.getByRole("button", { name: "Tambah keuskupan" });
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
    await expect(this.page).toHaveURL(/\/sys\/diocese\/create$/);
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

  async submitForm() {
    await this.saveButton.click();
  }

  async cancelForm() {
    await this.cancelButton.click();
  }

  async openEditPage(name: string) {
    await this.editButton(name).click();
    await expect(this.page).toHaveURL(/\/sys\/diocese\/update\/\d+$/);
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

  async expectDioceseVisible(name: string) {
    await expect(this.row(name)).toBeVisible();
  }

  async expectDioceseWithStatusVisible(name: string, statusText: string) {
    const rowLocator = this.row(name);
    await expect(rowLocator).toBeVisible();
    await expect(
      rowLocator.getByText(statusText, { exact: true }),
    ).toBeVisible();
  }

  async expectDioceseNotVisible(name: string) {
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
