import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  listBuckets,
  createBucket,
  listFiles,
  uploadFile,
  deleteFile,
  downloadUrl,
  type Bucket,
  type FileItem,
} from "@/api/storage";
import { useAuth } from "@/hooks/useAuth";
import { PageHeader } from "@/components/PageHeader";
import { LoadingTable } from "@/components/LoadingTable";
import { EmptyState } from "@/components/EmptyState";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Download, Trash2, UploadCloud } from "lucide-react";

export function Storage() {
  const { projectId } = useAuth();
  const queryClient = useQueryClient();
  const [bucketName, setBucketName] = useState("");
  const [selectedBucket, setSelectedBucket] = useState<string | null>(null);

  const { data: buckets = [], isLoading: bucketsLoading } = useQuery({
    queryKey: ["buckets", projectId],
    queryFn: listBuckets,
    enabled: !!projectId,
  });

  const { data: files = [], isLoading: filesLoading } = useQuery({
    queryKey: ["files", selectedBucket],
    queryFn: () => listFiles(selectedBucket!),
    enabled: !!selectedBucket,
  });

  const createBucketMutation = useMutation({
    mutationFn: createBucket,
    onSuccess: () => {
      toast.success("Bucket created");
      queryClient.invalidateQueries({ queryKey: ["buckets"] });
      setBucketName("");
    },
  });

  const uploadMutation = useMutation({
    mutationFn: ({ bucketId, file }: { bucketId: string; file: File }) =>
      uploadFile(bucketId, file),
    onSuccess: () => {
      toast.success("File uploaded");
      queryClient.invalidateQueries({ queryKey: ["files", selectedBucket] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: ({ bucketId, fileId }: { bucketId: string; fileId: string }) =>
      deleteFile(bucketId, fileId),
    onSuccess: () => {
      toast.success("File deleted");
      queryClient.invalidateQueries({ queryKey: ["files", selectedBucket] });
    },
  });

  const handleCreateBucket = (e: React.FormEvent) => {
    e.preventDefault();
    createBucketMutation.mutate({ name: bucketName });
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file && selectedBucket) {
      uploadMutation.mutate({ bucketId: selectedBucket, file });
      e.target.value = "";
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Storage" description="Buckets and files" />

      <Card>
        <CardHeader>
          <CardTitle>Create bucket</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreateBucket} className="flex flex-col gap-4 md:flex-row md:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="bucketName">Bucket name</Label>
              <Input
                id="bucketName"
                value={bucketName}
                onChange={(e) => setBucketName(e.target.value)}
                placeholder="uploads"
                required
              />
            </div>
            <Button type="submit" disabled={createBucketMutation.isPending}>
              {createBucketMutation.isPending ? "Creating..." : "Create"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Buckets</CardTitle>
          <CardDescription>Click a bucket to manage files</CardDescription>
        </CardHeader>
        <CardContent>
          {bucketsLoading ? (
            <div className="flex gap-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="h-10 w-28 rounded-md bg-muted animate-pulse" />
              ))}
            </div>
          ) : buckets.length === 0 ? (
            <EmptyState title="No buckets" description="Create a bucket to upload files." />
          ) : (
            <div className="flex flex-wrap gap-2">
              {buckets.map((b: Bucket) => (
                <Button
                  key={b.id}
                  variant={selectedBucket === b.id ? "default" : "outline"}
                  onClick={() => setSelectedBucket(b.id)}
                >
                  {b.name}
                </Button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {selectedBucket && (
        <Card>
          <CardHeader>
            <CardTitle>Files</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center gap-4">
              <UploadCloud className="h-5 w-5 text-muted-foreground" />
              <Input type="file" onChange={handleFileChange} className="max-w-sm" />
              {uploadMutation.isPending && <span className="text-sm text-muted-foreground">Uploading...</span>}
            </div>

            {filesLoading ? (
              <LoadingTable columns={5} />
            ) : files.length === 0 ? (
              <EmptyState title="No files" description="Upload a file to this bucket." />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Size</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead className="w-32"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {files.map((f: FileItem) => (
                    <TableRow key={f.id}>
                      <TableCell className="font-medium">{f.name}</TableCell>
                      <TableCell>{formatBytes(f.size)}</TableCell>
                      <TableCell>
                        <Badge variant="outline">{f.mime_type}</Badge>
                      </TableCell>
                      <TableCell>{new Date(f.created_at).toLocaleString()}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <a
                            href={downloadUrl(selectedBucket, f.id)}
                            target="_blank"
                            rel="noreferrer"
                          >
                            <Button variant="ghost" size="icon">
                              <Download className="h-4 w-4" />
                            </Button>
                          </a>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() =>
                              deleteMutation.mutate({ bucketId: selectedBucket, fileId: f.id })
                            }
                            disabled={deleteMutation.isPending}
                          >
                            <Trash2 className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}
