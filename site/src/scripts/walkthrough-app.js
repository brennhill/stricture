const scenarios = [
  {
    id: "logistics",
    label: "Logistics ETA Projection",
    domain: "logistics",
    summary: "Shipment ETA projection with external carrier enrichment.",
    endpoint: {
      method: "get",
      path: "/v1/shipments/{shipment_id}",
      operationId: "GetShipmentEtaProjection",
      responseField: "eta_projection",
      responseType: "object",
    },
    field: {
      fieldId: "response_logistics_eta_projection",
      fieldPath: "response.logistics.eta_projection",
      sourceSystem: "LogisticsGateway",
      sourceVersion: "v2026.06",
      minSupportedSourceVersion: "v2026.01",
      transformType: "normalize",
      mergeStrategy: "priority",
      breakPolicy: "strict",
      confidence: "declared",
      dataClassification: "internal",
      owner: "team.logistics",
      escalation: "pagerduty:logistics-oncall",
      contractTestId: "ci://contracts/logistics/eta_projection",
      introducedAt: "2026-01-11",
      flow: "from @RoutingCore enriched @self",
      note: "normalized by EtaNormalizer.Apply; spec=https://specs.example.com/logistics/eta_projection",
      sources: [
        {
          kind: "api",
          target: "RoutingCore.GetEtaProjection",
          path: "response.eta_projection",
          scope: "cross_repo",
          contractRef: "git+https://github.com/acme/routing-core//openapi.yaml@r06",
          upstreamSystem: "RoutingCore",
        },
        {
          kind: "api",
          target: "Fedex.GetEtaProjection",
          path: "response.eta_projection",
          scope: "external",
          asOf: "2026-02-14",
          providerId: "fedex",
          contractRef: "https://api.fedex.example.com/eta_projection",
          upstreamSystem: "Fedex",
        },
        {
          kind: "event",
          target: "LogisticsEvents.EtaProjectionChanged",
          path: "payload.eta_projection",
          scope: "internal",
          contractRef: "internal://events/logistics/eta_projection_changed",
          upstreamSystem: "LogisticsEvents",
        },
      ],
    },
    services: [
      { id: "LogisticsGateway", owner: "team.logistics" },
      { id: "RoutingCore", owner: "team.routing" },
      { id: "Fedex", owner: "external.provider" },
      { id: "LogisticsEvents", owner: "team.logistics" },
    ],
    steps: buildDefaultSteps({
      domainLabel: "Logistics ETA projection",
      sourceVersionFrom: "v2026.06",
      sourceVersionTo: "v2026.07",
      externalAsOfFrom: "2026-02-14",
      externalAsOfTo: "2025-11-02",
      fieldPathFrom: "response.logistics.eta_projection",
      fieldPathTo: "response.logistics.eta_estimate_projection",
      contractTestFrom: "ci://contracts/logistics/eta_projection",
      contractTestTo: "ci://contracts/logistics/eta_projection_v2",
      escalationOwner: "team.logistics",
      escalationChain: [
        { depth: 0, systemId: "LogisticsGateway", owner: "team.logistics", contact: "pagerduty:logistics-oncall", reason: "reported_bad_data" },
        { depth: 1, systemId: "RoutingCore", owner: "team.routing", contact: "slack:#routing-oncall", reason: "upstream_of:LogisticsGateway" },
        { depth: 1, systemId: "Fedex", owner: "external.provider", contact: "vendor:fedex-support", reason: "upstream_of:LogisticsGateway" },
      ],
    }),
  },
  {
    id: "fintech",
    label: "Fintech Authorization Decision",
    domain: "fintech",
    summary: "Payment authorization response with fraud and provider overlays.",
    endpoint: {
      method: "post",
      path: "/v1/payments/authorize",
      operationId: "AuthorizePayment",
      responseField: "authorization_decision",
      responseType: "object",
    },
    field: {
      fieldId: "response_fintech_authorization_decision",
      fieldPath: "response.fintech.authorization_decision",
      sourceSystem: "FintechGateway",
      sourceVersion: "v2026.04",
      minSupportedSourceVersion: "v2026.01",
      transformType: "aggregate",
      mergeStrategy: "priority",
      breakPolicy: "strict",
      confidence: "declared",
      dataClassification: "sensitive",
      owner: "team.fintech",
      escalation: "pagerduty:fintech-oncall",
      contractTestId: "ci://contracts/fintech/authorization_decision",
      introducedAt: "2026-01-19",
      flow: "from @LedgerCore enriched @self",
      note: "resolved in DecisionOrchestrator.Apply; spec=https://specs.example.com/fintech/authorization",
      sources: [
        {
          kind: "api",
          target: "LedgerCore.GetAuthorizationDecision",
          path: "response.authorization_decision",
          scope: "cross_repo",
          contractRef: "git+https://github.com/acme/ledger-core//openapi.yaml@r04",
          upstreamSystem: "LedgerCore",
        },
        {
          kind: "api",
          target: "Stripe.GetAuthorizationDecision",
          path: "response.authorization_decision",
          scope: "external",
          asOf: "2026-02-12",
          providerId: "stripe",
          contractRef: "https://api.stripe.example.com/authorization_decision",
          upstreamSystem: "Stripe",
        },
      ],
    },
    services: [
      { id: "FintechGateway", owner: "team.fintech" },
      { id: "LedgerCore", owner: "team.ledger" },
      { id: "Stripe", owner: "external.provider" },
    ],
    steps: buildDefaultSteps({
      domainLabel: "Fintech authorization decision",
      sourceVersionFrom: "v2026.04",
      sourceVersionTo: "v2026.05",
      externalAsOfFrom: "2026-02-12",
      externalAsOfTo: "2025-10-20",
      fieldPathFrom: "response.fintech.authorization_decision",
      fieldPathTo: "response.fintech.auth_decision",
      contractTestFrom: "ci://contracts/fintech/authorization_decision",
      contractTestTo: "ci://contracts/fintech/authorization_decision_v2",
      escalationOwner: "team.fintech",
      escalationChain: [
        { depth: 0, systemId: "FintechGateway", owner: "team.fintech", contact: "pagerduty:fintech-oncall", reason: "reported_bad_data" },
        { depth: 1, systemId: "LedgerCore", owner: "team.ledger", contact: "slack:#ledger-oncall", reason: "upstream_of:FintechGateway" },
        { depth: 1, systemId: "Stripe", owner: "external.provider", contact: "vendor:stripe-support", reason: "upstream_of:FintechGateway" },
      ],
    }),
  },
  {
    id: "media",
    label: "Media Metadata Unification",
    domain: "media",
    summary: "Track metadata stitched from internal and external providers.",
    endpoint: {
      method: "get",
      path: "/v1/tracks/{track_id}",
      operationId: "GetUnifiedTrack",
      responseField: "track",
      responseType: "object",
    },
    field: {
      fieldId: "response_media_track_metadata",
      fieldPath: "response.media.track_metadata",
      sourceSystem: "MediaGateway",
      sourceVersion: "v2026.08",
      minSupportedSourceVersion: "v2026.01",
      transformType: "join",
      mergeStrategy: "priority",
      breakPolicy: "additive_only",
      confidence: "declared",
      dataClassification: "internal",
      owner: "team.media",
      escalation: "pagerduty:media-oncall",
      contractTestId: "ci://contracts/media/track_metadata",
      introducedAt: "2026-01-22",
      flow: "from @CatalogCore enriched @self",
      note: "mapped in TrackMetadataMapper; spec=https://specs.example.com/media/track_metadata",
      sources: [
        {
          kind: "api",
          target: "CatalogCore.GetTrackMetadata",
          path: "response.track_metadata",
          scope: "cross_repo",
          contractRef: "git+https://github.com/acme/catalog-core//openapi.yaml@r08",
          upstreamSystem: "CatalogCore",
        },
        {
          kind: "api",
          target: "Spotify.GetTrack",
          path: "response.track",
          scope: "external",
          asOf: "2026-02-13",
          providerId: "spotify",
          contractRef: "https://developer.spotify.com/reference/get-track",
          upstreamSystem: "Spotify",
        },
      ],
    },
    services: [
      { id: "MediaGateway", owner: "team.media" },
      { id: "CatalogCore", owner: "team.catalog" },
      { id: "Spotify", owner: "external.provider" },
    ],
    steps: buildDefaultSteps({
      domainLabel: "Media metadata unification",
      sourceVersionFrom: "v2026.08",
      sourceVersionTo: "v2026.09",
      externalAsOfFrom: "2026-02-13",
      externalAsOfTo: "2025-12-01",
      fieldPathFrom: "response.media.track_metadata",
      fieldPathTo: "response.media.track_meta",
      contractTestFrom: "ci://contracts/media/track_metadata",
      contractTestTo: "ci://contracts/media/track_metadata_v2",
      escalationOwner: "team.media",
      escalationChain: [
        { depth: 0, systemId: "MediaGateway", owner: "team.media", contact: "pagerduty:media-oncall", reason: "reported_bad_data" },
        { depth: 1, systemId: "CatalogCore", owner: "team.catalog", contact: "slack:#catalog-oncall", reason: "upstream_of:MediaGateway" },
        { depth: 1, systemId: "Spotify", owner: "external.provider", contact: "vendor:spotify-support", reason: "upstream_of:MediaGateway" },
      ],
    }),
  },
  {
    id: "ecommerce",
    label: "Ecommerce Promotion Resolution",
    domain: "ecommerce",
    summary: "Promotion eligibility synthesized across internal and provider systems.",
    endpoint: {
      method: "post",
      path: "/v1/cart/promotion-resolution",
      operationId: "ResolvePromotionEligibility",
      responseField: "promotion_resolution",
      responseType: "object",
    },
    field: {
      fieldId: "response_ecommerce_promotion_resolution",
      fieldPath: "response.ecommerce.promotion_resolution",
      sourceSystem: "CommerceGateway",
      sourceVersion: "v2026.07",
      minSupportedSourceVersion: "v2026.01",
      transformType: "aggregate",
      mergeStrategy: "priority",
      breakPolicy: "strict",
      confidence: "declared",
      dataClassification: "internal",
      owner: "team.ecommerce",
      escalation: "pagerduty:ecommerce-oncall",
      contractTestId: "ci://contracts/ecommerce/promotion_resolution",
      introducedAt: "2026-01-25",
      flow: "from @PromotionCore enriched @self",
      note: "resolved in PromotionResolver.Apply; spec=https://specs.example.com/ecommerce/promotion_resolution",
      sources: [
        {
          kind: "api",
          target: "PromotionCore.GetPromotionResolution",
          path: "response.promotion_resolution",
          scope: "cross_repo",
          contractRef: "git+https://github.com/acme/promotion-core//openapi.yaml@r07",
          upstreamSystem: "PromotionCore",
        },
        {
          kind: "api",
          target: "Shopify.GetPromotionEligibilityResolution",
          path: "response.promotion_eligibility_resolution",
          scope: "external",
          asOf: "2026-02-16",
          providerId: "shopify",
          contractRef: "https://api.shopify.example.com/promotion_eligibility_resolution",
          upstreamSystem: "Shopify",
        },
      ],
    },
    services: [
      { id: "CommerceGateway", owner: "team.ecommerce" },
      { id: "PromotionCore", owner: "team.promotion" },
      { id: "Shopify", owner: "external.provider" },
    ],
    steps: buildDefaultSteps({
      domainLabel: "Ecommerce promotion resolution",
      sourceVersionFrom: "v2026.07",
      sourceVersionTo: "v2026.08",
      externalAsOfFrom: "2026-02-16",
      externalAsOfTo: "2025-09-28",
      fieldPathFrom: "response.ecommerce.promotion_resolution",
      fieldPathTo: "response.ecommerce.promotion_eligibility_resolution",
      contractTestFrom: "ci://contracts/ecommerce/promotion_resolution",
      contractTestTo: "ci://contracts/ecommerce/promotion_resolution_v2",
      escalationOwner: "team.ecommerce",
      escalationChain: [
        { depth: 0, systemId: "CommerceGateway", owner: "team.ecommerce", contact: "pagerduty:ecommerce-oncall", reason: "reported_bad_data" },
        { depth: 1, systemId: "PromotionCore", owner: "team.promotion", contact: "slack:#promotion-oncall", reason: "upstream_of:CommerceGateway" },
        { depth: 1, systemId: "Shopify", owner: "external.provider", contact: "vendor:shopify-support", reason: "upstream_of:CommerceGateway" },
      ],
    }),
  },
  {
    id: "governance",
    label: "Governance Disclosure Readiness",
    domain: "governance",
    summary: "Disclosure readiness signal with policy and external filing dependencies.",
    endpoint: {
      method: "get",
      path: "/v1/disclosures/readiness",
      operationId: "GetDisclosureReadiness",
      responseField: "disclosure_readiness",
      responseType: "object",
    },
    field: {
      fieldId: "response_governance_disclosure_readiness",
      fieldPath: "response.governance.disclosure_readiness",
      sourceSystem: "GovernanceHub",
      sourceVersion: "v2026.03",
      minSupportedSourceVersion: "v2026.01",
      transformType: "derive",
      mergeStrategy: "priority",
      breakPolicy: "strict",
      confidence: "declared",
      dataClassification: "regulated",
      owner: "team.governance",
      escalation: "pagerduty:governance-oncall",
      contractTestId: "ci://contracts/governance/disclosure_readiness",
      introducedAt: "2026-01-28",
      flow: "from @ControlCore validated @self",
      note: "computed in DisclosureReadinessEngine; spec=https://specs.example.com/governance/disclosure_readiness",
      sources: [
        {
          kind: "api",
          target: "ControlCore.GetDisclosureReadiness",
          path: "response.disclosure_readiness",
          scope: "cross_repo",
          contractRef: "git+https://github.com/acme/control-core//openapi.yaml@r03",
          upstreamSystem: "ControlCore",
        },
        {
          kind: "api",
          target: "SecFilings.GetDisclosureReadiness",
          path: "response.disclosure_readiness",
          scope: "external",
          asOf: "2026-02-11",
          providerId: "sec",
          contractRef: "https://api.sec.example.com/disclosure_readiness",
          upstreamSystem: "SecFilings",
        },
      ],
    },
    services: [
      { id: "GovernanceHub", owner: "team.governance" },
      { id: "ControlCore", owner: "team.controls" },
      { id: "SecFilings", owner: "external.provider" },
    ],
    steps: buildDefaultSteps({
      domainLabel: "Governance disclosure readiness",
      sourceVersionFrom: "v2026.03",
      sourceVersionTo: "v2026.04",
      externalAsOfFrom: "2026-02-11",
      externalAsOfTo: "2025-08-15",
      fieldPathFrom: "response.governance.disclosure_readiness",
      fieldPathTo: "response.governance.disclosure_ready_state",
      contractTestFrom: "ci://contracts/governance/disclosure_readiness",
      contractTestTo: "ci://contracts/governance/disclosure_readiness_v2",
      escalationOwner: "team.governance",
      escalationChain: [
        { depth: 0, systemId: "GovernanceHub", owner: "team.governance", contact: "pagerduty:governance-oncall", reason: "reported_bad_data" },
        { depth: 1, systemId: "ControlCore", owner: "team.controls", contact: "slack:#controls-oncall", reason: "upstream_of:GovernanceHub" },
        { depth: 1, systemId: "SecFilings", owner: "external.provider", contact: "vendor:sec-filings-desk", reason: "upstream_of:GovernanceHub" },
      ],
    }),
  },
];

