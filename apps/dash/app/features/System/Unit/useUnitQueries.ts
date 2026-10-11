import type { components } from "@domus/openapi";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "~/lib/api";

export type Unit = components["schemas"]["model.UnitResponse"];
export type CreateUnitInput = components["schemas"]["model.CreateUnitRequest"];
export type UpdateUnitInput = components["schemas"]["model.UpdateUnitRequest"];
export type SearchUnitInput = components["schemas"]["model.SearchUnitRequest"];
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

export function useUnitsSearchQuery(filters?: SearchUnitInput) {
  const { POST } = useApi();
  return useQuery<Unit[], ApiError>({
    queryKey: ["units", filters],
    queryFn: async () => {
      const { data, error, response } = await POST("/units/search", {
        body: filters ?? {},
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data ?? [];
    },
  });
}

export function useUnitDetailQuery(id: number) {
  const { GET } = useApi();
  return useQuery<Unit | undefined, ApiError>({
    queryKey: ["units", id],
    queryFn: async () => {
      const { data, error, response } = await GET("/units/{id}", {
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

export function useCreateUnitMutation() {
  const { POST } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Unit | undefined, ApiError, CreateUnitInput>({
    mutationFn: async (body: CreateUnitInput) => {
      const { data, error, response } = await POST("/units", { body });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["units"] });
    },
  });
}

export function useUpdateUnitMutation(id: number) {
  const { PUT } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Unit | undefined, ApiError, UpdateUnitInput>({
    mutationFn: async (body: UpdateUnitInput) => {
      const { data, error, response } = await PUT("/units/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["units"] });
      queryClient.invalidateQueries({ queryKey: ["units", id] });
    },
  });
}

export function useDeleteUnitMutation() {
  const { DELETE } = useApi();
  const queryClient = useQueryClient();
  return useMutation<void, ApiError, number>({
    mutationFn: async (id: number) => {
      const { error, response } = await DELETE("/units/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["units"] });
    },
  });
}
