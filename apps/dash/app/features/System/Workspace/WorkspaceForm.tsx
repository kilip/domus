import { ArrowLeftOutlined, SaveOutlined } from "@ant-design/icons";
import {
  Button,
  Form,
  type FormInstance,
  Input,
  Select,
  Space,
  Switch,
} from "antd";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useParishesListQuery } from "../Parish/useParishQueries";

export interface WorkspaceFormValues {
  name: string;
  parishId: number;
  active?: boolean;
  description?: string;
}

export interface WorkspaceFormProps {
  form?: FormInstance<WorkspaceFormValues>;
  initialValues?: Partial<WorkspaceFormValues>;
  currentWorkspaceId?: number;
  onFinish: (values: WorkspaceFormValues) => Promise<void> | void;
  onCancel: () => void;
  loading?: boolean;
  formError?: string | null;
  submitText?: string;
  cancelText?: string;
  autoFocus?: boolean;
}

export function WorkspaceForm({
  form: formProp,
  initialValues,
  onFinish,
  onCancel,
  loading = false,
  formError,
  submitText,
  cancelText,
  autoFocus = true,
}: WorkspaceFormProps) {
  const { t } = useTranslation();
  const [internalForm] = Form.useForm<WorkspaceFormValues>();
  const form = formProp ?? internalForm;

  const { data: parishes = [], isLoading: parishesLoading } =
    useParishesListQuery();

  useEffect(() => {
    if (initialValues) {
      form.setFieldsValue(initialValues);
    }
  }, [initialValues, form]);

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={onFinish}
      requiredMark={false}
      initialValues={initialValues}
    >
      <Form.Item
        name="name"
        label={t("workspace.nameLabel")}
        rules={[
          {
            required: true,
            whitespace: true,
            message: t("workspace.nameRequired"),
          },
        ]}
        extra={
          formError ? (
            <span style={{ color: "var(--ant-color-error)" }}>{formError}</span>
          ) : undefined
        }
      >
        <Input
          autoFocus={autoFocus}
          maxLength={200}
          placeholder={t("workspace.nameLabel")}
        />
      </Form.Item>

      <Form.Item
        name="parishId"
        label={t("workspace.parishLabel")}
        rules={[
          {
            required: true,
            message: t("workspace.parishRequired"),
          },
        ]}
      >
        <Select
          allowClear
          showSearch
          loading={parishesLoading}
          placeholder={t("workspace.parishPlaceholder")}
          optionFilterProp="label"
          options={parishes.map((p) => ({
            value: p.id,
            label: p.name,
          }))}
        />
      </Form.Item>

      <Form.Item
        name="active"
        label={t("common.status")}
        valuePropName="checked"
        initialValue={true}
      >
        <Switch
          checkedChildren={t("common.active")}
          unCheckedChildren={t("common.inactive")}
        />
      </Form.Item>

      <Form.Item name="description" label={t("workspace.descriptionLabel")}>
        <Input.TextArea
          rows={3}
          maxLength={500}
          placeholder={t("workspace.descriptionPlaceholder")}
        />
      </Form.Item>

      <Form.Item className="mb-0">
        <Space>
          <Button
            type="primary"
            htmlType="submit"
            icon={<SaveOutlined />}
            loading={loading}
          >
            {submitText ?? t("common.save")}
          </Button>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={onCancel}
            disabled={loading}
          >
            {cancelText ?? t("common.cancel")}
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
}
