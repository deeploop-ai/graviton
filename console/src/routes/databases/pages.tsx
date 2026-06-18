import { useCallback, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Plus } from "lucide-react";
import {
  listDatabases,
  getDatabase,
  createDatabase,
  deleteDatabase,
  listCollections,
  getCollection,
  createCollection,
  deleteCollection,
  type Database,
  type Collection,
} from "@/api/databases";
import { useAuth } from "@/hooks/useAuth";
import { ResourceListPage } from "@/components/list/ResourceListPage";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { ColumnDef } from "@/components/list/DataTable";
import {
  FormPageWrapper,
  FormField,
  DetailPageWrapper,
  DetailGrid,
  DetailSkeleton,
  NotFound,
  BulkDeleteButton,
  RowDeleteButton,
  DeleteButton,
} from "@/components/resource/shared";

const dbColumns: ColumnDef<Database>[] = [
  {
    key: "id",
    header: "ID",
    className: "font-mono text-xs max-w-[140px] truncate",
    cell: (d) => d.id,
  },
  { key: "name", header: "名称", cell: (d) => d.name },
  {
    key: "created",
    header: "创建时间",
    cell: (d) => new Date(d.created_at).toLocaleString(),
  },
];

export function DatabasesListPage() {
  const { projectId } = useAuth();
  const queryClient = useQueryClient();
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const { data: databases = [], isLoading } = useQuery({
    queryKey: ["databases", projectId],
    queryFn: listDatabases,
    enabled: !!projectId,
  });

  const remove = useMutation({
    mutationFn: deleteDatabase,
    onSuccess: () => {
      toast.success("Database 已删除");
      queryClient.invalidateQueries({ queryKey: ["databases"] });
    },
  });

  const getSearchText = useCallback((d: Database) => `${d.id} ${d.name}`, []);

  const handleBulkDelete = async (selected: Database[], clear: () => void) => {
    setBulkDeleting(true);
    try {
      await Promise.all(selected.map((d) => deleteDatabase(d.id)));
      toast.success(`已删除 ${selected.length} 个 Database`);
      queryClient.invalidateQueries({ queryKey: ["databases"] });
      clear();
    } finally {
      setBulkDeleting(false);
    }
  };

  return (
    <ResourceListPage
      title="Databases"
      description="管理数据库与集合"
      searchPlaceholder="搜索数据库名称或 ID..."
      isLoading={isLoading}
      items={databases}
      columns={dbColumns}
      getSearchText={getSearchText}
      detailPath={(d) => `/console/databases/${d.id}`}
      toolbarActions={
        <Button asChild>
          <Link to="/console/databases/new">
            <Plus className="h-4 w-4 mr-2" />
            新建 Database
          </Link>
        </Button>
      }
      selectionActions={(selected, clear) => (
        <BulkDeleteButton
          count={selected.length}
          loading={bulkDeleting}
          onConfirm={() => handleBulkDelete(selected, clear)}
        />
      )}
      rowActions={(d) => (
        <RowDeleteButton onConfirm={() => remove.mutate(d.id)} loading={remove.isPending} />
      )}
      emptyTitle="暂无 Database"
      emptyDescription="创建第一个 Database"
      emptyAction={
        <Button asChild>
          <Link to="/console/databases/new">新建 Database</Link>
        </Button>
      }
    />
  );
}

export function DatabaseNewPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [id, setId] = useState("");

  const mutation = useMutation({
    mutationFn: createDatabase,
    onSuccess: (db) => {
      toast.success("Database 创建成功");
      queryClient.invalidateQueries({ queryKey: ["databases"] });
      navigate(`/console/databases/${db.id}`);
    },
  });

  return (
    <FormPageWrapper
      title="新建 Database"
      backTo="/console/databases"
      submitLabel="创建"
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate({
          id: id || name.toLowerCase().replace(/\s+/g, "_"),
          name,
        });
      }}
      loading={mutation.isPending}
    >
      <FormField id="name" label="名称" value={name} onChange={setName} required placeholder="Production DB" />
      <FormField id="id" label="ID（可选）" value={id} onChange={setId} placeholder="production" />
    </FormPageWrapper>
  );
}

