import { listQuery, type HttpTransport } from "../http.js";
import type {
  Attribute,
  Collection,
  Database,
  Document,
  Index,
  ListParams,
} from "../types.js";

export class ServerDatabasesService {
  constructor(private readonly http: HttpTransport) {}

  async createDatabase(input: { id: string; name: string }): Promise<Database> {
    return this.http.request<Database>("POST", "/v1/server/databases", {
      auth: "apiKey",
      body: input,
    });
  }

  async listDatabases(params?: ListParams): Promise<Database[]> {
    const res = await this.http.request<{ databases: Database[] }>("GET", "/v1/server/databases", {
      auth: "apiKey",
      query: listQuery(params),
    });
    return res.databases ?? [];
  }

  async getDatabase(id: string): Promise<Database> {
    return this.http.request<Database>("GET", `/v1/server/databases/${id}`, { auth: "apiKey" });
  }

  async deleteDatabase(id: string): Promise<void> {
    await this.http.request<void>("DELETE", `/v1/server/databases/${id}`, { auth: "apiKey" });
  }

  async createCollection(
    databaseId: string,
    input: {
      id: string;
      name: string;
      permissions?: string[];
      attributes?: Array<{
        key: string;
        type: string;
        size?: number;
        required?: boolean;
        array?: boolean;
      }>;
    }
  ): Promise<Collection> {
    return this.http.request<Collection>(
      "POST",
      `/v1/server/databases/${databaseId}/collections`,
      { auth: "apiKey", body: { database_id: databaseId, ...input } }
    );
  }

  async listCollections(databaseId: string, params?: ListParams): Promise<Collection[]> {
    const res = await this.http.request<{ collections: Collection[] }>(
      "GET",
      `/v1/server/databases/${databaseId}/collections`,
      { auth: "apiKey", query: listQuery(params) }
    );
    return res.collections ?? [];
  }

  async getCollection(databaseId: string, collectionId: string): Promise<Collection> {
    return this.http.request<Collection>(
      "GET",
      `/v1/server/databases/${databaseId}/collections/${collectionId}`,
      { auth: "apiKey" }
    );
  }

  async deleteCollection(databaseId: string, collectionId: string): Promise<void> {
    await this.http.request<void>(
      "DELETE",
      `/v1/server/databases/${databaseId}/collections/${collectionId}`,
      { auth: "apiKey" }
    );
  }

  async createAttribute(
    databaseId: string,
    collectionId: string,
    input: { key: string; type: string; size?: number; required?: boolean; array?: boolean }
  ): Promise<Attribute> {
    return this.http.request<Attribute>(
      "POST",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/attributes`,
      {
        auth: "apiKey",
        body: { database_id: databaseId, collection_id: collectionId, ...input },
      }
    );
  }

  async createIndex(
    databaseId: string,
    collectionId: string,
    input: { type: string; attributes: string[]; orders?: string[] }
  ): Promise<Index> {
    return this.http.request<Index>(
      "POST",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/indexes`,
      {
        auth: "apiKey",
        body: { database_id: databaseId, collection_id: collectionId, ...input },
      }
    );
  }

  async createDocument(
    databaseId: string,
    collectionId: string,
    input: {
      document_id?: string;
      data: Record<string, unknown>;
      permissions?: string[];
    }
  ): Promise<Document> {
    return this.http.request<Document>(
      "POST",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents`,
      {
        auth: "apiKey",
        body: {
          database_id: databaseId,
          collection_id: collectionId,
          document_id: input.document_id ?? "",
          data: input.data,
          permissions: input.permissions,
        },
      }
    );
  }

  async listDocuments(
    databaseId: string,
    collectionId: string,
    params?: ListParams
  ): Promise<Document[]> {
    const res = await this.http.request<{ documents: Document[] }>(
      "GET",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents`,
      { auth: "apiKey", query: listQuery(params) }
    );
    return res.documents ?? [];
  }

  async getDocument(
    databaseId: string,
    collectionId: string,
    documentId: string
  ): Promise<Document> {
    return this.http.request<Document>(
      "GET",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents/${documentId}`,
      { auth: "apiKey" }
    );
  }

  async updateDocument(
    databaseId: string,
    collectionId: string,
    documentId: string,
    data: Record<string, unknown>
  ): Promise<Document> {
    return this.http.request<Document>(
      "PATCH",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents/${documentId}`,
      {
        auth: "apiKey",
        body: {
          database_id: databaseId,
          collection_id: collectionId,
          document_id: documentId,
          data,
        },
      }
    );
  }

  async deleteDocument(
    databaseId: string,
    collectionId: string,
    documentId: string
  ): Promise<void> {
    await this.http.request<void>(
      "DELETE",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents/${documentId}`,
      { auth: "apiKey" }
    );
  }

  async countDocuments(
    databaseId: string,
    collectionId: string,
    params?: Pick<ListParams, "queries">
  ): Promise<number> {
    const res = await this.http.request<{ count: number }>(
      "GET",
      `/v1/server/databases/${databaseId}/collections/${collectionId}/documents/count`,
      { auth: "apiKey", query: listQuery(params) }
    );
    return res.count ?? 0;
  }
}
