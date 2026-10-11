import { BankOutlined } from "@ant-design/icons";
import { Select } from "antd";
import { useTranslation } from "react-i18next";
import { useFinance } from "./FinanceContext";

export function FinanceUnitSelector() {
  const { t } = useTranslation();
  const { activeUnitId, units, switchUnit, isLoading } = useFinance();

  if (isLoading || units.length === 0) {
    return null;
  }

  return (
    <div
      className="flex items-center gap-1.5"
      data-testid="finance-unit-selector"
    >
      <Select
        size="middle"
        className="min-w-44"
        prefix={<BankOutlined style={{ color: "var(--domus-text-muted)" }} />}
        placeholder={t("modules.fin.selectUnit")}
        value={activeUnitId ?? undefined}
        onChange={(val) => void switchUnit(val)}
        options={units.map((u) => ({
          value: u.id,
          label: u.code ? `${u.name} (${u.code})` : u.name,
        }))}
      />
    </div>
  );
}
