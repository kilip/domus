import type { paths } from "@domus/openapi";
import createClient from "openapi-fetch";
import { clearAccessToken, getAccessToken } from "./token";

/** Base URL API Go. Override lewat `VITE_API_URL` di `.env`. */
export const API_BASE_URL: string =
  import.meta.env.VITE_API_URL ?? "http://localhost:8002";

function redirectToLogin(): never {
  clearAccessToken();
  const { pathname, search } = window.location;
  const target = `${pathname}${search}`;
  const to =
    target === "/" ? "/login" : `/login?redirect=${encodeURIComponent(target)}`;
  window.location.assign(to);
  throw new Error("Session expired");
}

function withToken(request: Request, token: string): Request {
  const next = new Request(request);
  next.headers.set("Authorization", `Bearer ${token}`);
  return next;
}

/**
 * `fetch` yang menyisipkan JWT dan memperbarui token bila perlu:
 * 1. token dicek/diperbarui sebelum request (lihat `getAccessToken`);
 * 2. kalau API tetap membalas 401 (mis. token dicabut), minta token baru
 *    lalu ulangi request satu kali;
 * 3. kalau sesi memang sudah habis, arahkan ke `/login`.
 */
const authFetch: typeof fetch = async (input, init) => {
  const request = new Request(input, init);
  // Body hanya bisa dibaca sekali; simpan salinan untuk percobaan ulang.
  const retryable = request.clone();

  const token = await getAccessToken();
  if (!token) redirectToLogin();

  const response = await fetch(withToken(request, token));
  if (response.status !== 401) return response;

  const fresh = await getAccessToken(true);
  if (!fresh) redirectToLogin();

  const retried = await fetch(withToken(retryable, fresh));
  if (retried.status === 401) redirectToLogin();
  return retried;
};

export const api = createClient<paths>({
  baseUrl: API_BASE_URL,
  fetch: authFetch,
});

export type ApiClient = typeof api;
