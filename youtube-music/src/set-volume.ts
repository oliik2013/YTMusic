import { showHUD, showToast, Toast } from "@raycast/api";
import { initClient } from "./lib/client";
import { postPlayerVolume } from "./generated/sdk.gen";

export default async function Command(props: {
  arguments: { volume: string };
}) {
  await initClient();
  const vol = parseInt(props.arguments.volume, 10);

  if (isNaN(vol) || vol < 0 || vol > 100) {
    await showToast(Toast.Style.Failure, "Volume must be 0-100");
    return;
  }

  try {
    const { data, error } = await postPlayerVolume({ body: { volume: vol } });

    if (error) {
      throw new Error(
        (error as unknown as Record<string, string>).error ||
          "Failed to set volume",
      );
    }

    await showHUD(`🔊 Volume: ${data?.volume}%`);
  } catch (e) {
    await showToast(Toast.Style.Failure, "Error", String(e));
  }
}
