import {
  Action,
  ActionPanel,
  List,
  Icon,
  Toast,
  showToast,
} from "@raycast/api";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { initClient } from "./lib/client";
import { QueryProvider } from "./components/QueryProvider";
import {
  getSearchOptions,
  postPlayerPlayMutation,
  postQueueAddMutation,
} from "./generated/@tanstack/react-query.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <List isLoading />;

  return (
    <QueryProvider>
      <SearchView />
    </QueryProvider>
  );
}

function SearchView() {
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState("songs");

  const { data, isLoading, error } = useQuery({
    ...getSearchOptions({ query: { q: query, filter, limit: 20 } }),
    enabled: query.trim().length > 0,
  });

  const playMutation = useMutation({
    ...postPlayerPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing track"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play", String(err)),
  });

  const queueMutation = useMutation({
    ...postQueueAddMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Added to queue"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to add to queue", String(err)),
  });

  if (error) {
    showToast(Toast.Style.Failure, "Search Failed", String(error));
  }

  return (
    <List
      isLoading={isLoading}
      searchBarPlaceholder="Search YouTube Music..."
      onSearchTextChange={setQuery}
      throttle
      searchBarAccessory={
        <List.Dropdown
          tooltip="Filter Type"
          storeValue={true}
          onChange={setFilter}
        >
          <List.Dropdown.Item title="Songs" value="songs" />
          <List.Dropdown.Item title="Albums" value="albums" />
          <List.Dropdown.Item title="Artists" value="artists" />
          <List.Dropdown.Item title="Playlists" value="playlists" />
          <List.Dropdown.Item title="Videos" value="videos" />
        </List.Dropdown>
      }
    >
      {data?.results?.map((result, i) => {
        const type = result.result_type;
        const track = result.track;

        if (type === "song" && track) {
          return (
            <List.Item
              key={`${track.video_id}-${i}`}
              title={track.title ?? "Unknown"}
              subtitle={`${track.artist} • ${track.album}`}
              icon={track.thumbnail_url ?? Icon.Music}
              actions={
                <ActionPanel>
                  <Action
                    title="Play Now"
                    icon={Icon.Play}
                    onAction={() =>
                      playMutation.mutate({
                        body: { video_id: track.video_id! },
                      })
                    }
                  />
                  <Action
                    title="Add to Queue"
                    icon={Icon.Plus}
                    onAction={() =>
                      queueMutation.mutate({
                        body: { video_id: track.video_id! },
                      })
                    }
                  />
                </ActionPanel>
              }
            />
          );
        }

        // Add other result types (album/artist/playlist) as needed,
        // for now just showing basic info
        return (
          <List.Item
            key={`other-${i}`}
            title={
              result.album?.title ||
              result.artist?.name ||
              result.playlist?.title ||
              "Unknown"
            }
            subtitle={type}
            icon={Icon.MagnifyingGlass}
          />
        );
      })}
    </List>
  );
}
