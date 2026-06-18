import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/api/projects";
import { listAPIKeys } from "@/api/apiKeys";
import { listUsers } from "@/api/users";
import { listBuckets } from "@/api/storage";
import { listDatabases } from "@/api/databases";
import { useAuth } from "@/hooks/useAuth";
import { PageHeader } from "@/components/PageHeader";
import { Skeleton } from "@/components/ui/skeleton";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export function Dashboard() {
  const { projectId, selectProject } = useAuth();

  const { data: projects = [], isLoading: projectsLoading } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  const selectedProjectId = projectId || "";

  const { data: apiKeys = [] } = useQuery({
    queryKey: ["api-keys", selectedProjectId],
    queryFn: listAPIKeys,
    enabled: !!selectedProjectId,
  });

  const { data: users = [] } = useQuery({
    queryKey: ["users", selectedProjectId],
    queryFn: listUsers,
    enabled: !!selectedProjectId,
  });

  const { data: buckets = [] } = useQuery({
    queryKey: ["buckets", selectedProjectId],
    queryFn: listBuckets,
    enabled: !!selectedProjectId,
  });

  const { data: databases = [] } = useQuery({
    queryKey: ["databases", selectedProjectId],
    queryFn: listDatabases,
    enabled: !!selectedProjectId,
  });

  const handleSelectProject = (id: string) => {
    selectProject(id);
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Dashboard" description="Overview of your Fleet workspace" />

      <Card>
        <CardHeader>
          <CardTitle>Active project</CardTitle>
          <CardDescription>
            Server-side resources are scoped to the selected project
          </CardDescription>
        </CardHeader>
        <CardContent>
          {projectsLoading ? (
            <Skeleton className="h-10 w-64" />
          ) : (
            <Select value={selectedProjectId} onValueChange={handleSelectProject}>
              <SelectTrigger className="w-64">
                <SelectValue placeholder="Select project" />
              </SelectTrigger>
              <SelectContent>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Projects" value={projects.length} isLoading={projectsLoading} />
        <StatCard title="API Keys" value={apiKeys.length} />
        <StatCard title="Users" value={users.length} />
        <StatCard title="Buckets" value={buckets.length} />
        <StatCard title="Databases" value={databases.length} />
      </div>
    </div>
  );
}

function StatCard({ title, value, isLoading }: { title: string; value: number; isLoading?: boolean }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardDescription>{title}</CardDescription>
        {isLoading ? (
          <Skeleton className="h-9 w-16" />
        ) : (
          <CardTitle className="text-4xl">{value}</CardTitle>
        )}
      </CardHeader>
    </Card>
  );
}
