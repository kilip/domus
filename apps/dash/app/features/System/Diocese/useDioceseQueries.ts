import type { components } from "@domus/openapi";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "~/lib/api";

export type Diocese = components["schemas"]["model.DioceseResponse"];
export type DioceseInput = components["schemas"]["model.CreateDioceseRequest"];
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

export function useDiocesesListQuery() {
  const { GET } = useApi();
  return useQuery<Diocese[], ApiError>({
    queryKey: ["dioceses"],
    queryFn: async () => {
      const { data, error, response } = await GET("/dioceses");
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data ?? [];
    },
  });
}

export function useDioceseDetailQuery(id: number) {
  const { GET } = useApi();
  return useQuery<Diocese | undefined, ApiError>({
    queryKey: ["dioceses", id],
    queryFn: async () => {
      const { data, error, response } = await GET("/dioceses/{id}", {
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

export function useCreateDioceseMutation() {
  const { POST } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Diocese | undefined, ApiError, DioceseInput>({
    mutationFn: async (body: DioceseInput) => {
      const { data, error, response } = await POST("/dioceses", { body });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dioceses"] });
    },
  });
}

export function useUpdateDioceseMutation(id: number) {
  const { PUT } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Diocese | undefined, ApiError, DioceseInput>({
    mutationFn: async (body: DioceseInput) => {
      const { data, error, response } = await PUT("/dioceses/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dioceses"] });
      queryClient.invalidateQueries({ queryKey: ["dioceses", id] });
    },
  });
}

export function useDeleteDioceseMutation() {
  const { DELETE } = useApi();
  const queryClient = useQueryClient();
  return useMutation<void, ApiError, number>({
    mutationFn: async (id: number) => {
      const { error, response } = await DELETE("/dioceses/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["dioceses"] });
    },
  });
}