const state = {
  scenarioId: "logistics",
  stepIndex: 0,
  profile: "stricture",
  sourceLanguage: "go",
  showAliases: false,
  changedOnly: false,
  sourceExpanded: false,
  autoplay: false,
  autoplayTimer: null,
};

const selectors = {
  runtimeStatus: document.querySelector("#wt-runtime-status"),
  gate: document.querySelector("#walkthrough-gate"),
  scenario: document.querySelector("#wt-scenario"),
  profile: document.querySelector("#wt-profile"),
  language: document.querySelector("#wt-language"),
  aliases: document.querySelector("#wt-aliases"),
  changedOnly: document.querySelector("#wt-changed-only"),
  sourceView: document.querySelector("#wt-source-view"),
  prev: document.querySelector("#wt-prev"),
  next: document.querySelector("#wt-next"),
  autoplay: document.querySelector("#wt-autoplay"),
  copyLink: document.querySelector("#wt-copy-link"),
  steps: document.querySelector("#wt-steps"),
  stepDescription: document.querySelector("#wt-step-description"),
  contract: document.querySelector("#wt-contract"),
  source: document.querySelector("#wt-source"),
  graph: document.querySelector("#wt-graph"),
  edges: document.querySelector("#wt-edges"),
  summary: document.querySelector("#wt-summary"),
  findings: document.querySelector("#wt-findings"),
  escalation: document.querySelector("#wt-escalation"),
  paneA: document.querySelector("#wt-pane-a-status"),
  paneB: document.querySelector("#wt-pane-b-status"),
  paneD: document.querySelector("#wt-pane-d-status"),
};

