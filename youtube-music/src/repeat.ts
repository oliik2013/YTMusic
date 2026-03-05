import { showHUD } from "@raycast/api";
import { getPlayerState, postPlayerRepeat } from "./generated/sdk.gen";
import { initClient } from "./lib/client";

export default async function Command() {
  await initClient();

  const { data, error: getError } = await getPlayerState();
  if (getError) {
    console.error("Failed to get player state:", getError);
    return;
  }

  const current = data?.repeat || "off";
  const next = current === "off" ? "all" : current === "all" ? "one" : "off";

  const { error } = await postPlayerRepeat({ body: { repeat: next } });
  if (error) {
    console.error("Failed to set repeat:", error);
    return;
  }

  const label = next === "one" ? "One" : next === "all" ? "All" : "Off";
  await showHUD(`Repeat: ${label}`);
}
