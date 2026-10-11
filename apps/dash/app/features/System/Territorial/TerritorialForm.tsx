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
import { useTerritoriesSearchQuery } from "./useTerritoryQueries";

export interface TerritorialFormValues {
  name: string;
  code?: string;
  kind: string;
  parishId: number;
  parentId?: number;
  active?: boolean;
  description?: string;
}

export interface TerritorialFormProps {
  form?: FormInstance<TerritorialFormValues>;
  initialValues?: Partial<TerritorialFormValues>;
  currentTerritoryId?: number;
  onFinish: (values: TerritorialFormValues) => Promise<void> | void;
  onCancel: () => void;
  loading?: boolean;
  formError?: string | null;
  submitText?: string;
  cancelText?: string;
  autoFocus?: boolean;
}

export function TerritorialForm({
  form: formProp,
  initialValues,
  currentTerritoryId,
  onFinish,
  onCancel,
  loading = false,
  formError,
  submitText,
  cancelText,
  autoFocus = true,
}: TerritorialFormProps) {
  const { t } = useTranslation();
  const [internalForm] = Form.useForm<TerritorialFormValues>();
  const form = formProp ?? internalForm;

  const { data: parishes = [], isLoading: parishesLoading } =
    useParishesListQuery();

  const watchedParishId = Form.useWatch("parishId", form);

  // Fetch territories belonging to the selected parish to populate parent options
  const { data: parishTerritories = [], isLoading: territoriesLoading } =
    useTerritoriesSearchQuery(
      watchedParishId ? { parishId: watchedParishId } : undefined,
    );

  useEffect(() => {
    if (initialValues) {
      form.setFieldsValue(initialValues);
    }
  }, [initialValues, form]);

  // Filter out self when editing
  const parentOptions = parishTerritories
    .filter((item) =>
      currentTerritoryId ? item.id !== currentTerritoryId : true,
    )
    .map((item) => ({
      value: item.id,
      label: item.code ? `${item.name} (${item.code})` : item.name,
    }));

  const kindOptions = [
    { value: "region", label: t("territory.kinds.region") },
    { value: "station", label: t("territory.kinds.station") },
    { value: "neighborhood", label: t("territory.kinds.neighborhood") },
  ];

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
        label={t("territory.nameLabel")}
        rules={[
          {
            required: true,
            whitespace: true,
            message: t("territory.nameRequired"),
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
          placeholder={t("territory.nameLabel")}
        />
      </Form.Item>

      <Form.Item name="code" label={t("territory.codeLabel")}>
        <Input maxLength={50} placeholder={t("territory.codePlaceholder")} />
      </Form.Item>

      <Form.Item
        name="kind"
        label={t("territory.kindLabel")}
        rules={[
          {
            required: true,
            message: t("territory.kindRequired"),
          },
        ]}
      >
        <Select placeholder={t("territory.kindLabel")} options={kindOptions} />
      </Form.Item>

      <Form.Item
        name="parishId"
        label={t("territory.parishLabel")}
        rules={[
          {
            required: true,
            message: t("territory.parishRequired"),
          },
        ]}
      >
        <Select
          allowClear
          showSearch
          loading={parishesLoading}
          placeholder={t("territory.parishPlaceholder")}
          optionFilterProp="label"
          onChange={() => {
            // Reset parent if parish changes
            form.setFieldValue("parentId", undefined);
          }}
          options={parishes.map((p) => ({
            value: p.id,
            label: p.name,
          }))}
        />
      </Form.Item>

      <Form.Item name="parentId" label={t("territory.parentLabel")}>
        <Select
          allowClear
          showSearch
          disabled={!watchedParishId}
          loading={territoriesLoading}
          placeholder={t("territory.parentPlaceholder")}
          optionFilterProp="label"
          options={parentOptions}
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

      <Form.Item name="description" label={t("territory.descriptionLabel")}>
        <Input.TextArea
          rows={3}
          maxLength={500}
          placeholder={t("territory.descriptionPlaceholder")}
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
