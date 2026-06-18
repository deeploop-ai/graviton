import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/api/projects";
import { useAuth } from "@/hooks/useAuth";

/** Ensures admin console requests have X-Fleet-Project before project-scoped APIs run. */
export function ProjectBootstrap() {
  const { projectId, selectProject } = useAuth();

  const { data: projects = [] } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
    enabled: !projectId,
  });

  useEffect(() => {
    if (!projectId && projects[0]?.id) {
      selectProject(projects[0].id);
    }
  }, [projectId, projects, selectProject]);

  return null;
}
