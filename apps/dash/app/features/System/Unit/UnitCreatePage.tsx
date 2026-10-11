import { Card, Form, message } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { UnitForm, type UnitFormValues } from "./UnitForm";
import { type ApiError, useCreateUnitMutation } from "./useUnitQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("unit.add")} | ${i18n.t("modules.sys.menu.unit")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function UnitCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [form] = Form.useForm<UnitFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const createMutation = useCreateUnitMutation();
  const [formError, setFormError] = useState<string | null>(null);

  async function handleFinish(values: UnitFormValues) {
    setFormError(null);
    try {
      await createMutation.mutateAsync({
        name: values.name.trim(),
        code: values.code?.trim() || undefined,
        kind: values.kind,
        parishId: Number(values.parishId),
        parentId: values.parentId ? Number(values.parentId) : undefined,
        territoryId:
          values.kind === "territorial" && values.territoryId
            ? Number(values.territoryId)
            : undefined,
        status: values.active ? "active" : "inactive",
        description: values.description?.trim() || undefined,
      });
      messageApi.success(t("unit.created"));
      navigate(modulePath("sys", "unit"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      if (apiErr.fields && "name" in apiErr.fields) {
        form.setFields([{ name: "name", errors: [t("unit.nameRequired")] }]);
      } else if (apiErr.fields && "kind" in apiErr.fields) {
        form.setFields([{ name: "kind", errors: [t("unit.kindRequired")] }]);
      } else if (apiErr.fields && "parishId" in apiErr.fields) {
        form.setFields([
          { name: "parishId", errors: [t("unit.parishRequired")] },
        ]);
      } else if (apiErr.fields && "territoryId" in apiErr.fields) {
        form.setFields([
          { name: "territoryId", errors: [t("unit.territoryRequired")] },
        ]);
      } else {
        const errorMsg =
          apiErr.status === 403 ? t("unit.forbidden") : t("unit.saveFailed");
        setFormError(errorMsg);
        messageApi.error(errorMsg);
      }
    }
  }

  function handleCancel() {
    navigate(modulePath("sys", "unit"));
  }

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          {
            title: t("modules.sys.menu.unit"),
            to: modulePath("sys", "unit"),
          },
          { title: t("unit.add") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("unit.add")}>
          <UnitForm
            form={form}
            onFinish={handleFinish}
            onCancel={handleCancel}
            loading={createMutation.isPending}
            formError={formError}
          />
        </Card>
      </section>
    </>
  );
}
