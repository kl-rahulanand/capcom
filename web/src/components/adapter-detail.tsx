"use client"

import * as React from "react"
import { useMutation, useQueries, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import { AddInstanceDialog } from "@/components/add-instance-dialog"
import { AgentDrilldownDrawer } from "@/components/agent-drilldown-drawer"
import {
  AgentTableRow,
} from "@/components/agent-table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  Collapsible,
  CollapsibleContent,
} from "@/components/ui/collapsible"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  SyncDialog,
  type SyncDialogSubmit,
} from "@/components/sync-dialog"
import {
  buildAdaptersModel,
  statusClass,
  type AdapterInstance,
  type AdapterModel,
} from "@/lib/adapters"
import { capcomApi } from "@/lib/api-client"
import {
  queryKeys,
  usePersistedAgentsQuery,
  useRuntimeInstanceAgentsQuery,
  useRuntimeInstancesQuery,
  useSyncRuntimeInstanceMutation,
  useTestRuntimeInstanceMutation,
} from "@/lib/api-hooks"
import type {
  PersistedAgent,
  RuntimeCapabilities,
  RuntimeConnectionTestResult,
  RuntimeInstance,
  RuntimeType,
  RuntimeSyncRun,
} from "@/lib/api-types"
import { cn } from "@/lib/utils"

const AGENT_PREVIEW_LIMIT = 8

export function AdapterDetail({ adapterId }: { adapterId: string }) {
  const queryClient = useQueryClient()
  const [syncAllOpen, setSyncAllOpen] = React.useState(false)
  const [addInstanceOpen, setAddInstanceOpen] = React.useState(false)
  const [selectedAgent, setSelectedAgent] = React.useState<PersistedAgent | null>(
    null
  )
  const runtimeInstancesQuery = useRuntimeInstancesQuery()
  const agentsQuery = usePersistedAgentsQuery()
  const now = React.useMemo(() => {
    const timestamp = Math.max(
      runtimeInstancesQuery.dataUpdatedAt,
      agentsQuery.dataUpdatedAt,
      0
    )
    return timestamp > 0 ? new Date(timestamp) : new Date()
  }, [agentsQuery.dataUpdatedAt, runtimeInstancesQuery.dataUpdatedAt])
  const adaptersModel = React.useMemo(
    () =>
      buildAdaptersModel(
        runtimeInstancesQuery.data ?? [],
        agentsQuery.data ?? [],
        now
      ),
    [agentsQuery.data, now, runtimeInstancesQuery.data]
  )
  const adapter = adaptersModel.adapters.find((item) => item.id === adapterId)
  const loading = runtimeInstancesQuery.isLoading || agentsQuery.isLoading
  const syncAllMutation = useMutation<
    RuntimeSyncRun[],
    Error,
    SyncDialogSubmit
  >({
    mutationFn: async (payload) => {
      if (!adapter) {
        return []
      }
      return Promise.all(
        adapter.instances.map((item) =>
          capcomApi.syncRuntimeInstance(item.instance.id, payload)
        )
      )
    },
    onSuccess: async (runs) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.runtimeInstances }),
        queryClient.invalidateQueries({ queryKey: queryKeys.persistedAgents() }),
        ...runs.map((run) =>
          queryClient.invalidateQueries({
            queryKey: queryKeys.runtimeInstanceAgents(
              run.runtime_connection_id
            ),
          })
        ),
        ...runs.map((run) =>
          queryClient.invalidateQueries({
            queryKey: queryKeys.runtimeInstanceSyncRuns(
              run.runtime_connection_id
            ),
          })
        ),
      ])
      setSyncAllOpen(false)
      toast.success(
        `${runs.length} instance${runs.length === 1 ? "" : "s"} re-imported`
      )
    },
  })

  if (loading) {
    return <AdapterDetailSkeleton />
  }

  if (!adapter) {
    return (
      <EmptyState
        title="Adapter not found"
        description={`No runtime instances are connected for ${adapterId}.`}
      />
    )
  }

  const styles = statusClass(adapter.status)

  return (
    <section className="flex flex-col gap-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="text-[22px] font-bold leading-tight tracking-[-0.02em] text-[var(--tx)]">
              {adapter.name}
            </h1>
            <Badge className={cn("font-hud text-[11px]", styles.badge)}>
              {adapter.badge}
            </Badge>
          </div>
          <p className="mt-1 text-[13px] text-[var(--mu)]">
            {adapter.instanceCount} instances connected · {adapter.agentCount} agents running · state re-imported automatically every{" "}
            {syncIntervalLabel(adapter.instances)}
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            className="hover:border-[var(--ac)] hover:text-[var(--ac)]"
            onClick={() => setAddInstanceOpen(true)}
          >
            + Add instance
          </Button>
          <Button
            variant="outline"
            className="hover:border-[var(--ac)] hover:text-[var(--ac)]"
            onClick={() => toast.info("Adapter settings arrive in a later stage.")}
          >
            Adapter settings
          </Button>
          <Button
            className="shadow-[0_0_0_3px_var(--glow)] hover:brightness-[1.08]"
            onClick={() => setSyncAllOpen(true)}
          >
            ↻ Re-import all instances
          </Button>
        </div>
      </div>

      <div className="flex flex-col gap-3.5">
        {adapter.instances.map((item, index) => (
          <InstanceGroup
            key={item.instance.id}
            item={item}
            defaultOpen={index === 0}
            onAgentClick={setSelectedAgent}
          />
        ))}
      </div>

      <PageFooter adapter={adapter} />

      <AgentDrilldownDrawer
        agent={selectedAgent}
        open={Boolean(selectedAgent)}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) {
            setSelectedAgent(null)
          }
        }}
      />

      <SyncDialog
        open={syncAllOpen}
        targets={adapter.instances.map((item) => ({
          id: item.instance.id,
          name: item.instance.display_name || item.instance.name,
        }))}
        pending={syncAllMutation.isPending}
        defaultReason={`Re-import all ${adapter.name} runtime instances`}
        onOpenChange={setSyncAllOpen}
        onSubmit={(payload) => syncAllMutation.mutate(payload)}
      />

      <AddInstanceDialog
        open={addInstanceOpen}
        defaultAdapterId={runtimeTypeFromRoute(adapterId)}
        onOpenChange={setAddInstanceOpen}
      />
    </section>
  )
}

