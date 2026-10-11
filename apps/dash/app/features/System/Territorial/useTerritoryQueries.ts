import type { components } from "@domus/openapi";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "~/lib/api";

export type Territory = components["schemas"]["model.TerritoryResponse"];
export type CreateTerritoryInput =
  components["schemas"]["model.CreateTerritoryRequest"];
export type UpdateTerritoryInput =
  components["schemas"]["model.UpdateTerritoryRequest"];
export type SearchTerritoryInput =
  components["schemas"]["model.SearchTerritoryRequest"];
type ErrorResponse = components["schemas"]["httpx.ErrorResponse"];

export interface ApiError {
  status: number;
  message?: string;
  fields: Record<string, string>;
}

export function toApiError(status: number, body?: ErrorResponse): ApiError {
  const fields: Record<string, string> = {};
  for (const f of body?.error?.fields ?? []) {
    if (!f.field) continue;
    fields[f.field.charAt(0).toLowerCase() + f.field.slice(1)] = f.rule ?? "";
  }
  return { status, message: body?.error?.message, fields };
}

export function useTerritoriesSearchQuery(filters?: SearchTerritoryInput) {
  const { POST } = useApi();
  return useQuery<Territory[], ApiError>({
    queryKey: ["territories", filters],
    queryFn: async () => {
      const { data, error, response } = await POST("/territories/search", {
        body: filters ?? {},
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data ?? [];
    },
  });
}

export function useTerritoryDetailQuery(id: number) {
  const { GET } = useApi();
  return useQuery<Territory | undefined, ApiError>({
    queryKey: ["territories", id],
    queryFn: async () => {
      const { data, error, response } = await GET("/territories/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    enabled: Boolean(id) && !Number.isNaN(id),
  });
}

export function useCreateTerritoryMutation() {
  const { POST } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Territory | undefined, ApiError, CreateTerritoryInput>({
    mutationFn: async (body: CreateTerritoryInput) => {
      const { data, error, response } = await POST("/territories", { body });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["territories"] });
    },
  });
}

export function useUpdateTerritoryMutation(id: number) {
  const { PUT } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Territory | undefined, ApiError, UpdateTerritoryInput>({
    mutationFn: async (body: UpdateTerritoryInput) => {
      const { data, error, response } = await PUT("/territories/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["territories"] });
      queryClient.invalidateQueries({ queryKey: ["territories", id] });
    },
  });
}

export function useDeleteTerritoryMutation() {
  const { DELETE } = useApi();
  const queryClient = useQueryClient();
  return useMutation<void, ApiError, number>({
    mutationFn: async (id: number) => {
      const { error, response } = await DELETE("/territories/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["territories"] });
    },
  });
}
