import { DEMO_PACK } from "./generated/demo-pack";

export interface Env {
  DEMO_SESSIONS: DurableObjectNamespace;
  ASSETS: Fetcher;
}

type Severity = "info" | "low" | "medium" | "high";
type GateMode = "warn" | "block";
type GateDecision = "PASS" | "BLOCK";

type MutationType =
  | "type_changed"
  | "enum_changed"
  | "field_removed"
  | "source_version_changed"
  | "external_as_of_stale"
  | "annotation_missing";

interface ServiceNode {
  id: string;
  name: string;
  domain: string;
  kind: "internal" | "external";
  owner: string;
  escalation: string;
  runbookURL?: string;
  docRoot?: string;
  flowCount: number;
  flows?: string[];
}

interface FlowDefinition {
  id: string;
  name: string;
  level: number;
  owner?: string;
}

interface EdgeLink {
  id: string;
  from: string;
  to: string;
  fieldId: string;
  label: string;
  status: "healthy" | "warning" | "blocked";
}

interface FindingContextEdge {
  serviceId?: string;
  service?: string;
  api?: string;
}

interface FindingTypeDelta {
  change?: string;
  beforeContractRef?: string;
  afterContractRef?: string;
  beforeLabel?: string;
  afterLabel?: string;
}

interface Finding {
  id: string;
  severity: Severity;
  serviceId: string;
  fieldId: string;
  changeType: string;
  summary: string;
  remediation: string;
  source?: FindingContextEdge;
  impact?: FindingContextEdge;
  typeDelta?: FindingTypeDelta;
  modifiedBy?: string[];
  validation?: string;
  suggestion?: string;
  policyRationale?: string;
  timestamp: string;
}

interface Mutation {
  id: string;
  type: MutationType;
  serviceId: string;
  fieldId: string;
  timestamp: string;
}

interface Override {
  id: string;
  fieldId: string;
  changeType: string;
  expires: string;
  reason: string;
  ticket?: string;
  createdAt: string;
}

interface Policy {
  mode: GateMode;
  failOn: Severity;
  pack: {
    schema_version: number;
    policy_id: string;
    lineage: {
      findings: {
        require_downstream_impact: boolean;
        flow_criticality: {
          enabled: boolean;
          critical_flow_ids: string[];
          critical_flow_block_reason: string;
        };
      };
    };
  };
  coverage: {
    implementedFields: string[];
    supportedFields: string[];
  };
}

interface RunSummary {
  runCount: number;
  gate: GateDecision;
  mode: GateMode;
  findingCount: number;
  blockedCount: number;
  warningCount: number;
  policyRationale?: string;
  generatedAt: string;
}

interface DemoSnapshot {
  services: ServiceNode[];
  flows: FlowDefinition[];
  edges: EdgeLink[];
  findings: Finding[];
  mutations: Mutation[];
  overrides: Override[];
  policy: Policy;
  runSummary: RunSummary;
  truth: {
    supportedFlows: number;
    annotatedFlows: number;
    annotationCoveragePct: number;
    truthVersion: string;
    lineageChecksum: string;
  };
  mutationTypes: MutationType[];
  fieldsByMutation: Record<string, string[]>;
}

interface EscalationStep {
  depth: number;
  system_id: string;
  owner?: string;
  runbook_url?: string;
  doc_root?: string;
  reason: string;
  contacts?: Array<{ role: string; name?: string; channel?: string }>;
}

interface DemoPackShape {
  flows: FlowDefinition[];
  services: ServiceNode[];
  edges: Array<Omit<EdgeLink, "status">>;
  truth: DemoSnapshot["truth"];
  mutation_types: MutationType[];
  fields_by_mutation: Record<string, string[]>;
  field_metadata: Record<string, { serviceId: string; domain: string }>;
  mutation_scenarios: Record<
    string,
    Record<
      string,
      {
        changes: Array<{
          severity: Severity;
          change_type: string;
          field_id: string;
          message: string;
          source?: { service?: string; api?: string };
          impact?: { service?: string; api?: string };
          type_delta?: {
            change?: string;
            before_contract_ref?: string;
            after_contract_ref?: string;
            before_label?: string;
            after_label?: string;
          };
          modified_by?: string[];
          validation?: string;
          suggestion?: string;
        }>;
      }
    >
  >;
  escalation_by_system: Record<string, EscalationStep[]>;
}

const PACK = DEMO_PACK as unknown as DemoPackShape;