function InstanceGroup({
  item,
  defaultOpen,
  onAgentClick,
}: {
  item: AdapterInstance
  defaultOpen: boolean
  onAgentClick: (agent: PersistedAgent) => void
}) {
  const [open, setOpen] = React.useState(defaultOpen)
  const [syncOpen, setSyncOpen] = React.useState(false)
  const agentsQuery = useRuntimeInstanceAgentsQuery(item.instance.id)
  const syncMutation = useSyncRuntimeInstanceMutation(item.instance.id)
  const agents = agentsQuery.data ?? []
  const shownAgents = agents.slice(0, AGENT_PREVIEW_LIMIT)
  const styles = statusClass(item.status)
  const updateLabel =
    item.status === "failed"
      ? `import failed · ${item.updated}`
      : `updated ${item.updated}`

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <Card className="gap-0 border border-[var(--hl)] bg-[var(--el)] py-0 shadow-[var(--chi)]">
        <div
          className="flex cursor-pointer flex-wrap items-center gap-3 px-[18px] py-3.5 transition hover:bg-[var(--sl)]"
          onClick={() => setOpen((value) => !value)}
        >
          <button
            type="button"
            className="font-hud text-[11px] text-[var(--fa)]"
            aria-label={open ? "Collapse instance" : "Expand instance"}
            onClick={(event) => {
              event.stopPropagation()
              setOpen((value) => !value)
            }}
          >
            {open ? "▾" : "▸"}
          </button>
          <span className={cn("size-2 rounded-full", styles.dot)} />
          <span className="font-hud text-sm font-semibold text-[var(--tx)]">
            {item.instance.display_name || item.instance.name}
          </span>
          <Badge
            variant="outline"
            className="border-[var(--hl)] bg-transparent font-hud text-[11px] text-[var(--mu)]"
          >
            {item.instance.environment || "unspecified"}
          </Badge>
          <span className="font-hud text-[12px] text-[var(--mu)]">
            {agentsQuery.isLoading ? (
              "loading agents"
            ) : (
              <>
                {agents.length} agents · <InstanceSkillTotal agents={agents} /> skills
              </>
            )}
          </span>
          <div className="ml-auto flex items-center gap-3">
            <span
              className={cn(
                "font-hud text-[12px]",
                item.status === "ok" ? "text-[var(--mu)]" : styles.text
              )}
            >
              {updateLabel}
            </span>
            <Button
              variant="outline"
              size="sm"
              className="font-hud text-xs text-[var(--ac)] hover:border-[var(--ac)] hover:text-[var(--ac)]"
              disabled={syncMutation.isPending}
              onClick={(event) => {
                event.stopPropagation()
                setSyncOpen(true)
              }}
            >
              ↻ Re-import
            </Button>
          </div>
        </div>

        <CollapsibleContent>
          <div className="border-t border-[var(--sl)]">
            <AgentSubTable
              loading={agentsQuery.isLoading}
              agents={shownAgents}
              totalAgents={agents.length}
              onAgentClick={onAgentClick}
            />
            <InstanceCapabilityPanel instance={item.instance} />
          </div>
        </CollapsibleContent>
      </Card>

      <SyncDialog
        open={syncOpen}
        targets={[
          {
            id: item.instance.id,
            name: item.instance.display_name || item.instance.name,
          },
        ]}
        pending={syncMutation.isPending}
        defaultReason={`Re-import ${item.instance.display_name || item.instance.name}`}
        onOpenChange={setSyncOpen}
        onSubmit={(payload) => {
          syncMutation.mutate(payload, {
            onSuccess: (run) => {
              setSyncOpen(false)
              toast.success(syncSummary(run))
            },
          })
        }}
      />
    </Collapsible>
  )
}

