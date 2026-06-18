import { useCallback, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Plus, Download, UploadCloud } from "lucide-react";
import {
  listBuckets,
  getBucket,
  createBucket,
  deleteBucket,
  listFiles,
  getFile,
  uploadFile,
  deleteFile,
  downloadUrl,
  type Bucket,
  type FileItem,
} from "@/api/storage";
import { useAuth } from "@/hooks/useAuth";
import { ResourceListPage } from "@/components/list/ResourceListPage";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { ColumnDef } from "@/components/list/DataTable";
import {
  FormPageWrapper,
  FormField,
  DetailPageWrapper,
  DetailGrid,
  DetailSkeleton,
  NotFound,
  BulkDeleteButton,
  RowDeleteButton,
  DeleteButton,
} from "@/components/resource/shared";

const bucketColumns: ColumnDef<Bucket>[] = [
  {
    key: "id",
    header: "ID",
    className: "font-mono text-xs max-w-[140px] truncate",
    cell: (b) => b.id,
  },
  { key: "name", header: "名称", cell: (b) => b.name },
  {
    key: "created",
    header: "创建时间",
    cell: (b) => new Date(b.created_at).toLocaleString(),
  },
];

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

export function StorageListPage() {
  const { projectId } = useAuth();
  const queryClient = useQueryClient();
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const { data: buckets = [], isLoading } = useQuery({
    queryKey: ["buckets", projectId],
    queryFn: listBuckets,
    enabled: !!projectId,
  });

  const remove = useMutation({
    mutationFn: deleteBucket,
    onSuccess: () => {
      toast.success("Bucket 已删除");
      queryClient.invalidateQueries({ queryKey: ["buckets"] });
    },
  });

  const getSearchText = useCallback((b: Bucket) => `${b.id} ${b.name}`, []);

  const handleBulkDelete = async (selected: Bucket[], clear: () => void) => {
    setBulkDeleting(true);
    try {
      await Promise.all(selected.map((b) => deleteBucket(b.id)));
      toast.success(`已删除 ${selected.length} 个 Bucket`);
      queryClient.invalidateQueries({ queryKey: ["buckets"] });
      clear();
    } finally {
      setBulkDeleting(false);
    }
  };

  return (
    <ResourceListPage
      title="Storage"
      description="管理存储 Bucket"
      searchPlaceholder="搜索 Bucket 名称或 ID..."
      isLoading={isLoading}
      items={buckets}
      columns={bucketColumns}
      getSearchText={getSearchText}
      detailPath={(b) => `/console/storage/${b.id}`}
      toolbarActions={
        <Button asChild>
          <Link to="/console/storage/new">
            <Plus className="h-4 w-4 mr-2" />
            新建 Bucket
          </Link>
        </Button>
      }
      selectionActions={(selected, clear) => (
        <BulkDeleteButton
          count={selected.length}
          loading={bulkDeleting}
          onConfirm={() => handleBulkDelete(selected, clear)}
        />
      )}
      rowActions={(b) => (
        <RowDeleteButton onConfirm={() => remove.mutate(b.id)} loading={remove.isPending} />
      )}
      emptyTitle="暂无 Bucket"
      emptyDescription="创建 Bucket 以上传文件"
      emptyAction={
        <Button asChild>
          <Link to="/console/storage/new">新建 Bucket</Link>
        </Button>
      }
    />
  );
}

export function BucketNewPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");

  const mutation = useMutation({
    mutationFn: createBucket,
    onSuccess: (bucket) => {
      toast.success("Bucket 创建成功");
      queryClient.invalidateQueries({ queryKey: ["buckets"] });
      navigate(`/console/storage/${bucket.id}`);
    },
  });

  return (
    <FormPageWrapper
      title="新建 Bucket"
      backTo="/console/storage"
      submitLabel="创建"
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate({ name });
      }}
      loading={mutation.isPending}
    >
      <FormField id="name" label="Bucket 名称" value={name} onChange={setName} required placeholder="uploads" />
    </FormPageWrapper>
  );
}

