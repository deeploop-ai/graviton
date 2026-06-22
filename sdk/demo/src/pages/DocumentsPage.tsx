import { useState } from "react";
import { Link } from "react-router-dom";
import { useFleet } from "@/lib/fleet-context";
import { suffix } from "@/lib/storage";
import { ErrorBanner, JsonPanel, MethodTag, PageHeader } from "@/components/Ui";

const PERM_TYPES = ["read", "create", "update", "delete"] as const;

export function DocumentsPage() {
  const { client, settings, updateSettings, serverFleet, run, lastError } = useFleet();
  const [result, setResult] = useState<unknown>(null);
  const [loading, setLoading] = useState(false);
  const [title, setTitle] = useState("Hello from SDK Playground");
  const [views, setViews] = useState(1);
  const [collPermissions, setCollPermissions] = useState<string[]>([
    "read:any",
    "create:users",
    "update:users",
    "delete:users",
  ]);
  const [docPermissions, setDocPermissions] = useState<string[]>([]);
  const [permType, setPermType] = useState<string>("read");
  const [permRole, setPermRole] = useState<string>("");
  const [permTarget, setPermTarget] = useState<"collection" | "document">("collection");

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

  function addPerm() {
    const r = permRole.trim();
    if (!r) return;
    const entry = `${permType}:${r}`;
    if (permTarget === "collection") {
      if (!collPermissions.includes(entry)) setCollPermissions([...collPermissions, entry]);
    } else {
      if (!docPermissions.includes(entry)) setDocPermissions([...docPermissions, entry]);
    }
    setPermRole("");
  }

  function removePerm(entry: string) {
    if (permTarget === "collection") {
      setCollPermissions(collPermissions.filter((p) => p !== entry));
    } else {
      setDocPermissions(docPermissions.filter((p) => p !== entry));
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
          permissions: collPermissions,
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

  const activePerms = permTarget === "collection" ? collPermissions : docPermissions;

  return (
    <div>
      <PageHeader
        title="Documents API"
        description="使用 Client SDK 创建与列出文档。支持 Collection 与 Document 级权限配置演示。"
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

      {/* Permission configuration panel */}
      <div className="mb-4 rounded-xl border border-fleet-border bg-fleet-panel/50 p-4">
        <div className="mb-3 flex items-center gap-2">
          <h3 className="text-sm font-medium text-slate-200">权限配置</h3>
          <div className="flex gap-1">
            <button
              type="button"
              className={`rounded px-2 py-0.5 text-xs ${permTarget === "collection" ? "bg-fleet-accent text-black" : "bg-fleet-panel text-fleet-muted"}`}
              onClick={() => setPermTarget("collection")}
            >
              Collection
            </button>
            <button
              type="button"
              className={`rounded px-2 py-0.5 text-xs ${permTarget === "document" ? "bg-fleet-accent text-black" : "bg-fleet-panel text-fleet-muted"}`}
              onClick={() => setPermTarget("document")}
            >
              Document
            </button>
          </div>
        </div>

        <p className="mb-3 text-xs text-fleet-muted">
          格式 <code className="text-cyan-200">操作:角色</code>，例如 <code className="text-cyan-200">read:any</code>、<code className="text-cyan-200">update:user:&lt;id&gt;</code>。Collection 权限在初始化时生效；Document 权限在 createDocument 时生效。
        </p>

        {activePerms.length > 0 && (
          <div className="mb-3 flex flex-wrap gap-1.5">
            {activePerms.map((p) => (
              <span key={p} className="inline-flex items-center gap-1 rounded bg-fleet-panel px-2 py-0.5 font-mono text-[11px] text-cyan-200">
                {p}
                <button type="button" onClick={() => removePerm(p)} className="text-red-300 hover:text-red-200">×</button>
              </span>
            ))}
          </div>
        )}

        <div className="flex items-end gap-2">
          <label className="space-y-1">
            <span className="text-xs text-fleet-muted">操作</span>
            <select
              value={permType}
              onChange={(e) => setPermType(e.target.value)}
              className="field h-9"
            >
              {PERM_TYPES.map((t) => (
                <option key={t} value={t}>{t}</option>
              ))}
            </select>
          </label>
          <label className="flex-1 space-y-1">
            <span className="text-xs text-fleet-muted">角色</span>
            <input
              className="field"
              value={permRole}
              onChange={(e) => setPermRole(e.target.value)}
              onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); addPerm(); } }}
              placeholder="any / users / keys / admin / user:<id>"
            />
          </label>
          <button type="button" className="btn-secondary h-9" onClick={addPerm}>添加</button>
        </div>

        <div className="mt-2 flex flex-wrap gap-1.5">
          {["any", "users", "keys", "admin"].map((preset) => (
            <button
              key={preset}
              type="button"
              className="rounded border border-fleet-border px-2 py-0.5 text-[11px] text-fleet-muted hover:text-cyan-200"
              onClick={() => {
                const entry = `${permType}:${preset}`;
                if (permTarget === "collection") {
                  if (!collPermissions.includes(entry)) setCollPermissions([...collPermissions, entry]);
                } else {
                  if (!docPermissions.includes(entry)) setDocPermissions([...docPermissions, entry]);
                }
              }}
            >
              + {permType}:{preset}
            </button>
          ))}
        </div>
      </div>

      <div className="mb-4 flex flex-wrap gap-2">
        <button
          type="button"
          className="btn-primary"
          disabled={loading || !settings.apiKey}
          onClick={bootstrapDemoEnv}
        >
          初始化演示库（带 Collection 权限）
        </button>
        <button
          type="button"
          className="btn-secondary"
          disabled={loading || !dbId}
          onClick={() =>
            exec("databases.createDocument() with permissions", () =>
              client.databases.createDocument(dbId, collId, {
                data: { title, views: Number(views) },
                permissions: docPermissions.length > 0 ? docPermissions : undefined,
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
