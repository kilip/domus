import { ArrowLeftOutlined, SaveOutlined } from "@ant-design/icons";
import {
  Button,
  Card,
  Form,
  Input,
  message,
  Skeleton,
  Space,
  Switch,
} from "antd";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import {
  type ApiError,
  useDioceseDetailQuery,
  useUpdateDioceseMutation,
} from "./useDioceseQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("diocese.editTitle")} | ${i18n.t("modules.sys.menu.diocese")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

export default function DioceseUpdatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const params = useParams<{ id: string }>();
  const id = Number(params.id);

  const [form] = Form.useForm<{ name: string; active?: boolean }>();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: diocese,
    isLoading,
    error: loadError,
  } = useDioceseDetailQuery(id);
  const updateMutation = useUpdateDioceseMutation(id);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    if (diocese) {
      form.setFieldsValue({
        name: diocese.name,
        active: diocese.active ?? true,
      });
    }
  }, [diocese, form]);

  async function handleFinish(values: { name: string; active?: boolean }) {
    setFormError(null);
    try {
      await updateMutation.mutateAsync({
        name: values.name.trim(),
        active: values.active ?? true,
      });
      messageApi.success(t("diocese.updated"));
      navigate(modulePath("sys", "diocese"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      if (apiErr.fields && "name" in apiErr.fields) {
        form.setFields([{ name: "name", errors: [t("diocese.nameRequired")] }]);
      } else {
        const errorMsg =
          apiErr.status === 403
            ? t("diocese.forbidden")
            : t("diocese.saveFailed");
        setFormError(errorMsg);
        messageApi.error(errorMsg);
      }
    }
  }

  function handleCancel() {
    navigate(modulePath("sys", "diocese"));
  }

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          {
            title: t("modules.sys.menu.diocese"),
            to: modulePath("sys", "diocese"),
          },
          { title: t("diocese.editTitle") },
        ]}
      />
      <section className="mx-auto w-full max-w-2xl px-4 py-5">
        <Card title={t("diocese.editTitle")}>
          {isLoading ? (
            <Skeleton active paragraph={{ rows: 3 }} />
          ) : loadError ? (
            <div role="alert" style={{ color: "var(--ant-color-error)" }}>
              {t("diocese.loadFailed")}
            </div>
          ) : (
            <Form
              form={form}
              layout="vertical"
              onFinish={handleFinish}
              requiredMark={false}
              initialValues={{
                name: diocese?.name ?? "",
                active: diocese?.active ?? true,
              }}
            >
              <Form.Item
                name="name"
                label={t("diocese.nameLabel")}
                rules={[
                  {
                    required: true,
                    whitespace: true,
                    message: t("diocese.nameRequired"),
                  },
                ]}
                extra={
                  formError ? (
                    <span style={{ color: "var(--ant-color-error)" }}>
                      {formError}
                    </span>
                  ) : undefined
                }
              >
                <Input
                  autoFocus
                  maxLength={200}
                  placeholder={t("diocese.nameLabel")}
                />
              </Form.Item>
              <Form.Item
                name="active"
                label={t("common.status")}
                valuePropName="checked"
              >
                <Switch
                  checkedChildren={t("common.active")}
                  unCheckedChildren={t("common.inactive")}
                />
              </Form.Item>
              <Form.Item className="mb-0">
                <Space>
                  <Button
                    type="primary"
                    htmlType="submit"
                    icon={<SaveOutlined />}
                    loading={updateMutation.isPending}
                  >
                    {t("common.save")}
                  </Button>
                  <Button
                    icon={<ArrowLeftOutlined />}
                    onClick={handleCancel}
                    disabled={updateMutation.isPending}
                  >
                    {t("common.cancel")}
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          )}
        </Card>
      </section>
    </>
  );
}
