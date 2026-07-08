"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { getItems, getSources, getTopics, type Item, type Source, type Topic } from "@/lib/api";

const SOURCE_COLORS: Record<string, string> = {
  reddit: "text-orange-600 bg-orange-50 dark:text-orange-400 dark:bg-orange-950",
  hackernews: "text-amber-700 bg-amber-50 dark:text-amber-400 dark:bg-amber-950",
  rss: "text-emerald-700 bg-emerald-50 dark:text-emerald-400 dark:bg-emerald-950",
  secedgar: "text-indigo-700 bg-indigo-50 dark:text-indigo-400 dark:bg-indigo-950",
};

const SOURCE_BORDER: Record<string, string> = {
  reddit: "ring-orange-400",
  hackernews: "ring-amber-400",
  rss: "ring-emerald-400",
  secedgar: "ring-indigo-400",
};

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function isSelected(selected: Set<number>, sourceID: number): boolean {
  return selected.size === 0 || selected.has(sourceID);
}

export default function Home() {
  const [items, setItems] = useState<Item[]>([]);
  const [sources, setSources] = useState<Source[]>([]);
  const [topics, setTopics] = useState<Topic[]>([]);
  const [selectedSources, setSelectedSources] = useState<Set<number>>(new Set());
  const [selectedTopic, setSelectedTopic] = useState<number>(0);
  const [initialLoading, setInitialLoading] = useState(true);
  const [totalCount, setTotalCount] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loadingMore, setLoadingMore] = useState(false);
  const isFirstLoad = useRef(true);

  const loadConfig = useCallback(async () => {
    try {
      const [sourcesData, topicsData] = await Promise.all([getSources(), getTopics()]);
      setSources(sourcesData);
      setTopics(topicsData);
      return { sourcesData, topicsData };
    } catch (err) {
      console.error(err);
      return { sourcesData: null, topicsData: null };
    }
  }, []);

  const resolveQuery = useCallback(() => {
    let sourceId: number | undefined;
    let sourceType: string | undefined;

    if (sources.length === 0) return { sourceId, sourceType };

    if (selectedSources.size === 1) {
      sourceId = [...selectedSources][0];
    } else if (selectedSources.size > 1) {
      const types = new Set(
        sources.filter((s) => selectedSources.has(s.id)).map((s) => s.type)
      );
      if (types.size === 1) {
        sourceType = [...types][0];
      }
    }
    return { sourceId, sourceType };
  }, [selectedSources, sources]);

  const fetchItems = useCallback(async (currentOffset: number = 0) => {
    try {
      if (currentOffset !== 0) setLoadingMore(true);

      const { sourceId, sourceType } = resolveQuery();

      const result = await getItems({
        limit: 50,
        offset: currentOffset,
        source_id: sourceId,
        source_type: sourceType,
        topic_id: selectedTopic || undefined,
      });

      if (currentOffset === 0) {
        setItems(result.items);
      } else {
        setItems((prev) => [...prev, ...result.items]);
      }
      setTotalCount(result.total);
      setOffset(currentOffset + result.items.length);
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingMore(false);
    }
  }, [selectedTopic, resolveQuery]);

  const fetchData = useCallback(async (currentOffset: number = 0) => {
    if (currentOffset === 0 && isFirstLoad.current) {
      setInitialLoading(true);
      await loadConfig();
      await fetchItems(0);
      setInitialLoading(false);
      isFirstLoad.current = false;
    } else {
      await fetchItems(currentOffset);
    }
  }, [loadConfig, fetchItems]);

  useEffect(() => {
    fetchData(0);
    const interval = setInterval(() => fetchData(0), 30000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const enabledSources = sources.filter((s) => s.enabled);
  const redditSourceIDs = enabledSources
    .filter((s) => s.url.includes("reddit.com"))
    .map((s) => s.id);
  const redditSources = enabledSources.filter((s) => s.url.includes("reddit.com"));
  const edgarSourceIDs = enabledSources
    .filter((s) => s.type === "secedgar")
    .map((s) => s.id);
  const edgarSources = enabledSources.filter((s) => s.type === "secedgar");
  const nonRedditSources = enabledSources.filter((s) => !s.url.includes("reddit.com") && s.type !== "secedgar").sort((a, b) => {
    if (a.type === "hackernews") return -1;
    if (b.type === "hackernews") return 1;
    return 0;
  });
  const anyRedditSelected = redditSourceIDs.some((id) => selectedSources.has(id));
  const redditPillHighlighted = redditSourceIDs.length > 0 && (selectedSources.size === 0 || anyRedditSelected);
  const showRedditPills = anyRedditSelected;
  const anyEdgarSelected = edgarSourceIDs.some((id) => selectedSources.has(id));
  const edgarPillHighlighted = edgarSourceIDs.length > 0 && (selectedSources.size === 0 || anyEdgarSelected);
  const showEdgarPills = anyEdgarSelected;

  const toggleSource = (id: number) => {
    setSelectedSources((prev) => {
      const isReddit = redditSourceIDs.includes(id);
      const allRedditSelected = redditSourceIDs.every((rid) => prev.has(rid));

      if (isReddit && allRedditSelected) {
        const next = new Set(prev);
        redditSourceIDs.forEach((rid) => next.delete(rid));
        next.add(id);
        return next;
      }

      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const selectReddit = () => {
    if (anyRedditSelected) {
      setSelectedSources((prev) => {
        const next = new Set(prev);
        redditSourceIDs.forEach((id) => next.delete(id));
        return next;
      });
    } else if (selectedSources.size === 0) {
      setSelectedSources(new Set(redditSourceIDs));
    } else {
      setSelectedSources((prev) => {
        const next = new Set(prev);
        redditSourceIDs.forEach((id) => next.add(id));
        return next;
      });
    }
  };

  const selectEdgar = () => {
    if (anyEdgarSelected) {
      setSelectedSources((prev) => {
        const next = new Set(prev);
        edgarSourceIDs.forEach((id) => next.delete(id));
        return next;
      });
    } else if (selectedSources.size === 0) {
      setSelectedSources(new Set(edgarSourceIDs));
    } else {
      setSelectedSources((prev) => {
        const next = new Set(prev);
        edgarSourceIDs.forEach((id) => next.add(id));
        return next;
      });
    }
  };

  const pillClass = (active: boolean, color?: string, border?: string) =>
    `shrink-0 text-xs font-medium px-3 py-1.5 rounded-full border transition-colors ${
      active
        ? `${color || "bg-zinc-900 text-white border-zinc-900 dark:bg-zinc-100 dark:text-zinc-900 dark:border-zinc-100"} ${border ? `ring-2 ${border} border-transparent` : ""}`
        : "bg-white dark:bg-zinc-900 text-zinc-500 border-zinc-200 dark:border-zinc-700 hover:border-zinc-400"
    }`;

  return (
    <div className="max-w-2xl mx-auto py-8 px-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">
          {selectedTopic
            ? `Topic: ${topics.find((t) => t.id === selectedTopic)?.name}`
            : "Your Timeline"}
        </h2>
        {!initialLoading && (
          <p className="text-xs text-zinc-400 dark:text-zinc-500 mt-1">
            Showing {items.length} of {totalCount} items
          </p>
        )}
        {topics.length > 0 && (
          <select
            value={selectedTopic}
            onChange={(e) => {
              setSelectedTopic(Number(e.target.value));
            }}
            className="text-sm border border-zinc-300 dark:border-zinc-700 rounded-md px-2 py-1 bg-white dark:bg-zinc-900"
          >
            <option value={0}>All topics</option>
            {topics.map((t) => (
              <option key={t.id} value={t.id}>
                {t.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {enabledSources.length > 0 && (
        <div className="mb-5 space-y-2.5">
          <div className="flex flex-wrap gap-2.5">
            <button
              onClick={() => { setSelectedSources(new Set()); }}
              className={pillClass(selectedSources.size === 0)}
            >
              All
            </button>
            {nonRedditSources.map((source) => (
              <button
                key={source.id}
                onClick={() => toggleSource(source.id)}
                className={pillClass(
                  isSelected(selectedSources, source.id),
                  SOURCE_COLORS[source.type],
                  SOURCE_BORDER[source.type]
                )}
              >
                {source.name}
              </button>
            ))}
            {redditSourceIDs.length > 0 && (
              <button onClick={selectReddit} className={pillClass(redditPillHighlighted, "text-red-600 bg-red-50 dark:text-red-400 dark:bg-red-950", "ring-red-400")}>
                Reddit
              </button>
            )}
            {edgarSourceIDs.length > 0 && (
              <button onClick={selectEdgar} className={pillClass(edgarPillHighlighted, "text-indigo-600 bg-indigo-50 dark:text-indigo-400 dark:bg-indigo-950", "ring-indigo-400")}>
                SEC EDGAR
              </button>
            )}
          </div>
          {showRedditPills && (
            <div className="flex flex-wrap gap-2.5 max-h-[12rem] overflow-y-auto p-1 -mx-1 [&::-webkit-scrollbar]:hidden" style={{ scrollbarWidth: "none" }}>
              {redditSources.map((source) => (
                <button
                  key={source.id}
                  onClick={() => toggleSource(source.id)}
                  className={pillClass(
                    isSelected(selectedSources, source.id),
                    SOURCE_COLORS[source.type],
                    SOURCE_BORDER[source.type]
                  )}
                >
                  {source.name}
                </button>
          ))}
        </div>
          )}
          {showEdgarPills && (
            <div className="flex flex-wrap gap-2.5 max-h-[12rem] overflow-y-auto p-1 -mx-1 [&::-webkit-scrollbar]:hidden" style={{ scrollbarWidth: "none" }}>
              {edgarSources.map((source) => {
                const ticker = source.name.includes(" - ") ? source.name.split(" - ")[0] : source.name;
                return (
                <button
                  key={source.id}
                  onClick={() => toggleSource(source.id)}
                  className={pillClass(
                    isSelected(selectedSources, source.id),
                    SOURCE_COLORS[source.type],
                    SOURCE_BORDER[source.type]
                  )}
                >
                  {ticker}
                </button>
                );
              })}
            </div>
          )}
        </div>
      )}

      {initialLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="animate-pulse rounded-lg border border-zinc-200 dark:border-zinc-800 p-4">
              <div className="h-4 bg-zinc-200 dark:bg-zinc-800 rounded w-3/4 mb-2" />
              <div className="h-3 bg-zinc-200 dark:bg-zinc-800 rounded w-1/2" />
            </div>
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-zinc-500 dark:text-zinc-400">
            No items yet. Add sources and wait for them to populate.
          </p>
        </div>
      ) : (
        <div className="space-y-3 transition-opacity duration-350">
          {items.map((item) => (
            <a
              key={item.id}
              href={item.url}
              target="_blank"
              rel="noopener noreferrer"
              className="block rounded-lg border border-zinc-200 dark:border-zinc-800 p-4 hover:border-zinc-400 dark:hover:border-zinc-600 transition-colors bg-white dark:bg-zinc-900"
            >
              <div className="flex items-center gap-2 mb-1">
                <span
                  className={`text-xs font-medium px-2 py-0.5 rounded ${SOURCE_COLORS[item.source_type] || "text-zinc-600 bg-zinc-100 dark:text-zinc-400 dark:bg-zinc-800"}`}
                >
                  {item.source_name}
                </span>
                <span className="text-xs text-zinc-400">{timeAgo(item.published_at)}</span>
                {item.source_type === "hackernews" && (() => {
                  try {
                    const m = JSON.parse(item.metadata);
                    const threadURL = m.thread_url || item.id.replace(/^hn-/, "https://news.ycombinator.com/item?id=");
                    return (
                      <>
                        <span className="text-xs px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500">
                          {m.score} pts
                        </span>
                        <span
                          onClick={(e) => {
                            e.stopPropagation();
                            e.preventDefault();
                            window.open(threadURL, "_blank", "noopener,noreferrer");
                          }}
                          className="text-xs px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 hover:text-amber-600 dark:hover:text-amber-400 transition-colors cursor-pointer"
                        >
                          {m.descendants} comments
                        </span>
                      </>
                    );
                  } catch {
                    return null;
                  }
                })()}
                {item.source_type === "secedgar" && (() => {
                  try {
                    const m = JSON.parse(item.metadata);
                    return (
                      <span className="text-xs px-2 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400">
                        {m.filing_label}
                      </span>
                    );
                  } catch {
                    return null;
                  }
                })()}
              </div>
              <h3 className="text-sm font-medium text-zinc-900 dark:text-zinc-100 leading-snug">
                {item.title}
              </h3>
              {item.body && !item.url.includes("reddit.com") && (
                <p className="text-xs text-zinc-500 dark:text-zinc-400 mt-1 line-clamp-2">
                  {item.body}
                </p>
              )}
            </a>
          ))}
          {items.length < totalCount && (
            <div className="pt-4">
              <button
                onClick={() => fetchData(offset)}
                disabled={loadingMore}
                className="w-full py-3 px-4 text-sm font-medium text-zinc-500 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg hover:border-zinc-400 dark:hover:border-zinc-600 hover:text-zinc-700 dark:hover:text-zinc-300 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {loadingMore ? "Loading..." : "Show more"}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
