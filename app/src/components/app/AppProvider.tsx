import { ThemeProvider } from "@/components/ThemeProvider.tsx";
import DialogManager from "@/dialogs";
import { useEffectAsync } from "@/utils/hook.ts";
import { bindMarket, getApiPlans } from "@/api/v1.ts";
import { useDispatch } from "react-redux";
import {
  stack,
  updateMessageUsage,
  updateMessageToolStatus,
  updateMasks,
  updateSupportModels,
  useMessageActions,
} from "@/store/chat.ts";
import { dispatchSubscriptionData, setTheme } from "@/store/globals.ts";
import { infoEvent } from "@/events/info.ts";
import { setForm } from "@/store/info.ts";
import { themeEvent } from "@/events/theme.ts";
import { useEffect } from "react";
import { deductQuota } from "@/store/quota.ts";

function AppProvider({ children }: { children?: React.ReactNode }) {
  const dispatch = useDispatch();
  const { receive } = useMessageActions();

  useEffect(() => {
    infoEvent.bind((data) => dispatch(setForm(data)));
    themeEvent.bind((theme) => dispatch(setTheme(theme)));

    stack.setCallback(async (id, message: any) => {
      if (message.keyword === "usage") {
        dispatch(updateMessageUsage({ id, data: {
          cacheHitTokens: message.cacheHitTokens ?? 0,
          cacheMissTokens: message.cacheMissTokens ?? 0,
          completionTokens: message.completionTokens ?? 0,
          cost: message.quota ?? 0,
        }}));
        dispatch(deductQuota(message.quota ?? 0));
      } else if (message.keyword === "tool_status") {
        dispatch(updateMessageToolStatus({ id, data: {
          status: message.status ?? "",
          toolName: message.toolName ?? "",
        }}));
      } else {
        await receive(id, message);
      }
    });
  }, []);

  useEffectAsync(async () => {
    updateSupportModels(dispatch, await bindMarket());
    dispatchSubscriptionData(dispatch, await getApiPlans());
    await updateMasks(dispatch);
  }, []);

  return (
    <ThemeProvider>
      <DialogManager />
      {children}
    </ThemeProvider>
  );
}

export default AppProvider;
