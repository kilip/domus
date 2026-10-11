import { DeleteOutlined, EditOutlined, PlusOutlined } from "@ant-design/icons";
import { Button, message, Popconfirm, Space, Table, Tag, Tooltip } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import i18n from "~/i18n";
import { ControlPanel } from "~/shared/layout/ControlPanel";
import { modulePath } from "~/shared/modules/registry";
import {
  type ApiError,
  type Diocese,
  useDeleteDioceseMutation,
  useDiocesesListQuery,
} from "./useDioceseQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("modules.sys.menu.diocese")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar keuskupan dengan aksi tambah, ubah, dan hapus. */
export default function DioceseListPage() {
  const { t, i18n: i18nInstance } = useTranslation();
  const navigate = useNavigate();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: items = [],
    isLoading: loading,
    error: loadError,
    refetch,
  } = useDiocesesListQuery();
  const deleteMutation = useDeleteDioceseMutation();

  const [search, setSearch] = useState("");

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase();
    return q ? items.filter((d) => d.name?.toLowerCase().includes(q)) : items;
  }, [items, search]);

  async function handleDelete(diocese: Diocese) {
    if (!diocese.id) return;
    try {
      await deleteMutation.mutateAsync(diocese.id);
      messageApi.success(t("diocese.deleted"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      messageApi.error(
        apiErr.status === 403
          ? t("diocese.forbidden")
          : t("diocese.deleteFailed"),
      );
    }
  }

  const columns: ColumnsType<Diocese> = [
    {
      title: t("common.name"),
      dataIndex: "name",
      sorter: (a, b) => (a.name ?? "").localeCompare(b.name ?? ""),
      defaultSortOrder: "ascend",
    },
    {
      title: t("common.status"),
      dataIndex: "active",
      width: 120,
      render: (active?: boolean) => (
        <Tag color={active ? "success" : "default"}>
          {active ? t("common.active") : t("common.inactive")}
        </Tag>
      ),
    },
    {
      title: t("diocese.updatedAt"),
      dataIndex: "updatedAt",
      responsive: ["md"],
      width: 200,
      render: (value?: string) =>
        value
          ? new Intl.DateTimeFormat(i18nInstance.language, {
              dateStyle: "medium",
              timeStyle: "short",
            }).format(new Date(value))
          : "-",
    },
    {
      title: t("common.actions"),
      key: "actions",
      width: 120,
      align: "right",
      render: (_, diocese) => (
        <Space size="small">
          <Tooltip title={t("common.edit")}>
            <Button
              type="text"
              icon={<EditOutlined />}
              aria-label={t("common.edit")}
              onClick={() =>
                navigate(modulePath("sys", `diocese/update/${diocese.id}`))
              }
            />
          </Tooltip>
          <Popconfirm
            title={t("diocese.deleteConfirm", { name: diocese.name })}
            okText={t("common.delete")}
            okButtonProps={{ danger: true }}
            cancelText={t("common.cancel")}
            onConfirm={() => handleDelete(diocese)}
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
          { title: t("modules.sys.menu.diocese") },
        ]}
        searchPlaceholder={t("common.searchPlaceholder")}
        onSearch={setSearch}
        actions={
          <Link to={modulePath("sys", "diocese/create")}>
            <Button type="primary" icon={<PlusOutlined />}>
              {t("diocese.add")}
            </Button>
          </Link>
        }
      />
      <section className="mx-auto w-full max-w-5xl px-4 py-5">
        {loadError && (
          <div
            className="mb-3 flex items-center justify-between gap-2"
            role="alert"
            style={{ color: "var(--ant-color-error)" }}
          >
            <span>
              {loadError?.status === 403
                ? t("diocese.forbidden")
                : t("diocese.loadFailed")}
            </span>
            <Button size="small" onClick={() => void refetch()}>
              {t("common.reload")}
            </Button>
          </div>
        )}
        <Table<Diocese>
          rowKey="id"
          size="middle"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: t("diocese.empty") }}
        />
      </section>
    </>
  );
}
