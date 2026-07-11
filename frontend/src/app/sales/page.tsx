"use client";

import { useEffect, useState, useCallback } from "react";
import {
  getWatchItems,
  getSaleAlerts,
  getAuthStatus,
  login,
  logout,
  createWatchItem,
  updateWatchItem,
  deleteWatchItem,
  dismissSaleAlert,
  type WatchItem,
  type SaleAlert,
} from "@/lib/api";

const SOURCE_COLORS: Record<string, string> = {
  ebay: "text-amber-700 bg-amber-50 dark:text-amber-400 dark:bg-amber-950",
  slickdeals: "text-sky-700 bg-sky-50 dark:text-sky-400 dark:bg-sky-950",
  reddit: "text-orange-600 bg-orange-50 dark:text-orange-400 dark:bg-orange-950",
};

function fmt(n: number | null | undefined): string {
  if (n === null || n === undefined || n <= 0) return "—";
  return `$${n.toFixed(2)}`;
}

function bestPrice(item: WatchItem): number | null {
  const prices = [item.ebay_price, item.slickdeals_price, item.reddit_price].filter(
    (p): p is number => p !== null && p > 0
  );
  return prices.length ? Math.min(...prices) : null;
}

const emptyForm = { name: "", search_term: "", threshold: "", floor: "", category: "" };

