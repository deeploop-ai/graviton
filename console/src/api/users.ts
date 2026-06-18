import { api } from "./client";

export interface User {
  id: string;
  email: string;
  name: string;
  status: string;
  email_verified: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListUsersResponse {
  users: User[];
  meta?: { total_count?: number; page_size?: number };
}

export async function listUsers(): Promise<User[]> {
  const res = await api.get<ListUsersResponse>("/server/users");
  return res.data.users ?? [];
}

export async function getUser(id: string): Promise<User> {
  const res = await api.get<User>(`/server/users/${id}`);
  return res.data;
}

export async function updateUser(
  id: string,
  input: { status?: string }
): Promise<User> {
  const res = await api.patch<User>(`/server/users/${id}`, input);
  return res.data;
}

export async function deleteUser(id: string): Promise<void> {
  await api.delete(`/server/users/${id}`);
}
