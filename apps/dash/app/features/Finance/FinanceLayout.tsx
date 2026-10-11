import { Button, Result, Spin } from "antd";
import { useTranslation } from "react-i18next";
import { Outlet, useNavigate } from "react-router";
import { useFinance } from "./FinanceContext";

export default function FinanceLayout() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isLoading, hasNoAccess } = useFinance();

  if (isLoading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center p-8">
        <Spin size="large" />
      </div>
    );
  }

  if (hasNoAccess) {
    return (
      <div
        className="mx-auto max-w-xl px-4 py-12"
        data-testid="finance-no-access"
      >
        <Result
          status="warning"
          title={t("modules.fin.noAccessTitle")}
          subTitle={t("modules.fin.noAccessDescription")}
          extra={
            <Button type="primary" onClick={() => navigate("/")}>
              {t("common.backToHome")}
            </Button>
          }
        />
      </div>
    );
  }

  return <Outlet />;
}
