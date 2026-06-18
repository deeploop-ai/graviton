import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { listAPIKeys, createAPIKey, deleteAPIKey, type APIKey } from "@/api/apiKeys";
import { PageHeader } from "@/components/PageHeader";
import { LoadingTable } from "@/components/LoadingTable";
import { EmptyState } from "@/components/EmptyState";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Copy, Trash2 } from "lucide-react";

export function ApiKeys() {
  const queryClient = useQueryClient();
  const [name, setName] = useState("");
  const [scopes, setScopes] = useState("");
  const [createdSecret, setCreatedSecret] = useState<string | null>(null);

  const { data: keys = [], isLoading } = useQuery({
    queryKey: ["api-keys"],
    queryFn: listAPIKeys,
  });

  const create = useMutation({
    mutationFn: createAPIKey,
    onSuccess: (data) => {
      setCreatedSecret(data.secret);
      toast.success("API key created. Copy the secret now — it won't be shown again.");
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
      setName("");
      setScopes("");
    },
  });

  const remove = useMutation({
    mutationFn: deleteAPIKey,
    onSuccess: () => {
      toast.success("API key deleted");
      queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    create.mutate({
      name,
      scopes: scopes.split(",").map((s) => s.trim()).filter(Boolean),
    });
  };

  const copySecret = () => {
    if (createdSecret) {
      navigator.clipboard.writeText(createdSecret);
      toast.success("Secret copied to clipboard");
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="API Keys" description="Manage server API keys for the selected project" />

      <Card>
        <CardHeader>
          <CardTitle>Create API key</CardTitle>
          <CardDescription>The secret is shown only once after creation</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <form onSubmit={handleSubmit} className="flex flex-col gap-4 md:flex-row md:items-end">
            <div className="flex-1 space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Production API Key"
                required
              />
            </div>
            <div className="flex-[2] space-y-2">
              <Label htmlFor="scopes">Scopes (comma separated)</Label>
              <Input
                id="scopes"
                value={scopes}
                onChange={(e) => setScopes(e.target.value)}
                placeholder="users.read, users.write"
              />
            </div>
            <Button type="submit" disabled={create.isPending}>
              {create.isPending ? "Creating..." : "Create"}
            </Button>
          </form>
          {createdSecret && (
            <div className="rounded-md bg-muted p-3 flex items-center justify-between gap-4">
              <code className="break-all text-xs flex-1">{createdSecret}</code>
              <Button variant="secondary" size="sm" onClick={copySecret}>
                <Copy className="h-4 w-4 mr-1" />
                Copy
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>API keys</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <LoadingTable columns={5} />
          ) : keys.length === 0 ? (
            <EmptyState title="No API keys" description="Create an API key to access the server API." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Scopes</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="w-24"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {keys.map((k: APIKey) => (
                  <TableRow key={k.id}>
                    <TableCell className="font-mono text-xs max-w-[120px] truncate">{k.id}</TableCell>
                    <TableCell className="font-medium">{k.name}</TableCell>
                    <TableCell>{k.scopes.join(", ")}</TableCell>
                    <TableCell>
                      <Badge variant={k.enabled ? "default" : "secondary"}>
                        {k.enabled ? "Active" : "Disabled"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => remove.mutate(k.id)}
                        disabled={remove.isPending}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
