import { Link } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/PageHeader";

interface FormPageProps {
  title: string;
  description?: string;
  backTo: string;
  backLabel?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export function FormPage({
  title,
  description,
  backTo,
  backLabel = "返回列表",
  actions,
  children,
}: FormPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader title={title} description={description} actions={actions} />
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link to={backTo}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            {backLabel}
          </Link>
        </Button>
      </div>
      {children}
    </div>
  );
}

interface DetailPageProps {
  title: string;
  description?: string;
  backTo: string;
  backLabel?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export function DetailPage({
  title,
  description,
  backTo,
  backLabel = "返回列表",
  actions,
  children,
}: DetailPageProps) {
  return (
    <div className="space-y-6">
      <PageHeader title={title} description={description} actions={actions} />
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link to={backTo}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            {backLabel}
          </Link>
        </Button>
      </div>
      {children}
    </div>
  );
}
