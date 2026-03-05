import { LocalStorage } from "@raycast/api";

const TOKEN_KEY = "ytm-session-token";
const EXPIRES_KEY = "ytm-token-expires";

export async function saveToken(
  token: string,
  expiresAt: string,
): Promise<void> {
  await LocalStorage.setItem(TOKEN_KEY, token);
  await LocalStorage.setItem(EXPIRES_KEY, expiresAt);
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

export async function clearToken(): Promise<void> {
  await LocalStorage.removeItem(TOKEN_KEY);
  await LocalStorage.removeItem(EXPIRES_KEY);
}