const requiredSelectorKeys = [
  "gate",
  "scenario",
  "profile",
  "language",
  "aliases",
  "changedOnly",
  "sourceView",
  "prev",
  "next",
  "autoplay",
  "copyLink",
  "steps",
  "stepDescription",
  "contract",
  "source",
  "graph",
  "edges",
  "summary",
  "findings",
  "escalation",
  "paneA",
  "paneB",
  "paneD",
];

function setRuntimeStatus(kind, message) {
  if (!selectors.runtimeStatus) {
    return;
  }
  selectors.runtimeStatus.classList.remove("walkthrough-status-loading", "walkthrough-status-ready", "walkthrough-status-error");
  selectors.runtimeStatus.classList.add(`walkthrough-status-${kind}`);
  selectors.runtimeStatus.textContent = message;
}

function buildDefaultSteps(params) {
  return [
    {
      id: "baseline",
      title: "Baseline Contract",
      description: `${params.domainLabel}: all contracts and annotations align. Gate remains green.`,
      status: "ok",
      patch: {},
      changes: [],
      findings: [],
      escalationChain: [],
    },
    {
      id: "profile-view",
      title: "Profile Overlay",
      description: "Same canonical meaning, now shown with profile aliases for easier adoption.",
      status: "ok",
      patch: {},
      changes: [{ key: "profile_aliases", label: "Aliases exposed", from: "hidden", to: "visible" }],
      findings: [],
      escalationChain: [],
    },
    {
      id: "source-version-advance",
      title: "Source Version Advance",
      description: "Producer version advanced without baseline approval. Warn mode before hard block.",
      status: "warn",
      patch: { sourceVersion: params.sourceVersionTo },
      changes: [{ key: "source_version", label: "source_version", from: params.sourceVersionFrom, to: params.sourceVersionTo }],
      findings: [
        {
          severity: "medium",
          changeType: "source_version_changed",
          summary: `source_version changed from ${params.sourceVersionFrom} to ${params.sourceVersionTo}`,
          remediation: "Refresh baseline and validate downstream compatibility tests.",
        },
      ],
      escalationChain: [],
    },
    {
      id: "external-stale",
      title: "External As-Of Stale",
      description: "External provider data timestamp regressed below allowed freshness window.",
      status: "block",
      patch: { externalAsOf: params.externalAsOfTo },
      changes: [{ key: "external_as_of", label: "external as_of", from: params.externalAsOfFrom, to: params.externalAsOfTo }],
      findings: [
        {
          severity: "high",
          changeType: "external_as_of_stale",
          summary: `external as_of regressed from ${params.externalAsOfFrom} to ${params.externalAsOfTo}`,
          remediation: "Update provider as_of reference and verify current external contract snapshot.",
        },
      ],
      escalationChain: params.escalationChain,
    },
    {
      id: "field-path-drift",
      title: "Field Path Drift",
      description: "A field-path rename landed without explicit migration/override policy.",
      status: "block",
      patch: { fieldPath: params.fieldPathTo, contractTestId: params.contractTestTo },
      changes: [
        { key: "field", label: "field path", from: params.fieldPathFrom, to: params.fieldPathTo },
        { key: "contract_test_id", label: "contract_test_id", from: params.contractTestFrom, to: params.contractTestTo },
      ],
      findings: [
        {
          severity: "high",
          changeType: "field_path_changed",
          summary: `field path changed from ${params.fieldPathFrom} to ${params.fieldPathTo}`,
          remediation: "Use renamed_from migration metadata and coordinated dual-read rollout.",
        },
        {
          severity: "medium",
          changeType: "contract_test_id_changed",
          summary: "Contract test reference changed with no approved migration ticket.",
          remediation: "Attach approved contract suite evidence before release.",
        },
      ],
      escalationChain: params.escalationChain,
    },
    {
      id: "fix-and-green",
      title: "Fix Applied",
      description: "Field path, external freshness, and contract evidence restored. Gate returns green.",
      status: "ok",
      patch: {
        sourceVersion: params.sourceVersionFrom,
        externalAsOf: params.externalAsOfFrom,
        fieldPath: params.fieldPathFrom,
        contractTestId: params.contractTestFrom,
      },
      changes: [{ key: "gate", label: "gate", from: "block", to: "ok" }],
      findings: [],
      escalationChain: [],
    },
  ];
}