function AgentSubTable({
  loading,
  agents,
  totalAgents,
  onAgentClick,
}: {
  loading: boolean
  agents: PersistedAgent[]
  totalAgents: number
  onAgentClick: (agent: PersistedAgent) => void
}) {
  return (
    <div>
      <Table className="table-fixed">
        <colgroup>
          <col className="w-[34%]" />
          <col className="w-[13%]" />
          <col className="w-[38%]" />
          <col className="w-[15%]" />
        </colgroup>
        <TableHeader>
          <TableRow className="border-[var(--sl)] hover:bg-transparent">
            <TableHead className="capcom-eyebrow h-9 px-[18px]">Agent</TableHead>
            <TableHead className="capcom-eyebrow h-9 px-[18px]">Skills</TableHead>
            <TableHead className="capcom-eyebrow h-9 px-[18px]">Can access</TableHead>
            <TableHead className="capcom-eyebrow h-9 px-[18px]">Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            <AgentSkeletonRows columns={4} />
          ) : agents.length ? (
            agents.map((agent) => (
              <AgentTableRow
                key={agent.id}
                agent={agent}
                onAgentClick={onAgentClick}
              />
            ))
          ) : (
            <TableRow className="border-[var(--sl)] hover:bg-transparent">
              <TableCell
                colSpan={4}
                className="px-[18px] py-8 text-center text-[13px] text-[var(--mu)]"
              >
                No imported agents. Run a sync for this runtime instance.
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
      {totalAgents > agents.length ? (
        <div className="border-t border-[var(--sl)] px-[18px] py-3 font-hud text-[12px] text-[var(--fa)]">
          Showing {agents.length} of {totalAgents} agents
        </div>
      ) : null}
    </div>
  )
}

function InstanceSkillTotal({ agents }: { agents: PersistedAgent[] }) {
  const skillQueries = useQueries({
    queries: agents.map((agent) => ({
      queryKey: queryKeys.agentSkills(agent.id),
      queryFn: () => capcomApi.listAgentSkills(agent.id),
      staleTime: 30_000,
    })),
  })
  const pending = skillQueries.some((query) => query.isLoading)
  const total = skillQueries.reduce(
    (sum, query) => sum + (query.data?.length ?? 0),
    0
  )

  if (!agents.length) {
    return <span>0</span>
  }

  return <span>{pending ? "..." : total}</span>
}

function InstanceCapabilityPanel({ instance }: { instance: RuntimeInstance }) {
  const testMutation = useTestRuntimeInstanceMutation(instance.id)
  const result = testMutation.data

  return (
    <div className="flex flex-col gap-3 border-t border-[var(--sl)] px-[18px] py-3 md:flex-row md:items-center md:justify-between">
      <div className="min-w-0">
        <div className="capcom-eyebrow">Test connection</div>
        <div className="mt-1 text-[12px] text-[var(--mu)]">
          {result?.message ?? "Capabilities pending"}
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <CapabilityChips result={result} />
        <Button
          variant="outline"
          size="sm"
          className="font-hud text-xs hover:border-[var(--ac)] hover:text-[var(--ac)]"
          disabled={testMutation.isPending}
          onClick={() => {
            testMutation.mutate(undefined, {
              onSuccess: (data) => {
                toast.success(`Connection test ${data.status}`)
              },
            })
          }}
        >
          {testMutation.isPending ? "Testing" : "Test connection"}
        </Button>
      </div>
    </div>
  )
}

