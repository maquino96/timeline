"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

const LINKS = [
  { href: "/", label: "Feed" },
  { href: "/sales", label: "Sales" },
];

const SETTINGS_LINKS = [
  { href: "/sources", label: "Sources" },
  { href: "/topics", label: "Topics" },
];

function NavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const onSettingsRoute = SETTINGS_LINKS.some((l) => l.href === pathname);
  const [settingsOpen, setSettingsOpen] = useState(onSettingsRoute);

  const renderLink = ({ href, label }: { href: string; label: string }) => {
    const active = pathname === href;
    return (
      <Link
        key={href}
        href={href}
        onClick={onNavigate}
        className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
          active
            ? "bg-zinc-200 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-100"
            : "text-zinc-600 hover:bg-zinc-200 dark:text-zinc-400 dark:hover:bg-zinc-800"
        }`}
      >
        {label}
      </Link>
    );
  };
  return (
    <>
      {LINKS.map(renderLink)}
      <button
        onClick={() => setSettingsOpen((o) => !o)}
        aria-expanded={settingsOpen}
        className="mt-4 flex items-center justify-between px-3 py-1 text-xs font-semibold uppercase tracking-wide text-zinc-400 dark:text-zinc-600 hover:text-zinc-600 dark:hover:text-zinc-400 transition-colors"
      >
        <span>Settings</span>
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          className={`transition-transform ${settingsOpen ? "rotate-90" : ""}`}
        >
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </button>
      {settingsOpen && <div className="flex flex-col gap-2 mt-1">{SETTINGS_LINKS.map(renderLink)}</div>}
    </>
  );
}

export default function Nav() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <nav className="hidden md:flex w-48 shrink-0 border-r border-zinc-200 dark:border-zinc-800 p-4 flex-col gap-2">
        <h1 className="text-lg font-semibold mb-2 text-zinc-900 dark:text-zinc-100">
          Timeline
        </h1>
        <NavLinks />
      </nav>

      <header className="md:hidden fixed top-0 inset-x-0 z-30 flex items-center gap-3 h-14 px-4 border-b border-zinc-200 dark:border-zinc-800 bg-zinc-50/90 dark:bg-zinc-950/90 backdrop-blur">
        <button
          onClick={() => setOpen(true)}
          aria-label="Open menu"
          className="p-2 -ml-2 rounded-md text-zinc-600 hover:bg-zinc-200 dark:text-zinc-400 dark:hover:bg-zinc-800 transition-colors"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </button>
        <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
          Timeline
        </h1>
      </header>

      {open && (
        <div className="md:hidden fixed inset-0 z-40">
          <div
            className="absolute inset-0 bg-black/40"
            onClick={() => setOpen(false)}
          />
          <nav className="absolute top-0 left-0 h-full w-64 max-w-[80%] flex flex-col gap-2 p-4 bg-zinc-50 dark:bg-zinc-950 border-r border-zinc-200 dark:border-zinc-800 shadow-xl">
            <div className="flex items-center justify-between mb-2">
              <h1 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
                Timeline
              </h1>
              <button
                onClick={() => setOpen(false)}
                aria-label="Close menu"
                className="p-2 -mr-2 rounded-md text-zinc-600 hover:bg-zinc-200 dark:text-zinc-400 dark:hover:bg-zinc-800 transition-colors"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            <NavLinks onNavigate={() => setOpen(false)} />
          </nav>
        </div>
      )}
    </>
  );
}
