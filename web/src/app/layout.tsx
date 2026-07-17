import type { Metadata } from "next"
import { GeistSans } from "geist/font/sans"
import { GeistMono } from "geist/font/mono"

import { AppProviders } from "@/components/app-providers"

import "./globals.css"

export const metadata: Metadata = {
  title: "Capcom Console",
  description: "Runtime-agnostic AgentOps control plane",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${GeistSans.variable} ${GeistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full">
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  )
}
