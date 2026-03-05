import { getPreferenceValues } from "@raycast/api";
import { client } from "../generated/client.gen";
import { getToken } from "./auth";

interface Preferences {
  apiUrl: string;
  authMode: "localhost" | "cookies";
}

export async function initClient(): Promise<void> {
  const { apiUrl, authMode } = getPreferenceValues<Preferences>();
  const baseUrl = apiUrl.replace(/\/+$/, "");

  client.setConfig({ baseUrl });

  const isLocalhost = /localhost|127\.0\.0\.1/.test(baseUrl);

  // If we are using cookie auth, or not on localhost, attach the token
  if (authMode !== "localhost" || !isLocalhost) {
    const token = await getToken();
    if (token) {
      client.interceptors.request.use((request) => {
        request.headers.set("X-Session-Token", token);
        return request;
      });
    } else {
      // Warning user when auth token is missing for actions that require it.
      // Commands handle this by capturing HTTP 401s and bubbling up to UI.
    }
  }
}
