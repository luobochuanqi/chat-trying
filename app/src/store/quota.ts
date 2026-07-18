import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { RootState } from "./index.ts";
import { getQuota, QuotaResponse } from "@/api/quota.ts";

export const quotaSlice = createSlice({
  name: "quota",
  initialState: {
    quota: 0,
    credit_money: 0,
    draw_count: 0,
  },
  reducers: {
    setQuota: (state, action) => {
      state.quota = action.payload as number;
    },
    deductQuota: (state, action) => {
      state.quota = Math.max(0, state.quota - (action.payload as number));
    },
  },
  extraReducers: (builder) => {
    builder.addCase(refreshQuota.fulfilled, (state, action) => {
      state.quota = action.payload.quota;
      state.credit_money = action.payload.credit_money;
      state.draw_count = action.payload.draw_count;
    });
  },
});

export const { setQuota, deductQuota } = quotaSlice.actions;
export default quotaSlice.reducer;

export const quotaSelector = (state: RootState): number => state.quota.quota;
export const creditMoneySelector = (state: RootState): number => state.quota.credit_money;
export const drawCountSelector = (state: RootState): number => state.quota.draw_count;

export const refreshQuota = createAsyncThunk("quota/refreshQuota", async () => {
  const data = await getQuota();
  return data || { status: false, quota: 0, credit_money: 0, draw_count: 0 } as QuotaResponse;
});