const severityRank: Record<Severity, number> = {
  info: 0,
  low: 1,
  medium: 2,
  high: 3,
};

const remediationByChangeType: Record<string, string> = {
  field_removed: "Reintroduce field or provide compatibility adapter before deploy.",
  source_version_changed: "Update min supported version and regenerate baseline.",
  external_as_of_rollback: "Refresh external source as-of and verify provider contract.",
  source_contract_ref_changed: "Update contract tests and reference docs with versioned rollout.",
  merge_strategy_changed: "Validate merge semantics with downstream consumers and tests.",
  source_removed: "Restore source mapping or annotate accepted migration override.",
};

function nowISO(): string {
  return new Date().toISOString();
}

function jsonResponse(payload: unknown, init?: ResponseInit): Response {
  const headers = new Headers(init?.headers);
  headers.set("content-type", "application/json; charset=utf-8");
  return new Response(JSON.stringify(payload, null, 2), { ...init, headers });
}

function textResponse(message: string, status = 400): Response {
  return new Response(message, { status, headers: { "content-type": "text/plain; charset=utf-8" } });
}

function parseDateDay(value: string): Date {
  return new Date(`${value}T23:59:59.000Z`);
}

function isOverrideActive(override: Override, finding: Finding, now: Date): boolean {
  if (override.fieldId !== finding.fieldId) {
    return false;
  }
  if (override.changeType !== "*" && override.changeType !== finding.changeType) {
    return false;
  }
  return parseDateDay(override.expires).getTime() >= now.getTime();
}

function isBlocked(severity: Severity, policy: Policy): boolean {
  return severityRank[severity] >= severityRank[policy.failOn];
}

function normalizeID(value: string): string {
  return value.trim().toLowerCase();
}

function resolveServiceID(token: string | undefined, services: ServiceNode[]): string {
  const value = normalizeID(String(token || "").replace(/[^a-zA-Z0-9]/g, ""));
  if (!value) {
    return "";
  }
  const match = services.find((service) => {
    const id = normalizeID(service.id.replace(/[^a-zA-Z0-9]/g, ""));
    const name = normalizeID(service.name.replace(/[^a-zA-Z0-9]/g, ""));
    return id === value || name === value;
  });
  return match?.id || "";
}

function edgesWithHealthyStatus(): EdgeLink[] {
  return PACK.edges.map((edge) => ({
    ...edge,
    status: "healthy",
  }));
}

function defaultSnapshot(): DemoSnapshot {
  return {
    services: PACK.services,
    flows: PACK.flows || [],
    edges: edgesWithHealthyStatus(),
    findings: [],
    mutations: [],
    overrides: [],
    policy: {
      mode: "warn",
      failOn: "high",
      pack: {
        schema_version: 1,
        policy_id: "production_standard",
        lineage: {
          findings: {
            require_downstream_impact: true,
            flow_criticality: {
              enabled: true,
              critical_flow_ids: ["checkout"],
              critical_flow_block_reason: "Risk of order loss",
            },
          },
        },
      },
      coverage: {
        implementedFields: [
          "lineage.findings.require_downstream_impact",
          "lineage.findings.flow_criticality.enabled",
          "lineage.findings.flow_criticality.critical_flow_ids",
          "lineage.findings.flow_criticality.critical_flow_block_reason",
        ],
        supportedFields: [
          "lineage.require.*",
          "lineage.defaults.*",
          "lineage.severity_overrides.*",
          "lineage.findings.require_downstream_impact",
          "lineage.findings.unknown_impact_severity",
          "lineage.findings.self_only.*",
          "lineage.findings.flow_criticality.*",
        ],
      },
    },
    runSummary: {
      runCount: 0,
      gate: "PASS",
      mode: "warn",
      findingCount: 0,
      blockedCount: 0,
      warningCount: 0,
      generatedAt: nowISO(),
    },
    truth: PACK.truth,
    mutationTypes: PACK.mutation_types,
    fieldsByMutation: PACK.fields_by_mutation,
  };
}

function compactSnapshot(snapshot: DemoSnapshot): DemoSnapshot {
  return {
    ...snapshot,
    overrides: snapshot.overrides.filter((item) => parseDateDay(item.expires).getTime() >= Date.now()),
  };
}

function remediationFor(changeType: string): string {
  return remediationByChangeType[changeType] || "Inspect diff details and update source annotations/contracts.";
}

function serviceFlows(snapshot: DemoSnapshot, serviceId: string): string[] {
  if (!serviceId) return [];
  const service = snapshot.services.find((row) => row.id === serviceId);
  return Array.isArray(service?.flows) ? service.flows : [];
}

