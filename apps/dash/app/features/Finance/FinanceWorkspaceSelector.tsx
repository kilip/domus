import { BankOutlined } from "@ant-design/icons";
import { Select } from "antd";
import { useTranslation } from "react-i18next";
import { useFinance } from "./FinanceContext";

export function FinanceWorkspaceSelector() {
  const { t } = useTranslation();
  const { activeWorkspaceId, workspaces, switchWorkspace, isLoading } =
    useFinance();

  if (isLoading || workspaces.length === 0) {
    return null;
  }

  return (
    <div
      className="flex items-center gap-1.5"
      data-testid="finance-workspace-selector"
    >
      <Select
        size="middle"
        className="min-w-44"
        prefix={<BankOutlined style={{ color: "var(--domus-text-muted)" }} />}
        placeholder={t("modules.fin.selectWorkspace")}
        value={activeWorkspaceId ?? undefined}
        onChange={(val) => void switchWorkspace(val)}
        options={workspaces.map((w) => ({
          value: w.id,
          label: w.name,
        }))}
      />
    </div>
  );
}

// Backwards compatibility export
export const FinanceUnitSelector = FinanceWorkspaceSelector;
