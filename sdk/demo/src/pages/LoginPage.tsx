import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useOrionid } from "@/lib/orionid-context";

export function LoginPage() {
  const { client, setAuth, run } = useOrionid();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await run(() => client.account.signIn({ email, password }));
      setAuth({
        accessToken: res.tokens.access_token,
        refreshToken: res.tokens.refresh_token,
        email: res.account.email,
        name: res.account.name,
        userId: res.account.id,
      });
      navigate("/app");
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <form className="panel space-y-4 p-6" onSubmit={onSubmit}>
      <h2 className="text-lg font-semibold text-white">登录</h2>
      {error ? (
        <div className="rounded-lg border border-red-500/40 bg-red-500/10 px-3 py-2 text-sm text-red-200">
          {error}
        </div>
      ) : null}
      <label className="block space-y-1">
        <span className="text-xs text-orionid-muted">邮箱</span>
        <input
          className="field"
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@example.com"
        />
      </label>
      <label className="block space-y-1">
        <span className="text-xs text-orionid-muted">密码</span>
        <input
          className="field"
          type="password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="••••••••"
        />
      </label>
      <button type="submit" className="btn-primary w-full" disabled={loading}>
        {loading ? "登录中…" : "登录"}
      </button>
      <p className="text-center text-sm text-orionid-muted">
        还没有账号？{" "}
        <Link className="text-orionid-accent hover:underline" to="/register">
          注册
        </Link>
      </p>
    </form>
  );
}
