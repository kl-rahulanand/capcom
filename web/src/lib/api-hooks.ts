"use client"

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseQueryOptions,
} from "@tanstack/react-query"

import { capcomApi } from "@/lib/api-client"
import type {
  CreateRuntimeInstanceRequest,
  CreateSecretRequest,
  HealthResponse,
  PersistedAgent,
  RuntimeAgentAccess,
  RuntimeAgentSkill,
  RuntimeConnectionTestResult,
  RuntimeInstance,
  RuntimeSyncRun,
  ReconcileAccessRequest,
  SubagentExecution,
  SyncRuntimeRequest,
  ControlAction,
} from "@/lib/api-types"

export const queryKeys = {
  health: ["health"] as const,
  runtimeInstances: ["runtime-instances"] as const,
  runtimeInstance: (id: string) => ["runtime-instances", id] as const,
  runtimeInstanceTest: (id: string) =>
    ["runtime-instances", id, "test"] as const,
  runtimeInstanceSyncRuns: (id: string) =>
    ["runtime-instances", id, "sync-runs"] as const,
  runtimeInstanceAgents: (id: string) =>
    ["runtime-instances", id, "agents"] as const,
  runtimeInstanceSubagentExecutions: (id: string, agentId?: string) =>
    ["runtime-instances", id, "subagent-executions", agentId ?? "all"] as const,
  persistedAgents: (runtimeConnectionId?: string) =>
    ["agents", runtimeConnectionId ?? "fleet"] as const,
  persistedAgent: (id: string) => ["agents", id] as const,
  agentSkills: (id: string) => ["agents", id, "skills"] as const,
  agentAccess: (id: string) => ["agents", id, "access"] as const,
  subagentExecutions: (runtimeConnectionId?: string, agentId?: string) =>
    [
      "subagent-executions",
      runtimeConnectionId ?? "all",
      agentId ?? "all",
    ] as const,
}

export function useHealthQuery(
  options?: Omit<UseQueryOptions<HealthResponse>, "queryKey" | "queryFn">
) {
  return useQuery({
    queryKey: queryKeys.health,
    queryFn: capcomApi.health,
    refetchInterval: 30_000,
    ...options,
  })
}

export function useRuntimeInstancesQuery(enabled = true) {
  return useQuery<RuntimeInstance[]>({
    queryKey: queryKeys.runtimeInstances,
    queryFn: capcomApi.listRuntimeInstances,
    enabled,
  })
}

export function useRuntimeInstanceQuery(id: string | undefined) {
  return useQuery<RuntimeInstance>({
    queryKey: queryKeys.runtimeInstance(id ?? ""),
    queryFn: () => capcomApi.getRuntimeInstance(id ?? ""),
    enabled: Boolean(id),
  })
}

export function useTestRuntimeInstanceMutation(id: string) {
  return useMutation<RuntimeConnectionTestResult>({
    mutationKey: queryKeys.runtimeInstanceTest(id),
    mutationFn: () => capcomApi.testRuntimeInstance(id),
  })
}

export function usePersistedAgentsQuery(runtimeConnectionId?: string) {
  return useQuery<PersistedAgent[]>({
    queryKey: queryKeys.persistedAgents(runtimeConnectionId),
    queryFn: () => capcomApi.listPersistedAgents(runtimeConnectionId),
  })
}

export function useRuntimeInstanceAgentsQuery(id: string | undefined) {
  return useQuery<PersistedAgent[]>({
    queryKey: queryKeys.runtimeInstanceAgents(id ?? ""),
    queryFn: () => capcomApi.listRuntimeInstanceAgents(id ?? ""),
    enabled: Boolean(id),
  })
}

export function usePersistedAgentQuery(id: string | undefined) {
  return useQuery<PersistedAgent>({
    queryKey: queryKeys.persistedAgent(id ?? ""),
    queryFn: () => capcomApi.getPersistedAgent(id ?? ""),
    enabled: Boolean(id),
  })
}

export function useAgentSkillsQuery(id: string | undefined) {
  return useQuery<RuntimeAgentSkill[]>({
    queryKey: queryKeys.agentSkills(id ?? ""),
    queryFn: () => capcomApi.listAgentSkills(id ?? ""),
    enabled: Boolean(id),
  })
}

export function useAgentAccessQuery(id: string | undefined) {
  return useQuery<RuntimeAgentAccess>({
    queryKey: queryKeys.agentAccess(id ?? ""),
    queryFn: () => capcomApi.getAgentAccess(id ?? ""),
    enabled: Boolean(id),
  })
}

export function useReconcileAgentAccessMutation(id: string) {
  const queryClient = useQueryClient()
  return useMutation<ControlAction, Error, ReconcileAccessRequest>({
    mutationFn: (body) => capcomApi.reconcileAgentAccess(id, body),
    onSuccess: async (action) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.agentAccess(id) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.persistedAgent(id) }),
        queryClient.invalidateQueries({ queryKey: queryKeys.persistedAgents() }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.runtimeInstanceAgents(
            action.runtime_connection_id
          ),
        }),
      ])
    },
  })
}

export function useSubagentExecutionsQuery(params: {
  runtimeConnectionId?: string
  agentId?: string
}) {
  return useQuery<SubagentExecution[]>({
    queryKey: queryKeys.subagentExecutions(
      params.runtimeConnectionId,
      params.agentId
    ),
    queryFn: () => capcomApi.listSubagentExecutions(params),
  })
}

export function useRuntimeInstanceSubagentExecutionsQuery(
  id: string | undefined,
  agentId?: string
) {
  return useQuery<SubagentExecution[]>({
    queryKey: queryKeys.runtimeInstanceSubagentExecutions(id ?? "", agentId),
    queryFn: () => capcomApi.listRuntimeInstanceSubagentExecutions(id ?? "", agentId),
    enabled: Boolean(id),
  })
}

export function useRuntimeInstanceSyncRunsQuery(id: string | undefined) {
  return useQuery<RuntimeSyncRun[]>({
    queryKey: queryKeys.runtimeInstanceSyncRuns(id ?? ""),
    queryFn: () => capcomApi.listRuntimeInstanceSyncRuns(id ?? ""),
    enabled: Boolean(id),
  })
}

export function useSyncRuntimeInstanceMutation(id: string) {
  const queryClient = useQueryClient()
  return useMutation<RuntimeSyncRun, Error, SyncRuntimeRequest>({
    mutationFn: (body) => capcomApi.syncRuntimeInstance(id, body),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.runtimeInstances }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.runtimeInstance(id),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.runtimeInstanceAgents(id),
        }),
        queryClient.invalidateQueries({ queryKey: queryKeys.persistedAgents() }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.runtimeInstanceSyncRuns(id),
        }),
      ])
    },
  })
}

export type AddRuntimeInstanceInput = {
  secret: CreateSecretRequest
  runtimeInstance: CreateRuntimeInstanceRequest
}

export function useAddRuntimeInstanceMutation() {
  const queryClient = useQueryClient()

  return useMutation<RuntimeInstance, Error, AddRuntimeInstanceInput>({
    mutationFn: async ({ secret, runtimeInstance }) => {
      await capcomApi.createSecret(secret)
      return capcomApi.createRuntimeInstance(runtimeInstance)
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.runtimeInstances }),
        queryClient.invalidateQueries({ queryKey: queryKeys.persistedAgents() }),
      ])
      await Promise.all([
        queryClient.refetchQueries({ queryKey: queryKeys.runtimeInstances }),
        queryClient.refetchQueries({ queryKey: queryKeys.persistedAgents() }),
      ])
    },
  })
}
