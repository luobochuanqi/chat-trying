import axios from "axios";

export interface QuotaResponse {
  status: boolean;
  quota: number;
  credit_money: number;
  draw_count: number;
}

export async function getQuota(): Promise<QuotaResponse> {
  try {
    const response = await axios.get("/quota");
    return response.data as QuotaResponse;
  } catch (e) {
    console.debug(e);
    return { status: false, quota: 0, credit_money: 0, draw_count: 0 };
  }
}