function getScenario() {
  return scenarios.find((entry) => entry.id === state.scenarioId) || scenarios[0];
}

function getStep(scenario) {
  return scenario.steps[Math.max(0, Math.min(state.stepIndex, scenario.steps.length - 1))];
}

function cloneDeep(value) {
  return JSON.parse(JSON.stringify(value));
}

function applyStepField(baseField, stepPatch) {
  const field = cloneDeep(baseField);
  if (!stepPatch) {
    return field;
  }
  if (stepPatch.sourceVersion) {
    field.sourceVersion = stepPatch.sourceVersion;
  }
  if (stepPatch.fieldPath) {
    field.fieldPath = stepPatch.fieldPath;
  }
  if (stepPatch.contractTestId) {
    field.contractTestId = stepPatch.contractTestId;
  }
  if (stepPatch.externalAsOf) {
    const external = field.sources.find((entry) => entry.scope === "external");
    if (external) {
      external.asOf = stepPatch.externalAsOf;
    }
  }
  return field;
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
}

function changedKeySet(step) {
  const keys = new Set();
  step.changes.forEach((entry) => keys.add(entry.key));
  return keys;
}

function formatValue(key, value, changed) {
  const safe = escapeHtml(value);
  if (!changed.has(key)) {
    return safe;
  }
  return `<span class="wt-changed">${safe}</span>`;
}

