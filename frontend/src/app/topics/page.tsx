"use client";

import { useEffect, useState } from "react";
import { getTopics, createTopic, updateTopic, deleteTopic, getAuthStatus, login, logout, type Topic } from "@/lib/api";

export default function TopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: "", keywords: "" });
  const [authed, setAuthed] = useState(false);
  const [authRequired, setAuthRequired] = useState(false);
  const [showLogin, setShowLogin] = useState(false);
  const [password, setPassword] = useState("");
  const [authError, setAuthError] = useState("");

  const canManage = authed || !authRequired;

  const fetchTopics = async () => {
    try {
      const [topicsData, auth] = await Promise.all([getTopics(), getAuthStatus()]);
      setTopics(topicsData);
      setAuthed(auth.authenticated);
      setAuthRequired(auth.auth_required);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTopics();
  }, []);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setAuthError("");
    try {
      await login(password);
      setPassword("");
      setShowLogin(false);
      await fetchTopics();
    } catch {
      setAuthError("Invalid password");
    }
  };

  const handleLogout = async () => {
    await logout();
    await fetchTopics();
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    await createTopic(form);
    setShowForm(false);
    setForm({ name: "", keywords: "" });
    await fetchTopics();
  };

  const handleToggle = async (topic: Topic) => {
    await updateTopic(topic.id, { enabled: !topic.enabled });
    await fetchTopics();
  };

  const handleDelete = async (id: number) => {
    await deleteTopic(id);
    await fetchTopics();
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
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">Topics</h2>
        <div className="flex items-center gap-2">
          {canManage && (
            <button
              onClick={() => setShowForm(!showForm)}
              className="text-sm font-medium px-3 py-1.5 rounded-md bg-zinc-900 text-white dark:bg-zinc-100 dark:text-zinc-900 hover:opacity-90 transition-opacity"
            >
              {showForm ? "Cancel" : "Add Topic"}
            </button>
          )}
          {authRequired && authed && (
            <button
              onClick={handleLogout}
              className="text-sm font-medium px-3 py-1.5 rounded-md text-zinc-500 hover:bg-zinc-200 dark:hover:bg-zinc-800 transition-colors"
            >
              Log out
            </button>
          )}
          {authRequired && !authed && (
            <button
              onClick={() => setShowLogin(!showLogin)}
              className="text-sm font-medium px-3 py-1.5 rounded-md text-zinc-500 hover:bg-zinc-200 dark:hover:bg-zinc-800 transition-colors"
            >
              {showLogin ? "Cancel" : "Log in to manage"}
            </button>
          )}
        </div>
      </div>

      {showLogin && !authed && (
        <form
          onSubmit={handleLogin}
          className="mb-6 p-4 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 flex items-center gap-2"
        >
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            className="flex-1 text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
            autoFocus
          />
          <button
            type="submit"
            className="text-sm font-medium px-3 py-1.5 rounded-md bg-teal-600 text-white hover:bg-teal-700 transition-colors"
          >
            Log in
          </button>
          {authError && <span className="text-xs text-red-500">{authError}</span>}
        </form>
      )}

      {showForm && canManage && (
        <form onSubmit={handleCreate} className="mb-6 p-4 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 space-y-3">
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="AI / ML"
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              required
            />
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">
              Keywords (comma-separated)
            </label>
            <input
              type="text"
              value={form.keywords}
              onChange={(e) => setForm({ ...form, keywords: e.target.value })}
              placeholder="ai, llm, gpt, machine learning"
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              required
            />
          </div>
          <button
            type="submit"
            className="text-sm font-medium px-3 py-1.5 rounded-md bg-emerald-600 text-white hover:bg-emerald-700 transition-colors"
          >
            Create Topic
          </button>
        </form>
      )}

      {topics.length === 0 ? (
        <p className="text-zinc-500 dark:text-zinc-400">
          No topics configured. Create topics to filter your timeline by keyword.
        </p>
      ) : (
        <div className="space-y-2">
          {topics.map((topic) => (
            <div
              key={topic.id}
              className="flex items-center justify-between p-3 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900"
            >
              <div className="flex-1 min-w-0">
                <span className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                  {topic.name}
                </span>
                <div className="flex flex-wrap gap-1 mt-1">
                  {topic.keywords.split(",").map((kw) => (
                    <span
                      key={kw}
                      className="text-xs px-1.5 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500"
                    >
                      {kw.trim()}
                    </span>
                  ))}
                </div>
              </div>
              <div className="flex items-center gap-1.5 shrink-0 ml-4">
                {canManage ? (
                  <>
                    <button
                      onClick={() => handleToggle(topic)}
                      className={`text-xs px-2 py-1 rounded transition-colors ${topic.enabled ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300" : "bg-zinc-100 text-zinc-400 dark:bg-zinc-800"}`}
                    >
                      {topic.enabled ? "On" : "Off"}
                    </button>
                    <button
                      onClick={() => handleDelete(topic.id)}
                      className="text-xs px-2 py-1 rounded text-red-500 hover:bg-red-50 dark:hover:bg-red-950 transition-colors"
                    >
                      Delete
                    </button>
                  </>
                ) : (
                  <span className={`text-xs px-2 py-1 rounded ${topic.enabled ? "bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300" : "bg-zinc-100 text-zinc-400 dark:bg-zinc-800"}`}>
                    {topic.enabled ? "On" : "Off"}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