function policyRationaleForFinding(snapshot: DemoSnapshot, finding: Finding): string | undefined {
  const policy = snapshot.policy;
  const flowPolicy = policy.pack.lineage.findings.flow_criticality;
  if (!flowPolicy.enabled) {
    return undefined;
  }
  if (finding.changeType !== "enum_changed") {
    return undefined;
  }
  const impactedID = finding.impact?.serviceId || finding.serviceId || "";
  const impactedFlows = serviceFlows(snapshot, impactedID).map(normalizeID);
  const criticalFlows = (flowPolicy.critical_flow_ids || []).map(normalizeID);
  const isCritical = impactedFlows.some((flow) => criticalFlows.includes(flow));
  if (!isCritical) {
    return undefined;
  }
  return flowPolicy.critical_flow_block_reason || "Risk of order loss";
}

function computeFindings(snapshot: DemoSnapshot): Finding[] {
  const rows: Finding[] = [];

  snapshot.mutations.forEach((mutation) => {
    const scenario = PACK.mutation_scenarios[mutation.fieldId]?.[mutation.type];
    if (!scenario) {
      return;
    }

    scenario.changes.forEach((change, idx) => {
      const metadata = PACK.field_metadata[change.field_id] || PACK.field_metadata[mutation.fieldId];
      const sourceServiceID = resolveServiceID(change.source?.service, snapshot.services);
      const impactServiceID = resolveServiceID(change.impact?.service, snapshot.services);
      const findingServiceID = impactServiceID || metadata?.serviceId || mutation.serviceId;
      rows.push({
        id: `${mutation.id}:${idx}`,
        severity: change.severity,
        serviceId: findingServiceID,
        fieldId: change.field_id,
        changeType: change.change_type,
        summary: change.message,
        remediation: remediationFor(change.change_type),
        source: change.source ? {
          serviceId: sourceServiceID || undefined,
          service: change.source.service,
          api: change.source.api,
        } : undefined,
        impact: change.impact ? {
          serviceId: impactServiceID || undefined,
          service: change.impact.service,
          api: change.impact.api,
        } : undefined,
        typeDelta: change.type_delta ? {
          change: change.type_delta.change,
          beforeContractRef: change.type_delta.before_contract_ref,
          afterContractRef: change.type_delta.after_contract_ref,
          beforeLabel: change.type_delta.before_label,
          afterLabel: change.type_delta.after_label,
        } : undefined,
        modifiedBy: Array.isArray(change.modified_by) ? change.modified_by : undefined,
        validation: change.validation,
        suggestion: change.suggestion,
        timestamp: mutation.timestamp,
      });
    });
  });

  rows.sort((left, right) => {
    const lr = severityRank[left.severity];
    const rr = severityRank[right.severity];
    if (lr !== rr) {
      return rr - lr;
    }
    if (left.fieldId !== right.fieldId) {
      return left.fieldId.localeCompare(right.fieldId);
    }
    return left.changeType.localeCompare(right.changeType);
  });

  return rows;
}

function computeEdgeStatus(edges: EdgeLink[], findings: Finding[]): EdgeLink[] {
  const byField: Record<string, "healthy" | "warning" | "blocked"> = {};

  findings.forEach((finding) => {
    const current = byField[finding.fieldId] || "healthy";
    const next = finding.severity === "high" ? "blocked" : "warning";
    if (current === "blocked") {
      return;
    }
    if (current === "warning" && next === "warning") {
      return;
    }
    byField[finding.fieldId] = next;
  });

  return edges.map((edge) => ({
    ...edge,
    status: byField[edge.fieldId] || "healthy",
  }));
}

function mutationIDFromFindingID(findingID: string): string {
  const sep = findingID.indexOf(":");
  if (sep <= 0) {
    return "";
  }
  return findingID.slice(0, sep);
}

function findingHasDownstreamImpact(snapshot: DemoSnapshot, finding: Finding): boolean {
  const mutationID = mutationIDFromFindingID(finding.id);
  const mutation = mutationID ? snapshot.mutations.find((row) => row.id === mutationID) : undefined;
  const sourceServiceID = finding.source?.serviceId || mutation?.serviceId || "";
  const impactServiceID = finding.impact?.serviceId || finding.serviceId || "";

  if (sourceServiceID && impactServiceID && sourceServiceID !== impactServiceID) {
    return true;
  }
  if (!sourceServiceID) {
    return true;
  }
  return snapshot.edges.some((edge) => {
    return edge.fieldId === finding.fieldId && edge.from === sourceServiceID && edge.to !== sourceServiceID;
  });
}

