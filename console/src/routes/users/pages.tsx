import { useCallback, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { listUsers, getUser, updateUser, deleteUser, type User } from "@/api/users";
import { useAuth } from "@/hooks/useAuth";
import { ResourceListPage } from "@/components/list/ResourceListPage";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { ColumnDef } from "@/components/list/DataTable";
import {
  FormPageWrapper,
  DetailPageWrapper,
  DetailGrid,
  DetailSkeleton,
  NotFound,
  BulkDeleteButton,
  RowDeleteButton,
  DeleteButton,
} from "@/components/resource/shared";

const columns: ColumnDef<User>[] = [
  {
    key: "id",
    header: "ID",
    className: "font-mono text-xs max-w-[140px] truncate",
    cell: (u) => u.id,
  },
  { key: "email", header: "邮箱", cell: (u) => u.email },
  { key: "name", header: "名称", cell: (u) => u.name },
  {
    key: "status",
    header: "状态",
    cell: (u) => (
      <Badge variant={u.status === "active" ? "default" : "secondary"}>{u.status}</Badge>
    ),
  },
  { key: "verified", header: "已验证", cell: (u) => (u.email_verified ? "是" : "否") },
  {
    key: "created",
    header: "创建时间",
    cell: (u) => new Date(u.created_at).toLocaleString(),
  },
];

export function UsersListPage() {
  const { projectId } = useAuth();
  const queryClient = useQueryClient();
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const { data: users = [], isLoading } = useQuery({
    queryKey: ["users", projectId],
    queryFn: listUsers,
    enabled: !!projectId,
  });

  const remove = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      toast.success("用户已删除");
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });

  const getSearchText = useCallback(
    (u: User) => `${u.id} ${u.email} ${u.name} ${u.status}`,
    []
  );

  const handleBulkDelete = async (selected: User[], clear: () => void) => {
    setBulkDeleting(true);
    try {
      await Promise.all(selected.map((u) => deleteUser(u.id)));
      toast.success(`已删除 ${selected.length} 个用户`);
      queryClient.invalidateQueries({ queryKey: ["users"] });
      clear();
    } finally {
      setBulkDeleting(false);
    }
  };

  return (
    <ResourceListPage
      title="Users"
      description="当前项目的注册用户"
      searchPlaceholder="搜索邮箱、名称或 ID..."
      isLoading={isLoading}
      items={users}
      columns={columns}
      getSearchText={getSearchText}
      detailPath={(u) => `/console/users/${u.id}`}
      editPath={(u) => `/console/users/${u.id}/edit`}
      selectionActions={(selected, clear) => (
        <BulkDeleteButton
          count={selected.length}
          loading={bulkDeleting}
          onConfirm={() => handleBulkDelete(selected, clear)}
        />
      )}
      rowActions={(u) => (
        <RowDeleteButton onConfirm={() => remove.mutate(u.id)} loading={remove.isPending} />
      )}
      emptyTitle="暂无用户"
      emptyDescription="用户注册后将显示在此"
    />
  );
}

export function UserDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: user, isLoading } = useQuery({
    queryKey: ["users", id],
    queryFn: () => getUser(id!),
    enabled: !!id,
  });

  const remove = useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      toast.success("用户已删除");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      navigate("/console/users");
    },
  });

  if (isLoading) return <DetailSkeleton />;
  if (!user) return <NotFound backTo="/console/users" />;

  return (
    <DetailPageWrapper
      title={user.name || user.email}
      description="用户详情"
      backTo="/console/users"
      actions={
        <div className="flex gap-2">
          <Button asChild variant="outline" size="sm">
            <Link to={`/console/users/${user.id}/edit`}>编辑</Link>
          </Button>
          <DeleteButton onConfirm={() => remove.mutate(user.id)} loading={remove.isPending} />
        </div>
      }
    >
      <DetailGrid
        items={[
          { label: "ID", value: user.id, mono: true },
          { label: "邮箱", value: user.email },
          { label: "名称", value: user.name },
          { label: "状态", value: user.status },
          { label: "邮箱已验证", value: user.email_verified ? "是" : "否" },
          { label: "创建时间", value: new Date(user.created_at).toLocaleString() },
          { label: "更新时间", value: new Date(user.updated_at).toLocaleString() },
        ]}
      />
    </DetailPageWrapper>
  );
}

export function UserEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [status, setStatus] = useState<string>("");

  const { data: user, isLoading } = useQuery({
    queryKey: ["users", id],
    queryFn: () => getUser(id!),
    enabled: !!id,
  });

  const mutation = useMutation({
    mutationFn: (input: { status: string }) => updateUser(id!, input),
    onSuccess: () => {
      toast.success("用户已更新");
      queryClient.invalidateQueries({ queryKey: ["users"] });
      navigate(`/console/users/${id}`);
    },
  });

  if (isLoading) return <DetailSkeleton />;
  if (!user) return <NotFound backTo="/console/users" />;

  const currentStatus = status || user.status;

  return (
    <FormPageWrapper
      title="编辑用户"
      description={user.email}
      backTo={`/console/users/${id}`}
      backLabel="返回详情"
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate({ status: currentStatus });
      }}
      loading={mutation.isPending}
    >
      <div className="space-y-2">
        <Label htmlFor="status">状态</Label>
        <Select value={currentStatus} onValueChange={setStatus}>
          <SelectTrigger id="status">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="active">active</SelectItem>
            <SelectItem value="inactive">inactive</SelectItem>
            <SelectItem value="blocked">blocked</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </FormPageWrapper>
  );
}
