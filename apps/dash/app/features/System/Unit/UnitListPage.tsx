import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import {
  Button,
  message,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import { useParishesListQuery } from "../Parish/useParishQueries";
import {
  type ApiError,
  type Unit,
  useDeleteUnitMutation,
  useUnitsSearchQuery,
} from "./useUnitQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("modules.sys.menu.unit")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar unit organisasi dengan aksi tambah, ubah, dan hapus. */
export default function UnitListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [messageApi, messageContext] = message.useMessage();

  const [search, setSearch] = useState("");
  const [filterParishId, setFilterParishId] = useState<number | undefined>();
  const [filterKind, setFilterKind] = useState<string | undefined>();

  const { data: parishes = [] } = useParishesListQuery();

  const searchFilters = useMemo(() => {
    return {
      parishId: filterParishId,
      kind: filterKind,
    };
  }, [filterParishId, filterKind]);

  const {
    data: items = [],
    isLoading: loading,
    error: loadError,
    refetch,
  } = useUnitsSearchQuery(searchFilters);

  const deleteMutation = useDeleteUnitMutation();

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter(
      (item) =>
        item.name?.toLowerCase().includes(q) ||
        item.code?.toLowerCase().includes(q) ||
        item.parish?.name?.toLowerCase().includes(q) ||
        item.parent?.name?.toLowerCase().includes(q) ||
        item.territory?.name?.toLowerCase().includes(q),
    );
  }, [items, search]);

  async function handleDelete(unit: Unit) {
    if (!unit.id) return;
    try {
      await deleteMutation.mutateAsync(unit.id);
      messageApi.success(t("unit.deleted"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      messageApi.error(
        apiErr.status === 403
          ? t("unit.forbidden")
          : apiErr.message || t("unit.deleteFailed"),
      );
    }
  }

  function getKindLabel(kind: string): string {
    switch (kind) {
      case "council":
        return t("unit.kinds.council");
      case "territorial":
        return t("unit.kinds.territorial");
      case "category":
        return t("unit.kinds.category");
      case "section":
        return t("unit.kinds.section");
      case "committee":
        return t("unit.kinds.committee");
      default:
        return kind;
    }
  }

  function getKindColor(kind: string): string {
    switch (kind) {
      case "council":
        return "gold";
      case "territorial":
        return "blue";
      case "category":
        return "green";
      case "section":
        return "purple";
      case "committee":
        return "magenta";
      default:
        return "default";
    }
  }

  const columns: ColumnsType<Unit> = [
    {
      title: t("common.name"),
      dataIndex: "name",
      sorter: (a, b) => (a.name ?? "").localeCompare(b.name ?? ""),
      defaultSortOrder: "ascend",
    },
    {
      title: t("unit.codeLabel"),
      dataIndex: "code",
      width: 120,
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("unit.kindLabel"),
      dataIndex: "kind",
      width: 130,
      render: (kind: string) => (
        <Tag color={getKindColor(kind)}>{getKindLabel(kind)}</Tag>
      ),
    },
    {
      title: t("unit.parishLabel"),
      dataIndex: ["parish", "name"],
      responsive: ["md"],
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("unit.parentColumn"),
      dataIndex: ["parent", "name"],
      responsive: ["lg"],
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("unit.territoryColumn"),
      dataIndex: ["territory", "name"],
      responsive: ["lg"],
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("common.status"),
      dataIndex: "status",
      width: 110,
      render: (status?: string) => {
        const isActive = status === "active";
        return (
          <Tag color={isActive ? "success" : "default"}>
            {isActive ? t("common.active") : t("common.inactive")}
          </Tag>
        );
      },
    },
    {
      title: t("common.actions"),
      key: "actions",
      width: 110,
      align: "right",
      render: (_, unit) => (
        <Space size="small">
          <Tooltip title={t("common.edit")}>
            <Button
              type="text"
              icon={<EditOutlined />}
              aria-label={t("common.edit")}
              onClick={() =>
                navigate(modulePath("sys", `unit/update/${unit.id}`))
              }
            />
          </Tooltip>
          <Popconfirm
            title={t("unit.deleteConfirm", { name: unit.name })}
            okText={t("common.delete")}
            okButtonProps={{ danger: true }}
            cancelText={t("common.cancel")}
            onConfirm={() => handleDelete(unit)}
          >
            <Tooltip title={t("common.delete")}>
              <Button
                type="text"
                danger
                icon={<DeleteOutlined />}
                aria-label={t("common.delete")}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      {messageContext}
      <ControlPanel
        breadcrumbs={[
          { title: t("modules.sys.name"), to: modulePath("sys") },
          { title: t("modules.sys.menu.unit") },
        ]}
        searchPlaceholder={t("common.searchPlaceholder")}
        onSearch={setSearch}
        actions={
          <Link to={modulePath("sys", "unit/create")}>
            <Button type="primary" icon={<PlusOutlined />}>
              {t("unit.add")}
            </Button>
          </Link>
        }
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <div className="mb-4 flex flex-wrap gap-3">
          <Select
            allowClear
            showSearch
            placeholder={t("unit.filterByParish")}
            className="w-56"
            value={filterParishId}
            onChange={(val) => setFilterParishId(val)}
            optionFilterProp="label"
            options={parishes.map((p) => ({
              value: p.id,
              label: p.name,
            }))}
          />
          <Select
            allowClear
            placeholder={t("unit.filterByKind")}
            className="w-44"
            value={filterKind}
            onChange={(val) => setFilterKind(val)}
            options={[
              { value: "council", label: t("unit.kinds.council") },
              { value: "territorial", label: t("unit.kinds.territorial") },
              { value: "category", label: t("unit.kinds.category") },
              { value: "section", label: t("unit.kinds.section") },
              { value: "committee", label: t("unit.kinds.committee") },
            ]}
          />
        </div>

        {loadError && (
          <div
            className="mb-3 flex items-center justify-between gap-2"
            role="alert"
            style={{ color: "var(--ant-color-error)" }}
          >
            <span>
              {loadError?.status === 403
                ? t("unit.forbidden")
                : t("unit.loadFailed")}
            </span>
            <Button size="small" onClick={() => void refetch()}>
              {t("common.reload")}
            </Button>
          </div>
        )}

        <Table<Unit>
          rowKey="id"
          size="middle"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: t("unit.empty") }}
        />
      </section>
    </>
  );
}