function runEngine(snapshot: DemoSnapshot): DemoSnapshot {
  const now = new Date();
  const allFindings = computeFindings(snapshot);
  const requireImpact = snapshot.policy.pack.lineage.findings.require_downstream_impact;
  const activeFindings = allFindings.filter((finding) => {
    if (snapshot.overrides.some((override) => isOverrideActive(override, finding, now))) {
      return false;
    }
    return requireImpact ? findingHasDownstreamImpact(snapshot, finding) : true;
  });
  const findingsWithPolicy = activeFindings.map((finding) => {
    const rationale = policyRationaleForFinding(snapshot, finding);
    if (!rationale) {
      return finding;
    }
    return {
      ...finding,
      policyRationale: rationale,
    };
  });

  const hardBlocked = findingsWithPolicy.filter((finding) => !!finding.policyRationale);
  const hardBlockedIDs = new Set(hardBlocked.map((finding) => finding.id));
  findingsWithPolicy
    .filter((finding) => finding.severity === "high")
    .forEach((finding) => hardBlockedIDs.add(finding.id));

  const blockedCount = hardBlockedIDs.size;
  const warningCount = Math.max(0, findingsWithPolicy.length - blockedCount);
  const shouldBlock = hardBlocked.length > 0
    || (snapshot.policy.mode === "block" && findingsWithPolicy.some((finding) => isBlocked(finding.severity, snapshot.policy)));
  const policyRationale = hardBlocked.length
    ? `Hard block policy: ${[...new Set(hardBlocked.map((finding) => finding.policyRationale))].join("; ")}.`
    : undefined;

  return {
    ...snapshot,
    findings: findingsWithPolicy,
    edges: computeEdgeStatus(edgesWithHealthyStatus(), findingsWithPolicy),
    runSummary: {
      runCount: snapshot.runSummary.runCount + 1,
      gate: shouldBlock ? "BLOCK" : "PASS",
      mode: snapshot.policy.mode,
      findingCount: findingsWithPolicy.length,
      blockedCount,
      warningCount,
      policyRationale,
      generatedAt: nowISO(),
    },
  };
}

function parseSessionPath(pathname: string): { sessionId: string; tail: string } | null {
  const match = /^\/api\/session\/([^/]+)(\/.*)?$/.exec(pathname);
  if (!match) {
    return null;
  }
  return {
    sessionId: match[1],
    tail: match[2] || "/",
  };
}

async function parseJSON(request: Request): Promise<Record<string, unknown>> {
  try {
    const body = (await request.json()) as Record<string, unknown>;
    return body;
  } catch {
    return {};
  }
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === "/api/session" && request.method === "POST") {
      const sessionId = crypto.randomUUID();
      const stub = env.DEMO_SESSIONS.get(env.DEMO_SESSIONS.idFromName(sessionId));
      const response = await stub.fetch(new Request("https://session/init", { method: "POST" }));
      if (!response.ok) {
        return textResponse("failed to initialize session", 500);
      }
      const initialized = (await response.json()) as { snapshot: DemoSnapshot };
      return jsonResponse({ sessionId, snapshot: initialized.snapshot });
    }

    const parsed = parseSessionPath(url.pathname);
    if (parsed) {
      const stub = env.DEMO_SESSIONS.get(env.DEMO_SESSIONS.idFromName(parsed.sessionId));
      const proxiedURL = new URL(request.url);
      proxiedURL.pathname = parsed.tail;
      return stub.fetch(new Request(proxiedURL.toString(), request));
    }

    return env.ASSETS.fetch(request);
  },
};

export class DemoSession {
  private readonly ctx: DurableObjectState;

