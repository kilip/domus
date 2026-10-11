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
  type Parish,
  useDeleteParishMutation,
  useParishesListQuery,
} from "./useParishQueries";

export function meta() {
  return [
    {
      title: `${i18n.t("modules.sys.menu.parish")} | ${i18n.t("modules.sys.name")} | Domus`,
    },
  ];
}

/** Daftar paroki dengan aksi tambah, ubah, dan hapus (tanpa modal). */
export default function ParishListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [messageApi, messageContext] = message.useMessage();
  const {
    data: items = [],
    isLoading: loading,
    error: loadError,
    refetch,
  } = useParishesListQuery();
  const deleteMutation = useDeleteParishMutation();

  const [search, setSearch] = useState("");

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return items;
    return items.filter(
      (p) =>
        p.name?.toLowerCase().includes(q) ||
        p.church?.toLowerCase().includes(q) ||
        p.diocese?.name?.toLowerCase().includes(q),
    );
  }, [items, search]);

  async function handleDelete(parish: Parish) {
    if (!parish.id) return;
    try {
      await deleteMutation.mutateAsync(parish.id);
      messageApi.success(t("parish.deleted"));
    } catch (err: unknown) {
      const apiErr = err as ApiError;
      messageApi.error(
        apiErr.status === 403
          ? t("parish.forbidden")
          : t("parish.deleteFailed"),
      );
    }
  }

  const columns: ColumnsType<Parish> = [
    {
      title: t("common.name"),
      dataIndex: "name",
      sorter: (a, b) => (a.name ?? "").localeCompare(b.name ?? ""),
      defaultSortOrder: "ascend",
    },
    {
      title: t("parish.churchLabel"),
      dataIndex: "church",
      render: (value?: string) => value ?? "-",
    },
    {
      title: t("parish.dioceseLabel"),
      dataIndex: ["diocese", "name"],
      responsive: ["md"],
      render: (value?: string) => value ?? "-",
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
      title: t("common.actions"),
      key: "actions",
      width: 120,
      align: "right",
      onCell: () => ({
        onClick: (e) => e.stopPropagation(),
      }),
      render: (_, parish) => (
        <Space size="small">
          <Tooltip title={t("common.edit")}>
            <Button
              type="text"
              icon={<EditOutlined />}
              aria-label={t("common.edit")}
              onClick={() =>
                navigate(modulePath("sys", `parish/update/${parish.id}`))
              }
            />
          </Tooltip>
          <Popconfirm
            title={t("parish.deleteConfirm", { name: parish.name })}
            okText={t("common.delete")}
            okButtonProps={{ danger: true }}
            cancelText={t("common.cancel")}
            onConfirm={() => handleDelete(parish)}
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
          { title: t("modules.sys.menu.parish") },
        ]}
        searchPlaceholder={t("common.searchPlaceholder")}
        onSearch={setSearch}
        actions={
          <Link to={modulePath("sys", "parish/create")}>
            <Button type="primary" icon={<PlusOutlined />}>
              {t("parish.add")}
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
                ? t("parish.forbidden")
                : t("parish.loadFailed")}
            </span>
            <Button size="small" onClick={() => void refetch()}>
              {t("common.reload")}
            </Button>
          </div>
        )}
        <Table<Parish>
          rowKey="id"
          size="middle"
          loading={loading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, hideOnSinglePage: true }}
          locale={{ emptyText: t("parish.empty") }}
          onRow={(record) => ({
            onClick: () => {
              if (record.id) {
                navigate(modulePath("sys", `parish/update/${record.id}`));
              }
            },
            className: "cursor-pointer",
          })}
        />
      </section>
    </>
  );
}
