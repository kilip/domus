import { expect } from "@playwright/test";
import { BasePage } from "../BasePage";

export class WorkspacePage extends BasePage {
  readonly path = "/sys/workspace";
  readonly createPath = "/sys/workspace/create";

  get addButton() {
    return this.page.getByRole("button", {
      name: "Tambah workspace",
    });
  }

  get searchInput() {
    return this.page.getByPlaceholder(/cari/i);
  }

  get nameInput() {
    return this.page.locator("#name");
  }

  get parishSelect() {
    return this.page.locator("#parishId");
  }

  get activeSwitch() {
    return this.page.locator("#active");
  }

  get descriptionInput() {
    return this.page.locator("#description");
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
    await expect(this.page).toHaveURL(/\/sys\/workspace\/create$/);
  }

  async fillName(name: string) {
    await this.nameInput.fill(name);
  }

  async selectParish(parishName: string) {
    await this.parishSelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: parishName })
      .first()
      .click();
  }

  async setActive(checked: boolean) {
    const isChecked =
      (await this.activeSwitch.getAttribute("aria-checked")) === "true";
    if (isChecked !== checked) {
      await this.activeSwitch.click();
    }
  }

  async fillDescription(description: string) {
    await this.descriptionInput.fill(description);
  }

  async submitForm() {
    await this.saveButton.click();
  }

  async cancelForm() {
    await this.cancelButton.click();
  }

  async openEditPage(name: string) {
    await this.editButton(name).click();
    await expect(this.page).toHaveURL(/\/sys\/workspace\/update\/\d+$/);
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
  }

  async expectWorkspaceVisible(name: string) {
    await expect(this.row(name)).toBeVisible();
  }

  async expectWorkspaceWithStatusVisible(name: string, statusText: string) {
    const row = this.row(name);
    await expect(row).toBeVisible();
    await expect(row).toContainText(statusText);
  }

  async expectWorkspaceNotVisible(name: string) {
    await expect(this.row(name)).not.toBeVisible();
  }

  async expectSuccessToast(text?: string) {
    await expect(this.toastSuccess).toBeVisible();
    if (text) {
      await expect(this.toastSuccess).toContainText(text);
    }
  }

  async expectValidationError(text: string) {
    await expect(this.validationError.filter({ hasText: text })).toBeVisible();
  }
}