function profileAliasMap(profile, field) {
  const aliases = [];
  if (profile === "stricture") {
    return aliases;
  }
  aliases.push(["field_path", field.fieldPath]);
  aliases.push(["service_name", field.sourceSystem]);
  aliases.push(["service_version", field.sourceVersion]);
  aliases.push(["owner_team", field.owner]);
  if (profile === "otel") {
    aliases.push(["service.name", field.sourceSystem]);
    aliases.push(["service.version", field.sourceVersion]);
    aliases.push(["owner.team", field.owner]);
  }
  if (profile === "openlineage") {
    aliases.push(["openlineage_job_name", field.sourceSystem]);
    aliases.push(["openlineage.job.version", field.sourceVersion]);
  }
  if (profile === "openapi") {
    aliases.push(["openapi_field_path", field.fieldPath]);
    aliases.push(["x-strict-source", "embedded"]);
  }
  if (profile === "asyncapi") {
    aliases.push(["asyncapi_field_path", field.fieldPath]);
  }
  return aliases;
}

function buildContractPane(scenario, field, step) {
  const changed = changedKeySet(step);
  if (state.changedOnly) {
    if (step.changes.length === 0) {
      return `# No contract deltas in this step.\n# Baseline contract remains unchanged.`;
    }
    return `# Contract deltas\n${step.changes
      .map((entry) => `- ${entry.label}: ${entry.from} -> ${entry.to}`)
      .join("\n")}`;
  }

  const aliases = state.showAliases ? profileAliasMap(state.profile, field) : [];
  const lines = [
    "openapi: 3.1.0",
    "info:",
    `  title: ${scenario.label}`,
    `  version: ${formatValue("source_version", field.sourceVersion, changed)}`,
    "paths:",
    `  ${scenario.endpoint.path}:`,
    `    ${scenario.endpoint.method}:`,
    `      operationId: ${scenario.endpoint.operationId}`,
    "      responses:",
    '        "200":',
    "          content:",
    "            application/json:",
    "              schema:",
    "                type: object",
    "                properties:",
    `                  ${scenario.endpoint.responseField}:`,
    `                    type: ${scenario.endpoint.responseType}`,
    "                    x-strict-source:",
    `                      field_id: ${escapeHtml(field.fieldId)}`,
    `                      field: ${formatValue("field", field.fieldPath, changed)}`,
    `                      source_system: ${escapeHtml(field.sourceSystem)}`,
    `                      source_version: ${formatValue("source_version", field.sourceVersion, changed)}`,
    `                      contract_test_id: ${formatValue("contract_test_id", field.contractTestId, changed)}`,
  ];
  aliases.forEach((entry) => {
    lines.push(`                      ${entry[0]}: ${escapeHtml(entry[1])}`);
  });
  return lines.join("\n");
}

function sourceRefToRaw(ref) {
  const scopeSegment = ref.scope === "external" ? `${ref.scope}!${ref.asOf || ""}` : ref.scope;
  const query = [];
  query.push(`contract_ref=${ref.contractRef}`);
  if (ref.providerId) {
    query.push(`provider_id=${ref.providerId}`);
  }
  if (ref.upstreamSystem) {
    query.push(`upstream_system=${ref.upstreamSystem}`);
  }
  return `${ref.kind}:${ref.target}#${ref.path}@${scopeSegment}?${query.join("&")}`;
}

function sourceLanguageSpec(language) {
  switch (language) {
    case "typescript":
      return { path: "handlers/lineage-handler.ts", commentPrefix: "//" };
    case "javascript":
      return { path: "handlers/lineage-handler.js", commentPrefix: "//" };
    case "python":
      return { path: "handlers/lineage_handler.py", commentPrefix: "#" };
    case "java":
      return { path: "src/main/java/com/acme/LineageHandler.java", commentPrefix: "//" };
    case "go":
    default:
      return { path: "handlers/lineage_handler.go", commentPrefix: "//" };
  }
}

function renderAnnotationPair(entry, changed) {
  const key = `<span class="tok-key">${escapeHtml(entry.key)}</span>`;
  const baseClass = entry.quoted ? "tok-string" : "tok-value";
  let value = `<span class="${baseClass}">${escapeHtml(entry.value)}</span>`;
  if (changed.has(entry.key)) {
    value = `<span class="wt-changed">${value}</span>`;
  }
  if (entry.quoted) {
    return `${key}="${value}"`;
  }
  return `${key}=${value}`;
}

