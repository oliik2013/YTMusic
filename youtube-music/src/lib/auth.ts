import { LocalStorage } from "@raycast/api";

const TOKEN_KEY = "ytm-session-token";
const EXPIRES_KEY = "ytm-token-expires";
const USES_PRESEEDED_KEY = "ytm-uses-preseeded";

export type ErrorCode =
  | "SESSION_EXPIRED"
  | "COOKIES_EXPIRED"
  | "CONFIG_UNCHANGED"
  | "NO_PRESEEDED_COOKIES";

export interface AuthError {
  error: string;
  code: number;
  error_code?: ErrorCode;
}

export interface RefreshResponse {
  token: string;
  expires_at: string;
  used_preseeded: boolean;
}

export async function saveToken(
  token: string,
  expiresAt: string,
  usesPreseeded: boolean = false,
): Promise<void> {
  await LocalStorage.setItem(TOKEN_KEY, token);
  await LocalStorage.setItem(EXPIRES_KEY, expiresAt);
  await LocalStorage.setItem(USES_PRESEEDED_KEY, usesPreseeded ? "true" : "false");
}

export async function getToken(): Promise<string | null> {
  const token = await LocalStorage.getItem<string>(TOKEN_KEY);
  const expiresAt = await LocalStorage.getItem<string>(EXPIRES_KEY);

  if (!token || !expiresAt) {
    return null;
  }

  const expiryDate = new Date(expiresAt);
  if (expiryDate <= new Date()) {
    await clearToken();
    return null;
  }

  return token;
}

export async function getUsesPreseeded(): Promise<boolean> {
  const value = await LocalStorage.getItem<string>(USES_PRESEEDED_KEY);
  return value === "true";
}

export async function clearToken(): Promise<void> {
  await LocalStorage.removeItem(TOKEN_KEY);
  await LocalStorage.removeItem(EXPIRES_KEY);
  await LocalStorage.removeItem(USES_PRESEEDED_KEY);
}

export function isAuthError(error: unknown): error is AuthError {
  return (
    typeof error === "object" &&
    error !== null &&
    "error" in error &&
    "code" in error
  );
}

export function needsCookieRefresh(error: AuthError): boolean {
  return (
    error.code === 401 &&
    (error.error_code === "COOKIES_EXPIRED" ||
      error.error_code === "CONFIG_UNCHANGED" ||
      error.error_code === "NO_PRESEEDED_COOKIES")
  );
}
