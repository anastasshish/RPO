export function fmtDate(v) {
  if (!v) return '—';
  try {
    return new Date(v).toLocaleString('ru-RU');
  } catch {
    return String(v);
  }
}
