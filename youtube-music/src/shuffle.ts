import { showHUD } from "@raycast/api";
import { getPlayerState, postPlayerShuffle } from "./generated/sdk.gen";
import { initClient } from "./lib/client";

export default async function Command() {
  await initClient();

  const { data, error: getError } = await getPlayerState();
  if (getError) {
    console.error("Failed to get player state:", getError);
    return;
  }

  const { error } = await postPlayerShuffle();
  if (error) {
    console.error("Failed to toggle shuffle:", error);
    return;
  }

  const newState = !(data?.shuffle ?? false);
  await showHUD(`Shuffle: ${newState ? "On" : "Off"}`);
}
