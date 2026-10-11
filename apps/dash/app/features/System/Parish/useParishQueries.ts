import type { components } from "@domus/openapi";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "~/lib/api";

export type Parish = components["schemas"]["model.ParishResponse"];
export type CreateParishInput =
  components["schemas"]["model.CreateParishRequest"];
export type UpdateParishInput =
  components["schemas"]["model.UpdateParishRequest"];
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

export function useParishesListQuery() {
  const { GET } = useApi();
  return useQuery<Parish[], ApiError>({
    queryKey: ["parishes"],
    queryFn: async () => {
      const { data, error, response } = await GET("/parishes");
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data ?? [];
    },
  });
}

export function useParishDetailQuery(id: number) {
  const { GET } = useApi();
  return useQuery<Parish | undefined, ApiError>({
    queryKey: ["parishes", id],
    queryFn: async () => {
      const { data, error, response } = await GET("/parishes/{id}", {
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

export function useCreateParishMutation() {
  const { POST } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Parish | undefined, ApiError, CreateParishInput>({
    mutationFn: async (body: CreateParishInput) => {
      const { data, error, response } = await POST("/parishes", { body });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["parishes"] });
    },
  });
}

export function useUpdateParishMutation(id: number) {
  const { PUT } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Parish | undefined, ApiError, UpdateParishInput>({
    mutationFn: async (body: UpdateParishInput) => {
      const { data, error, response } = await PUT("/parishes/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["parishes"] });
      queryClient.invalidateQueries({ queryKey: ["parishes", id] });
    },
  });
}

export function useDeleteParishMutation() {
  const { DELETE } = useApi();
  const queryClient = useQueryClient();
  return useMutation<void, ApiError, number>({
    mutationFn: async (id: number) => {
      const { error, response } = await DELETE("/parishes/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["parishes"] });
    },
  });
}
