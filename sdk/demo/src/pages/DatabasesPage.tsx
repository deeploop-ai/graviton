import { useState } from "react";
import { Link } from "react-router-dom";
import { useFleet } from "@/lib/fleet-context";
import { suffix } from "@/lib/storage";
import { ErrorBanner, JsonPanel, MethodTag, PageHeader } from "@/components/Ui";

const DEFAULT_COLL_PERMS = [
  "read:any",
  "create:users",
  "update:users",
  "delete:users",
] as const;

type StepResult = { step: string; ok: boolean; data?: unknown; error?: string };

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-6 rounded-xl border border-fleet-border bg-fleet-panel/40 p-4">
      <h3 className="mb-3 text-sm font-medium text-slate-200">{title}</h3>
      <div className="flex flex-wrap gap-2">{children}</div>
    </section>
  );
}

function ActionButton({
  label,
  method,
  disabled,
  onClick,
}: {
  label: string;
  method: string;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <button type="button" className="btn-secondary text-xs" disabled={disabled} onClick={onClick}>
      <MethodTag method={method} /> {label}
    </button>
  );
}

export function DatabasesPage() {
  const { client, settings, updateSettings, serverFleet, run, lastError } = useFleet();
  const [result, setResult] = useState<unknown>(null);
  const [loading, setLoading] = useState(false);
  const [title, setTitle] = useState("SDK demo document");
  const [views, setViews] = useState(1);

  const dbId = settings.demoDbId;
  const collId = settings.demoCollId || "posts";
  const docId = settings.demoDocId;
  const indexId = settings.demoIndexId;

  const hasEnv = Boolean(dbId && collId);
  const hasApiKey = Boolean(settings.apiKey);

  async function exec(label: string, fn: () => Promise<unknown>) {
    setLoading(true);
    try {
      const data = await run(fn);
      setResult({ action: label, data });
      return data;
    } catch {
      return null;
    } finally {
      setLoading(false);
    }
  }

  async function bootstrapEnv() {
    setLoading(true);
    try {
      const fleet = serverFleet();
      const newDbId = `sdk_db_${suffix()}`;
      const newCollId = "posts";
      const data = await run(async () => {
        await fleet.server.databases.createDatabase({
          id: newDbId,
          name: `SDK DB ${suffix()}`,
        });
        await fleet.server.databases.createCollection(newDbId, {
          id: newCollId,
          name: "Posts",
          permissions: [...DEFAULT_COLL_PERMS],
        });
        await fleet.server.databases.createAttribute(newDbId, newCollId, {
          key: "title",
          type: "string",
          size: 256,
        });
        await fleet.server.databases.createAttribute(newDbId, newCollId, {
          key: "views",
          type: "integer",
        });
        const index = await fleet.server.databases.createIndex(newDbId, newCollId, {
          id: "idx_title",
          type: "key",
          attributes: ["title"],
        });
        const doc = await fleet.server.databases.createDocument(newDbId, newCollId, {
          data: { title: "Seed document", views: 0 },
        });
        return {
          databaseId: newDbId,
          collectionId: newCollId,
          indexId: index.id,
          seedDocument: doc,
        };
      });
      const payload = data as {
        databaseId: string;
        collectionId: string;
        indexId: string;
        seedDocument: { id: string };
      };
      updateSettings({
        ...settings,
        demoDbId: payload.databaseId,
        demoCollId: payload.collectionId,
        demoDocId: payload.seedDocument.id,
        demoIndexId: payload.indexId,
      });
      setResult({ action: "bootstrapEnv()", data });
    } catch {
      /* banner */
    } finally {
      setLoading(false);
    }
  }

  async function runFullVerification() {
    setLoading(true);
    const steps: StepResult[] = [];
    const fleet = serverFleet();

    async function step(name: string, fn: () => Promise<unknown>) {
      try {
        const data = await fn();
        steps.push({ step: name, ok: true, data });
        return data;
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        steps.push({ step: name, ok: false, error: message });
        throw err;
      }
    }

    const testDbId = `sdk_verify_${suffix()}`;
    const testCollId = "items";
    const testIndexId = "idx_views";
    let serverDocId = "";
    let clientDocId = "";
    let extraDocId = "";

    try {
      await step("server.createDatabase", () =>
        fleet.server.databases.createDatabase({ id: testDbId, name: "Verify DB" })
      );
      await step("server.getDatabase", () => fleet.server.databases.getDatabase(testDbId));
      await step("server.listDatabases", () => fleet.server.databases.listDatabases({ page_size: 5 }));

      await step("server.createCollection", () =>
        fleet.server.databases.createCollection(testDbId, {
          id: testCollId,
          name: "Items",
          permissions: [...DEFAULT_COLL_PERMS],
        })
      );
      await step("server.listCollections", () =>
        fleet.server.databases.listCollections(testDbId)
      );
      await step("server.getCollection", () =>
        fleet.server.databases.getCollection(testDbId, testCollId)
      );
      await step("server.updateCollection", () =>
        fleet.server.databases.updateCollection(testDbId, testCollId, {
          name: "Items Updated",
        })
      );

      await step("server.createAttribute(title)", () =>
        fleet.server.databases.createAttribute(testDbId, testCollId, {
          key: "title",
          type: "string",
          size: 128,
        })
      );
      await step("server.createAttribute(views)", () =>
        fleet.server.databases.createAttribute(testDbId, testCollId, {
          key: "views",
          type: "integer",
        })
      );
      const index = (await step("server.createIndex", () =>
        fleet.server.databases.createIndex(testDbId, testCollId, {
          id: testIndexId,
          type: "key",
          attributes: ["views"],
          orders: ["ASC"],
        })
      )) as { id: string };

      const serverDoc = (await step("server.createDocument", () =>
        fleet.server.databases.createDocument(testDbId, testCollId, {
          data: { title: "Server doc", views: 1 },
        })
      )) as { id: string };
      serverDocId = serverDoc.id;

      await step("server.listDocuments", () =>
        fleet.server.databases.listDocuments(testDbId, testCollId)
      );
      await step("server.getDocument", () =>
        fleet.server.databases.getDocument(testDbId, testCollId, serverDocId)
      );
      await step("server.updateDocument", () =>
        fleet.server.databases.updateDocument(testDbId, testCollId, serverDocId, {
          data: { title: "Server doc updated" },
          increment: { views: 2 },
        })
      );
      await step("server.countDocuments", () =>
        fleet.server.databases.countDocuments(testDbId, testCollId)
      );

      const clientDoc = (await step("client.createDocument", () =>
        client.databases.createDocument(testDbId, testCollId, {
          data: { title: "Client doc", views: 5 },
        })
      )) as { id: string };
      clientDocId = clientDoc.id;

      const extraDoc = (await step("client.createDocument (bulk target)", () =>
        client.databases.createDocument(testDbId, testCollId, {
          data: { title: "Bulk target", views: 0 },
        })
      )) as { id: string };
      extraDocId = extraDoc.id;

      await step("client.listDocuments", () =>
        client.databases.listDocuments(testDbId, testCollId, { page_size: 10 })
      );
      await step("client.getDocument", () =>
        client.databases.getDocument(testDbId, testCollId, clientDocId)
      );
      await step("client.updateDocument", () =>
        client.databases.updateDocument(testDbId, testCollId, clientDocId, {
          data: { title: "Client doc updated" },
          increment: { views: 3 },
        })
      );
      await step("client.countDocuments", () =>
        client.databases.countDocuments(testDbId, testCollId)
      );

      await step("server.bulkUpdateDocuments", () =>
        fleet.server.databases.bulkUpdateDocuments(testDbId, testCollId, {
          document_ids: [extraDocId],
          data: { title: "Bulk updated" },
        })
      );
      await step("server.bulkDeleteDocuments", () =>
        fleet.server.databases.bulkDeleteDocuments(testDbId, testCollId, [extraDocId])
      );

      await step("client.deleteDocument", () =>
        client.databases.deleteDocument(testDbId, testCollId, clientDocId)
      );
      await step("server.deleteDocument", () =>
        fleet.server.databases.deleteDocument(testDbId, testCollId, serverDocId)
      );

      await step("server.deleteAttribute(views)", () =>
        fleet.server.databases.deleteAttribute(testDbId, testCollId, "views")
      );
      await step("server.deleteIndex", () =>
        fleet.server.databases.deleteIndex(testDbId, testCollId, index.id)
      );
      await step("server.deleteCollection", () =>
        fleet.server.databases.deleteCollection(testDbId, testCollId)
      );
      await step("server.deleteDatabase", () =>
        fleet.server.databases.deleteDatabase(testDbId)
      );

      setResult({
        action: "runFullVerification()",
        passed: steps.filter((s) => s.ok).length,
        total: steps.length,
        steps,
      });
    } catch {
      setResult({
        action: "runFullVerification() — failed",
        passed: steps.filter((s) => s.ok).length,
        total: steps.length,
        steps,
      });
    } finally {
      setLoading(false);
    }
  }

  const disabled = loading || !hasEnv;
  const serverDisabled = loading || !hasApiKey;

  return (
    <div>
      <PageHeader
        title="Databases API"
        description="Server API（库/集合/属性/索引/文档/Bulk）与 Client API（文档 CRUD）全功能验证。"
        actions={
          <button
            type="button"
            className="btn-primary"
            disabled={serverDisabled}
            onClick={runFullVerification}
          >
            运行全量验证
          </button>
        }
      />
      <ErrorBanner message={lastError} />

      {!hasApiKey ? (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
          Server API 需要 API Key。请先到{" "}
          <Link className="text-fleet-accent underline" to="/app/settings">
            设置
          </Link>{" "}
          填写。
        </div>
      ) : null}

      <div className="mb-4 grid gap-3 rounded-xl border border-fleet-border bg-fleet-panel/50 p-4 text-sm md:grid-cols-2 lg:grid-cols-4">
        <div>
          <div className="text-fleet-muted">databaseId</div>
          <div className="truncate font-mono text-cyan-100">{dbId || "（未初始化）"}</div>
        </div>
        <div>
          <div className="text-fleet-muted">collectionId</div>
          <div className="font-mono text-cyan-100">{collId}</div>
        </div>
        <div>
          <div className="text-fleet-muted">documentId</div>
          <div className="truncate font-mono text-cyan-100">{docId || "—"}</div>
        </div>
        <div>
          <div className="text-fleet-muted">indexId</div>
          <div className="font-mono text-cyan-100">{indexId || "—"}</div>
        </div>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        <button
          type="button"
          className="btn-primary"
          disabled={serverDisabled}
          onClick={bootstrapEnv}
        >
          初始化演示环境
        </button>
      </div>

      <div className="mb-4 grid gap-3 md:grid-cols-2">
        <label className="block space-y-1">
          <span className="text-xs text-fleet-muted">data.title</span>
          <input className="field" value={title} onChange={(e) => setTitle(e.target.value)} />
        </label>
        <label className="block space-y-1">
          <span className="text-xs text-fleet-muted">data.views</span>
          <input
            className="field"
            type="number"
            value={views}
            onChange={(e) => setViews(Number(e.target.value))}
          />
        </label>
      </div>

      <Section title="Server — Database">
        <ActionButton
          label="createDatabase()"
          method="POST"
          disabled={serverDisabled}
          onClick={() => {
            const id = `sdk_db_${suffix()}`;
            exec("server.databases.createDatabase()", async () => {
              const db = await serverFleet().server.databases.createDatabase({
                id,
                name: `DB ${suffix()}`,
              });
              updateSettings({ ...settings, demoDbId: db.id });
              return db;
            });
          }}
        />
        <ActionButton
          label="listDatabases()"
          method="GET"
          disabled={serverDisabled}
          onClick={() => exec("server.databases.listDatabases()", () => serverFleet().server.databases.listDatabases())}
        />
        <ActionButton
          label="getDatabase()"
          method="GET"
          disabled={serverDisabled || !dbId}
          onClick={() => exec("server.databases.getDatabase()", () => serverFleet().server.databases.getDatabase(dbId))}
        />
        <ActionButton
          label="deleteDatabase()"
          method="DELETE"
          disabled={serverDisabled || !dbId}
          onClick={() =>
            exec("server.databases.deleteDatabase()", async () => {
              await serverFleet().server.databases.deleteDatabase(dbId);
              updateSettings({ ...settings, demoDbId: "", demoCollId: "posts", demoDocId: "", demoIndexId: "" });
            })
          }
        />
      </Section>

      <Section title="Server — Collection / Attribute / Index">
        <ActionButton
          label="createCollection()"
          method="POST"
          disabled={serverDisabled || !dbId}
          onClick={() =>
            exec("server.databases.createCollection()", async () => {
              const id = `coll_${suffix()}`;
              const coll = await serverFleet().server.databases.createCollection(dbId, {
                id,
                name: "Collection",
                permissions: [...DEFAULT_COLL_PERMS],
              });
              updateSettings({ ...settings, demoCollId: coll.id });
              return coll;
            })
          }
        />
        <ActionButton
          label="listCollections()"
          method="GET"
          disabled={serverDisabled || !dbId}
          onClick={() => exec("server.databases.listCollections()", () => serverFleet().server.databases.listCollections(dbId))}
        />
        <ActionButton
          label="getCollection()"
          method="GET"
          disabled={disabled}
          onClick={() => exec("server.databases.getCollection()", () => serverFleet().server.databases.getCollection(dbId, collId))}
        />
        <ActionButton
          label="updateCollection()"
          method="PATCH"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.updateCollection()", () =>
              serverFleet().server.databases.updateCollection(dbId, collId, { name: `Updated ${suffix()}` })
            )
          }
        />
        <ActionButton
          label="deleteCollection()"
          method="DELETE"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.deleteCollection()", async () => {
              await serverFleet().server.databases.deleteCollection(dbId, collId);
              updateSettings({ ...settings, demoCollId: "posts", demoDocId: "", demoIndexId: "" });
            })
          }
        />
        <ActionButton
          label="createAttribute()"
          method="POST"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.createAttribute()", () =>
              serverFleet().server.databases.createAttribute(dbId, collId, {
                key: `field_${suffix()}`,
                type: "string",
                size: 64,
              })
            )
          }
        />
        <ActionButton
          label="deleteAttribute()"
          method="DELETE"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.deleteAttribute()", () =>
              serverFleet().server.databases.deleteAttribute(dbId, collId, "views")
            )
          }
        />
        <ActionButton
          label="createIndex()"
          method="POST"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.createIndex()", async () => {
              const idx = await serverFleet().server.databases.createIndex(dbId, collId, {
                id: `idx_${suffix()}`,
                type: "key",
                attributes: ["title"],
              });
              updateSettings({ ...settings, demoIndexId: idx.id });
              return idx;
            })
          }
        />
        <ActionButton
          label="deleteIndex()"
          method="DELETE"
          disabled={disabled || !indexId}
          onClick={() =>
            exec("server.databases.deleteIndex()", async () => {
              await serverFleet().server.databases.deleteIndex(dbId, collId, indexId);
              updateSettings({ ...settings, demoIndexId: "" });
            })
          }
        />
      </Section>

      <Section title="Server — Documents">
        <ActionButton
          label="createDocument()"
          method="POST"
          disabled={disabled}
          onClick={() =>
            exec("server.databases.createDocument()", async () => {
              const doc = await serverFleet().server.databases.createDocument(dbId, collId, {
                data: { title, views: Number(views) },
              });
              updateSettings({ ...settings, demoDocId: doc.id });
              return doc;
            })
          }
        />
        <ActionButton
          label="listDocuments()"
          method="GET"
          disabled={disabled}
          onClick={() => exec("server.databases.listDocuments()", () => serverFleet().server.databases.listDocuments(dbId, collId))}
        />
        <ActionButton
          label="getDocument()"
          method="GET"
          disabled={disabled || !docId}
          onClick={() => exec("server.databases.getDocument()", () => serverFleet().server.databases.getDocument(dbId, collId, docId))}
        />
        <ActionButton
          label="updateDocument()"
          method="PATCH"
          disabled={disabled || !docId}
          onClick={() =>
            exec("server.databases.updateDocument()", () =>
              serverFleet().server.databases.updateDocument(dbId, collId, docId, {
                data: { title: `${title} (server)` },
                increment: { views: 1 },
              })
            )
          }
        />
        <ActionButton
          label="deleteDocument()"
          method="DELETE"
          disabled={disabled || !docId}
          onClick={() =>
            exec("server.databases.deleteDocument()", async () => {
              await serverFleet().server.databases.deleteDocument(dbId, collId, docId);
              updateSettings({ ...settings, demoDocId: "" });
            })
          }
        />
        <ActionButton
          label="countDocuments()"
          method="GET"
          disabled={disabled}
          onClick={() => exec("server.databases.countDocuments()", () => serverFleet().server.databases.countDocuments(dbId, collId))}
        />
        <ActionButton
          label="bulkUpdateDocuments()"
          method="PATCH"
          disabled={disabled || !docId}
          onClick={() =>
            exec("server.databases.bulkUpdateDocuments()", () =>
              serverFleet().server.databases.bulkUpdateDocuments(dbId, collId, {
                document_ids: [docId],
                data: { title: "Bulk updated" },
              })
            )
          }
        />
        <ActionButton
          label="bulkDeleteDocuments()"
          method="POST"
          disabled={disabled || !docId}
          onClick={() =>
            exec("server.databases.bulkDeleteDocuments()", async () => {
              const res = await serverFleet().server.databases.bulkDeleteDocuments(dbId, collId, [docId]);
              updateSettings({ ...settings, demoDocId: "" });
              return res;
            })
          }
        />
      </Section>

      <Section title="Client — Documents">
        <ActionButton
          label="createDocument()"
          method="POST"
          disabled={disabled}
          onClick={() =>
            exec("client.databases.createDocument()", async () => {
              const doc = await client.databases.createDocument(dbId, collId, {
                data: { title, views: Number(views) },
              });
              updateSettings({ ...settings, demoDocId: doc.id });
              return doc;
            })
          }
        />
        <ActionButton
          label="listDocuments()"
          method="GET"
          disabled={disabled}
          onClick={() => exec("client.databases.listDocuments()", () => client.databases.listDocuments(dbId, collId, { page_size: 20 }))}
        />
        <ActionButton
          label="getDocument()"
          method="GET"
          disabled={disabled || !docId}
          onClick={() => exec("client.databases.getDocument()", () => client.databases.getDocument(dbId, collId, docId))}
        />
        <ActionButton
          label="updateDocument()"
          method="PATCH"
          disabled={disabled || !docId}
          onClick={() =>
            exec("client.databases.updateDocument()", () =>
              client.databases.updateDocument(dbId, collId, docId, {
                data: { title: `${title} (client)` },
                increment: { views: 1 },
              })
            )
          }
        />
        <ActionButton
          label="deleteDocument()"
          method="DELETE"
          disabled={disabled || !docId}
          onClick={() =>
            exec("client.databases.deleteDocument()", async () => {
              await client.databases.deleteDocument(dbId, collId, docId);
              updateSettings({ ...settings, demoDocId: "" });
            })
          }
        />
        <ActionButton
          label="countDocuments()"
          method="GET"
          disabled={disabled}
          onClick={() => exec("client.databases.countDocuments()", () => client.databases.countDocuments(dbId, collId))}
        />
      </Section>

      <JsonPanel title="SDK 响应" data={result} />
    </div>
  );
}
