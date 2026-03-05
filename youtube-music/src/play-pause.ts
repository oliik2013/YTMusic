import { showHUD, showToast, Toast } from "@raycast/api";
import { initClient } from "./lib/client";
import { postPlayerPause } from "./generated/sdk.gen";

export default async function Command() {
  await initClient();
  try {
    const { data, error } = await postPlayerPause();

    if (error) {
      throw new Error(
        (error as unknown as Record<string, string>).error ||
          "Failed to toggle playback",
      );
    }

    const status = data?.is_paused ? "⏸ Paused" : "▶️ Playing";
    await showHUD(status);
  } catch (e) {
    await showToast(Toast.Style.Failure, "Error", String(e));
  }
}
