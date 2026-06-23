import { Link } from "react-router-dom";
import { useFleet } from "@/lib/fleet-context";
import { PageHeader } from "@/components/Ui";

const cards = [
  {
    title: "Account API",
    path: "/app/account",
    desc: "me / prefs / sessions / refresh",
    code: "client.account.*",
  },
  {
    title: "Databases API",
    path: "/app/databases",
    desc: "Server + Client Databases 全功能验证",
    code: "databases.*",
  },
  {
    title: "Teams API",
    path: "/app/teams",
    desc: "建队、邀请成员、列表查询",
    code: "client.teams.*",
  },
  {
    title: "Server API",
    path: "/app/server",
    desc: "Health / Projects / Users / Databases",
    code: "fleet.server.*",
  },
];

export function HomePage() {
  const { auth, settings } = useFleet();

  return (
    <div>
      <PageHeader
        title={`你好，${auth?.name || auth?.email}`}
        description="这是 Fleet TypeScript SDK 的交互式演示站点。每个页面都会直接调用 SDK 并展示 JSON 响应。"
      />

      <div className="mb-6 grid gap-3 rounded-xl border border-fleet-border bg-fleet-panel/60 p-4 text-sm md:grid-cols-2">
        <div>
          <div className="text-fleet-muted">Endpoint</div>
          <div className="font-mono text-cyan-100">{settings.endpoint}</div>
        </div>
        <div>
          <div className="text-fleet-muted">Project</div>
          <div className="font-mono text-cyan-100">{settings.projectId}</div>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {cards.map((card) => (
          <Link
            key={card.path}
            to={card.path}
            className="panel block p-5 transition hover:border-fleet-accent"
          >
            <div className="font-mono text-xs text-fleet-accent">{card.code}</div>
            <h3 className="mt-2 text-lg font-semibold text-white">{card.title}</h3>
            <p className="mt-1 text-sm text-fleet-muted">{card.desc}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
