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
  flowCount: number;
}

interface EdgeLink {
  id: string;
  from: string;
  to: string;
  fieldId: string;
  label: string;
  status: "healthy" | "warning" | "blocked";
}

interface Finding {
  id: string;
  severity: Severity;
  serviceId: string;
  fieldId: string;
  changeType: string;
  summary: string;
  remediation: string;
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
}

interface RunSummary {
  runCount: number;
  gate: GateDecision;
  mode: GateMode;
  findingCount: number;
  blockedCount: number;
  warningCount: number;
  generatedAt: string;
}

interface DemoSnapshot {
  services: ServiceNode[];
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
  reason: string;
  contacts?: Array<{ role: string; name?: string; channel?: string }>;
}

interface DemoPackShape {
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

function edgesWithHealthyStatus(): EdgeLink[] {
  return PACK.edges.map((edge) => ({
    ...edge,
    status: "healthy",
  }));
}

function defaultSnapshot(): DemoSnapshot {
  return {
    services: PACK.services,
    edges: edgesWithHealthyStatus(),
    findings: [],
    mutations: [],
    overrides: [],
    policy: {
      mode: "warn",
      failOn: "high",
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

function computeFindings(snapshot: DemoSnapshot): Finding[] {
  const rows: Finding[] = [];

  snapshot.mutations.forEach((mutation) => {
    const scenario = PACK.mutation_scenarios[mutation.fieldId]?.[mutation.type];
    if (!scenario) {
      return;
    }

    scenario.changes.forEach((change, idx) => {
      const metadata = PACK.field_metadata[change.field_id] || PACK.field_metadata[mutation.fieldId];
      rows.push({
        id: `${mutation.id}:${idx}`,
        severity: change.severity,
        serviceId: metadata?.serviceId || mutation.serviceId,
        fieldId: change.field_id,
        changeType: change.change_type,
        summary: change.message,
        remediation: remediationFor(change.change_type),
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

function runEngine(snapshot: DemoSnapshot): DemoSnapshot {
  const now = new Date();
  const allFindings = computeFindings(snapshot);
  const activeFindings = allFindings.filter((finding) => {
    return !snapshot.overrides.some((override) => isOverrideActive(override, finding, now));
  });

  const blockedCount = activeFindings.filter((finding) => finding.severity === "high").length;
  const warningCount = activeFindings.filter((finding) => finding.severity !== "high").length;
  const shouldBlock = snapshot.policy.mode === "block" && activeFindings.some((finding) => isBlocked(finding.severity, snapshot.policy));

  return {
    ...snapshot,
    findings: activeFindings,
    edges: computeEdgeStatus(edgesWithHealthyStatus(), activeFindings),
    runSummary: {
      runCount: snapshot.runSummary.runCount + 1,
      gate: shouldBlock ? "BLOCK" : "PASS",
      mode: snapshot.policy.mode,
      findingCount: activeFindings.length,
      blockedCount,
      warningCount,
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
      const serviceId = metadata?.serviceId || normalizeID(serviceIdRaw);
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

      if ((mode !== "warn" && mode !== "block") || !(failOn in severityRank)) {
        return textResponse("policy requires valid mode and failOn", 400);
      }

      const next = {
        ...snapshot,
        policy: {
          mode,
          failOn,
        },
      };
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