function sourceFunctionLines(language, commentPrefix) {
  const cmt = `<span class="tok-comment">${escapeHtml(commentPrefix)}</span>`;
  switch (language) {
    case "typescript":
    case "javascript":
      return [
        `<span class="tok-kw">function</span> <span class="tok-fn">mapResponse</span>() {`,
        `  ${cmt} mapping logic`,
        `}`,
      ];
    case "python":
      return [
        `<span class="tok-kw">def</span> <span class="tok-fn">map_response</span>():`,
        `    <span class="tok-kw">pass</span>`,
      ];
    case "java":
      return [
        `<span class="tok-kw">class</span> LineageHandler {`,
        `  <span class="tok-kw">void</span> <span class="tok-fn">mapResponse</span>() {}`,
        `}`,
      ];
    case "go":
    default:
      return [
        `<span class="tok-kw">func</span> <span class="tok-fn">mapResponse</span>() {`,
        `  ${cmt} mapping logic`,
        `}`,
      ];
  }
}

function buildSourcePane(field, step) {
  const changed = changedKeySet(step);
  const { path, commentPrefix } = sourceLanguageSpec(state.sourceLanguage);
  const sourceLines = field.sources.map((source) => sourceRefToRaw(source));
  if (state.changedOnly) {
    if (step.changes.length === 0) {
      return `${commentPrefix} No annotation deltas in this step.\n`;
    }
    return `${commentPrefix} Annotation deltas\n${step.changes
      .map((entry) => `${commentPrefix} ${entry.label}: ${entry.from} -> ${entry.to}`)
      .join("\n")}`;
  }

  const compactPairs = [
    { key: "field", value: field.fieldPath },
    { key: "source_system", value: field.sourceSystem },
    { key: "source_version", value: field.sourceVersion },
    { key: "sources", value: sourceLines.join(",") },
  ];
  const compactParts = compactPairs.map((entry) => renderAnnotationPair(entry, changed));
  const prefix = `<span class="tok-comment">${escapeHtml(commentPrefix)}</span>`;
  const compactLine = `${prefix} <span class="tok-comment">strict-source</span> ${compactParts.join(" ")}`;

  if (!state.sourceExpanded) {
    return [compactLine, ...sourceFunctionLines(state.sourceLanguage, commentPrefix)].join("\n");
  }

  const aliases = state.showAliases ? profileAliasMap(state.profile, field) : [];
  const canonicalPairs = [
    { key: "annotation_schema_version", value: "1" },
    { key: "field_id", value: field.fieldId },
    { key: "field", value: field.fieldPath },
    { key: "source_system", value: field.sourceSystem },
    { key: "source_version", value: field.sourceVersion },
    { key: "min_supported_source_version", value: field.minSupportedSourceVersion },
    { key: "transform", value: field.transformType },
    { key: "merge", value: field.mergeStrategy },
    { key: "break_policy", value: field.breakPolicy },
    { key: "confidence", value: field.confidence },
    { key: "data_classification", value: field.dataClassification },
    { key: "owner", value: field.owner },
    { key: "escalation", value: field.escalation },
    { key: "contract_test_id", value: field.contractTestId },
    { key: "introduced_at", value: field.introducedAt },
    { key: "flow", value: field.flow, quoted: true },
    { key: "note", value: field.note, quoted: true },
    { key: "sources", value: sourceLines.join(",") },
  ];
  aliases.forEach((entry) => {
    canonicalPairs.push({ key: entry[0], value: String(entry[1]) });
  });
  const canonicalLines = canonicalPairs.map((entry) => `${prefix}   ${renderAnnotationPair(entry, changed)}`);

  return [
    `${prefix} ${escapeHtml(path)}`,
    `${prefix} compact: ${compactParts.join(" ")}`,
    `${prefix} strict-source`,
    ...canonicalLines,
    ...sourceFunctionLines(state.sourceLanguage, commentPrefix),
  ].join("\n");
}

function buildGraphPane(scenario, step) {
  selectors.graph.innerHTML = "";
  const severityOrder = { high: 3, medium: 2, low: 1, info: 0 };
  let maxSeverity = 0;
  step.findings.forEach((finding) => {
    maxSeverity = Math.max(maxSeverity, severityOrder[finding.severity] || 0);
  });
  const ownerService = scenario.field.sourceSystem;
  scenario.services.forEach((service) => {
    const node = document.createElement("article");
    let statusClass = "node-ok";
    if (step.status === "warn" && service.id === ownerService) {
      statusClass = "node-warn";
    }
    if (step.status === "block" && service.id === ownerService) {
      statusClass = "node-block";
    }
    if (step.status === "block" && maxSeverity >= 3 && service.id !== ownerService && service.id === scenario.field.sources.find((s) => s.scope === "external")?.upstreamSystem) {
      statusClass = "node-block";
    }
    node.className = `node ${statusClass}`;
    node.innerHTML = `
      <h3>${escapeHtml(service.id)}</h3>
      <p>owner=${escapeHtml(service.owner)}</p>
      <span class="node-badge">${step.status.toUpperCase()}</span>
    `;
    selectors.graph.appendChild(node);
  });
}

function buildEdgesPane(field, step) {
  selectors.edges.innerHTML = "";
  const changed = changedKeySet(step);
  field.sources.forEach((source) => {
    const edge = document.createElement("article");
    let status = "healthy";
    if (step.status === "warn") {
      status = "warning";
    }
    if (step.status === "block") {
      status = source.scope === "external" || changed.has("field") ? "blocked" : "warning";
    }
    edge.className = `edge edge-${status}`;
    edge.textContent = `${source.upstreamSystem || source.target} -> ${field.sourceSystem} | ${source.kind}:${source.path}`;
    selectors.edges.appendChild(edge);
  });
}

