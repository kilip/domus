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
import { useTerritoriesSearchQuery } from "../Territorial/useTerritoryQueries";
import { useUnitsSearchQuery } from "./useUnitQueries";

export type UnitKind =
  | "council"
  | "territorial"
  | "category"
  | "section"
  | "committee";

export interface UnitFormValues {
  name: string;
  code?: string;
  kind: UnitKind;
  parishId: number;
  territoryId?: number;
  parentId?: number;
  active?: boolean;
  description?: string;
}

export interface UnitFormProps {
  form?: FormInstance<UnitFormValues>;
  initialValues?: Partial<UnitFormValues>;
  currentUnitId?: number;
  onFinish: (values: UnitFormValues) => Promise<void> | void;
  onCancel: () => void;
  loading?: boolean;
  formError?: string | null;
  submitText?: string;
  cancelText?: string;
  autoFocus?: boolean;
}

export function UnitForm({
  form: formProp,
  initialValues,
  currentUnitId,
  onFinish,
  onCancel,
  loading = false,
  formError,
  submitText,
  cancelText,
  autoFocus = true,
}: UnitFormProps) {
  const { t } = useTranslation();
  const [internalForm] = Form.useForm<UnitFormValues>();
  const form = formProp ?? internalForm;

  const { data: parishes = [], isLoading: parishesLoading } =
    useParishesListQuery();

  const watchedParishId = Form.useWatch("parishId", form);
  const watchedKind = Form.useWatch("kind", form);

  // Fetch territories belonging to the selected parish to populate territory options
  const { data: parishTerritories = [], isLoading: territoriesLoading } =
    useTerritoriesSearchQuery(
      watchedParishId ? { parishId: watchedParishId } : undefined,
    );

  // Fetch units belonging to the selected parish to populate parent options
  const { data: parishUnits = [], isLoading: unitsLoading } =
    useUnitsSearchQuery(
      watchedParishId ? { parishId: watchedParishId } : undefined,
    );

  useEffect(() => {
    if (initialValues) {
      form.setFieldsValue(initialValues);
    }
  }, [initialValues, form]);

  // Filter out self when editing parent
  const parentOptions = parishUnits
    .filter((item) => (currentUnitId ? item.id !== currentUnitId : true))
    .map((item) => ({
      value: item.id,
      label: item.code ? `${item.name} (${item.code})` : item.name,
    }));

  const territoryOptions = parishTerritories.map((item) => ({
    value: item.id,
    label: item.code ? `${item.name} (${item.code})` : item.name,
  }));

  const kindOptions = [
    { value: "council", label: t("unit.kinds.council") },
    { value: "territorial", label: t("unit.kinds.territorial") },
    { value: "category", label: t("unit.kinds.category") },
    { value: "section", label: t("unit.kinds.section") },
    { value: "committee", label: t("unit.kinds.committee") },
  ];

  const isTerritorial = watchedKind === "territorial";

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
        label={t("unit.nameLabel")}
        rules={[
          {
            required: true,
            whitespace: true,
            message: t("unit.nameRequired"),
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
          placeholder={t("unit.nameLabel")}
        />
      </Form.Item>

      <Form.Item name="code" label={t("unit.codeLabel")}>
        <Input maxLength={50} placeholder={t("unit.codePlaceholder")} />
      </Form.Item>

      <Form.Item
        name="kind"
        label={t("unit.kindLabel")}
        rules={[
          {
            required: true,
            message: t("unit.kindRequired"),
          },
        ]}
      >
        <Select
          placeholder={t("unit.kindLabel")}
          options={kindOptions}
          onChange={(newKind) => {
            if (newKind !== "territorial") {
              form.setFieldValue("territoryId", undefined);
            }
          }}
        />
      </Form.Item>

      <Form.Item
        name="parishId"
        label={t("unit.parishLabel")}
        rules={[
          {
            required: true,
            message: t("unit.parishRequired"),
          },
        ]}
      >
        <Select
          allowClear
          showSearch
          loading={parishesLoading}
          placeholder={t("unit.parishPlaceholder")}
          optionFilterProp="label"
          onChange={() => {
            // Reset parent and territory if parish changes
            form.setFieldValue("parentId", undefined);
            form.setFieldValue("territoryId", undefined);
          }}
          options={parishes.map((p) => ({
            value: p.id,
            label: p.name,
          }))}
        />
      </Form.Item>

      {isTerritorial && (
        <Form.Item
          name="territoryId"
          label={t("unit.territoryLabel")}
          rules={[
            {
              required: true,
              message: t("unit.territoryRequired"),
            },
          ]}
        >
          <Select
            allowClear
            showSearch
            disabled={!watchedParishId}
            loading={territoriesLoading}
            placeholder={t("unit.territoryPlaceholder")}
            optionFilterProp="label"
            options={territoryOptions}
          />
        </Form.Item>
      )}

      <Form.Item name="parentId" label={t("unit.parentLabel")}>
        <Select
          allowClear
          showSearch
          disabled={!watchedParishId}
          loading={unitsLoading}
          placeholder={t("unit.parentPlaceholder")}
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

      <Form.Item name="description" label={t("unit.descriptionLabel")}>
        <Input.TextArea
          rows={3}
          maxLength={500}
          placeholder={t("unit.descriptionPlaceholder")}
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
