import type { Metadata } from "next";
import Link from "next/link";
import { Geist, Geist_Mono } from "next/font/google";
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
  title: "CivicLens — Belediye Operasyon Konsolu",
  description: "KVKK-safe urban maintenance operations console",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="tr" className={`${geistSans.variable} ${geistMono.variable}`}>
      <body className="min-h-screen bg-slate-50 font-sans text-slate-900 antialiased">
        <nav className="sticky top-0 z-10 border-b border-slate-200 bg-white/90 backdrop-blur-sm">
          <div className="mx-auto flex max-w-7xl items-center px-4 py-3 sm:px-6 lg:px-8">
            <Link
              href="/"
              className="text-sm font-bold tracking-tight text-slate-900 transition-colors hover:text-slate-700"
            >
              CivicLens
            </Link>
          </div>
        </nav>
        {children}
      </body>
    </html>
  );
}
