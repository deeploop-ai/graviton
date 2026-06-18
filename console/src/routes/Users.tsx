import { useQuery } from "@tanstack/react-query";
import { listUsers, type User } from "@/api/users";
import { PageHeader } from "@/components/PageHeader";
import { LoadingTable } from "@/components/LoadingTable";
import { EmptyState } from "@/components/EmptyState";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

export function Users() {
  const { data: users = [], isLoading } = useQuery({
    queryKey: ["users"],
    queryFn: listUsers,
  });

  return (
    <div className="space-y-6">
      <PageHeader title="Users" description="Users registered in the selected project" />
      <Card>
        <CardHeader>
          <CardTitle>Project users</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <LoadingTable columns={6} />
          ) : users.length === 0 ? (
            <EmptyState title="No users" description="Users will appear after they sign up." />
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Email</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Verified</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((u: User) => (
                  <TableRow key={u.id}>
                    <TableCell className="font-mono text-xs max-w-[120px] truncate">{u.id}</TableCell>
                    <TableCell>{u.email}</TableCell>
                    <TableCell>{u.name}</TableCell>
                    <TableCell>
                      <Badge variant={u.status === "active" ? "default" : "secondary"}>{u.status}</Badge>
                    </TableCell>
                    <TableCell>{u.email_verified ? "Yes" : "No"}</TableCell>
                    <TableCell>{new Date(u.created_at).toLocaleString()}</TableCell>
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
