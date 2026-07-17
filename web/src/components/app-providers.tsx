"use client"

import * as React from "react"
import {
  MutationCache,
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query"
import { ThemeProvider } from "next-themes"
import { toast } from "sonner"

import { AppShell } from "@/components/app-shell"
import { Toaster } from "@/components/ui/sonner"
import { TooltipProvider } from "@/components/ui/tooltip"
import { ApiError } from "@/lib/api-client"

export function AppProviders({ children }: { children: React.ReactNode }) {
  const [queryClient] = React.useState(
    () =>
      new QueryClient({
        queryCache: new QueryCache({
          onError: showRequestError,
        }),
        mutationCache: new MutationCache({
          onError: showRequestError,
        }),
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            refetchOnWindowFocus: false,
          },
        },
      })
  )

  return (
    <ThemeProvider
      attribute="class"
      defaultTheme="dark"
      enableSystem={false}
      disableTransitionOnChange
    >
      <QueryClientProvider client={queryClient}>
        <TooltipProvider delay={250}>
          <AppShell>{children}</AppShell>
          <Toaster position="bottom-right" richColors={false} />
        </TooltipProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}

function showRequestError(error: unknown) {
  if (error instanceof ApiError) {
    toast.error(error.message)
    return
  }
  if (error instanceof Error) {
    toast.error(error.message)
    return
  }
  toast.error("Request failed")
}
