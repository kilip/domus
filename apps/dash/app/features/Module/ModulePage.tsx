import { useTranslation } from "react-i18next";
import { data } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { UnderConstruction } from "~/shared/layout/UnderConstruction";
import { findMenuItem, getModule, modulePath } from "~/shared/modules/registry";
import { useActiveModule } from "~/shared/modules/useActiveModule";
import type { Route } from "../../routes/+types/_app.$moduleId.$page";

export function meta({ params }: Route.MetaArgs) {
  const mod = getModule(params.moduleId);
  const item = mod && findMenuItem(mod, params.page);
  return [
    {
      title:
        mod && item
          ? `${i18n.t(item.label)} | ${i18n.t(mod.name)} | Domus`
          : "Domus",
    },
  ];
}

/** Halaman yang tidak ada di menu modul berujung ke 404. */
export function clientLoader({ params }: Route.ClientLoaderArgs) {
  const mod = getModule(params.moduleId);
  if (!mod || !findMenuItem(mod, params.page)) {
    throw data("Not Found", { status: 404 });
  }
  return null;
}

/**
 * Placeholder untuk semua halaman modul yang belum dibuat.
 * Ganti dengan route di `features/<Modul>/routes.ts` saat halamannya siap.
 */
export default function ModulePage() {
  const { t } = useTranslation();
  const { activeModule, activeItem } = useActiveModule();
  if (!activeModule || !activeItem) return null;

  return (
    <>
      <ControlPanel
        breadcrumbs={[
          { title: t(activeModule.name), to: modulePath(activeModule.id) },
          ...(activeItem.group ? [{ title: t(activeItem.group) }] : []),
          { title: t(activeItem.label) },
        ]}
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <UnderConstruction
          title={t(activeItem.label)}
          backTo={modulePath(activeModule.id)}
        />
      </section>
    </>
  );
}
