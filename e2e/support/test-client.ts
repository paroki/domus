import { config } from "./config";

export class TestClient {
  private token: string | null = null;
  private parokiId: string | null = null;

  constructor(private readonly baseUrl = config.apiBaseUrl) {}

  setToken(token: string | null): void {
    this.token = token;
  }

  setParokiId(parokiId: string | null): void {
    this.parokiId = parokiId;
  }

  async request<T = unknown>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<{ status: number; data: T }> {
    const headers: Record<string, string> = {};
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }
    if (this.parokiId) {
      headers["X-Paroki-Id"] = this.parokiId;
    }
    if (body !== undefined) {
      headers["Content-Type"] = "application/json";
    }

    const res = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    const text = await res.text();
    const data = (text ? JSON.parse(text) : null) as T;
    return { status: res.status, data };
  }

  get<T = unknown>(path: string) {
    return this.request<T>("GET", path);
  }

  post<T = unknown>(path: string, body?: unknown) {
    return this.request<T>("POST", path, body);
  }

  put<T = unknown>(path: string, body?: unknown) {
    return this.request<T>("PUT", path, body);
  }

  patch<T = unknown>(path: string, body?: unknown) {
    return this.request<T>("PATCH", path, body);
  }

  delete<T = unknown>(path: string) {
    return this.request<T>("DELETE", path);
  }
}