export function BucketDetailPage() {
  const { bucketId } = useParams<{ bucketId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const { data: bucket, isLoading: bucketLoading } = useQuery({
    queryKey: ["buckets", bucketId],
    queryFn: () => getBucket(bucketId!),
    enabled: !!bucketId,
  });

  const { data: files = [], isLoading: filesLoading } = useQuery({
    queryKey: ["files", bucketId],
    queryFn: () => listFiles(bucketId!),
    enabled: !!bucketId,
  });

  const removeBucket = useMutation({
    mutationFn: deleteBucket,
    onSuccess: () => {
      toast.success("Bucket 已删除");
      queryClient.invalidateQueries({ queryKey: ["buckets"] });
      navigate("/console/storage");
    },
  });

  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadFile(bucketId!, file),
    onSuccess: () => {
      toast.success("文件上传成功");
      queryClient.invalidateQueries({ queryKey: ["files", bucketId] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (fileId: string) => deleteFile(bucketId!, fileId),
    onSuccess: () => {
      toast.success("文件已删除");
      queryClient.invalidateQueries({ queryKey: ["files", bucketId] });
    },
  });

  const fileColumns: ColumnDef<FileItem>[] = [
    { key: "name", header: "文件名", cell: (f) => f.name },
    { key: "size", header: "大小", cell: (f) => formatBytes(f.size) },
    { key: "type", header: "类型", cell: (f) => f.mime_type },
    {
      key: "created",
      header: "上传时间",
      cell: (f) => new Date(f.created_at).toLocaleString(),
    },
  ];

  const getFileSearchText = useCallback(
    (f: FileItem) => `${f.id} ${f.name} ${f.mime_type}`,
    []
  );

  const handleBulkDeleteFiles = async (selected: FileItem[], clear: () => void) => {
    setBulkDeleting(true);
    try {
      await Promise.all(selected.map((f) => deleteFile(bucketId!, f.id)));
      toast.success(`已删除 ${selected.length} 个文件`);
      queryClient.invalidateQueries({ queryKey: ["files", bucketId] });
      clear();
    } finally {
      setBulkDeleting(false);
    }
  };

  if (bucketLoading) return <DetailSkeleton />;
  if (!bucket) return <NotFound backTo="/console/storage" />;

  return (
    <div className="space-y-6">
      <DetailPageWrapper
        title={bucket.name}
        description="Bucket 详情与文件管理"
        backTo="/console/storage"
        actions={
          <DeleteButton
            onConfirm={() => removeBucket.mutate(bucket.id)}
            loading={removeBucket.isPending}
          />
        }
      >
        <DetailGrid
          items={[
            { label: "ID", value: bucket.id, mono: true },
            { label: "名称", value: bucket.name },
            { label: "创建时间", value: new Date(bucket.created_at).toLocaleString() },
          ]}
        />
      </DetailPageWrapper>

      <ResourceListPage
        title=""
        cardTitle="文件列表"
        searchPlaceholder="搜索文件名..."
        isLoading={filesLoading}
        items={files}
        columns={fileColumns}
        getSearchText={getFileSearchText}
        detailPath={(f) => `/console/storage/${bucketId}/files/${f.id}`}
        toolbarActions={
          <div className="flex items-center gap-2">
            <UploadCloud className="h-4 w-4 text-muted-foreground" />
            <Input
              type="file"
              className="max-w-xs"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) {
                  uploadMutation.mutate(file);
                  e.target.value = "";
                }
              }}
            />
          </div>
        }
        selectionActions={(selected, clear) => (
          <BulkDeleteButton
            count={selected.length}
            loading={bulkDeleting}
            onConfirm={() => handleBulkDeleteFiles(selected, clear)}
          />
        )}
        rowActions={(f) => (
          <>
            <a href={downloadUrl(bucketId!, f.id)} target="_blank" rel="noreferrer">
              <Button variant="ghost" size="icon" title="下载">
                <Download className="h-4 w-4" />
              </Button>
            </a>
            <RowDeleteButton
              onConfirm={() => deleteMutation.mutate(f.id)}
              loading={deleteMutation.isPending}
            />
          </>
        )}
        emptyTitle="暂无文件"
        emptyDescription="上传文件到此 Bucket"
      />
    </div>
  );
}

export function FileDetailPage() {
  const { bucketId, fileId } = useParams<{ bucketId: string; fileId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: file, isLoading } = useQuery({
    queryKey: ["files", bucketId, fileId],
    queryFn: () => getFile(bucketId!, fileId!),
    enabled: !!bucketId && !!fileId,
  });

  const remove = useMutation({
    mutationFn: () => deleteFile(bucketId!, fileId!),
    onSuccess: () => {
      toast.success("文件已删除");
      queryClient.invalidateQueries({ queryKey: ["files", bucketId] });
      navigate(`/console/storage/${bucketId}`);
    },
  });

  if (isLoading) return <DetailSkeleton />;
  if (!file) return <NotFound backTo={`/console/storage/${bucketId}`} />;

  return (
    <DetailPageWrapper
      title={file.name}
      description="文件详情"
      backTo={`/console/storage/${bucketId}`}
      backLabel="返回 Bucket"
      actions={
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm">
            <a href={downloadUrl(bucketId!, file.id)} target="_blank" rel="noreferrer">
              <Download className="h-4 w-4 mr-2" />
              下载
            </a>
          </Button>
          <DeleteButton onConfirm={() => remove.mutate()} loading={remove.isPending} />
        </div>
      }
    >
      <DetailGrid
        items={[
          { label: "ID", value: file.id, mono: true },
          { label: "文件名", value: file.name },
          { label: "大小", value: formatBytes(file.size) },
          { label: "MIME 类型", value: file.mime_type },
          { label: "Bucket ID", value: file.bucket_id, mono: true },
          { label: "创建时间", value: new Date(file.created_at).toLocaleString() },
          { label: "更新时间", value: new Date(file.updated_at).toLocaleString() },
        ]}
      />
    </DetailPageWrapper>
  );
}
