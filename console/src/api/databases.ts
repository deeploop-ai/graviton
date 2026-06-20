import { api } from "./client";

export interface Database {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface Attribute {
  id: string;
  key: string;
  type: string;
  size?: number;
  required: boolean;
  array: boolean;
}

export interface Index {
  id: string;
  type: string;
  attributes: string[];
  orders: string[];
}

export interface Collection {
  id: string;
  database_id: string;
  name: string;
  permissions: string[];
  attributes: Attribute[];
  indexes: Index[];
  created_at: string;
  updated_at: string;
}

function normalizeIndex(index: Index): Index {
  return {
    ...index,
    attributes: index.attributes ?? [],
    orders: index.orders ?? [],
  };
}

function normalizeCollection(collection: Collection): Collection {
  return {
    ...collection,
    permissions: collection.permissions ?? [],
    attributes: collection.attributes ?? [],
    indexes: (collection.indexes ?? []).map(normalizeIndex),
  };
}

export async function listDatabases(): Promise<Database[]> {
  const res = await api.get<{ databases: Database[] }>("/server/databases");
  return res.data.databases ?? [];
}

export async function getDatabase(id: string): Promise<Database> {
  const res = await api.get<Database>(`/server/databases/${id}`);
  return res.data;
}

export async function createDatabase(input: {
  id: string;
  name: string;
}): Promise<Database> {
  const res = await api.post<Database>("/server/databases", input);
  return res.data;
}

export async function deleteDatabase(id: string): Promise<void> {
  await api.delete(`/server/databases/${id}`);
}

export async function listCollections(databaseId: string): Promise<Collection[]> {
  const res = await api.get<{ collections: Collection[] }>(
    `/server/databases/${databaseId}/collections`
  );
  return (res.data.collections ?? []).map(normalizeCollection);
}

export async function getCollection(
  databaseId: string,
  collectionId: string
): Promise<Collection> {
  const res = await api.get<Collection>(
    `/server/databases/${databaseId}/collections/${collectionId}`
  );
  return normalizeCollection(res.data);
}

export async function createCollection(
  databaseId: string,
  input: { id: string; name: string }
): Promise<Collection> {
  const res = await api.post<Collection>(
    `/server/databases/${databaseId}/collections`,
    input
  );
  return normalizeCollection(res.data);
}

export async function deleteCollection(
  databaseId: string,
  collectionId: string
): Promise<void> {
  await api.delete(
    `/server/databases/${databaseId}/collections/${collectionId}`
  );
}

export async function createAttribute(
  databaseId: string,
  collectionId: string,
  input: {
    key: string;
    type: string;
    size?: number;
    required?: boolean;
    array?: boolean;
  }
): Promise<Attribute> {
  const res = await api.post<Attribute>(
    `/server/databases/${databaseId}/collections/${collectionId}/attributes`,
    input
  );
  return res.data;
}

export async function createIndex(
  databaseId: string,
  collectionId: string,
  input: {
    id: string;
    type: string;
    attributes: string[];
    orders?: string[];
  }
): Promise<Index> {
  const res = await api.post<Index>(
    `/server/databases/${databaseId}/collections/${collectionId}/indexes`,
    input
  );
  return res.data;
}
