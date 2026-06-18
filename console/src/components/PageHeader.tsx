import { Link, useLocation } from "react-router-dom";
import { ChevronRight, Home } from "lucide-react";

const routeNames: Record<string, string> = {
  "": "Dashboard",
  projects: "Projects",
  "api-keys": "API Keys",
  users: "Users",
  storage: "Storage",
  databases: "Databases",
};

export function PageHeader({ title, description }: { title: string; description?: string }) {
  const location = useLocation();
  const segments = location.pathname.replace("/console", "").split("/").filter(Boolean);

  return (
    <div className="mb-8 space-y-2">
      <nav className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link to="/console" className="flex items-center gap-1 hover:text-foreground">
          <Home className="h-3.5 w-3.5" />
          <span className="sr-only">Dashboard</span>
        </Link>
        {segments.map((segment, idx) => {
          const path = "/console/" + segments.slice(0, idx + 1).join("/");
          const isLast = idx === segments.length - 1;
          return (
            <div key={path} className="flex items-center gap-2">
              <ChevronRight className="h-3.5 w-3.5" />
              {isLast ? (
                <span className="text-foreground font-medium">{routeNames[segment] || segment}</span>
              ) : (
                <Link to={path} className="hover:text-foreground">
                  {routeNames[segment] || segment}
                </Link>
              )}
            </div>
          );
        })}
      </nav>
      <div>
        <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
        {description && <p className="text-muted-foreground">{description}</p>}
      </div>
    </div>
  );
}
