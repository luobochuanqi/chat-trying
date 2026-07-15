import { useEffect, useState } from "react";
import axios from "axios";
import { useSelector } from "react-redux";
import { selectAuthenticated } from "@/store/auth";
import { Loader2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

type MarketModel = {
  id: string;
  name: string;
  description: string;
  default: boolean;
  tag?: string[];
};

type QuotaInfo = {
  credit_money: number;
  draw_count: number;
  quota: number;
};

function MarketPage() {
  const [models, setModels] = useState<MarketModel[]>([]);
  const [quota, setQuota] = useState<QuotaInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const auth = useSelector(selectAuthenticated);

  useEffect(() => {
    async function load() {
      try {
        const res = await axios.get("/api/v1/market");
        if (Array.isArray(res.data)) {
          setModels(res.data);
        }
      } catch { /* ignore */ }

      try {
        if (auth) {
          const res = await axios.get("/api/quota");
          if (res.data) setQuota(res.data);
        }
      } catch { /* ignore */ }

      setLoading(false);
    }
    load();
  }, [auth]);

  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  return (
    <div className="p-4 md:p-8 max-w-4xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">模型市场</h1>

      {auth && quota && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-lg">我的额度</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex gap-8">
              <div>
                <p className="text-sm text-muted-foreground">¥余额</p>
                <p className="text-xl font-bold">{(quota.credit_money ?? 0).toFixed(2)}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">生图次数</p>
                <p className="text-xl font-bold">{quota.draw_count ?? 0}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {models.map((model) => (
          <Card key={model.id}>
            <CardHeader>
              <CardTitle className="text-base">{model.name}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground">{model.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

export default MarketPage;
