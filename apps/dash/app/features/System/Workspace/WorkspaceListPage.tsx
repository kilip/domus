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
  useDeleteWorkspaceMutation,
  useWorkspacesSearchQuery,
  type Workspace,
} from "./useWorkspaceQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("modules.sys.menu.workspace")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar workspace dengan aksi tambah, ubah, dan hapus. */
export default function WorkspaceListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [messageApi, messageContext] = message.useMessage();

  const [search, setSearch] = useState("");
  const [filterParishId, setFilterParishId] = useState<number | undefined>();

  const { data: parishes = [] } = useParishesListQuery();

  const searchFilters = useMemo(() => {
    return {
      parishId: filterParishId,
    };
  }, [filterParishId]);

  const {
    data: items = [],
    isLoading: loading,
    error: loadError,
    refetch,
  } = useWorkspacesSearchQuery(searchFilters);

  const deleteMutation = useDeleteWorkspaceMutation();

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter(
      (item) =>
        item.name?.toLowerCase().includes(q) ||
        item.parish?.name?.toLowerCase().includes(q) ||
        item.description?.toLowerCase().includes(q),
    );
  }, [items, search]);

  async function handleDelete(workspace: Workspace) {
    if (!workspace.id) return;
    try {
      await deleteMutation.mutateAsync(workspace.id);
      messageApi.success(t("workspace.deleted"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      messageApi.error(
        apiErr.status === 403
          ? t("workspace.forbidden")
          : apiErr.message || t("workspace.deleteFailed"),
      );
    }
  }

  const columns: ColumnsType<Workspace> = [
    {
      title: t("common.name"),
      dataIndex: "name",
      sorter: (a, b) => (a.name ?? "").localeCompare(b.name ?? ""),
      defaultSortOrder: "ascend",
    },
    {
      title: t("workspace.parishLabel"),
      dataIndex: ["parish", "name"],
      responsive: ["md"],
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("workspace.descriptionLabel"),
      dataIndex: "description",
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
      render: (_, record) => (
        <Space size="small">
          <Tooltip title={t("common.edit")}>
            <Button
              type="text"
              icon={<EditOutlined />}
              aria-label={t("common.edit")}
              onClick={() =>
                navigate(modulePath("sys", `workspace/update/${record.id}`))
              }
            />
          </Tooltip>
          <Popconfirm
            title={t("workspace.deleteConfirm", { name: record.name })}
            okText={t("common.delete")}
            okButtonProps={{ danger: true }}
            cancelText={t("common.cancel")}
            onConfirm={() => handleDelete(record)}
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
          { title: t("modules.sys.menu.workspace") },
        ]}
        searchPlaceholder={t("common.searchPlaceholder")}
        onSearch={setSearch}
        actions={
          <Link to={modulePath("sys", "workspace/create")}>
            <Button type="primary" icon={<PlusOutlined />}>
              {t("workspace.add")}
            </Button>
          </Link>
        }
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        <div className="mb-4 flex flex-wrap gap-3">
          <Select
            allowClear
            showSearch
            placeholder={t("workspace.filterByParish")}
            className="w-56"
            value={filterParishId}
            onChange={(val) => setFilterParishId(val)}
            optionFilterProp="label"
            options={parishes.map((p) => ({
              value: p.id,
              label: p.name,
            }))}
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
                ? t("workspace.forbidden")
                : t("workspace.loadFailed")}
            </span>
            <Button size="small" onClick={() => void refetch()}>
              {t("common.reload")}
            </Button>
          </div>
        )}

        <Table<Workspace>
          rowKey="id"
          size="middle"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: t("workspace.empty") }}
          onRow={(record) => ({
            onClick: () => {
              if (record.id) {
                navigate(modulePath("sys", `workspace/update/${record.id}`));
              }
            },
            className: "cursor-pointer",
          })}
        />
      </section>
    </>
  );
}
