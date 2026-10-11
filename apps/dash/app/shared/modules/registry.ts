import {
  ApartmentOutlined,
  BookOutlined,
  CalendarOutlined,
  GlobalOutlined,
  TeamOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import type { DomusModule, FlatMenuItem, ModuleMenuItem, TKey } from "./types";

/**
 * Registry modul Domus (gaya Odoo).
 *
 * `name`, `description`, dan `label` adalah key terjemahan (lihat `~/i18n/locales`).
 * Tambah modul baru = tambah entri di sini + key-nya di `id.ts` dan `en.ts`.
 *
 * Tambah modul baru cukup dengan menambah satu entri di sini: launcher,
 * navbar, dan halaman placeholder-nya otomatis ikut. Kalau modul sudah
 * punya halaman sungguhan, bikin `features/<Modul>/routes.ts` lalu daftarkan
 * di `app/routes.ts` (route modul menang atas route dinamis `$moduleId/$page`).
 *
 * Belum ada filter role: semua modul dan menu tampil untuk semua user.
 */
export const modules: readonly DomusModule[] = [
  {
    id: "sys",
    name: "modules.sys.name",
    description: "modules.sys.description",
    icon: ApartmentOutlined,
    color: "var(--domus-app-sys)",
    menu: [
      { label: "modules.sys.menu.diocese", slug: "diocese" },
      { label: "modules.sys.menu.parish", slug: "parish" },
      { label: "modules.sys.menu.territorial", slug: "territorial" },
    ],
  },
  {
    id: "web",
    name: "modules.web.name",
    description: "modules.web.description",
    icon: GlobalOutlined,
    color: "var(--domus-app-web)",
    menu: [
      { label: "modules.web.menu.pages", slug: "pages" },
      { label: "modules.web.menu.news", slug: "news" },
      { label: "modules.web.menu.announcements", slug: "announcements" },
      { label: "modules.web.menu.gallery", slug: "gallery" },
      { label: "modules.web.menu.settings", slug: "settings" },
    ],
  },
  {
    id: "sacra",
    name: "modules.sacra.name",
    description: "modules.sacra.description",
    icon: BookOutlined,
    color: "var(--domus-app-sacra)",
    menu: [
      { label: "modules.sacra.menu.baptism", slug: "baptism" },
      { label: "modules.sacra.menu.first-communion", slug: "first-communion" },
      { label: "modules.sacra.menu.confirmation", slug: "confirmation" },
      { label: "modules.sacra.menu.marriage", slug: "marriage" },
      { label: "modules.sacra.menu.reports", slug: "reports" },
    ],
  },
  {
    id: "fin",
    name: "modules.fin.name",
    description: "modules.fin.description",
    icon: WalletOutlined,
    color: "var(--domus-app-fin)",
    menu: [
      {
        label: "modules.fin.groups.transactions",
        children: [
          { label: "modules.fin.menu.income", slug: "income" },
          { label: "modules.fin.menu.expenses", slug: "expenses" },
        ],
      },
      { label: "modules.fin.menu.collections", slug: "collections" },
      { label: "modules.fin.menu.budget", slug: "budget" },
      { label: "modules.fin.menu.reports", slug: "reports" },
    ],
  },
  {
    id: "par",
    name: "modules.par.name",
    description: "modules.par.description",
    icon: TeamOutlined,
    color: "var(--domus-app-par)",
    menu: [
      { label: "modules.par.menu.list", slug: "list" },
      { label: "modules.par.menu.families", slug: "families" },
      { label: "modules.par.menu.neighborhoods", slug: "neighborhoods" },
      { label: "modules.par.menu.committee", slug: "committee" },
    ],
  },
  {
    id: "act",
    name: "modules.act.name",
    description: "modules.act.description",
    icon: CalendarOutlined,
    color: "var(--domus-app-act)",
    menu: [
      { label: "modules.act.menu.calendar", slug: "calendar" },
      { label: "modules.act.menu.mass-schedule", slug: "mass-schedule" },
      { label: "modules.act.menu.events", slug: "events" },
      { label: "modules.act.menu.committees", slug: "committees" },
    ],
  },
];

const byId = new Map(modules.map((m) => [m.id, m]));

export function getModule(id: string): DomusModule | undefined {
  return byId.get(id);
}

export function modulePath(moduleId: string, slug?: string): string {
  return slug ? `/${moduleId}/${slug}` : `/${moduleId}`;
}

export function flattenMenu(
  items: ModuleMenuItem[],
  group?: TKey,
): FlatMenuItem[] {
  return items.flatMap((item) => {
    if (item.children) return flattenMenu(item.children, item.label);
    if (item.slug) return [{ slug: item.slug, label: item.label, group }];
    return [];
  });
}

export function findMenuItem(
  mod: DomusModule,
  slug: string,
): FlatMenuItem | undefined {
  return flattenMenu(mod.menu).find((item) => item.slug === slug);
}
