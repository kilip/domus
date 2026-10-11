import type { components } from "@domus/openapi";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApi } from "~/lib/api";

export type Workspace = components["schemas"]["model.WorkspaceResponse"];
export type CreateWorkspaceInput =
  components["schemas"]["model.CreateWorkspaceRequest"];
export type UpdateWorkspaceInput =
  components["schemas"]["model.UpdateWorkspaceRequest"];
export type SearchWorkspaceInput =
  components["schemas"]["model.SearchWorkspaceRequest"];
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

export function useWorkspacesSearchQuery(filters?: SearchWorkspaceInput) {
  const { POST } = useApi();
  return useQuery<Workspace[], ApiError>({
    queryKey: ["workspaces", filters],
    queryFn: async () => {
      const { data, error, response } = await POST("/workspaces/search", {
        body: filters ?? {},
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data ?? [];
    },
  });
}

export function useWorkspaceDetailQuery(id: number) {
  const { GET } = useApi();
  return useQuery<Workspace | undefined, ApiError>({
    queryKey: ["workspaces", id],
    queryFn: async () => {
      const { data, error, response } = await GET("/workspaces/{id}", {
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

export function useCreateWorkspaceMutation() {
  const { POST } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Workspace | undefined, ApiError, CreateWorkspaceInput>({
    mutationFn: async (body: CreateWorkspaceInput) => {
      const { data, error, response } = await POST("/workspaces", { body });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

export function useUpdateWorkspaceMutation(id: number) {
  const { PUT } = useApi();
  const queryClient = useQueryClient();
  return useMutation<Workspace | undefined, ApiError, UpdateWorkspaceInput>({
    mutationFn: async (body: UpdateWorkspaceInput) => {
      const { data, error, response } = await PUT("/workspaces/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
      return data?.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      queryClient.invalidateQueries({ queryKey: ["workspaces", id] });
    },
  });
}

export function useDeleteWorkspaceMutation() {
  const { DELETE } = useApi();
  const queryClient = useQueryClient();
  return useMutation<void, ApiError, number>({
    mutationFn: async (id: number) => {
      const { error, response } = await DELETE("/workspaces/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) {
        throw toApiError(response.status, error);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}
