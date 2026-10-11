import { token as fetchToken } from "@domus/better-auth/client";

/** Token dianggap kedaluwarsa sedikit lebih awal supaya tidak kadaluarsa di tengah request. */
const EXPIRY_SKEW_MS = 30_000;

interface CachedToken {
  value: string;
  /** Epoch ms. `0` kalau klaim `exp` tidak terbaca (token selalu diperbarui). */
  expiresAt: number;
}

let cached: CachedToken | null = null;
let inflight: Promise<string | null> | null = null;

/** Ambil klaim `exp` (detik) dari payload JWT tanpa verifikasi; verifikasi tugas API. */
function readExpiry(jwt: string): number {
  try {
    const payload = jwt.split(".")[1];
    if (!payload) return 0;
    const base64 = payload.replace(/-/g, "+").replace(/_/g, "/");
    const json = JSON.parse(atob(base64)) as { exp?: number };
    return typeof json.exp === "number" ? json.exp * 1000 : 0;
  } catch {
    return 0;
  }
}

function isFresh(entry: CachedToken | null): entry is CachedToken {
  return entry !== null && entry.expiresAt - EXPIRY_SKEW_MS > Date.now();
}

async function requestToken(): Promise<string | null> {
  const { data } = await fetchToken().catch(() => ({ data: null }));
  const value = data?.token;
  if (!value) {
    cached = null;
    return null;
  }
  cached = { value, expiresAt: readExpiry(value) };
  return value;
}

/**
 * JWT untuk header `Authorization`. Memakai cache selama belum kedaluwarsa,
 * kalau sudah (atau `force`) minta token baru ke auth server. Request
 * bersamaan berbagi satu permintaan token supaya tidak membanjiri `/token`.
 * Mengembalikan `null` kalau sesi sudah tidak ada.
 */
export async function getAccessToken(force = false): Promise<string | null> {
  if (!force && isFresh(cached)) return cached.value;

  inflight ??= requestToken().finally(() => {
    inflight = null;
  });
  return inflight;
}

/** Buang token cache, mis. saat logout. */
export function clearAccessToken() {
  cached = null;
}
