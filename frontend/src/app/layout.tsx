import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Link from "next/link";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Timeline",
  description: "Your unified chronological feed",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex bg-zinc-50 dark:bg-zinc-950">
        <nav className="w-48 shrink-0 border-r border-zinc-200 dark:border-zinc-800 p-4 flex flex-col gap-2">
          <h1 className="text-lg font-semibold mb-2 text-zinc-900 dark:text-zinc-100">
            Timeline
          </h1>
          <NavLink href="/">Feed</NavLink>
          <NavLink href="/sources">Sources</NavLink>
          <NavLink href="/topics">Topics</NavLink>
        </nav>
        <main className="flex-1 min-w-0">{children}</main>
      </body>
    </html>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className="px-3 py-1.5 rounded-md text-sm font-medium text-zinc-600 hover:bg-zinc-200 dark:text-zinc-400 dark:hover:bg-zinc-800 transition-colors"
    >
      {children}
    </Link>
  );
}
