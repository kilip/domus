import { Card, Form, message } from "antd";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import {
  type ApiError,
  useCreateWorkspaceMutation,
} from "./useWorkspaceQueries";
import { WorkspaceForm, type WorkspaceFormValues } from "./WorkspaceForm";

export function meta() {
  return [
    {
      title: `${i18n.t("workspace.add")} | ${i18n.t("modules.sys.menu.workspace")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function WorkspaceCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [form] = Form.useForm<WorkspaceFormValues>();
  const [messageApi, messageContext] = message.useMessage();
  const createMutation = useCreateWorkspaceMutation();
  const [formError, setFormError] = useState<string | null>(null);

  async function handleFinish(values: WorkspaceFormValues) {
    setFormError(null);
    try {
      await createMutation.mutateAsync({
        name: values.name.trim(),
        parishId: Number(values.parishId),
        status: values.active ? "active" : "inactive",
        description: values.description?.trim() || undefined,
      });
      messageApi.success(t("workspace.created"));
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
          { title: t("workspace.add") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("workspace.add")}>
          <WorkspaceForm
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
