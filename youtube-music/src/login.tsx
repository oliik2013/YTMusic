import {
  Action,
  ActionPanel,
  Detail,
  Form,
  Toast,
  showToast,
} from "@raycast/api";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { initClient } from "./lib/client";
import { saveToken, clearToken } from "./lib/auth";
import { QueryProvider } from "./components/QueryProvider";
import { queryClient } from "./lib/query";
import {
  getAuthStatusOptions,
  postAuthLoginMutation,
  deleteAuthLogoutMutation,
} from "./generated/@tanstack/react-query.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <Detail isLoading />;

  return (
    <QueryProvider>
      <LoginView />
    </QueryProvider>
  );
}

function LoginView() {
  const { data: status, isLoading, refetch } = useQuery(getAuthStatusOptions());

  const loginMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: async (data) => {
      if (data.data?.token) {
        await saveToken(data.data.token, data.data.expires_at!);
        showToast(Toast.Style.Success, "Logged in successfully");
        // re-initialize client to use the new token
        await initClient();
        refetch();
      }
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Login failed", String(err)),
  });

  const logoutMutation = useMutation({
    ...deleteAuthLogoutMutation(),
    onSuccess: async () => {
      await clearToken();
      queryClient.clear();
      showToast(Toast.Style.Success, "Logged out");
      // re-initialize to clear interceptors
      await initClient();
      refetch();
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Logout failed", String(err)),
  });

  if (isLoading) return <Detail isLoading />;

  if (status?.authenticated) {
    return (
      <Detail
        markdown={`# You are authenticated\n\n**Account**: ${status.account_name}\n\n**Session Expires**: ${new Date(status.expires_at!).toLocaleString()}`}
        actions={
          <ActionPanel>
            <Action title="Logout" onAction={() => logoutMutation.mutate()} />
          </ActionPanel>
        }
      />
    );
  }

  return (
    <Form
      actions={
        <ActionPanel>
          <Action.SubmitForm
            title="Login"
            onSubmit={(values: { cookies: string }) =>
              loginMutation.mutate({ body: { cookies: values.cookies } })
            }
          />
        </ActionPanel>
      }
    >
      <Form.TextArea
        id="cookies"
        title="Browser Cookies"
        placeholder="Paste your 'Cookie' header string from the browser network tab..."
      />
    </Form>
  );
}
