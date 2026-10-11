import { Card, Form, message, Skeleton } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { UnitForm, type UnitFormValues, type UnitKind } from "./UnitForm";
import {
  type ApiError,
  useUnitDetailQuery,
  useUpdateUnitMutation,
} from "./useUnitQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("unit.editTitle")} | ${i18n.t("modules.sys.menu.unit")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function UnitUpdatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();
  const id = Number(params.id);

  const [form] = Form.useForm<UnitFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const { data: unit, isLoading, error: loadError } = useUnitDetailQuery(id);
  const updateMutation = useUpdateUnitMutation(id);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (unit) {
      form.setFieldsValue({
        name: unit.name ?? "",
        code: unit.code ?? undefined,
        kind: (unit.kind as UnitKind) ?? "council",
        parishId: unit.parishId,
        territoryId: unit.territoryId ?? undefined,
        parentId: unit.parentId ?? undefined,
        active: unit.status === "active",
        description: unit.description ?? undefined,
      });
    }
  }, [unit, form]);

  async function handleFinish(values: UnitFormValues) {
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
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
      messageApi.success(t("unit.updated"));
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
          { title: t("unit.editTitle") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("unit.editTitle")}>
          {isLoading ? (
            <Skeleton active paragraph={{ rows: 4 }} />
          ) : loadError ? (
            <div role="alert" style={{ color: "var(--ant-color-error)" }}>
              {t("unit.loadFailed")}
            </div>
          ) : (
            <UnitForm
              form={form}
              currentUnitId={id}
              initialValues={{
                name: unit?.name ?? "",
                code: unit?.code ?? undefined,
                kind: (unit?.kind as UnitKind) ?? "council",
                parishId: unit?.parishId ?? 0,
                territoryId: unit?.territoryId ?? undefined,
                parentId: unit?.parentId ?? undefined,
                active: unit?.status === "active",
                description: unit?.description ?? undefined,
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
