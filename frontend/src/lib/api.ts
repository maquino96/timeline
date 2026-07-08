const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export interface Item {
  id: string;
  source_id: number;
  source_type: string;
  source_name: string;
  title: string;
  body: string;
  url: string;
  author: string;
  published_at: string;
  fetched_at: string;
  metadata: string;
}

export interface Source {
  id: number;
  type: string;
  name: string;
  url: string;
  interval: number;
  enabled: boolean;
  created_at: string;
}

export interface Topic {
  id: number;
  name: string;
  keywords: string;
  enabled: boolean;
  created_at: string;
}

async function fetchJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, options);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export async function getItems(params: {
  limit?: number;
  offset?: number;
  source_id?: number | string;
  source_type?: string;
  topic_id?: number;
  q?: string;
} = {}): Promise<{ items: Item[]; total: number }> {
  const sp = new URLSearchParams();
  if (params.limit) sp.set("limit", String(params.limit));
  if (params.offset) sp.set("offset", String(params.offset));
  if (params.source_id) sp.set("source_id", String(params.source_id));
  if (params.source_type) sp.set("source_type", params.source_type);
  if (params.topic_id) sp.set("topic_id", String(params.topic_id));
  if (params.q) sp.set("q", params.q);
  const qs = sp.toString();
  const res = await fetch(`${API_BASE}/api/items${qs ? `?${qs}` : ""}`);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  const items = await res.json();
  const total = parseInt(res.headers.get("X-Total-Count") || "0", 10);
  return { items, total };
}

export function getSources(): Promise<Source[]> {
  return fetchJSON<Source[]>("/api/sources");
}

export function createSource(source: Partial<Source>): Promise<Source> {
  return fetchJSON<Source>("/api/sources", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(source),
  });
}

export function updateSource(id: number, source: Partial<Source>): Promise<Source> {
  return fetchJSON<Source>(`/api/sources/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(source),
  });
}

export function deleteSource(id: number): Promise<void> {
  return fetchJSON<void>(`/api/sources/${id}`, { method: "DELETE" });
}

export function pollSource(id: number): Promise<void> {
  return fetchJSON<void>(`/api/sources/${id}/poll`, { method: "POST" });
}

export function getTopics(): Promise<Topic[]> {
  return fetchJSON<Topic[]>("/api/topics");
}

export function createTopic(topic: Partial<Topic>): Promise<Topic> {
  return fetchJSON<Topic>("/api/topics", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(topic),
  });
}

export function updateTopic(id: number, topic: Partial<Topic>): Promise<Topic> {
  return fetchJSON<Topic>(`/api/topics/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(topic),
  });
}

export function deleteTopic(id: number): Promise<void> {
  return fetchJSON<void>(`/api/topics/${id}`, { method: "DELETE" });
}
