import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { useApi } from "~/lib/api";
import { useActiveModule } from "~/shared/modules/useActiveModule";
import type { FinanceUnit } from "./types";

export const ACTIVE_FINANCE_UNIT_ID_KEY = "activeFinanceUnitId";
export const FINANCE_UNITS_KEY = "financeUnits";

interface FinanceContextValue {
  activeUnitId: number | null;
  activeUnit: FinanceUnit | null;
  units: FinanceUnit[];
  isLoading: boolean;
  hasNoAccess: boolean;
  switchUnit: (unitId: number) => Promise<void>;
  refreshUnits: () => Promise<void>;
}

const FinanceContext = createContext<FinanceContextValue | null>(null);

export function FinanceProvider({ children }: { children: ReactNode }) {
  const api = useApi();
  const [activeUnitId, setActiveUnitId] = useState<number | null>(null);
  const [activeUnit, setActiveUnit] = useState<FinanceUnit | null>(null);
  const [units, setUnits] = useState<FinanceUnit[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [hasNoAccess, setHasNoAccess] = useState(false);

  const activateUnit = useCallback(
    async (unit: FinanceUnit) => {
      try {
        await api.POST("/fin/activate", {
          body: {
            parishId: unit.parishId,
            unitId: unit.id,
          },
        });
      } catch (err) {
        console.error("Failed to activate finance unit context:", err);
      }
    },
    [api],
  );

  const initUnits = useCallback(async () => {
    setIsLoading(true);
    setHasNoAccess(false);

    const storedUnitIdStr = localStorage.getItem(ACTIVE_FINANCE_UNIT_ID_KEY);
    const storedUnitsStr = localStorage.getItem(FINANCE_UNITS_KEY);

    let parsedStoredUnits: FinanceUnit[] | null = null;
    if (storedUnitsStr) {
      try {
        parsedStoredUnits = JSON.parse(storedUnitsStr) as FinanceUnit[];
      } catch {
        parsedStoredUnits = null;
      }
    }

    if (storedUnitIdStr && parsedStoredUnits && parsedStoredUnits.length > 0) {
      const storedId = Number(storedUnitIdStr);
      const matchedUnit = parsedStoredUnits.find((u) => u.id === storedId);

      if (matchedUnit) {
        setUnits(parsedStoredUnits);
        setActiveUnitId(matchedUnit.id);
        setActiveUnit(matchedUnit);
        await activateUnit(matchedUnit);
        setIsLoading(false);
        return;
      }
    }

    // Fallback: Fetch accessible units from API
    try {
      const res = await api.GET("/fin/units");
      const fetchedUnits = (res.data?.data ?? []) as FinanceUnit[];

      if (fetchedUnits.length === 0) {
        setUnits([]);
        setActiveUnitId(null);
        setActiveUnit(null);
        setHasNoAccess(true);
        setIsLoading(false);
        return;
      }

      const firstUnit = fetchedUnits[0];
      setUnits(fetchedUnits);
      setActiveUnitId(firstUnit.id);
      setActiveUnit(firstUnit);

      localStorage.setItem(ACTIVE_FINANCE_UNIT_ID_KEY, String(firstUnit.id));
      localStorage.setItem(FINANCE_UNITS_KEY, JSON.stringify(fetchedUnits));

      await activateUnit(firstUnit);
    } catch (err) {
      console.error("Failed to fetch finance units:", err);
      setHasNoAccess(true);
    } finally {
      setIsLoading(false);
    }
  }, [api, activateUnit]);

  const { activeModule } = useActiveModule();

  useEffect(() => {
    if (activeModule?.id === "fin") {
      void initUnits();
    } else {
      setIsLoading(false);
      setHasNoAccess(false);
    }
  }, [activeModule?.id, initUnits]);

  const switchUnit = useCallback(
    async (unitId: number) => {
      const targetUnit = units.find((u) => u.id === unitId);
      if (!targetUnit) return;

      setActiveUnitId(targetUnit.id);
      setActiveUnit(targetUnit);
      localStorage.setItem(ACTIVE_FINANCE_UNIT_ID_KEY, String(targetUnit.id));

      await activateUnit(targetUnit);
    },
    [units, activateUnit],
  );

  return (
    <FinanceContext.Provider
      value={{
        activeUnitId,
        activeUnit,
        units,
        isLoading,
        hasNoAccess,
        switchUnit,
        refreshUnits: initUnits,
      }}
    >
      {children}
    </FinanceContext.Provider>
  );
}

export function useFinance(): FinanceContextValue {
  const ctx = useContext(FinanceContext);
  if (!ctx) {
    throw new Error("useFinance must be used within a FinanceProvider");
  }
  return ctx;
}