export default function SalesPage() {
  const [items, setItems] = useState<WatchItem[]>([]);
  const [alerts, setAlerts] = useState<SaleAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [authed, setAuthed] = useState(false);
  const [authRequired, setAuthRequired] = useState(false);
  const [showLogin, setShowLogin] = useState(false);
  const [password, setPassword] = useState("");
  const [authError, setAuthError] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<WatchItem | null>(null);
  const [form, setForm] = useState(emptyForm);

  const load = useCallback(async () => {
    try {
      const [itemsData, alertsData, auth] = await Promise.all([
        getWatchItems(),
        getSaleAlerts(),
        getAuthStatus(),
      ]);
      setItems(itemsData);
      setAlerts(alertsData.alerts);
      setAuthed(auth.authenticated);
      setAuthRequired(auth.auth_required);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setAuthError("");
    try {
      await login(password);
      setPassword("");
      setShowLogin(false);
      await load();
    } catch {
      setAuthError("Invalid password");
    }
  };

  const handleLogout = async () => {
    await logout();
    await load();
  };

  const openAdd = () => {
    setEditing(null);
    setForm(emptyForm);
    setShowForm(true);
  };

  const openEdit = (item: WatchItem) => {
    setEditing(item);
    setForm({
      name: item.name,
      search_term: item.search_term,
      threshold: String(item.threshold),
      floor: String(item.floor),
      category: item.category,
    });
    setShowForm(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const payload = {
      name: form.name.trim(),
      search_term: form.search_term.trim() || form.name.trim(),
      threshold: parseFloat(form.threshold) || 0,
      floor: parseFloat(form.floor) || 0,
      category: form.category.trim(),
    };
    if (editing) {
      await updateWatchItem(editing.id, { ...payload, active: editing.active });
    } else {
      await createWatchItem(payload);
    }
    setShowForm(false);
    setForm(emptyForm);
    setEditing(null);
    await load();
  };

  const handleDelete = async (id: number) => {
    await deleteWatchItem(id);
    await load();
  };

  const handleDismiss = async (id: number) => {
    await dismissSaleAlert(id);
    setAlerts((prev) => prev.filter((a) => a.id !== id));
  };

  const canManage = authed || !authRequired;

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
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">Sales</h2>
        <div className="flex items-center gap-2">
          {canManage && (
            <button
              onClick={showForm ? () => setShowForm(false) : openAdd}
              className="text-sm font-medium px-3 py-1.5 rounded-md bg-teal-600 text-white hover:bg-teal-700 transition-colors"
            >
              {showForm ? "Cancel" : "Add Item"}
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
        <form
          onSubmit={handleSubmit}
          className="mb-6 p-4 rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 space-y-3"
        >
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">Name</label>
            <input
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="TeamGroup MP33 2TB"
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              required
            />
          </div>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">
              Search term (defaults to name)
            </label>
            <input
              type="text"
              value={form.search_term}
              onChange={(e) => setForm({ ...form, search_term: e.target.value })}
              placeholder="TeamGroup MP33 2TB"
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-medium text-zinc-500 block mb-1">
                Alert below ($)
              </label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={form.threshold}
                onChange={(e) => setForm({ ...form, threshold: e.target.value })}
                placeholder="80.00"
                className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
                required
              />
            </div>
            <div>
              <label className="text-xs font-medium text-zinc-500 block mb-1">
                Price floor ($)
              </label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={form.floor}
                onChange={(e) => setForm({ ...form, floor: e.target.value })}
                placeholder="40.00"
                className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
              />
            </div>
          </div>
          <p className="text-xs text-zinc-400">
            Floor rejects false positives — prices below it (shipping-only or sub-value
            listings) are ignored.
          </p>
          <div>
            <label className="text-xs font-medium text-zinc-500 block mb-1">
              Category (optional)
            </label>
            <input
              type="text"
              value={form.category}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              placeholder="NVMe 2TB"
              className="w-full text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-3 py-1.5 bg-white dark:bg-zinc-800"
            />
          </div>
          <button
            type="submit"
            className="text-sm font-medium px-3 py-1.5 rounded-md bg-teal-600 text-white hover:bg-teal-700 transition-colors"
          >
            {editing ? "Save Changes" : "Add Item"}
          </button>
        </form>
      )}

      <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wide mb-3">
        Recent Alerts
      </h3>
      {alerts.length === 0 ? (
        <p className="text-sm text-zinc-500 dark:text-zinc-400 mb-8">
          No price drops in the last 14 days.
        </p>
      ) : (
        <div className="space-y-3 mb-8">
          {alerts.map((alert) => (
            <div
              key={alert.id}
              className="rounded-lg border border-zinc-200 dark:border-zinc-800 p-4 bg-white dark:bg-zinc-900"
            >
              <div className="flex items-start justify-between gap-2">
                <span className="text-lg font-bold text-green-600 dark:text-green-500">
                  {fmt(alert.price)}
                </span>
                {canManage && (
                  <button
                    onClick={() => handleDismiss(alert.id)}
                    className="text-zinc-400 hover:text-red-500 text-lg leading-none px-1"
                    title="Dismiss"
                  >
                    &times;
                  </button>
                )}
              </div>
              <a
                href={alert.deal_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-medium text-zinc-900 dark:text-zinc-100 hover:underline block mt-1"
              >
                {alert.title}
              </a>
              <div className="flex items-center gap-2 mt-1.5">
                <span
                  className={`text-xs font-medium px-2 py-0.5 rounded ${SOURCE_COLORS[alert.source] || "text-zinc-600 bg-zinc-100 dark:text-zinc-400 dark:bg-zinc-800"}`}
                >
                  {alert.source}
                </span>
                <span className="text-xs text-zinc-400">{alert.created_at.slice(0, 10)}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wide mb-3">
        Watched Items
      </h3>
      {items.length === 0 ? (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">No items being watched yet.</p>
      ) : (
        <div className="space-y-3">
          {items.map((item) => {
            const best = bestPrice(item);
            const below = best !== null && best < item.threshold;
            return (
              <div
                key={item.id}
                className="rounded-lg border border-zinc-200 dark:border-zinc-800 p-4 bg-white dark:bg-zinc-900"
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100 truncate">
                      {item.name}
                    </p>
                    {item.category && (
                      <p className="text-xs text-zinc-400 mt-0.5">{item.category}</p>
                    )}
                  </div>
                  {canManage && (
                    <div className="flex items-center gap-1 shrink-0">
                      <button
                        onClick={() => openEdit(item)}
                        className="text-xs px-2 py-1 rounded hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(item.id)}
                        className="text-xs px-2 py-1 rounded text-red-500 hover:bg-red-50 dark:hover:bg-red-950 transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-2 text-xs text-zinc-500">
                  <span>eBay {fmt(item.ebay_price)}</span>
                  <span>Slickdeals {fmt(item.slickdeals_price)}</span>
                  <span>Reddit {fmt(item.reddit_price)}</span>
                </div>
                <div className="flex items-center gap-3 mt-2 text-xs">
                  <span className="text-zinc-400">
                    Alert &lt; ${item.threshold.toFixed(2)}
                  </span>
                  <span className="text-zinc-400">Floor ${item.floor.toFixed(2)}</span>
                  <span
                    className={`font-semibold ${below ? "text-green-600 dark:text-green-500" : "text-zinc-500"}`}
                  >
                    Best {fmt(best)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
