WITH ranked AS (
    SELECT id,
           row_number() OVER (
               PARTITION BY runtime_connection_id, orchestrator_runtime_agent_id, delegate_ref
               ORDER BY updated_at DESC, created_at DESC, id DESC
           ) AS position
    FROM agent_delegations
)
DELETE FROM agent_delegations
WHERE id IN (SELECT id FROM ranked WHERE position > 1);

UPDATE agent_delegations
SET delegate_key = 'ref:' || delegate_ref,
    updated_at = now()
WHERE delegate_key <> 'ref:' || delegate_ref;
