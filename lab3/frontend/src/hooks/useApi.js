import { useCallback, useMemo } from 'react';
import { API } from '../constants.js';

export function useApi(token) {
  const headers = useMemo(
    () => ({
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }),
    [token]
  );

  return useCallback(
    async (method, path, body) => {
      const opts = { method, headers };
      if (body !== undefined) opts.body = JSON.stringify(body);
      const r = await fetch(`${API}${path}`, opts);
      const text = await r.text();
      let json = null;
      try {
        json = text ? JSON.parse(text) : null;
      } catch {
        json = { raw: text };
      }
      if (!r.ok) {
        const base = json?.error || json?.message || r.statusText || 'Ошибка запроса';
        const err = new Error(json?.detail ? `${base}: ${json.detail}` : base);
        err.status = r.status;
        err.body = json;
        throw err;
      }
      return json;
    },
    [headers]
  );
}
