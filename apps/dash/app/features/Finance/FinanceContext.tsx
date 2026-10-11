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
import type { FinanceWorkspace } from "./types";

export const ACTIVE_FINANCE_WORKSPACE_ID_KEY = "activeFinanceWorkspaceId";
export const FINANCE_WORKSPACES_KEY = "financeWorkspaces";
// Aliases for compatibility
export const ACTIVE_FINANCE_UNIT_ID_KEY = ACTIVE_FINANCE_WORKSPACE_ID_KEY;
export const FINANCE_UNITS_KEY = FINANCE_WORKSPACES_KEY;

interface FinanceContextValue {
  activeWorkspaceId: number | null;
  activeWorkspace: FinanceWorkspace | null;
  workspaces: FinanceWorkspace[];
  // Backwards compatibility properties
  activeUnitId: number | null;
  activeUnit: FinanceWorkspace | null;
  units: FinanceWorkspace[];
  isLoading: boolean;
  hasNoAccess: boolean;
  switchWorkspace: (workspaceId: number) => Promise<void>;
  switchUnit: (unitId: number) => Promise<void>;
  refreshWorkspaces: () => Promise<void>;
  refreshUnits: () => Promise<void>;
}

const FinanceContext = createContext<FinanceContextValue | null>(null);

export function FinanceProvider({ children }: { children: ReactNode }) {
  const api = useApi();
  const [activeWorkspaceId, setActiveWorkspaceId] = useState<number | null>(
    null,
  );
  const [activeWorkspace, setActiveWorkspace] =
    useState<FinanceWorkspace | null>(null);
  const [workspaces, setWorkspaces] = useState<FinanceWorkspace[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [hasNoAccess, setHasNoAccess] = useState(false);

  const activateWorkspace = useCallback(
    async (workspace: FinanceWorkspace) => {
      try {
        await api.POST("/fin/activate", {
          body: {
            parishId: workspace.parishId,
            workspaceId: workspace.id,
          },
        });
      } catch (err) {
        console.error("Failed to activate finance workspace context:", err);
      }
    },
    [api],
  );

  const initWorkspaces = useCallback(async () => {
    setIsLoading(true);
    setHasNoAccess(false);

    const storedWorkspaceIdStr =
      localStorage.getItem(ACTIVE_FINANCE_WORKSPACE_ID_KEY) ||
      localStorage.getItem("activeFinanceUnitId");
    const storedWorkspacesStr =
      localStorage.getItem(FINANCE_WORKSPACES_KEY) ||
      localStorage.getItem("financeUnits");

    let parsedStoredWorkspaces: FinanceWorkspace[] | null = null;
    if (storedWorkspacesStr) {
      try {
        parsedStoredWorkspaces = JSON.parse(
          storedWorkspacesStr,
        ) as FinanceWorkspace[];
      } catch {
        parsedStoredWorkspaces = null;
      }
    }

    if (
      storedWorkspaceIdStr &&
      parsedStoredWorkspaces &&
      parsedStoredWorkspaces.length > 0
    ) {
      const storedId = Number(storedWorkspaceIdStr);
      const matched = parsedStoredWorkspaces.find((w) => w.id === storedId);

      if (matched) {
        setWorkspaces(parsedStoredWorkspaces);
        setActiveWorkspaceId(matched.id);
        setActiveWorkspace(matched);
        await activateWorkspace(matched);
        setIsLoading(false);
        return;
      }
    }

    // Fallback: Fetch accessible workspaces from API
    try {
      const res = await api.GET("/fin/workspaces");
      const fetchedWorkspaces = (res.data?.data ?? []) as FinanceWorkspace[];

      if (fetchedWorkspaces.length === 0) {
        setWorkspaces([]);
        setActiveWorkspaceId(null);
        setActiveWorkspace(null);
        setHasNoAccess(true);
        setIsLoading(false);
        return;
      }

      const firstWorkspace = fetchedWorkspaces[0];
      setWorkspaces(fetchedWorkspaces);
      setActiveWorkspaceId(firstWorkspace.id);
      setActiveWorkspace(firstWorkspace);

      localStorage.setItem(
        ACTIVE_FINANCE_WORKSPACE_ID_KEY,
        String(firstWorkspace.id),
      );
      localStorage.setItem(
        FINANCE_WORKSPACES_KEY,
        JSON.stringify(fetchedWorkspaces),
      );

      await activateWorkspace(firstWorkspace);
    } catch (err) {
      console.error("Failed to fetch finance workspaces:", err);
      setHasNoAccess(true);
    } finally {
      setIsLoading(false);
    }
  }, [api, activateWorkspace]);

  const { activeModule } = useActiveModule();

  useEffect(() => {
    if (activeModule?.id === "fin") {
      void initWorkspaces();
    } else {
      setIsLoading(false);
      setHasNoAccess(false);
    }
  }, [activeModule?.id, initWorkspaces]);

  const switchWorkspace = useCallback(
    async (workspaceId: number) => {
      const target = workspaces.find((w) => w.id === workspaceId);
      if (!target) return;

      setActiveWorkspaceId(target.id);
      setActiveWorkspace(target);
      localStorage.setItem(ACTIVE_FINANCE_WORKSPACE_ID_KEY, String(target.id));

      await activateWorkspace(target);
    },
    [workspaces, activateWorkspace],
  );

  return (
    <FinanceContext.Provider
      value={{
        activeWorkspaceId,
        activeWorkspace,
        workspaces,
        activeUnitId: activeWorkspaceId,
        activeUnit: activeWorkspace,
        units: workspaces,
        isLoading,
        hasNoAccess,
        switchWorkspace,
        switchUnit: switchWorkspace,
        refreshWorkspaces: initWorkspaces,
        refreshUnits: initWorkspaces,
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
