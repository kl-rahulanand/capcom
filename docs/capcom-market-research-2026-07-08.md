# Capcom Market Research - 2026-07-08

## Research Question

What solutions are similar to Capcom, and where is the differentiation?

## Short Answer

The market is crowded around AI observability, evals, tracing, prompt management, and gateways. It is less crowded around declarative desired state, drift detection against live agent runtime access, and safe runtime mutation across heterogeneous agents.

Capcom should position as:

> A declarative AgentOps control plane that owns approved agent desired state, compares it against live runtime state, and executes audited control actions through runtime adapters.

## Market Map

| Category | Products / Standards | What They Do | Gap Capcom Can Own |
|---|---|---|---|
| Agent observability | AgentOps, LangSmith, Langfuse, Phoenix, Datadog LLM Observability, New Relic AI Monitoring, Splunk AI Agent Monitoring | Trace LLM/agent calls, tool use, errors, cost, latency, evals | Usually observe behavior; do not own approved runtime desired state or reconcile access drift |
| AI gateways / guardrails | Portkey, Helicone, LiteLLM, Cloudflare AI Gateway, Humanloop guardrails | Route model calls, apply guardrails, track usage, enforce input/output checks | Mostly model/request-layer control, not full agent runtime registry and capability state |
| Identity / directory / agent registry | Microsoft Agent 365, AGNTCY, Entra-style agent identity, internal service catalogs | Register agents, identities, capabilities, relationships | Often registry/identity-first; not necessarily runtime adapter reconciliation |
| Eval platforms | LangSmith evals, Langfuse evals, Phoenix evals, Galileo | Measure quality, reliability, regressions | Evals do not replace operational desired-state governance |
| Security/runtime governance research | AgentSight, MI9, AgentBound, AI Trust OS | Runtime monitoring, behavioral governance, provenance, policy enforcement | Mostly research/emerging; Capcom can productize a simpler wedge |
| Agent frameworks | Gantry, LangGraph, CrewAI, AutoGen, OpenAI Agents SDK, Google ADK, Agno | Build and run agents | Framework-specific; enterprises will run several |

## Product Notes From Sources

### AgentOps

AgentOps describes itself as a platform for testing, debugging, and deploying AI agents and LLM apps, with integrations across agent frameworks such as CrewAI, AutoGen, LangChain, OpenAI Agents, and others. Its docs emphasize automatic tracing, sessions, spans, dashboard views, and framework/provider integrations.

Source: https://docs.agentops.ai/

Implication for Capcom:

- Strong competitor for developer observability.
- Weakness relative to Capcom: not positioned as an enterprise desired-state reconciler for runtime access.

### LangSmith

LangSmith Observability provides visibility from traces to production metrics, supports many frameworks/providers, and includes dashboards, alerts, automations, feedback, and online evaluations.

Source: https://docs.langchain.com/langsmith/observability

Implication for Capcom:

- Excellent trace/eval source.
- Better treated as a telemetry adapter or downstream export, not as the same category.

### Langfuse

Langfuse is open source and self-hostable. It covers observability, prompt management, evaluation, metrics, sessions, users, and agent graphs. It supports OpenTelemetry and many integrations.

Source: https://langfuse.com/docs

Implication for Capcom:

- Strong open-source AI engineering platform.
- Capcom should integrate/export rather than compete only on traces.

### Arize Phoenix

Phoenix focuses on AI observability and evaluation. It captures traces for model calls, retrieval, tool use, and custom logic, accepts traces over OpenTelemetry, and is built on OpenInference instrumentation.

Source: https://arize.com/docs/phoenix

Implication for Capcom:

- Useful telemetry/eval source.
- Capcom should use Phoenix/OpenInference-style traces as evidence for actual behavior and drift context.

### Datadog LLM Observability

Datadog has LLM Observability and broader AI/agentic platform links including Agent Directory, MCP Server, governance console, workflow automation, incidents, and APM/security adjacency.

Source: https://docs.datadoghq.com/llm_observability/

Implication for Capcom:

- Enterprise observability vendors can expand into agent registry and governance.
- Capcom must be sharper than dashboards: own desired state and runtime mutation.

### Portkey

