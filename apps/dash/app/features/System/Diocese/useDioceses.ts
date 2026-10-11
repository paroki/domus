import type { components } from "@domus/openapi";
import { useCallback, useEffect, useState } from "react";
import { useApi } from "~/lib/api";

export type Diocese = components["schemas"]["model.DioceseResponse"];
export type DioceseInput = components["schemas"]["model.CreateDioceseRequest"];
type ErrorResponse = components["schemas"]["httpx.ErrorResponse"];

/** Hasil mutasi: `ok`, atau error yang sudah dipilah supaya UI gampang menampilkannya. */
export type MutationResult =
  | { ok: true }
  | {
      ok: false;
      status: number;
      message?: string;
      /** Error validasi per field (422), kunci sudah dalam camelCase. */
      fields: Record<string, string>;
    };

function toFailure(status: number, body?: ErrorResponse): MutationResult {
  const fields: Record<string, string> = {};
  for (const f of body?.error?.fields ?? []) {
    if (!f.field) continue;
    // Validator Go melaporkan nama field struct ("Name"); form memakai "name".
    fields[f.field.charAt(0).toLowerCase() + f.field.slice(1)] = f.rule ?? "";
  }
  return { ok: false, status, message: body?.error?.message, fields };
}

/** CRUD keuskupan lewat `useApi`. Daftar dimuat otomatis dan disegarkan setelah mutasi. */
export function useDioceses() {
  const { GET, POST, PUT, DELETE } = useApi();
  const [items, setItems] = useState<Diocese[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  const reload = useCallback(async () => {
    setLoading(true);
    const { data, error } = await GET("/dioceses");
    setLoadError(Boolean(error) || !data);
    setItems(data?.data ?? []);
    setLoading(false);
  }, [GET]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const create = useCallback(
    async (body: DioceseInput): Promise<MutationResult> => {
      const { error, response } = await POST("/dioceses", { body });
      if (error || !response.ok) return toFailure(response.status, error);
      await reload();
      return { ok: true };
    },
    [POST, reload],
  );

  const update = useCallback(
    async (id: number, body: DioceseInput): Promise<MutationResult> => {
      const { error, response } = await PUT("/dioceses/{id}", {
        params: { path: { id } },
        body,
      });
      if (error || !response.ok) return toFailure(response.status, error);
      await reload();
      return { ok: true };
    },
    [PUT, reload],
  );

  const remove = useCallback(
    async (id: number): Promise<MutationResult> => {
      const { error, response } = await DELETE("/dioceses/{id}", {
        params: { path: { id } },
      });
      if (error || !response.ok) return toFailure(response.status, error);
      await reload();
      return { ok: true };
    },
    [DELETE, reload],
  );

  return { items, loading, loadError, reload, create, update, remove };
}
