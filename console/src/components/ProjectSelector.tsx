import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/api/projects";
import { useAuth } from "@/hooks/useAuth";
import { Skeleton } from "@/components/ui/skeleton";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export function ProjectSelector() {
  const { projectId, selectProject } = useAuth();

  const { data: projects = [], isLoading } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  const selectedProjectId = projectId || "";

  if (isLoading) {
    return <Skeleton className="h-9 w-full" />;
  }

  if (projects.length === 0) {
    return (
      <p className="text-xs text-muted-foreground px-1">暂无项目，请先在 Projects 中创建</p>
    );
  }

  return (
    <Select value={selectedProjectId} onValueChange={selectProject}>
      <SelectTrigger className="w-full">
        <SelectValue placeholder="选择项目" />
      </SelectTrigger>
      <SelectContent>
        {projects.map((p) => (
          <SelectItem key={p.id} value={p.id}>
            {p.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
