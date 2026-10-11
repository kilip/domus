import { Card, Form, message, Skeleton } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { TerritorialForm, type TerritorialFormValues } from "./TerritorialForm";
import {
  type ApiError,
  useTerritoryDetailQuery,
  useUpdateTerritoryMutation,
} from "./useTerritoryQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("territory.editTitle")} | ${i18n.t("modules.sys.menu.territorial")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function TerritorialUpdatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();
  const id = Number(params.id);

  const [form] = Form.useForm<TerritorialFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: territory,
    isLoading,
    error: loadError,
  } = useTerritoryDetailQuery(id);
  const updateMutation = useUpdateTerritoryMutation(id);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (territory) {
      form.setFieldsValue({
        name: territory.name ?? "",
        code: territory.code ?? undefined,
        kind: territory.kind ?? "",
        parishId: territory.parishId,
        parentId: territory.parentId ?? undefined,
        active: territory.status === "active",
        description: territory.description ?? undefined,
      });
    }
  }, [territory, form]);

  async function handleFinish(values: TerritorialFormValues) {
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
        name: values.name.trim(),
        code: values.code?.trim() || undefined,
        kind: values.kind,
        parishId: Number(values.parishId),
        parentId: values.parentId ? Number(values.parentId) : undefined,
        status: values.active ? "active" : "inactive",
        description: values.description?.trim() || undefined,
      });
      messageApi.success(t("territory.updated"));
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
          { title: t("territory.editTitle") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("territory.editTitle")}>
          {isLoading ? (
            <Skeleton active paragraph={{ rows: 4 }} />
          ) : loadError ? (
            <div role="alert" style={{ color: "var(--ant-color-error)" }}>
              {t("territory.loadFailed")}
            </div>
          ) : (
            <TerritorialForm
              form={form}
              currentTerritoryId={id}
              initialValues={{
                name: territory?.name ?? "",
                code: territory?.code ?? undefined,
                kind: territory?.kind ?? "",
                parishId: territory?.parishId ?? 0,
                parentId: territory?.parentId ?? undefined,
                active: territory?.status === "active",
                description: territory?.description ?? undefined,
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