Portkey positions guardrails as gateway-level controls over requests and responses, with actions such as deny, log, dataset creation, fallback, retry, and custom guardrails.

Source: https://portkey.ai/docs/product/guardrails

Implication for Capcom:

- Gateway/guardrails are complementary.
- Capcom can use a gateway as an enforcement adapter, but Capcom's core is fleet desired state and runtime reconciliation.

### AGNTCY

AGNTCY is an LF open-source effort for agent discovery, verification, secure communication, schema, identity, and observability/evaluation across frameworks and organizations.

Source: https://docs.agntcy.org/

Implication for Capcom:

- Important future interoperability layer.
- Capcom should map its Agent resource to AGNTCY/OASF where possible and treat AGNTCY directory as registry/discovery input.

### OpenTelemetry / OpenInference

OpenTelemetry maintains GenAI/MCP semantic conventions and Phoenix uses OpenInference on top of OpenTelemetry instrumentation.

Sources:

- https://opentelemetry.io/docs/specs/semconv/gen-ai/
- https://arize.com/docs/phoenix

Implication for Capcom:

- Capcom should not invent tracing semantics.
- Use OTel/OpenInference for behavior/event ingestion, then layer desired state and control on top.

### Microsoft Agent 365

Current reporting describes Microsoft Agent 365 as a control plane for enterprise agents with registry, Entra integration, access control, telemetry, dashboards, alerts, and third-party ecosystem support.

Sources:

- https://www.itpro.com/technology/artificial-intelligence/microsofts-new-agent-365-platform-is-a-one-stop-shop-for-deploying-securing-and-keeping-tabs-on-ai-agents
- https://www.theverge.com/news/822035/microsoft-agent-365-businesses-control-security

Implication for Capcom:

- This is the clearest large-platform validation of the category.
- Capcom cannot win by being "agent inventory for Microsoft shops."
- Capcom can win as a runtime-agnostic, developer/platform-engineering control plane with open adapters and GitOps-style desired state.

## Competitive Differentiation

Capcom should not say:

> We are an AI observability platform.

That invites comparison to LangSmith, Langfuse, Phoenix, AgentOps, Datadog, and Splunk.

Capcom should say:

> Observability tools tell you what happened. Capcom knows what was approved, detects what drifted, and safely changes runtime state.

## MVP Wedge

The strongest validation experiment remains:

1. Connect to Gantry.
2. Import agents and actual access.
3. Apply desired access state.
4. Detect capability drift.
5. Restrict access through Gantry.
6. Show immutable audit.

This is narrower than the observability platforms and more concrete than strategy-only governance.

## Strategic Risks

| Risk | Why It Matters | Response |
|---|---|---|
| Microsoft Agent 365 owns enterprise registry | They have Entra, M365, Defender, Purview distribution | Be runtime-agnostic and developer/platform-engineering-first |
| Observability vendors add agent inventory | Datadog/Splunk/New Relic already own dashboards and alerts | Own desired state, drift, and safe control actions |
| LangSmith/Langfuse/Phoenix become default agent trace layer | Developers already instrument there | Integrate with them instead of replacing them |
| Gateways own enforcement | Portkey/Helicone/LiteLLM sit in request path | Treat gateways as enforcement adapters |
| Agent frameworks add native governance | Gantry and others may add more controls | Capcom remains cross-runtime and policy/audit centered |

## Recommended Positioning

Capcom is:

- runtime-agnostic
- desired-state-first
- adapter-based
- enterprise governance oriented
- control-action audited
- telemetry-compatible
- GitOps/API/CLI/dashboard friendly

Capcom is not:

- a new agent framework
- a trace viewer
- a prompt management tool
- an eval-only tool
- a model gateway only
- a Microsoft-only registry

## Immediate Product Implications

- Keep Gantry as the proof runtime.
- Add OpenTelemetry/OpenInference ingestion early.
- Add export/import compatibility with LangSmith, Langfuse, and Phoenix rather than fighting them.
- Make "adapter control level" visible in the product.
- Treat MCP as a first-class access surface.
- Add identity metadata now, but defer full identity federation.
- Keep YAML desired state, but make dashboard apply/export equally important.

