import { Card, Form, message, Skeleton } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { ParishForm, type ParishFormValues } from "./ParishForm";
import {
  type ApiError,
  useParishDetailQuery,
  useUpdateParishMutation,
} from "./useParishQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("parish.editTitle")} | ${i18n.t("modules.sys.menu.parish")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function ParishUpdatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();
  const id = Number(params.id);

  const [form] = Form.useForm<ParishFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: parish,
    isLoading,
    error: loadError,
  } = useParishDetailQuery(id);
  const updateMutation = useUpdateParishMutation(id);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (parish) {
      form.setFieldsValue({
        name: parish.name ?? "",
        active: parish.active ?? true,
        church: parish.church ?? undefined,
        dioceseId: parish.diocese?.id,
        address: parish.address ?? undefined,
        phone: parish.phone ?? undefined,
        email: parish.email ?? undefined,
      });
    }
  }, [parish, form]);

  async function handleFinish(values: ParishFormValues) {
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
        name: values.name.trim(),
        active: values.active ?? true,
        church: values.church?.trim() || undefined,
        dioceseId: values.dioceseId ? Number(values.dioceseId) : undefined,
        address: values.address?.trim() || undefined,
        phone: values.phone?.trim() || undefined,
        email: values.email?.trim() || undefined,
      });
      messageApi.success(t("parish.updated"));
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
          { title: t("parish.editTitle") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("parish.editTitle")}>
          {isLoading ? (
            <Skeleton active paragraph={{ rows: 3 }} />
          ) : loadError ? (
            <div role="alert" style={{ color: "var(--ant-color-error)" }}>
              {t("parish.loadFailed")}
            </div>
          ) : (
            <ParishForm
              form={form}
              initialValues={{
                name: parish?.name ?? "",
                active: parish?.active ?? true,
                church: parish?.church ?? undefined,
                dioceseId: parish?.diocese?.id,
                address: parish?.address ?? undefined,
                phone: parish?.phone ?? undefined,
                email: parish?.email ?? undefined,
              }}
              onFinish={handleFinish}
              onCancel={handleCancel}
              loading={updateMutation.isPending}
              formError={formError}
            />
          )}
        </Card>
      </section>
    </>
  );
}