export function DatabaseDetailPage() {
  const { dbId } = useParams<{ dbId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [bulkDeleting, setBulkDeleting] = useState(false);

  const { data: database, isLoading: dbLoading } = useQuery({
    queryKey: ["databases", dbId],
    queryFn: () => getDatabase(dbId!),
    enabled: !!dbId,
  });

  const { data: collections = [], isLoading: collLoading } = useQuery({
    queryKey: ["collections", dbId],
    queryFn: () => listCollections(dbId!),
    enabled: !!dbId,
  });

  const removeDb = useMutation({
    mutationFn: deleteDatabase,
    onSuccess: () => {
      toast.success("Database 已删除");
      queryClient.invalidateQueries({ queryKey: ["databases"] });
      navigate("/console/databases");
    },
  });

  const removeColl = useMutation({
    mutationFn: (collId: string) => deleteCollection(dbId!, collId),
    onSuccess: () => {
      toast.success("Collection 已删除");
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
    },
  });

  const collColumns: ColumnDef<Collection>[] = [
    {
      key: "id",
      header: "ID",
      className: "font-mono text-xs max-w-[140px] truncate",
      cell: (c) => c.id,
    },
    { key: "name", header: "名称", cell: (c) => c.name },
    {
      key: "attributes",
      header: "Attributes",
      cell: (c) => <Badge variant="secondary">{c.attributes.length}</Badge>,
    },
    {
      key: "indexes",
      header: "Indexes",
      cell: (c) => <Badge variant="secondary">{c.indexes.length}</Badge>,
    },
  ];

  const getCollSearchText = useCallback(
    (c: Collection) => `${c.id} ${c.name}`,
    []
  );

  const handleBulkDeleteColl = async (selected: Collection[], clear: () => void) => {
    setBulkDeleting(true);
    try {
      await Promise.all(selected.map((c) => deleteCollection(dbId!, c.id)));
      toast.success(`已删除 ${selected.length} 个 Collection`);
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
      clear();
    } finally {
      setBulkDeleting(false);
    }
  };

  if (dbLoading) return <DetailSkeleton />;
  if (!database) return <NotFound backTo="/console/databases" />;

  return (
    <div className="space-y-6">
      <DetailPageWrapper
        title={database.name}
        description="Database 详情与 Collection 管理"
        backTo="/console/databases"
        actions={
          <DeleteButton
            onConfirm={() => removeDb.mutate(database.id)}
            loading={removeDb.isPending}
          />
        }
      >
        <DetailGrid
          items={[
            { label: "ID", value: database.id, mono: true },
            { label: "名称", value: database.name },
            { label: "创建时间", value: new Date(database.created_at).toLocaleString() },
          ]}
        />
      </DetailPageWrapper>

      <ResourceListPage
        title=""
        cardTitle="Collections"
        searchPlaceholder="搜索 Collection..."
        isLoading={collLoading}
        items={collections}
        columns={collColumns}
        getSearchText={getCollSearchText}
        detailPath={(c) => `/console/databases/${dbId}/collections/${c.id}`}
        toolbarActions={
          <Button asChild>
            <Link to={`/console/databases/${dbId}/collections/new`}>
              <Plus className="h-4 w-4 mr-2" />
              新建 Collection
            </Link>
          </Button>
        }
        selectionActions={(selected, clear) => (
          <BulkDeleteButton
            count={selected.length}
            loading={bulkDeleting}
            onConfirm={() => handleBulkDeleteColl(selected, clear)}
          />
        )}
        rowActions={(c) => (
          <RowDeleteButton
            onConfirm={() => removeColl.mutate(c.id)}
            loading={removeColl.isPending}
          />
        )}
        emptyTitle="暂无 Collection"
        emptyDescription="在此 Database 中创建 Collection"
        emptyAction={
          <Button asChild>
            <Link to={`/console/databases/${dbId}/collections/new`}>新建 Collection</Link>
          </Button>
        }
      />
    </div>
  );
}

export function CollectionNewPage() {
  const { dbId } = useParams<{ dbId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [id, setId] = useState("");

  const mutation = useMutation({
    mutationFn: () =>
      createCollection(dbId!, {
        id: id || name.toLowerCase().replace(/\s+/g, "_"),
        name,
      }),
    onSuccess: (coll) => {
      toast.success("Collection 创建成功");
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
      navigate(`/console/databases/${dbId}/collections/${coll.id}`);
    },
  });

  return (
    <FormPageWrapper
      title="新建 Collection"
      backTo={`/console/databases/${dbId}`}
      backLabel="返回 Database"
      submitLabel="创建"
      onSubmit={(e) => {
        e.preventDefault();
        mutation.mutate();
      }}
      loading={mutation.isPending}
    >
      <FormField id="name" label="名称" value={name} onChange={setName} required placeholder="posts" />
      <FormField id="id" label="ID（可选）" value={id} onChange={setId} placeholder="posts" />
    </FormPageWrapper>
  );
}

export function CollectionDetailPage() {
  const { dbId, collId } = useParams<{ dbId: string; collId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: collection, isLoading } = useQuery({
    queryKey: ["collections", dbId, collId],
    queryFn: () => getCollection(dbId!, collId!),
    enabled: !!dbId && !!collId,
  });

  const remove = useMutation({
    mutationFn: () => deleteCollection(dbId!, collId!),
    onSuccess: () => {
      toast.success("Collection 已删除");
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
      navigate(`/console/databases/${dbId}`);
    },
  });

  if (isLoading) return <DetailSkeleton />;
  if (!collection) return <NotFound backTo={`/console/databases/${dbId}`} />;

  return (
    <DetailPageWrapper
      title={collection.name}
      description="Collection 详情"
      backTo={`/console/databases/${dbId}`}
      backLabel="返回 Database"
      actions={
        <DeleteButton onConfirm={() => remove.mutate()} loading={remove.isPending} />
      }
    >
      <DetailGrid
        items={[
          { label: "ID", value: collection.id, mono: true },
          { label: "名称", value: collection.name },
          { label: "Database ID", value: collection.database_id, mono: true },
          { label: "Attributes", value: collection.attributes.length },
          { label: "Indexes", value: collection.indexes.length },
          { label: "创建时间", value: new Date(collection.created_at).toLocaleString() },
        ]}
      />

      {collection.attributes.length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-semibold mb-3">Attributes</h3>
          <div className="rounded-md border divide-y">
            {collection.attributes.map((attr) => (
              <div key={attr.id} className="px-4 py-3 flex items-center justify-between text-sm">
                <span className="font-mono">{attr.key}</span>
                <Badge variant="outline">{attr.type}{attr.array ? "[]" : ""}</Badge>
              </div>
            ))}
          </div>
        </div>
      )}

      {collection.indexes.length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-semibold mb-3">Indexes</h3>
          <div className="rounded-md border divide-y">
            {collection.indexes.map((idx) => (
              <div key={idx.id} className="px-4 py-3 flex items-center justify-between text-sm">
                <span className="font-mono">{idx.attributes.join(", ")}</span>
                <Badge variant="outline">{idx.type}</Badge>
              </div>
            ))}
          </div>
        </div>
      )}
    </DetailPageWrapper>
  );
}
