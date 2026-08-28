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
  title: "Health Bot",
  description: "Personal health & fitness tracker",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="es"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-gray-50">
        <nav className="bg-white shadow-sm sticky top-0 z-10">
          <div className="max-w-3xl mx-auto px-6 py-3 flex gap-6">
            <Link href="/" className="font-semibold text-gray-800 hover:text-green-600">
              Hoy
            </Link>
            <Link href="/progress" className="text-gray-600 hover:text-green-600">
              Progreso
            </Link>
            <Link href="/gym" className="text-gray-600 hover:text-green-600">
              Gym
            </Link>
          </div>
        </nav>
        <div className="flex-1">{children}</div>
      </body>
    </html>
  );
}
