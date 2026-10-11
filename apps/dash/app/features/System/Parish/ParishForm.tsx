import { ArrowLeftOutlined, SaveOutlined } from "@ant-design/icons";
import { Button, Form, type FormInstance, Input, Select, Space } from "antd";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useDiocesesListQuery } from "../Diocese/useDioceseQueries";

export interface ParishFormValues {
  name: string;
  church?: string;
  dioceseId?: number;
  address?: string;
  phone?: string;
  email?: string;
}

export interface ParishFormProps {
  form?: FormInstance<ParishFormValues>;
  initialValues?: Partial<ParishFormValues>;
  onFinish: (values: ParishFormValues) => Promise<void> | void;
  onCancel: () => void;
  loading?: boolean;
  formError?: string | null;
  submitText?: string;
  cancelText?: string;
  autoFocus?: boolean;
}

export function ParishForm({
  form: formProp,
  initialValues,
  onFinish,
  onCancel,
  loading = false,
  formError,
  submitText,
  cancelText,
  autoFocus = true,
}: ParishFormProps) {
  const { t } = useTranslation();
  const [internalForm] = Form.useForm<ParishFormValues>();
  const form = formProp ?? internalForm;
  const { data: dioceses = [], isLoading: diocesesLoading } =
    useDiocesesListQuery();

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
        label={t("parish.nameLabel")}
        rules={[
          {
            required: true,
            whitespace: true,
            message: t("parish.nameRequired"),
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
          placeholder={t("parish.nameLabel")}
        />
      </Form.Item>

      <Form.Item name="church" label={t("parish.churchLabel")}>
        <Input maxLength={200} placeholder={t("parish.churchPlaceholder")} />
      </Form.Item>

      <Form.Item name="dioceseId" label={t("parish.dioceseLabel")}>
        <Select
          allowClear
          showSearch
          loading={diocesesLoading}
          placeholder={t("parish.diocesePlaceholder")}
          optionFilterProp="label"
          options={dioceses.map((d) => ({
            value: d.id,
            label: d.name,
          }))}
        />
      </Form.Item>

      <Form.Item name="address" label={t("parish.addressLabel")}>
        <Input.TextArea
          rows={3}
          maxLength={500}
          placeholder={t("parish.addressPlaceholder")}
        />
      </Form.Item>

      <Form.Item name="phone" label={t("parish.phoneLabel")}>
        <Input maxLength={50} placeholder={t("parish.phonePlaceholder")} />
      </Form.Item>

      <Form.Item
        name="email"
        label={t("parish.emailLabel")}
        rules={[
          {
            type: "email",
            message: t("parish.emailInvalid"),
          },
        ]}
      >
        <Input
          maxLength={100}
          type="email"
          placeholder={t("parish.emailPlaceholder")}
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
