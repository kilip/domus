import { type ApiClient, api } from "./client";

/**
 * Client API bertipe (openapi-fetch) yang sudah mengurus JWT.
 *
 * ```ts
 * const { GET, POST } = useApi();
 * const { data, error } = await GET("/dioceses");
 * const res = await GET("/dioceses/{id}", { params: { path: { id: 1 } } });
 * ```
 *
 * Instance-nya singleton sehingga referensinya stabil dan aman dipakai
 * sebagai dependency `useEffect`/`useCallback`.
 */
export function useApi(): ApiClient {
  return api;
}
