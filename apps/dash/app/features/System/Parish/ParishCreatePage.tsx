import { Card, Form, message } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { ParishForm, type ParishFormValues } from "./ParishForm";
import { type ApiError, useCreateParishMutation } from "./useParishQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("parish.add")} | ${i18n.t("modules.sys.menu.parish")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function ParishCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [form] = Form.useForm<ParishFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const createMutation = useCreateParishMutation();
  const [formError, setFormError] = useState<string | null>(null);

  async function handleFinish(values: ParishFormValues) {
    setFormError(null);
    try {
      await createMutation.mutateAsync({
        name: values.name.trim(),
        church: values.church?.trim() || undefined,
        dioceseId: values.dioceseId ? Number(values.dioceseId) : undefined,
        address: values.address?.trim() || undefined,
        phone: values.phone?.trim() || undefined,
        email: values.email?.trim() || undefined,
      });
      messageApi.success(t("parish.created"));
      navigate(modulePath("sys", "parish"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      if (apiErr.fields && "name" in apiErr.fields) {
        form.setFields([{ name: "name", errors: [t("parish.nameRequired")] }]);
      } else if (apiErr.fields && "email" in apiErr.fields) {
        form.setFields([{ name: "email", errors: [t("parish.emailInvalid")] }]);
      } else {
        const errorMsg =
          apiErr.status === 403
            ? t("parish.forbidden")
            : t("parish.saveFailed");
        setFormError(errorMsg);
        messageApi.error(errorMsg);
      }
    }
  }

  function handleCancel() {
    navigate(modulePath("sys", "parish"));
  }

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          {
            title: t("modules.sys.menu.parish"),
            to: modulePath("sys", "parish"),
          },
          { title: t("parish.add") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("parish.add")}>
          <ParishForm
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
