import { useEffect, useState } from "react";
import axios from "axios";
import { Loader2, Check, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { PaginationAction } from "@/components/ui/pagination";

type GalleryItem = {
  id: number;
  prompt: string;
  image_url: string;
  author_name: string;
  status: string;
  created_at: string;
};

const statusLabels: Record<string, string> = {
  pending: "待审核",
  approved: "已通过",
  rejected: "已拒绝",
};

function AdminGallery() {
  const [items, setItems] = useState<GalleryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(0);
  const [total, setTotal] = useState(0);
  const [statusFilter, setStatusFilter] = useState("");

  useEffect(() => {
    fetchItems();
  }, [page, statusFilter]);

  async function fetchItems() {
    setLoading(true);
    try {
      const params = new URLSearchParams({ page: String(page), limit: "10" });
      if (statusFilter) params.set("status", statusFilter);
      const res = await axios.get(`/api/admin/gallery/list?${params}`);
      if (res.data.status) {
        setItems(res.data.data);
        setTotal(res.data.total);
      }
    } catch {
      toast.error("加载失败");
    }
    setLoading(false);
  }

  async function approve(id: number) {
    try {
      const res = await axios.post("/api/admin/gallery/approve", { id });
      if (res.data.status) {
        toast.success("已通过");
        fetchItems();
      }
    } catch {
      toast.error("操作失败");
    }
  }

  async function reject(id: number) {
    try {
      const res = await axios.post("/api/admin/gallery/reject", { id });
      if (res.data.status) {
        toast.success("已拒绝");
        fetchItems();
      }
    } catch {
      toast.error("操作失败");
    }
  }

  return (
    <div className="p-4">
      <div className="flex items-center gap-2 mb-4">
        <Button
          size="sm"
          variant={statusFilter === "" ? "default" : "outline"}
          onClick={() => { setStatusFilter(""); setPage(0); }}
        >
          全部
        </Button>
        <Button
          size="sm"
          variant={statusFilter === "pending" ? "default" : "outline"}
          onClick={() => { setStatusFilter("pending"); setPage(0); }}
        >
          待审核
        </Button>
        <Button
          size="sm"
          variant={statusFilter === "approved" ? "default" : "outline"}
          onClick={() => { setStatusFilter("approved"); setPage(0); }}
        >
          已通过
        </Button>
        <Button
          size="sm"
          variant={statusFilter === "rejected" ? "default" : "outline"}
          onClick={() => { setStatusFilter("rejected"); setPage(0); }}
        >
          已拒绝
        </Button>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin" />
        </div>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>作者</TableHead>
              <TableHead>Prompt</TableHead>
              <TableHead>图片</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>时间</TableHead>
              <TableHead>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.id}</TableCell>
                <TableCell>{item.author_name}</TableCell>
                <TableCell className="max-w-xs truncate">{item.prompt}</TableCell>
                <TableCell>
                  <img
                    src={item.image_url}
                    alt={item.prompt}
                    className="w-16 h-16 object-cover rounded"
                  />
                </TableCell>
                <TableCell>
                  <Badge
                    variant={
                      item.status === "approved"
                        ? "default"
                        : item.status === "rejected"
                          ? "destructive"
                          : "outline"
                    }
                  >
                    {statusLabels[item.status] || item.status}
                  </Badge>
                </TableCell>
                <TableCell className="whitespace-nowrap">{item.created_at}</TableCell>
                <TableCell>
                  {item.status === "pending" && (
                    <div className="flex gap-1">
                      <Button size="icon" variant="ghost" onClick={() => approve(item.id)}>
                        <Check className="h-4 w-4 text-green-500" />
                      </Button>
                      <Button size="icon" variant="ghost" onClick={() => reject(item.id)}>
                        <X className="h-4 w-4 text-red-500" />
                      </Button>
                    </div>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <PaginationAction
        current={page}
        total={total}
        onPageChange={setPage}
        offset
      />
    </div>
  );
}

export default AdminGallery;