function renderFindings(step) {
  selectors.findings.innerHTML = "";
  if (step.findings.length === 0) {
    const row = document.createElement("article");
    row.className = "list-item";
    row.textContent = "No drift findings for this step.";
    selectors.findings.appendChild(row);
    return;
  }
  step.findings.forEach((finding) => {
    const row = document.createElement("article");
    row.className = `list-item sev-${finding.severity}`;
    row.innerHTML = `
      <h3>${escapeHtml(finding.changeType)} (${escapeHtml(finding.severity)})</h3>
      <p>${escapeHtml(finding.summary)}</p>
      <p class="item-meta">Fix: ${escapeHtml(finding.remediation)}</p>
    `;
    selectors.findings.appendChild(row);
  });
}

function renderEscalation(step) {
  selectors.escalation.innerHTML = "";
  if (!step.escalationChain || step.escalationChain.length === 0) {
    const row = document.createElement("article");
    row.className = "list-item";
    row.textContent = "Escalation chain inactive for this step.";
    selectors.escalation.appendChild(row);
    return;
  }
  step.escalationChain.forEach((entry) => {
    const row = document.createElement("article");
    row.className = "list-item";
    row.innerHTML = `
      <h3>depth=${entry.depth} ${escapeHtml(entry.systemId)}</h3>
      <p class="item-meta">owner=${escapeHtml(entry.owner)} contact=${escapeHtml(entry.contact)}</p>
      <p>${escapeHtml(entry.reason)}</p>
    `;
    selectors.escalation.appendChild(row);
  });
}

function renderSummary(step) {
  const severityCount = { high: 0, medium: 0, low: 0, info: 0 };
  step.findings.forEach((finding) => {
    if (severityCount[finding.severity] !== undefined) {
      severityCount[finding.severity] += 1;
    }
  });
  selectors.summary.innerHTML = `
    <article class="summary-item"><h3>Gate</h3><p>${escapeHtml(step.status.toUpperCase())}</p></article>
    <article class="summary-item"><h3>Findings</h3><p>${step.findings.length}</p></article>
    <article class="summary-item"><h3>High</h3><p>${severityCount.high}</p></article>
    <article class="summary-item"><h3>Changed</h3><p>${step.changes.length}</p></article>
  `;
}

function renderStepRail(scenario, step) {
  selectors.steps.innerHTML = "";
  scenario.steps.forEach((entry, index) => {
    const button = document.createElement("button");
    button.className = "step-pill";
    if (index === state.stepIndex) {
      button.classList.add("step-pill-active");
      button.setAttribute("aria-current", "step");
    }
    button.textContent = `${index + 1}. ${entry.title}`;
    button.addEventListener("click", () => {
      state.stepIndex = index;
      stopAutoplay();
      render();
    });
    selectors.steps.appendChild(button);
  });
  selectors.stepDescription.textContent = step.description;
}

function renderGate(step) {
  selectors.gate.classList.remove("gate-ok", "gate-warn", "gate-block");
  const className = step.status === "block" ? "gate-block" : step.status === "warn" ? "gate-warn" : "gate-ok";
  selectors.gate.classList.add(className);
  selectors.gate.textContent = `Step ${state.stepIndex + 1} | ${step.status.toUpperCase()} | findings=${step.findings.length} | profile=${state.profile}`;
  selectors.paneD.textContent = `Gate: ${step.status}`;
  selectors.paneA.textContent = state.profile === "stricture" ? "OpenAPI" : `Profile: ${state.profile}`;
  const sourceMode = state.changedOnly ? "Changed Only" : state.sourceExpanded ? "Full" : "Compact";
  const sourceLabel = state.showAliases ? `${sourceMode} + Aliases` : sourceMode;
  selectors.paneB.textContent = `${sourceLabel} · ${state.sourceLanguage}`;
}

function updateUrl() {
  const url = new URL(window.location.href);
  url.searchParams.set("scenario", state.scenarioId);
  url.searchParams.set("step", String(state.stepIndex));
  url.searchParams.set("profile", state.profile);
  url.searchParams.set("lang", state.sourceLanguage);
  url.searchParams.set("aliases", state.showAliases ? "1" : "0");
  url.searchParams.set("changed", state.changedOnly ? "1" : "0");
  url.searchParams.set("source", state.sourceExpanded ? "full" : "compact");
  url.searchParams.set("autoplay", state.autoplay ? "1" : "0");
  history.replaceState({}, "", url);
}

function syncControls(scenario) {
  selectors.scenario.innerHTML = "";
  scenarios.forEach((entry) => {
    const option = document.createElement("option");
    option.value = entry.id;
    option.textContent = `${entry.label} (${entry.domain})`;
    selectors.scenario.appendChild(option);
  });
  selectors.scenario.value = scenario.id;
  selectors.profile.value = state.profile;
  selectors.language.value = state.sourceLanguage;
  selectors.aliases.checked = state.showAliases;
  selectors.changedOnly.checked = state.changedOnly;
  selectors.sourceView.textContent = state.sourceExpanded ? "Show Compact" : "Show Full";
  selectors.sourceView.disabled = state.changedOnly;
  selectors.autoplay.textContent = state.autoplay ? "Stop Auto" : "Auto Play";
}

function render() {
  const scenario = getScenario();
  if (state.stepIndex >= scenario.steps.length) {
    state.stepIndex = scenario.steps.length - 1;
  }
  const step = getStep(scenario);
  const field = applyStepField(scenario.field, step.patch);

  syncControls(scenario);
  renderGate(step);
  renderStepRail(scenario, step);
  selectors.contract.innerHTML = buildContractPane(scenario, field, step);
  selectors.source.classList.add("code-view-source");
  selectors.source.innerHTML = buildSourcePane(field, step);
  buildGraphPane(scenario, step);
  buildEdgesPane(field, step);
  renderSummary(step);
  renderFindings(step);
  renderEscalation(step);
  updateUrl();
}

