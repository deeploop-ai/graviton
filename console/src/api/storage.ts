import { api } from "./client";

export interface Bucket {
  id: string;
  name: string;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

export interface FileItem {
  id: string;
  bucket_id: string;
  name: string;
  mime_type: string;
  size: number;
  created_at: string;
  updated_at: string;
}

export async function listBuckets(): Promise<Bucket[]> {
  const res = await api.get<{ buckets: Bucket[] }>("/server/storage/buckets");
  return res.data.buckets ?? [];
}

export async function getBucket(id: string): Promise<Bucket> {
  const res = await api.get<Bucket>(`/server/storage/buckets/${id}`);
  return res.data;
}

export async function createBucket(input: {
  name: string;
}): Promise<Bucket> {
  const res = await api.post<Bucket>("/server/storage/buckets", input);
  return res.data;
}

export async function deleteBucket(id: string): Promise<void> {
  await api.delete(`/server/storage/buckets/${id}`);
}

export async function listFiles(bucketId: string): Promise<FileItem[]> {
  const res = await api.get<{ files: FileItem[] }>(
    `/server/storage/buckets/${bucketId}/files`
  );
  return res.data.files ?? [];
}

export async function getFile(bucketId: string, fileId: string): Promise<FileItem> {
  const res = await api.get<FileItem>(
    `/server/storage/buckets/${bucketId}/files/${fileId}`
  );
  return res.data;
}

export async function uploadFile(
  bucketId: string,
  file: File
): Promise<FileItem> {
  const form = new FormData();
  form.append("file", file);
  const res = await api.post<FileItem>(
    `/storage/buckets/${bucketId}/files`,
    form,
    {
      headers: { "Content-Type": "multipart/form-data" },
    }
  );
  return res.data;
}

export function downloadUrl(bucketId: string, fileId: string): string {
  return `/v1/storage/buckets/${bucketId}/files/${fileId}/download`;
}

export async function deleteFile(bucketId: string, fileId: string): Promise<void> {
  await api.delete(`/server/storage/buckets/${bucketId}/files/${fileId}`);
}