function CapabilityChips({
  result,
}: {
  result?: RuntimeConnectionTestResult
}) {
  if (!result) {
    return (
      <Badge
        variant="outline"
        className="border-[var(--hl)] bg-transparent font-hud text-[11px] text-[var(--fa)]"
      >
        capabilities pending
      </Badge>
    )
  }

  return (
    <>
      {capabilityEntries(result.capabilities).map(([key, label, enabled]) => (
        <Badge
          key={key}
          variant="outline"
          className={cn(
            "border-[var(--hl)] bg-transparent font-hud text-[11px]",
            enabled
              ? "border-[color-mix(in_srgb,var(--ac)_35%,var(--hl))] bg-[var(--acd)] text-[var(--ac)]"
              : "text-[var(--fa)]"
          )}
        >
          {label}
        </Badge>
      ))}
    </>
  )
}

function capabilityEntries(capabilities: RuntimeCapabilities) {
  return [
    ["read_agents", "Read agents", capabilities.read_agents],
    ["read_agent_hierarchy", "Agent hierarchy", capabilities.read_agent_hierarchy],
    ["read_agent_skills", "Read skills", capabilities.read_agent_skills],
    ["read_agent_access", "Read access", capabilities.read_agent_access],
    ["replace_agent_access", "Replace access", capabilities.replace_agent_access],
    [
      "read_subagent_executions",
      "Subagent executions",
      Boolean(capabilities.read_subagent_executions),
    ],
  ] as const
}

function PageFooter({ adapter }: { adapter: AdapterModel }) {
  const first = adapter.instances[0]?.instance

  return (
    <p className="font-hud text-[12px] text-[var(--fa)]">
      Connection: {first?.endpoint ?? "none"} · adapter v
      {adapterVersion(adapter.instances.map((item) => item.instance))} · freshness budget 5m per instance
    </p>
  )
}

function AdapterDetailSkeleton() {
  return (
    <section className="flex flex-col gap-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-2">
          <Skeleton className="h-8 w-44" />
          <Skeleton className="h-5 w-[420px]" />
        </div>
        <Skeleton className="h-8 w-72" />
      </div>
      <Skeleton className="h-[260px] rounded-xl" />
      <Skeleton className="h-[74px] rounded-xl" />
    </section>
  )
}

function AgentSkeletonRows({ columns }: { columns: number }) {
  return (
    <>
      {[0, 1, 2].map((row) => (
        <TableRow key={row} className="border-[var(--sl)] hover:bg-transparent">
          {Array.from({ length: columns }).map((_, column) => (
            <TableCell key={column} className="px-[18px] py-3">
              <Skeleton className="h-5 w-full" />
            </TableCell>
          ))}
        </TableRow>
      ))}
    </>
  )
}

function EmptyState({
  title,
  description,
  action,
}: {
  title: string
  description: string
  action?: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-[var(--hl)] bg-[var(--el)] p-5 shadow-[var(--chi)]">
      <h1 className="text-[22px] font-bold leading-tight text-[var(--tx)]">
        {title}
      </h1>
      <p className="mt-1 text-[13px] text-[var(--mu)]">{description}</p>
      {action ? <div className="mt-4">{action}</div> : null}
    </section>
  )
}

function syncIntervalLabel(instances: AdapterInstance[]) {
  const intervals = new Set(
    instances.map((item) => item.instance.sync_interval_seconds)
  )
  if (intervals.size !== 1) {
    return "mixed intervals"
  }
  const seconds = instances[0]?.instance.sync_interval_seconds ?? 0
  if (seconds <= 0) {
    return "manual sync"
  }
  if (seconds < 60) {
    return `${seconds}s`
  }
  const minutes = Math.round(seconds / 60)
  return `${minutes}m`
}

function adapterVersion(instances: RuntimeInstance[]) {
  for (const instance of instances) {
    const version =
      instance.labels.adapter_version ??
      instance.labels.runtime_version ??
      instance.labels.version
    if (version) {
      return version
    }
  }
  return "unknown"
}

function syncSummary(run: RuntimeSyncRun) {
  const agents = run.agents_seen ?? 0
  const skills = run.skills_seen ?? 0
  return `Instance sync completed: ${agents} agents and ${skills} skills imported`
}

function runtimeTypeFromRoute(adapterId: string): RuntimeType | undefined {
  if (
    adapterId === "gantry" ||
    adapterId === "langgraph" ||
    adapterId === "temporal" ||
    adapterId === "letta" ||
    adapterId === "crewai"
  ) {
    return adapterId
  }
  return undefined
}
