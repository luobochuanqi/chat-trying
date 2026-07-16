import { useState } from "react";
import axios from "axios";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Loader2, Send } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card.tsx";
import { toast } from "sonner";

function DrawInterface() {
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ url: string; prompt: string } | null>(null);
  const [submitted, setSubmitted] = useState(false);

  async function handleGenerate() {
    if (!prompt.trim() || loading) return;
    setLoading(true);
    setResult(null);
    try {
      const res = await axios.post("/api/draw", { prompt: prompt.trim() });
      if (res.data?.url) {
        setResult({ url: res.data.url, prompt: res.data.prompt });
      }
    } catch (err: any) {
      const msg = err?.response?.data?.error || "生成失败";
      toast.error(msg);
    }
    setLoading(false);
  }

  async function handleSubmitToGallery() {
    if (!result || submitted) return;
    try {
      await axios.post("/api/gallery/submit", {
        image_url: result.url,
        prompt: result.prompt,
      });
      setSubmitted(true);
      toast.success("已提交，等待管理员审核");
    } catch {
      toast.error("提交失败");
    }
  }

  return (
    <div className="flex flex-col items-center justify-center h-full w-full p-6 gap-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold mb-1">AI 生图</h2>
        <p className="text-sm text-muted-foreground">输入提示词，一键生成图片</p>
      </div>

      <div className="w-full max-w-lg flex gap-2">
        <Input
          placeholder="描述你想生成的画面..."
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleGenerate()}
          disabled={loading}
          className="flex-1"
        />
        <Button onClick={handleGenerate} disabled={loading || !prompt.trim()}>
          {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          <span className="ml-1">生成</span>
        </Button>
      </div>

      {result && (
        <div className="w-full max-w-lg">
          <Card className="overflow-hidden">
            <div className="aspect-square overflow-hidden bg-muted">
              <img
                src={result.url}
                alt={result.prompt}
                className="w-full h-full object-cover"
              />
            </div>
            <CardContent className="p-3">
              <p className="text-sm mb-2">{result.prompt}</p>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={handleSubmitToGallery}
                  disabled={submitted}
                >
                  {submitted ? "已提交" : "提交到作品墙"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}

export default DrawInterface;
