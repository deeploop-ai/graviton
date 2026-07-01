export class GravitonError extends Error {
  readonly status: number;
  readonly code?: string;
  readonly body?: unknown;

  constructor(message: string, status: number, code?: string, body?: unknown) {
    super(message);
    this.name = "GravitonError";
    this.status = status;
    this.code = code;
    this.body = body;
  }
}

export async function parseErrorResponse(res: Response): Promise<GravitonError> {
  let body: unknown;
  try {
    body = await res.json();
  } catch {
    body = undefined;
  }
  const errObj = body as { error?: { message?: string; code?: string } } | undefined;
  const message = errObj?.error?.message ?? res.statusText ?? "Request failed";
  const code = errObj?.error?.code;
  return new GravitonError(message, res.status, code, body);
}
