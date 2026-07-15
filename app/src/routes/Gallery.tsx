import { useEffect, useState } from "react";
import axios from "axios";
import { Loader2 } from "lucide-react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

type GalleryItem = {
  id: number;
  prompt: string;
  image_url: string;
  author_name: string;
  created_at: string;
};

function PublicGallery() {
  const [items, setItems] = useState<GalleryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const limit = 20;

  useEffect(() => {
    fetchItems(1);
  }, []);

  async function fetchItems(p: number) {
    setLoading(true);
    try {
      const res = await axios.get(`/api/gallery?page=${p}&limit=${limit}`);
      if (res.data.status) {
        setItems(p === 1 ? res.data.data : [...items, ...res.data.data]);
        setTotal(res.data.total);
        setPage(p);
      }
    } catch {
      console.error("failed to load gallery");
    }
    setLoading(false);
  }

  const hasMore = items.length < total;

  return (
    <div className="p-4 md:p-8 max-w-6xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">作品墙</h1>
      {items.length === 0 && loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {items.map((item) => (
              <Card key={item.id} className="overflow-hidden group">
                <div className="aspect-square overflow-hidden">
                  <img
                    src={item.image_url}
                    alt={item.prompt}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  />
                </div>
                <div className="p-3">
                  <p className="text-sm font-medium truncate">{item.prompt}</p>
                  <div className="flex justify-between items-center mt-1 text-xs text-muted-foreground">
                    <span>{item.author_name}</span>
                    <span>{item.created_at?.slice(0, 10)}</span>
                  </div>
                </div>
              </Card>
            ))}
          </div>
          {hasMore && (
            <div className="flex justify-center mt-8">
              <Button
                variant="outline"
                onClick={() => fetchItems(page + 1)}
                disabled={loading}
              >
                {loading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
                加载更多
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default PublicGallery;