  constructor(ctx: DurableObjectState, _env: Env) {
    this.ctx = ctx;
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    if (url.pathname === "/init" && request.method === "POST") {
      const initial = defaultSnapshot();
      await this.ctx.storage.put("snapshot", initial);
      return jsonResponse({ snapshot: initial });
    }

    const current = (await this.ctx.storage.get<DemoSnapshot>("snapshot")) || defaultSnapshot();
    const snapshot = compactSnapshot(current);

    if (url.pathname === "/snapshot" && request.method === "GET") {
      await this.ctx.storage.put("snapshot", snapshot);
      return jsonResponse({ snapshot });
    }

    if (url.pathname === "/mutations" && request.method === "POST") {
      const body = await parseJSON(request);
      const mutationType = String(body.type || "") as MutationType;
      const serviceIdRaw = String(body.serviceId || "");
      const fieldId = String(body.fieldId || "");

      const supportedFields = snapshot.fieldsByMutation[mutationType] || [];
      if (!snapshot.mutationTypes.includes(mutationType) || !supportedFields.includes(fieldId)) {
        return textResponse("mutation requires valid type and fieldId", 400);
      }

      const metadata = PACK.field_metadata[fieldId];
      const requestedServiceID = normalizeID(serviceIdRaw);
      const hasRequestedService = !!requestedServiceID && snapshot.services.some((service) => service.id === requestedServiceID);
      const serviceId = hasRequestedService ? requestedServiceID : (metadata?.serviceId || requestedServiceID);
      if (!serviceId) {
        return textResponse("mutation requires resolvable serviceId", 400);
      }

      const mutation: Mutation = {
        id: crypto.randomUUID(),
        type: mutationType,
        serviceId,
        fieldId,
        timestamp: nowISO(),
      };

      const next = {
        ...snapshot,
        mutations: [...snapshot.mutations, mutation],
      };
      await this.ctx.storage.put("snapshot", next);
      return jsonResponse({ snapshot: next });
    }

    if (url.pathname === "/run" && request.method === "POST") {
      const rerun = runEngine(snapshot);
      await this.ctx.storage.put("snapshot", rerun);
      return jsonResponse({ snapshot: rerun });
    }

    if (url.pathname === "/policy" && request.method === "POST") {
      const body = await parseJSON(request);
      const mode = String(body.mode || "") as GateMode;
      const failOn = String(body.failOn || "") as Severity;
      const requireImpact = Boolean(body.requireDownstreamImpact);
      const flowHardBlock = Boolean(body.flowHardBlock);
      const criticalFlowId = normalizeID(String(body.criticalFlowId || ""));
      const hardBlockReason = String(body.hardBlockReason || "").trim();

      if ((mode !== "warn" && mode !== "block") || !(failOn in severityRank)) {
        return textResponse("policy requires valid mode and failOn", 400);
      }
      if (!snapshot.flows.some((flow) => normalizeID(flow.id) === criticalFlowId)) {
        return textResponse("policy requires valid criticalFlowId", 400);
      }

      const updated = {
        ...snapshot,
        policy: {
          ...snapshot.policy,
          mode,
          failOn,
          pack: {
            ...snapshot.policy.pack,
            lineage: {
              ...snapshot.policy.pack.lineage,
              findings: {
                ...snapshot.policy.pack.lineage.findings,
                require_downstream_impact: requireImpact,
                flow_criticality: {
                  ...snapshot.policy.pack.lineage.findings.flow_criticality,
                  enabled: flowHardBlock,
                  critical_flow_ids: [criticalFlowId],
                  critical_flow_block_reason: hardBlockReason || "Risk of order loss",
                },
              },
            },
          },
        },
      };
      const next = runEngine(updated);
      await this.ctx.storage.put("snapshot", next);
      return jsonResponse({ snapshot: next });
    }

    if (url.pathname === "/override" && request.method === "POST") {
      const body = await parseJSON(request);
      const fieldId = String(body.fieldId || "").trim();
      const changeType = String(body.changeType || "*").trim();
      const expires = String(body.expires || "").trim();
      const reason = String(body.reason || "").trim();
      const ticket = String(body.ticket || "").trim();

      if (!fieldId || !expires || !reason) {
        return textResponse("override requires fieldId, expires, and reason", 400);
      }

      const entry: Override = {
        id: crypto.randomUUID(),
        fieldId,
        changeType,
        expires,
        reason,
        ticket: ticket || undefined,
        createdAt: nowISO(),
      };

      const next = {
        ...snapshot,
        overrides: [...snapshot.overrides, entry],
      };
      await this.ctx.storage.put("snapshot", next);
      return jsonResponse({ snapshot: next });
    }

    if (url.pathname === "/escalation" && request.method === "GET") {
      const serviceId = normalizeID(url.searchParams.get("serviceId") || "");
      if (!serviceId) {
        return textResponse("serviceId query parameter is required", 400);
      }
      return jsonResponse({ chain: PACK.escalation_by_system[serviceId] || [] });
    }

    return textResponse("not found", 404);
  }
}

export const __test = {
  defaultSnapshot,
  computeFindings,
  runEngine,
  findingHasDownstreamImpact,
};
