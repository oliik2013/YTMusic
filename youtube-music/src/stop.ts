import { showHUD, showToast, Toast } from "@raycast/api";
import { initClient } from "./lib/client";
import { postPlayerStop } from "./generated/sdk.gen";

export default async function Command() {
  await initClient();
  try {
    const { error } = await postPlayerStop();

    if (error) {
      throw new Error(
        (error as unknown as Record<string, string>).error ||
          "Failed to stop playback",
      );
    }

    await showHUD("⏹ Stopped");
  } catch (e) {
    await showToast(Toast.Style.Failure, "Error", String(e));
  }
}
