"use client";

import { useEffect, useState } from "react";
import { getSources, createSource, updateSource, deleteSource, pollSource, type Source } from "@/lib/api";

export default function SourcesPage() {
  const [sources, setSources] = useState<Source[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ type: "reddit", name: "", url: "", interval: 300 });

  const fetchSources = async () => {
    try {
      setSources(await getSources());
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSources();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    await createSource(form);
    setShowForm(false);
    setForm({ type: "reddit", name: "", url: "", interval: 300 });
    await fetchSources();
  };

  const handleToggle = async (source: Source) => {
    await updateSource(source.id, { enabled: !source.enabled });
    await fetchSources();
  };

  const handleDelete = async (id: number) => {
    await deleteSource(id);
    await fetchSources();
  };

  const handlePoll = async (id: number) => {
    await pollSource(id);
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto py-6 px-4 md:py-8 md:px-6">
        <p className="text-zinc-500">Loading...</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto py-6 px-4 md:py-8 md:px-6">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">Sources</h2>
        <button
          onClick={() => setShowForm(!showForm)}
          className="text-sm font-medium px-3 py-1.5 rounded-md bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900 hover:opacity-90 transition-opacity"
        >
          {showForm ? "Cancel" : "Add Source"}
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="mb-6 p-4 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 space-y-3">
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">Type</label>
            <select
              value={form.type}
              onChange={(e) => setForm({ ...form, type: e.target.value })}
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
            >
              <option value="reddit">Reddit</option>
              <option value="hackernews">Hacker News</option>
              <option value="rss">RSS</option>
              <option value="secedgar">SEC EDGAR</option>
            </select>
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder={form.type === "reddit" ? "r/golang" : form.type === "hackernews" ? "Hacker News" : form.type === "secedgar" ? "FDGRX - Growth Company Fund (316200104)" : "Ars Technica"}
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              required
            />
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">
              {form.type === "reddit" ? "Subreddit" : form.type === "hackernews" ? "List" : form.type === "secedgar" ? "CIK" : "Feed URL"}
            </label>
            <input
              type="text"
              value={form.url}
              onChange={(e) => setForm({ ...form, url: e.target.value })}
              placeholder={
                form.type === "reddit"
                  ? "golang"
                  : form.type === "hackernews"
                    ? "topstories"
                    : form.type === "secedgar"
                      ? "0000707823"
                      : "https://example.com/feed.xml"
              }
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              required
            />
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">Poll interval (seconds)</label>
            <input
              type="number"
              value={form.interval}
              onChange={(e) => setForm({ ...form, interval: Number(e.target.value) })}
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              min={30}
            />
          </div>
          <button
            type="submit"
            className="text-sm font-medium px-3 py-1.5 rounded-md bg-emerald-600 text-white hover:bg-emerald-700 transition-colors"
          >
            Create Source
          </button>
        </form>
      )}

      {sources.length === 0 ? (
        <p className="text-zinc-500 dark:text-zinc-400">No sources configured.</p>
      ) : (
        <div className="space-y-2">
          {sources.map((source) => (
            <div
              key={source.id}
              className="flex items-center justify-between p-3 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-xs font-mono px-1.5 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500">
                    {source.type}
                  </span>
                  <span className="text-sm font-medium text-zinc-900 dark:text-zinc-100 truncate">
                    {source.name}
                  </span>
                </div>
                <p className="text-xs text-zinc-400 mt-0.5 truncate">{source.url} · {source.interval}s</p>
              </div>
              <div className="flex items-center gap-1.5 shrink-0 ml-4">
                <button
                  onClick={() => handlePoll(source.id)}
                  className="text-xs px-2 py-1 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
                  title="Poll now"
                >
                  Refresh
                </button>
                <button
                  onClick={() => handleToggle(source)}
                  className={`text-xs px-2 py-1 rounded transition-colors ${source.enabled ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300" : "bg-zinc-100 text-zinc-400 dark:bg-zinc-800"}`}
                >
                  {source.enabled ? "On" : "Off"}
                </button>
                <button
                  onClick={() => handleDelete(source.id)}
                  className="text-xs px-2 py-1 rounded text-red-500 hover:bg-red-50 dark:hover:bg-red-950 transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