function nextStep() {
  const scenario = getScenario();
  if (state.stepIndex < scenario.steps.length - 1) {
    state.stepIndex += 1;
  } else {
    state.stepIndex = 0;
  }
  render();
}

function previousStep() {
  const scenario = getScenario();
  if (state.stepIndex > 0) {
    state.stepIndex -= 1;
  } else {
    state.stepIndex = scenario.steps.length - 1;
  }
  render();
}

function stopAutoplay() {
  state.autoplay = false;
  if (state.autoplayTimer) {
    window.clearInterval(state.autoplayTimer);
    state.autoplayTimer = null;
  }
}

function toggleAutoplay() {
  if (state.autoplay) {
    stopAutoplay();
    render();
    return;
  }
  state.autoplay = true;
  state.autoplayTimer = window.setInterval(() => {
    nextStep();
  }, 2500);
  render();
}

async function copyShareLink() {
  const url = window.location.href;
  try {
    await navigator.clipboard.writeText(url);
    selectors.copyLink.textContent = "Copied";
    setTimeout(() => {
      selectors.copyLink.textContent = "Copy Share Link";
    }, 1200);
  } catch (error) {
    selectors.copyLink.textContent = "Copy Failed";
    setTimeout(() => {
      selectors.copyLink.textContent = "Copy Share Link";
    }, 1200);
  }
}

function bindEvents() {
  selectors.scenario.addEventListener("change", () => {
    state.scenarioId = selectors.scenario.value;
    state.stepIndex = 0;
    stopAutoplay();
    render();
  });
  selectors.profile.addEventListener("change", () => {
    state.profile = selectors.profile.value;
    render();
  });
  selectors.language.addEventListener("change", () => {
    state.sourceLanguage = selectors.language.value;
    render();
  });
  selectors.aliases.addEventListener("change", () => {
    state.showAliases = selectors.aliases.checked;
    render();
  });
  selectors.changedOnly.addEventListener("change", () => {
    state.changedOnly = selectors.changedOnly.checked;
    render();
  });
  selectors.sourceView.addEventListener("click", () => {
    state.sourceExpanded = !state.sourceExpanded;
    render();
  });
  selectors.prev.addEventListener("click", () => {
    stopAutoplay();
    previousStep();
  });
  selectors.next.addEventListener("click", () => {
    stopAutoplay();
    nextStep();
  });
  selectors.autoplay.addEventListener("click", () => {
    toggleAutoplay();
  });
  selectors.copyLink.addEventListener("click", () => {
    copyShareLink();
  });
}

function parseInitialStateFromUrl() {
  const url = new URL(window.location.href);
  const scenarioId = url.searchParams.get("scenario");
  if (scenarioId && scenarios.some((entry) => entry.id === scenarioId)) {
    state.scenarioId = scenarioId;
  }
  const scenario = getScenario();
  const stepRaw = Number.parseInt(url.searchParams.get("step") || "0", 10);
  if (Number.isInteger(stepRaw) && stepRaw >= 0 && stepRaw < scenario.steps.length) {
    state.stepIndex = stepRaw;
  }
  const profile = (url.searchParams.get("profile") || "").toLowerCase();
  if (["stricture", "openlineage", "otel", "openapi", "asyncapi"].includes(profile)) {
    state.profile = profile;
  }
  const lang = (url.searchParams.get("lang") || "").toLowerCase();
  if (["go", "typescript", "javascript", "python", "java"].includes(lang)) {
    state.sourceLanguage = lang;
  }
  state.showAliases = url.searchParams.get("aliases") === "1";
  state.changedOnly = url.searchParams.get("changed") === "1";
  const source = (url.searchParams.get("source") || "").toLowerCase();
  if (source === "full") {
    state.sourceExpanded = true;
  }
  if (url.searchParams.get("autoplay") === "1") {
    state.autoplay = true;
  }
}

function renderInitFailure(error) {
  const message = "Walkthrough failed to initialize. Reload and try again.";
  setRuntimeStatus("error", `${message} (${error.message})`);
  if (selectors.gate) {
    selectors.gate.classList.remove("gate-ok", "gate-warn");
    selectors.gate.classList.add("gate-block");
    selectors.gate.textContent = "Walkthrough unavailable";
  }
  if (selectors.contract) {
    selectors.contract.textContent = `Initialization error:\n${error.message}`;
  }
  if (selectors.source) {
    selectors.source.textContent = "Source annotation view unavailable until initialization succeeds.";
  }
}

function init() {
  try {
    const missingSelectors = requiredSelectorKeys.filter((key) => !selectors[key]);
    if (missingSelectors.length > 0) {
      throw new Error(`missing DOM nodes: ${missingSelectors.join(", ")}`);
    }
    setRuntimeStatus("loading", "Loading walkthrough data...");
    parseInitialStateFromUrl();
    bindEvents();
    render();
    setRuntimeStatus("ready", "Walkthrough ready.");
    window.setTimeout(() => {
      setRuntimeStatus("ready", "Use Scenario and Profile to explore each lifecycle step.");
    }, 1500);
    if (state.autoplay) {
      state.autoplayTimer = window.setInterval(() => {
        nextStep();
      }, 2500);
    }
  } catch (error) {
    console.error("[walkthrough] init failed", error);
    renderInitFailure(error instanceof Error ? error : new Error("unknown initialization error"));
  }
}

init();
