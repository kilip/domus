import { Card, Form, message, Skeleton } from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import {
  type ApiError,
  useUpdateWorkspaceMutation,
  useWorkspaceDetailQuery,
} from "./useWorkspaceQueries";
import { WorkspaceForm, type WorkspaceFormValues } from "./WorkspaceForm";

export function meta() {
  return [
    {
      title: `${i18n.t("workspace.editTitle")} | ${i18n.t("modules.sys.menu.workspace")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function WorkspaceUpdatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();
  const id = Number(params.id);

  const [form] = Form.useForm<WorkspaceFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: workspace,
    isLoading,
    error: loadError,
  } = useWorkspaceDetailQuery(id);
  const updateMutation = useUpdateWorkspaceMutation(id);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (workspace) {
      form.setFieldsValue({
        name: workspace.name ?? "",
        parishId: workspace.parishId,
        active: workspace.status === "active",
        description: workspace.description ?? undefined,
      });
    }
  }, [workspace, form]);

  async function handleFinish(values: WorkspaceFormValues) {
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
        name: values.name.trim(),
        parishId: Number(values.parishId),
        status: values.active ? "active" : "inactive",
        description: values.description?.trim() || undefined,
      });
      messageApi.success(t("workspace.updated"));
      navigate(modulePath("sys", "workspace"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      if (apiErr.fields && "name" in apiErr.fields) {
        form.setFields([
          { name: "name", errors: [t("workspace.nameRequired")] },
        ]);
      } else if (apiErr.fields && "parishId" in apiErr.fields) {
        form.setFields([
          { name: "parishId", errors: [t("workspace.parishRequired")] },
        ]);
      } else {
        const errorMsg =
          apiErr.status === 403
            ? t("workspace.forbidden")
            : t("workspace.saveFailed");
        setFormError(errorMsg);
        messageApi.error(errorMsg);
      }
    }
  }

  function handleCancel() {
    navigate(modulePath("sys", "workspace"));
  }

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          {
            title: t("modules.sys.menu.workspace"),
            to: modulePath("sys", "workspace"),
          },
          { title: t("workspace.editTitle") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("workspace.editTitle")}>
          {isLoading ? (
            <Skeleton active paragraph={{ rows: 4 }} />
          ) : loadError ? (
            <div role="alert" style={{ color: "var(--ant-color-error)" }}>
              {t("workspace.loadFailed")}
            </div>
          ) : (
            <WorkspaceForm
              form={form}
              currentWorkspaceId={id}
              initialValues={{
                name: workspace?.name ?? "",
                parishId: workspace?.parishId ?? 0,
                active: workspace?.status === "active",
                description: workspace?.description ?? undefined,
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
