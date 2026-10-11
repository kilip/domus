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
  type Territory,
  useDeleteTerritoryMutation,
  useTerritoriesSearchQuery,
} from "./useTerritoryQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("modules.sys.menu.territorial")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar wilayah teritorial dengan aksi tambah, ubah, dan hapus. */
export default function TerritorialListPage() {
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
  } = useTerritoriesSearchQuery(searchFilters);

  const deleteMutation = useDeleteTerritoryMutation();

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter(
      (item) =>
        item.name?.toLowerCase().includes(q) ||
        item.code?.toLowerCase().includes(q) ||
        item.parish?.name?.toLowerCase().includes(q) ||
        item.parent?.name?.toLowerCase().includes(q),
    );
  }, [items, search]);

  async function handleDelete(territory: Territory) {
    if (!territory.id) return;
    try {
      await deleteMutation.mutateAsync(territory.id);
      messageApi.success(t("territory.deleted"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      messageApi.error(
        apiErr.status === 403
          ? t("territory.forbidden")
          : apiErr.message || t("territory.deleteFailed"),
      );
    }
  }

  function getKindLabel(kind: string): string {
    if (kind === "region") return t("territory.kinds.region");
    if (kind === "station") return t("territory.kinds.station");
    if (kind === "neighborhood") return t("territory.kinds.neighborhood");
    return kind;
  }

  function getKindColor(kind: string): string {
    switch (kind) {
      case "region":
        return "blue";
      case "station":
        return "purple";
      case "neighborhood":
        return "cyan";
      default:
        return "default";
    }
  }

  const columns: ColumnsType<Territory> = [
    {
      title: t("common.name"),
      dataIndex: "name",
      sorter: (a, b) => (a.name ?? "").localeCompare(b.name ?? ""),
      defaultSortOrder: "ascend",
    },
    {
      title: t("territory.codeLabel"),
      dataIndex: "code",
      width: 120,
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("territory.kindLabel"),
      dataIndex: "kind",
      width: 130,
      render: (kind: string) => (
        <Tag color={getKindColor(kind)}>{getKindLabel(kind)}</Tag>
      ),
    },
    {
      title: t("territory.parishLabel"),
      dataIndex: ["parish", "name"],
      responsive: ["md"],
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("territory.parentColumn"),
      dataIndex: ["parent", "name"],
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
      onCell: () => ({
        onClick: (e) => e.stopPropagation(),
      }),
      render: (_, territory) => (
        <Space size="small">
          <Tooltip title={t("common.edit")}>
            <Button
              type="text"
              icon={<EditOutlined />}
              aria-label={t("common.edit")}
              onClick={() =>
                navigate(
                  modulePath("sys", `territorial/update/${territory.id}`),
                )
              }
            />
          </Tooltip>
          <Popconfirm
            title={t("territory.deleteConfirm", { name: territory.name })}
            okText={t("common.delete")}
            okButtonProps={{ danger: true }}
            cancelText={t("common.cancel")}
            onConfirm={() => handleDelete(territory)}
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
          { title: t("modules.sys.menu.territorial") },
        ]}
        searchPlaceholder={t("common.searchPlaceholder")}
        onSearch={setSearch}
        actions={
          <Link to={modulePath("sys", "territorial/create")}>
            <Button type="primary" icon={<PlusOutlined />}>
              {t("territory.add")}
            </Button>
          </Link>
        }
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <div className="mb-4 flex flex-wrap gap-3">
          <Select
            allowClear
            showSearch
            placeholder={t("territory.filterByParish")}
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
            placeholder={t("territory.filterByKind")}
            className="w-44"
            value={filterKind}
            onChange={(val) => setFilterKind(val)}
            options={[
              { value: "region", label: t("territory.kinds.region") },
              { value: "station", label: t("territory.kinds.station") },
              {
                value: "neighborhood",
                label: t("territory.kinds.neighborhood"),
              },
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
                ? t("territory.forbidden")
                : t("territory.loadFailed")}
            </span>
            <Button size="small" onClick={() => void refetch()}>
              {t("common.reload")}
            </Button>
          </div>
        )}

        <Table<Territory>
          rowKey="id"
          size="middle"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: t("territory.empty") }}
          onRow={(record) => ({
            onClick: () => {
              if (record.id) {
                navigate(modulePath("sys", `territorial/update/${record.id}`));
              }
            },
            className: "cursor-pointer",
          })}
        />
      </section>
    </>
  );
}
