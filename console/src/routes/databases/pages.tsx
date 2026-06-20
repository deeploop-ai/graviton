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
  createAttribute,
  createIndex,
  type Database,
  type Collection,
  type Attribute,
} from "@/api/databases";
import { useAuth } from "@/hooks/useAuth";
import { ResourceListPage } from "@/components/list/ResourceListPage";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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

const ATTRIBUTE_TYPES = [
  { value: "string", label: "String" },
  { value: "integer", label: "Integer" },
  { value: "float", label: "Float" },
  { value: "boolean", label: "Boolean" },
  { value: "datetime", label: "Datetime" },
  { value: "email", label: "Email" },
  { value: "url", label: "URL" },
  { value: "json", label: "JSON" },
] as const;

const INDEX_TYPES = [
  { value: "key", label: "Key" },
  { value: "unique", label: "Unique" },
  { value: "fulltext", label: "Fulltext" },
] as const;

const STRING_LIKE_TYPES = new Set(["string", "email", "url"]);

function AttributeList({
  attributes,
  onAdd,
}: {
  attributes: Attribute[];
  onAdd: () => void;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-lg">Attributes</CardTitle>
        <Button size="sm" onClick={onAdd}>
          <Plus className="h-4 w-4 mr-2" />
          添加 Attribute
        </Button>
      </CardHeader>
      <CardContent>
        {attributes.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center">
            暂无字段定义，点击上方按钮添加第一个 Attribute
          </p>
        ) : (
          <div className="rounded-md border divide-y">
            {attributes.map((attr) => (
              <div key={attr.id} className="px-4 py-3 flex items-center justify-between text-sm gap-4">
                <div className="min-w-0">
                  <span className="font-mono">{attr.key}</span>
                  <div className="flex flex-wrap gap-1.5 mt-1">
                    {attr.required && <Badge variant="secondary">required</Badge>}
                    {attr.array && <Badge variant="secondary">array</Badge>}
                    {attr.size ? <Badge variant="secondary">size {attr.size}</Badge> : null}
                  </div>
                </div>
                <Badge variant="outline">
                  {attr.type}
                  {attr.array ? "[]" : ""}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function IndexList({
  indexes,
  onAdd,
  canAdd,
}: {
  indexes: Collection["indexes"];
  onAdd: () => void;
  canAdd: boolean;
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-lg">Indexes</CardTitle>
        <Button size="sm" onClick={onAdd} disabled={!canAdd}>
          <Plus className="h-4 w-4 mr-2" />
          添加 Index
        </Button>
      </CardHeader>
      <CardContent>
        {!canAdd && (
          <p className="text-sm text-muted-foreground mb-4">
            请先添加至少一个 Attribute，再创建 Index。
          </p>
        )}
        {indexes.length === 0 ? (
          <p className="text-sm text-muted-foreground py-4 text-center">
            {canAdd ? "暂无索引，点击上方按钮添加 Index" : "暂无索引"}
          </p>
        ) : (
          <div className="rounded-md border divide-y">
            {indexes.map((idx) => (
              <div key={idx.id} className="px-4 py-3 flex items-center justify-between text-sm gap-4">
                <div className="min-w-0">
                  <span className="font-mono text-xs text-muted-foreground">{idx.id}</span>
                  <p className="font-mono mt-1">{idx.attributes.join(", ")}</p>
                </div>
                <Badge variant="outline">{idx.type}</Badge>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function AddAttributeDialog({
  open,
  onOpenChange,
  loading,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  loading: boolean;
  onSubmit: (input: {
    key: string;
    type: string;
    size?: number;
    required: boolean;
    array: boolean;
  }) => void;
}) {
  const [key, setKey] = useState("");
  const [type, setType] = useState("string");
  const [size, setSize] = useState("");
  const [required, setRequired] = useState(false);
  const [array, setArray] = useState(false);

  const reset = () => {
    setKey("");
    setType("string");
    setSize("");
    setRequired(false);
    setArray(false);
  };

  const handleOpenChange = (next: boolean) => {
    if (!next) reset();
    onOpenChange(next);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>添加 Attribute</DialogTitle>
          <DialogDescription>为 Collection 定义字段类型与约束。</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit({
              key: key.trim(),
              type,
              size: size ? Number(size) : undefined,
              required,
              array,
            });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="attr-key">Key</Label>
            <Input
              id="attr-key"
              value={key}
              onChange={(e) => setKey(e.target.value)}
              placeholder="title"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Type</Label>
            <Select value={type} onValueChange={setType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ATTRIBUTE_TYPES.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {STRING_LIKE_TYPES.has(type) && (
            <div className="space-y-2">
              <Label htmlFor="attr-size">Size（可选）</Label>
              <Input
                id="attr-size"
                type="number"
                min={1}
                value={size}
                onChange={(e) => setSize(e.target.value)}
                placeholder="256"
              />
            </div>
          )}
          <div className="flex flex-wrap gap-6">
            <label className="flex items-center gap-2 text-sm">
              <Checkbox checked={required} onChange={(e) => setRequired(e.target.checked)} />
              Required
            </label>
            <label className="flex items-center gap-2 text-sm">
              <Checkbox checked={array} onChange={(e) => setArray(e.target.checked)} />
              Array
            </label>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={loading || !key.trim()}>
              {loading ? "添加中..." : "添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function AddIndexDialog({
  open,
  onOpenChange,
  loading,
  attributes,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  loading: boolean;
  attributes: Attribute[];
  onSubmit: (input: { id: string; type: string; attributes: string[] }) => void;
}) {
  const [id, setId] = useState("");
  const [type, setType] = useState("key");
  const [selected, setSelected] = useState<string[]>([]);

  const reset = () => {
    setId("");
    setType("key");
    setSelected([]);
  };

  const handleOpenChange = (next: boolean) => {
    if (!next) reset();
    onOpenChange(next);
  };

  const toggleAttribute = (key: string) => {
    setSelected((prev) =>
      prev.includes(key) ? prev.filter((item) => item !== key) : [...prev, key]
    );
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>添加 Index</DialogTitle>
          <DialogDescription>选择索引类型，并指定参与索引的 Attribute。</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit({ id: id.trim(), type, attributes: selected });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="index-id">ID</Label>
            <Input
              id="index-id"
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder="title_key"
              required
            />
          </div>
          <div className="space-y-2">
            <Label>Type</Label>
            <Select value={type} onValueChange={setType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {INDEX_TYPES.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    {item.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Attributes</Label>
            <div className="rounded-md border divide-y max-h-48 overflow-y-auto">
              {attributes.map((attr) => (
                <label
                  key={attr.key}
                  className="flex items-center gap-2 px-3 py-2 text-sm cursor-pointer hover:bg-muted/50"
                >
                  <Checkbox
                    checked={selected.includes(attr.key)}
                    onChange={() => toggleAttribute(attr.key)}
                  />
                  <span className="font-mono">{attr.key}</span>
                  <Badge variant="outline" className="ml-auto">
                    {attr.type}
                  </Badge>
                </label>
              ))}
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              取消
            </Button>
            <Button type="submit" disabled={loading || !id.trim() || selected.length === 0}>
              {loading ? "添加中..." : "添加"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

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
  const [attrDialogOpen, setAttrDialogOpen] = useState(false);
  const [indexDialogOpen, setIndexDialogOpen] = useState(false);

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

  const addAttribute = useMutation({
    mutationFn: (input: {
      key: string;
      type: string;
      size?: number;
      required: boolean;
      array: boolean;
    }) => createAttribute(dbId!, collId!, input),
    onSuccess: () => {
      toast.success("Attribute 已添加");
      queryClient.invalidateQueries({ queryKey: ["collections", dbId, collId] });
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
      setAttrDialogOpen(false);
    },
  });

  const addIndex = useMutation({
    mutationFn: (input: { id: string; type: string; attributes: string[] }) =>
      createIndex(dbId!, collId!, {
        ...input,
        orders: input.attributes.map(() => "asc"),
      }),
    onSuccess: () => {
      toast.success("Index 已添加");
      queryClient.invalidateQueries({ queryKey: ["collections", dbId, collId] });
      queryClient.invalidateQueries({ queryKey: ["collections", dbId] });
      setIndexDialogOpen(false);
    },
  });

  if (isLoading) return <DetailSkeleton />;
  if (!collection) return <NotFound backTo={`/console/databases/${dbId}`} />;

  return (
    <>
      <DetailPageWrapper
        title={collection.name}
        description="Collection 详情与 Schema 管理"
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

        <div className="mt-6 space-y-6">
          <AttributeList
            attributes={collection.attributes}
            onAdd={() => setAttrDialogOpen(true)}
          />
          <IndexList
            indexes={collection.indexes}
            canAdd={collection.attributes.length > 0}
            onAdd={() => setIndexDialogOpen(true)}
          />
        </div>
      </DetailPageWrapper>

      <AddAttributeDialog
        open={attrDialogOpen}
        onOpenChange={setAttrDialogOpen}
        loading={addAttribute.isPending}
        onSubmit={(input) => addAttribute.mutate(input)}
      />
      <AddIndexDialog
        open={indexDialogOpen}
        onOpenChange={setIndexDialogOpen}
        loading={addIndex.isPending}
        attributes={collection.attributes}
        onSubmit={(input) => addIndex.mutate(input)}
      />
    </>
  );
}
