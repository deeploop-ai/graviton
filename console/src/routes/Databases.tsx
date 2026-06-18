import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  listDatabases,
  createDatabase,
  listCollections,
  createCollection,
  type Database,
  type Collection,
} from "@/api/databases";
import { useAuth } from "@/hooks/useAuth";
import { PageHeader } from "@/components/PageHeader";
import { LoadingTable } from "@/components/LoadingTable";
import { EmptyState } from "@/components/EmptyState";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export function Databases() {
  const { projectId } = useAuth();
  const queryClient = useQueryClient();
  const [dbName, setDbName] = useState("");
  const [dbId, setDbId] = useState("");
  const [selectedDb, setSelectedDb] = useState<string | null>(null);
  const [collName, setCollName] = useState("");
  const [collId, setCollId] = useState("");

  const { data: databases = [], isLoading: dbLoading } = useQuery({
    queryKey: ["databases", projectId],
    queryFn: listDatabases,
    enabled: !!projectId,
  });

  const { data: collections = [], isLoading: collLoading } = useQuery({
    queryKey: ["collections", selectedDb],
    queryFn: () => listCollections(selectedDb!),
    enabled: !!selectedDb,
  });

  const createDb = useMutation({
    mutationFn: createDatabase,
    onSuccess: () => {
      toast.success("Database created");
      queryClient.invalidateQueries({ queryKey: ["databases"] });
      setDbName("");
      setDbId("");
    },
  });

  const createColl = useMutation({
    mutationFn: () => createCollection(selectedDb!, { id: collId, name: collName }),
    onSuccess: () => {
      toast.success("Collection created");
      queryClient.invalidateQueries({ queryKey: ["collections", selectedDb] });
      setCollName("");
      setCollId("");
    },
  });

  const handleCreateDb = (e: React.FormEvent) => {
    e.preventDefault();
    createDb.mutate({
      id: dbId || dbName.toLowerCase().replace(/\s+/g, "_"),
      name: dbName,
    });
  };

  const handleCreateColl = (e: React.FormEvent) => {
    e.preventDefault();
    createColl.mutate();
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Databases" description="Databases and collections" />

      <Card>
        <CardHeader>
          <CardTitle>Create database</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreateDb} className="flex flex-col gap-4 md:flex-row md:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="dbName">Name</Label>
              <Input
                id="dbName"
                value={dbName}
                onChange={(e) => setDbName(e.target.value)}
                placeholder="Production DB"
                required
              />
            </div>
            <div className="flex-1 space-y-2">
              <Label htmlFor="dbId">ID (optional)</Label>
              <Input
                id="dbId"
                value={dbId}
                onChange={(e) => setDbId(e.target.value)}
                placeholder="production"
              />
            </div>
            <Button type="submit" disabled={createDb.isPending}>
              {createDb.isPending ? "Creating..." : "Create"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Databases</CardTitle>
          <CardDescription>Select a database to view collections</CardDescription>
        </CardHeader>
        <CardContent>
          {dbLoading ? (
            <div className="flex gap-2">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="h-10 w-28 rounded-md bg-muted animate-pulse" />
              ))}
            </div>
          ) : databases.length === 0 ? (
            <EmptyState title="No databases" description="Create a database first." />
          ) : (
            <div className="flex flex-wrap gap-2">
              {databases.map((db: Database) => (
                <Button
                  key={db.id}
                  variant={selectedDb === db.id ? "default" : "outline"}
                  onClick={() => setSelectedDb(db.id)}
                >
                  {db.name}
                </Button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {selectedDb && (
        <Card>
          <CardHeader>
            <CardTitle>Collections</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <form onSubmit={handleCreateColl} className="flex flex-col gap-4 md:flex-row md:items-end">
              <div className="flex-1 space-y-2">
                <Label htmlFor="collName">Collection name</Label>
                <Input
                  id="collName"
                  value={collName}
                  onChange={(e) => setCollName(e.target.value)}
                  placeholder="posts"
                  required
                />
              </div>
              <div className="flex-1 space-y-2">
                <Label htmlFor="collId">ID (optional)</Label>
                <Input
                  id="collId"
                  value={collId}
                  onChange={(e) => setCollId(e.target.value)}
                  placeholder="posts"
                />
              </div>
              <Button type="submit" disabled={createColl.isPending}>
                {createColl.isPending ? "Creating..." : "Create"}
              </Button>
            </form>

            {collLoading ? (
              <LoadingTable columns={4} />
            ) : collections.length === 0 ? (
              <EmptyState title="No collections" description="Create a collection in this database." />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Attributes</TableHead>
                    <TableHead>Indexes</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {collections.map((c: Collection) => (
                    <TableRow key={c.id}>
                      <TableCell className="font-mono text-xs max-w-[120px] truncate">{c.id}</TableCell>
                      <TableCell className="font-medium">{c.name}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">{c.attributes.length}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary">{c.indexes.length}</Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
