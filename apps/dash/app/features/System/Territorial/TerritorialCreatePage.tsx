import { Card, Form, message } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { TerritorialForm, type TerritorialFormValues } from "./TerritorialForm";
import {
  type ApiError,
  useCreateTerritoryMutation,
} from "./useTerritoryQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("territory.add")} | ${i18n.t("modules.sys.menu.territorial")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function TerritorialCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [form] = Form.useForm<TerritorialFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const createMutation = useCreateTerritoryMutation();
  const [formError, setFormError] = useState<string | null>(null);

  async function handleFinish(values: TerritorialFormValues) {
    setFormError(null);
    try {
      await createMutation.mutateAsync({
        name: values.name.trim(),
        code: values.code?.trim() || undefined,
        kind: values.kind,
        parishId: Number(values.parishId),
        parentId: values.parentId ? Number(values.parentId) : undefined,
        status: values.active ? "active" : "inactive",
        description: values.description?.trim() || undefined,
      });
      messageApi.success(t("territory.created"));
      navigate(modulePath("sys", "territorial"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      if (apiErr.fields && "name" in apiErr.fields) {
        form.setFields([
          { name: "name", errors: [t("territory.nameRequired")] },
        ]);
      } else if (apiErr.fields && "kind" in apiErr.fields) {
        form.setFields([
          { name: "kind", errors: [t("territory.kindRequired")] },
        ]);
      } else if (apiErr.fields && "parishId" in apiErr.fields) {
        form.setFields([
          { name: "parishId", errors: [t("territory.parishRequired")] },
        ]);
      } else {
        const errorMsg =
          apiErr.status === 403
            ? t("territory.forbidden")
            : t("territory.saveFailed");
        setFormError(errorMsg);
        messageApi.error(errorMsg);
      }
    }
  }

  function handleCancel() {
    navigate(modulePath("sys", "territorial"));
  }

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          {
            title: t("modules.sys.menu.territorial"),
            to: modulePath("sys", "territorial"),
          },
          { title: t("territory.add") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("territory.add")}>
          <TerritorialForm
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
