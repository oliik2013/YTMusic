import { showHUD, showToast, Toast } from "@raycast/api";
import { initClient } from "./lib/client";
import { postPlayerNext } from "./generated/sdk.gen";

export default async function Command() {
  await initClient();
  try {
    const { data, error } = await postPlayerNext();

    if (error) {
      throw new Error(
        (error as unknown as Record<string, string>).error ||
          "Failed to skip track",
      );
    }

    await showHUD(`⏭ ${data?.current_track?.title || "Next track"}`);
  } catch (e) {
    await showToast(Toast.Style.Failure, "Error", String(e));
  }
}
