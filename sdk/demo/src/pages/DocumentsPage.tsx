import { useState } from "react";
import { Link } from "react-router-dom";
import { useFleet } from "@/lib/fleet-context";
import { suffix } from "@/lib/storage";
import { ErrorBanner, JsonPanel, MethodTag, PageHeader } from "@/components/Ui";

export function DocumentsPage() {
  const { client, settings, updateSettings, serverFleet, run, lastError } = useFleet();
  const [result, setResult] = useState<unknown>(null);
  const [loading, setLoading] = useState(false);
  const [title, setTitle] = useState("Hello from SDK Playground");
  const [views, setViews] = useState(1);

  const dbId = settings.demoDbId;
  const collId = settings.demoCollId || "posts";

  async function exec(label: string, fn: () => Promise<unknown>) {
    setLoading(true);
    try {
      const data = await run(fn);
      setResult({ action: label, data });
    } catch {
      /* banner */
    } finally {
      setLoading(false);
    }
  }

  async function bootstrapDemoEnv() {
    setLoading(true);
    try {
      const fleet = serverFleet();
      const newDbId = `sdk_web_${suffix()}`;
      const newCollId = "posts";
      const data = await run(async () => {
        await fleet.server.databases.createDatabase({
          id: newDbId,
          name: `SDK Web DB ${suffix()}`,
        });
        await fleet.server.databases.createCollection(newDbId, {
          id: newCollId,
          name: "Posts",
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
        const doc = await fleet.server.databases.createDocument(newDbId, newCollId, {
          data: { title: "Seed from Server SDK", views: 0 },
        });
        return { databaseId: newDbId, collectionId: newCollId, seedDocument: doc };
      });
      updateSettings({ ...settings, demoDbId: newDbId, demoCollId: newCollId });
      setResult({ action: "bootstrapDemoEnv()", data });
    } catch {
      /* banner */
    } finally {
      setLoading(false);
    }
  }

  return (
    <div>
      <PageHeader
        title="Documents API"
        description="使用 Client SDK 创建与列出文档。首次使用请先用 Server API Key 初始化演示库/集合。"
      />
      <ErrorBanner message={lastError} />

      {!settings.apiKey ? (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-100">
          需要 Server API Key 才能初始化演示环境。请先到{" "}
          <Link className="text-fleet-accent underline" to="/app/settings">
            设置
          </Link>{" "}
          填写。
        </div>
      ) : null}

      <div className="mb-4 grid gap-3 rounded-xl border border-fleet-border bg-fleet-panel/50 p-4 text-sm md:grid-cols-2">
        <div>
          <div className="text-fleet-muted">databaseId</div>
          <div className="font-mono text-cyan-100">{dbId || "（未初始化）"}</div>
        </div>
        <div>
          <div className="text-fleet-muted">collectionId</div>
          <div className="font-mono text-cyan-100">{collId}</div>
        </div>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        <button
          type="button"
          className="btn-primary"
          disabled={loading || !settings.apiKey}
          onClick={bootstrapDemoEnv}
        >
          初始化演示库（Server SDK）
        </button>
        <button
          type="button"
          className="btn-secondary"
          disabled={loading || !dbId}
          onClick={() =>
            exec("databases.createDocument()", () =>
              client.databases.createDocument(dbId, collId, {
                data: { title, views: Number(views) },
              })
            )
          }
        >
          <MethodTag method="POST" /> createDocument()
        </button>
        <button
          type="button"
          className="btn-secondary"
          disabled={loading || !dbId}
          onClick={() =>
            exec("databases.listDocuments()", () =>
              client.databases.listDocuments(dbId, collId, { page_size: 20 })
            )
          }
        >
          <MethodTag method="GET" /> listDocuments()
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

      <JsonPanel title="SDK 响应" data={result} />
    </div>
  );
}
