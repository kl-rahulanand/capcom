"use client"

import * as React from "react"

import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  TableCell,
  TableRow,
} from "@/components/ui/table"
import type { PersistedAgent, RuntimeInstance } from "@/lib/api-types"
import {
  useAgentAccessQuery,
  useAgentSkillsQuery,
} from "@/lib/api-hooks"
import { displayRuntimeType } from "@/lib/adapters"
import { cn } from "@/lib/utils"

export type AgentLocation = {
  adapterName: string
  instanceName: string
}

type AgentTableRowProps = {
  agent: PersistedAgent
  location?: AgentLocation
  onAgentClick?: (agent: PersistedAgent) => void
}

export function AgentTableRow({
  agent,
  location,
  onAgentClick,
}: AgentTableRowProps) {
  return (
    <TableRow
      data-agent-id={agent.id}
      className="cursor-pointer border-[var(--sl)] hover:bg-[var(--sl)]"
      onClick={() => onAgentClick?.(agent)}
    >
      <TableCell className="min-w-[220px] px-[18px] py-3 whitespace-normal">
        <AgentIdentity agent={agent} />
      </TableCell>
      {location ? (
        <TableCell className="min-w-[220px] px-[18px] py-3 whitespace-normal">
          <div className="font-hud text-[12px] text-[var(--mu)]">
            {location.adapterName} · {location.instanceName}
          </div>
        </TableCell>
      ) : null}
      <TableCell className="px-[18px] py-3">
        <AgentSkillCount agentId={agent.id} />
      </TableCell>
      <TableCell className="min-w-[240px] px-[18px] py-3 whitespace-normal">
        <AgentAccessChips agentId={agent.id} />
      </TableCell>
      <TableCell className="px-[18px] py-3">
        <AgentStatusPill agent={agent} />
      </TableCell>
    </TableRow>
  )
}

export function AgentIdentity({ agent }: { agent: PersistedAgent }) {
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <span className="truncate font-hud text-[13px] font-medium text-[var(--tx)]">
        {agent.name}
      </span>
      <span className="truncate font-hud text-[11px] text-[var(--fa)]">
        {agent.runtime_agent_id}
      </span>
    </div>
  )
}

export function AgentSkillCount({ agentId }: { agentId: string }) {
  const skillsQuery = useAgentSkillsQuery(agentId)

  if (skillsQuery.isLoading) {
    return <Skeleton className="h-5 w-10" />
  }

  return (
    <span className="font-hud text-[13px] tabular text-[var(--tx)]">
      {skillsQuery.data?.length ?? 0}
    </span>
  )
}

export function AgentAccessChips({ agentId }: { agentId: string }) {
  const accessQuery = useAgentAccessQuery(agentId)
  const selections =
    accessQuery.data?.selections.filter((selection) => selection.allowed) ?? []
  const visible = selections.slice(0, 3)
  const hiddenCount = Math.max(0, selections.length - visible.length)

  if (accessQuery.isLoading) {
    return (
      <div className="flex flex-wrap gap-1.5">
        <Skeleton className="h-5 w-20" />
        <Skeleton className="h-5 w-16" />
      </div>
    )
  }

  if (!visible.length) {
    return <span className="font-hud text-[11px] text-[var(--fa)]">none resolved</span>
  }

  return (
    <div className="flex flex-wrap gap-1.5">
      {visible.map((selection) => (
        <Badge
          key={`${selection.kind}:${selection.id}`}
          variant="outline"
          title={selection.name || selection.id}
          className="border-[var(--hl)] bg-[var(--sl)] font-hud text-[11px] text-[var(--mu)]"
        >
          {selection.kind}:{selection.name || selection.id}
        </Badge>
      ))}
      {hiddenCount > 0 ? (
        <Badge
          variant="outline"
          className="border-[var(--hl)] bg-[var(--sl)] font-hud text-[11px] text-[var(--fa)]"
        >
          +{hiddenCount}
        </Badge>
      ) : null}
    </div>
  )
}

export function AgentStatusPill({ agent }: { agent: PersistedAgent }) {
  const status = agentStatus(agent)
  const className =
    status === "running"
      ? "bg-[var(--acd)] text-[var(--ac)]"
      : status === "failed"
        ? "bg-[var(--dgd)] text-[var(--dg)]"
        : "bg-[var(--sl)] text-[var(--fa)]"

  return (
    <Badge className={cn("font-hud text-[11px]", className)}>
      {status === "failed" ? "failed" : status}
    </Badge>
  )
}

export function locationForAgent(
  agent: PersistedAgent,
  instances: RuntimeInstance[]
): AgentLocation {
  const instance = instances.find(
    (item) => item.id === agent.runtime_connection_id
  )

  return {
    adapterName: instance
      ? displayRuntimeType(instance.runtime_type)
      : "Unknown",
    instanceName: instance?.display_name || instance?.name || "unknown",
  }
}

function agentStatus(agent: PersistedAgent) {
  const value = (agent.runtime_status || agent.status || "").toLowerCase()
  if (value.includes("fail") || value.includes("error")) {
    return "failed"
  }
  if (
    value.includes("running") ||
    value.includes("active") ||
    value.includes("healthy")
  ) {
    return "running"
  }
  return "idle"
}
