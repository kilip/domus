import { expect } from "@playwright/test";
import { BasePage } from "../BasePage";

export class UnitPage extends BasePage {
  readonly path = "/sys/unit";
  readonly createPath = "/sys/unit/create";

  get addButton() {
    return this.page.getByRole("button", {
      name: "Tambah unit organisasi",
    });
  }

  get searchInput() {
    return this.page.getByPlaceholder(/cari/i);
  }

  get nameInput() {
    return this.page.locator("#name");
  }

  get codeInput() {
    return this.page.locator("#code");
  }

  get kindSelect() {
    return this.page.locator("#kind");
  }

  get parishSelect() {
    return this.page.locator("#parishId");
  }

  get territorySelect() {
    return this.page.locator("#territoryId");
  }

  get parentSelect() {
    return this.page.locator("#parentId");
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
    await expect(this.page).toHaveURL(/\/sys\/unit\/create$/);
  }

  async fillName(name: string) {
    await this.nameInput.fill(name);
  }

  async fillCode(code: string) {
    await this.codeInput.fill(code);
  }

  async selectKind(kindLabel: string) {
    await this.kindSelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: kindLabel })
      .first()
      .click();
  }

  async selectParish(parishName: string) {
    await this.parishSelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: parishName })
      .first()
      .click();
  }

  async selectTerritory(territoryName: string) {
    await this.territorySelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: territoryName })
      .first()
      .click();
  }

  async selectParent(parentName: string) {
    await this.parentSelect.click();
    await this.page
      .locator(".ant-select-item-option-content")
      .filter({ hasText: parentName })
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

  async fillDescription(desc: string) {
    await this.descriptionInput.fill(desc);
  }

  async submitForm() {
    await this.saveButton.click();
  }

  async cancelForm() {
    await this.cancelButton.click();
  }

  async openEditPage(name: string) {
    await this.editButton(name).click();
    await expect(this.page).toHaveURL(/\/sys\/unit\/update\/\d+$/);
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

  async expectUnitVisible(name: string) {
    await expect(this.row(name)).toBeVisible();
  }

  async expectUnitWithStatusVisible(name: string, statusText: string) {
    const rowLocator = this.row(name);
    await expect(rowLocator).toBeVisible();
    await expect(
      rowLocator.getByText(statusText, { exact: true }),
    ).toBeVisible();
  }

  async expectUnitWithKindVisible(name: string, kindText: string) {
    const rowLocator = this.row(name);
    await expect(rowLocator).toBeVisible();
    await expect(rowLocator.getByText(kindText, { exact: true })).toBeVisible();
  }

  async expectUnitNotVisible(name: string) {
    await expect(this.row(name)).not.toBeVisible();
  }

  async expectSuccessMessage(text: string) {
    await expect(this.toastSuccess.filter({ hasText: text })).toBeVisible();
  }

  async expectValidationError(text: string) {
    await expect(this.validationError.filter({ hasText: text })).toBeVisible();
  }

  async expectForbiddenOrError() {
    await expect(this.alertError).toBeVisible();
  }
}
