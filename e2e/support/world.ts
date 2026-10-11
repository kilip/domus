import {
  type IWorldOptions,
  setWorldConstructor,
  World,
} from "@cucumber/cucumber";
import type { BrowserContext, Page } from "@playwright/test";
import { LauncherPage } from "../pages/LauncherPage";
import { LoginPage } from "../pages/LoginPage";
import { DiocesePage } from "../pages/system/DiocesePage";
import { ParishPage } from "../pages/system/ParishPage";
import { TerritoryPage } from "../pages/system/TerritoryPage";
import { UnitPage } from "../pages/system/UnitPage";
import { createSession } from "./auth";

export class DomusWorld extends World {
  context!: BrowserContext;
  page!: Page;
  userIds: string[] = [];

  private _loginPage?: LoginPage;
  private _launcherPage?: LauncherPage;
  private _diocesePage?: DiocesePage;
  private _parishPage?: ParishPage;
  private _territoryPage?: TerritoryPage;
  private _unitPage?: UnitPage;

  constructor(options: IWorldOptions) {
    super(options);
  }

  get loginPage() {
    this._loginPage ??= new LoginPage(this.page);
    return this._loginPage;
  }

  get launcherPage() {
    this._launcherPage ??= new LauncherPage(this.page);
    return this._launcherPage;
  }

  get diocesePage() {
    this._diocesePage ??= new DiocesePage(this.page);
    return this._diocesePage;
  }

  get parishPage() {
    this._parishPage ??= new ParishPage(this.page);
    return this._parishPage;
  }

  get territoryPage() {
    this._territoryPage ??= new TerritoryPage(this.page);
    return this._territoryPage;
  }

  get unitPage() {
    this._unitPage ??= new UnitPage(this.page);
    return this._unitPage;
  }

  /** Login tanpa OAuth: bikin user via testUtils lalu pasang cookie sesi. */
  async loginAs(overrides: Record<string, unknown> = {}) {
    const { user, cookies } = await createSession(overrides);
    this.userIds.push(user.id);
    await this.context.addCookies(cookies);
    return user;
  }
}

setWorldConstructor(DomusWorld);
