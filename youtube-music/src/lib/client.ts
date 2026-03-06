import { getPreferenceValues } from "@raycast/api";
import { client } from "../generated/client.gen";
import { getToken, saveToken, clearToken, AuthError, needsCookieRefresh } from "./auth";

interface Preferences {
  apiUrl: string;
  authMode: "localhost" | "cookies";
}

let clientInitialized = false;

export async function initClient(): Promise<void> {
  const { apiUrl, authMode } = getPreferenceValues<Preferences>();
  const baseUrl = apiUrl.replace(/\/+$/, "");

  client.setConfig({ baseUrl });

  const isLocalhost = /localhost|127\.0\.0\.1/.test(baseUrl);

  if (authMode !== "localhost" || !isLocalhost) {
    const token = await getToken();
    if (token) {
      client.interceptors.request.use((request) => {
        request.headers.set("X-Session-Token", token);
        return request;
      });
    }
  }

  if (!clientInitialized) {
    client.interceptors.response.use(async (response, request) => {
      if (response.status === 401) {
        const newToken = response.headers.get("X-New-Session-Token");
        if (newToken) {
          await saveToken(newToken, new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(), true);
          return response;
        }
      }
      return response;
    });

    client.interceptors.error.use(async (error, response, request) => {
      if (response && response.status === 401) {
        const authError = error as AuthError;
        if (needsCookieRefresh(authError)) {
          await clearToken();
        }
      }
      return error;
    });

    clientInitialized = true;
  }
}
