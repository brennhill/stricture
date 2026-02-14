export type SupportLevel = "Yes" | "Partial" | "No";

export interface ComparisonMatrixRow {
  criterion: string;
  stricture: SupportLevel;
  tool: SupportLevel;
  note: string;
}

export interface ComparisonReference {
  label: string;
  href: string;
}

export interface ComparisonDiagram {
  title: string;
  description: string;
}

export interface ComparisonEntry {
  slug: string;
  tool: string;
  shortName: string;
  category: string;
  overlapScore: number;
  overlapSummary: string;
  oneLineDifference: string;
  bestAt: string[];
  overlapAreas: string[];
  strictureAdds: string[];
  toolAdds: string[];
  whenUseBoth: string[];
  whenUseToolOnly: string[];
  adoptionSteps: string[];
  overlaySnippet: string;
  matrix: ComparisonMatrixRow[];
  diagrams: ComparisonDiagram[];
  references: ComparisonReference[];
}

export const comparisons: ComparisonEntry[] = [
  {
    slug: "opentelemetry",
    tool: "OpenTelemetry",
    shortName: "OTel",
    category: "Observability Standard",
    overlapScore: 45,
    overlapSummary: "Moderate overlap around service identity and semantic field naming, but different primary purpose.",
    oneLineDifference: "OpenTelemetry explains runtime behavior; Stricture verifies contract and lineage correctness before and during deploy.",
    bestAt: [
      "Standardizing traces, metrics, and logs across languages and vendors",
      "Providing runtime context (span attributes, resource attributes, error telemetry)",
      "Powering SLOs, alerting, and production debugging"
    ],
    overlapAreas: [
      "Service/resource naming conventions",
      "Field naming and semantic conventions",
      "Cross-system identifiers that can enrich lineage evidence"
    ],
    strictureAdds: [
      "Field-level source provenance with multi-source annotations",
      "Spec drift gates (warn/block/override) in CI/CD",
      "Owner and escalation metadata tied to exact drifting fields",
      "Source system version and as-of drift checks"
    ],
    toolAdds: [
      "Deep runtime performance and reliability signals",
      "Mature collector/exporter ecosystem",
      "Broad APM/tooling integration"
    ],
    whenUseBoth: [
      "Use OpenTelemetry for runtime truth and Stricture for contract/lineage truth. Together, you know both what happened and why drift was allowed.",
      "Use Stricture overlay aliases so the same source annotations map to OTel-friendly keys without duplicate authoring.",
      "Keep OTel as system-of-record for telemetry pipelines, and Stricture as system-of-record for pre-deploy lineage and drift policy gates."
    ],
    whenUseToolOnly: [
      "You only need observability and do not require field provenance, cross-repo drift gating, or escalation ownership metadata.",
      "Your team can tolerate manual contract verification and only needs runtime diagnosis."
    ],
    adoptionSteps: [
      "Enable Stricture annotations on highest-risk response fields first.",
      "Map Stricture annotation keys to OTel semantic aliases using overlay profiles.",
      "Attach trace IDs and service IDs in Stricture reports for fast incident pivot.",
      "Gate deploys only on high-severity drift after one release of warn mode."
    ],
    overlaySnippet: `# Stricture annotation with OTel-aligned aliases\nx-stricture-lineage:\n  field: shipment.eta\n  from:\n    - system: CarrierGateway\n      version: 2026.01\n  aliases:\n    otel.attribute: shipment.eta\n    otel.resource.service.name: logistics-gateway`,
    matrix: [
      { criterion: "Runtime traces/metrics/logs", stricture: "Partial", tool: "Yes", note: "OTel is primary runtime telemetry stack." },
      { criterion: "Field-level API lineage", stricture: "Yes", tool: "No", note: "Stricture tracks source provenance per emitted field." },
      { criterion: "Cross-repo drift gate in CI", stricture: "Yes", tool: "No", note: "Stricture enforces warn/block/override policies." },
      { criterion: "Service identity conventions", stricture: "Partial", tool: "Yes", note: "Overlap exists via aliases and naming conventions." },
      { criterion: "Schema/version mismatch detection", stricture: "Yes", tool: "Partial", note: "OTel may observe effects, Stricture detects drift directly." },
      { criterion: "Escalation chain metadata", stricture: "Yes", tool: "No", note: "Stricture embeds contact routing in annotations." },
      { criterion: "External system as-of freshness", stricture: "Yes", tool: "No", note: "Stricture models stale provider metadata checks." },
      { criterion: "Vendor-neutral ecosystem", stricture: "Partial", tool: "Yes", note: "OTel has wider existing telemetry vendor support." }
    ],
    diagrams: [
      {
        title: "Runtime vs Contract Plane",
        description: "Two-layer diagram: OTel on runtime telemetry plane, Stricture on contract/lineage gate plane, linked by service IDs."
      },
      {
        title: "Incident Triage Path",
        description: "Flow from production alert (OTel) to field drift root cause and owner escalation (Stricture)."
      }
    ],
    references: [
      { label: "OpenTelemetry Documentation", href: "https://opentelemetry.io/docs/" },
      { label: "Stricture Overlay Overview", href: "/what-is-stricture" }
    ]
  },
  {
    slug: "openlineage",
    tool: "OpenLineage",
    shortName: "OpenLineage",
    category: "Lineage Standard",
    overlapScore: 70,
    overlapSummary: "High overlap in lineage intent; scope differs between job/dataset lineage and API field-level drift gating.",
    oneLineDifference: "OpenLineage tracks data-job lineage events; Stricture tracks API field provenance, policy gates, and escalation routing.",
    bestAt: [
      "Standard lineage events for jobs, runs, and datasets",
      "Warehouse and pipeline lineage integration",
      "Governance visibility across ETL/ELT systems"
    ],
    overlapAreas: [
      "Provenance concepts",
      "Source and destination graph modeling",
      "Change impact analysis"
    ],
    strictureAdds: [
      "Field-by-field API response lineage",
      "Release-time contract drift detection and blocking",
      "Owner/escalation paths on API-facing fields",
      "Cross-service source-system version compatibility checks"
    ],
    toolAdds: [
      "Job/run event model for data platform operations",
      "Dataset-centric ecosystem adoption",
      "Marquez-compatible lineage backends"
    ],
    whenUseBoth: [
      "Use OpenLineage to represent job/run/dataset movement and Stricture to represent API field provenance and release risk.",
      "Use overlays so shared metadata (system IDs, dataset names, versions) is authored once and emitted into both profiles.",
      "Keep OpenLineage as system-of-record for pipeline lineage and Stricture as system-of-record for contract drift policy and escalation."
    ],
    whenUseToolOnly: [
      "Your use case is primarily warehouse/pipeline lineage and not API contract enforcement.",
      "You do not need field-level API drift gates in CI/CD."
    ],
    adoptionSteps: [
      "Map existing OpenLineage namespace/name pairs to Stricture source-system IDs.",
      "Enable Stricture on externally consumed API fields first.",
      "Export Stricture lineage with the openlineage profile for downstream reuse.",
      "Turn on block mode for high-severity source-version drift after baselining."
    ],
    overlaySnippet: `# One annotation, two profiles\nx-stricture-lineage:\n  field: payout.amount\n  from:\n    - system: TreasuryLedger\n      dataset: ledger_entries\n      version: 2026.02\n  aliases:\n    openlineage.namespace: finance\n    openlineage.name: ledger_entries`,
    matrix: [
      { criterion: "Dataset/job lineage", stricture: "Partial", tool: "Yes", note: "OpenLineage is native for run/job facets." },
      { criterion: "Field-level API lineage", stricture: "Yes", tool: "Partial", note: "OpenLineage can model related info but not API-first by default." },
      { criterion: "CI drift gate", stricture: "Yes", tool: "No", note: "Stricture enforces deploy policy outcomes." },
      { criterion: "Override workflow", stricture: "Yes", tool: "No", note: "Stricture supports expiring override metadata." },
      { criterion: "Escalation contact routing", stricture: "Yes", tool: "No", note: "Stricture directly maps field failures to responders." },
      { criterion: "Pipeline ecosystem integrations", stricture: "Partial", tool: "Yes", note: "OpenLineage has broader lineage backend integrations today." },
      { criterion: "External API as-of checks", stricture: "Yes", tool: "No", note: "Stricture models freshness and provider version drift." },
      { criterion: "Cross-repo API contract coverage", stricture: "Yes", tool: "Partial", note: "Stricture is designed for multi-service API verification." }
    ],
    diagrams: [
      {
        title: "Job Lineage + API Lineage Overlay",
        description: "Top: OpenLineage dataset/job graph. Bottom: Stricture field graph. Shared IDs bridge both layers."
      },
      {
        title: "Deploy Gate Flow",
        description: "Show OpenLineage event ingestion and Stricture release gate decision path with warning/block outcomes."
      }
    ],
    references: [
      { label: "OpenLineage Documentation", href: "https://openlineage.io/docs/" },
      { label: "Marquez Project", href: "https://marquezproject.ai/" }
    ]
  },
  {
    slug: "openapi",
    tool: "OpenAPI",
    shortName: "OpenAPI",
    category: "API Contract Standard",
    overlapScore: 75,
    overlapSummary: "High overlap around API schema definitions; Stricture extends into lineage, ownership, and drift policy.",
    oneLineDifference: "OpenAPI defines interface shape; Stricture verifies provenance, source versions, and cross-system drift behavior.",
    bestAt: [
      "Contract-first API design",
      "Client/server code generation",
      "Schema documentation and tool interoperability"
    ],
    overlapAreas: [
      "Field definitions and schema semantics",
      "API operation identity",
      "Contract compatibility concerns"
    ],
    strictureAdds: [
      "Source-system provenance per response field",
      "Policy gating with severity thresholds",
      "Owner and escalation metadata on data contracts",
      "Cross-repo and cross-system drift audits"
    ],
    toolAdds: [
      "Large ecosystem of codegen, docs, and testing tools",
      "Broad developer familiarity",
      "Formal request/response modeling"
    ],
    whenUseBoth: [
      "Use OpenAPI as the contract source and Stricture as the enforceable provenance and drift guardrail.",
      "Use overlays so `x-stricture-*` metadata and OpenAPI extensions stay synchronized without writing duplicate annotations.",
      "Keep OpenAPI as system-of-record for interface shape and Stricture as system-of-record for data origin, ownership, and release gates."
    ],
    whenUseToolOnly: [
      "You only need contract docs/codegen and can tolerate manual drift review.",
      "No cross-service lineage obligations or compliance-driven provenance requirements."
    ],
    adoptionSteps: [
      "Add `x-stricture-lineage` to highest-risk OpenAPI response fields.",
      "Run Stricture in warn mode against your OpenAPI baseline.",
      "Attach escalation metadata for critical fields and provider dependencies.",
      "Promote high-severity checks to block mode after one stable release cycle."
    ],
    overlaySnippet: `components:\n  schemas:\n    PaymentResponse:\n      type: object\n      properties:\n        status:\n          type: string\n          x-stricture-lineage:\n            from:\n              - system: PaymentsCore\n                version: 2026.02\n            aliases:\n              openapi.extension: x-source-system`,
    matrix: [
      { criterion: "Schema/interface modeling", stricture: "Partial", tool: "Yes", note: "OpenAPI is primary contract format." },
      { criterion: "Field provenance metadata", stricture: "Yes", tool: "Partial", note: "OpenAPI can carry extensions, Stricture defines semantics and checks." },
      { criterion: "Automated drift gate", stricture: "Yes", tool: "No", note: "Stricture evaluates and enforces drift policies." },
      { criterion: "Code generation", stricture: "No", tool: "Yes", note: "OpenAPI ecosystems excel at SDK/server generation." },
      { criterion: "Escalation ownership metadata", stricture: "Yes", tool: "Partial", note: "OpenAPI extensions possible but no built-in enforcement model." },
      { criterion: "Cross-repo contract reconciliation", stricture: "Yes", tool: "Partial", note: "Stricture compares emitted artifacts across services." },
      { criterion: "Override controls with expiry", stricture: "Yes", tool: "No", note: "Stricture has policy override model." },
      { criterion: "Developer tooling breadth", stricture: "Partial", tool: "Yes", note: "OpenAPI has larger mature tool ecosystem." }
    ],
    diagrams: [
      {
        title: "OpenAPI Contract + Stricture Policy Gate",
        description: "Design-time OpenAPI spec feeds CI; Stricture evaluates lineage drift before deployment."
      },
      {
        title: "Field Card Expansion",
        description: "Single OpenAPI field expanded into provenance sources, transform note, owner, escalation chain."
      }
    ],
    references: [
      { label: "OpenAPI Initiative", href: "https://www.openapis.org/" },
      { label: "Swagger/OpenAPI Specification", href: "https://swagger.io/specification/" }
    ]
  },
  {
    slug: "asyncapi",
    tool: "AsyncAPI",
    shortName: "AsyncAPI",
    category: "Event Contract Standard",
    overlapScore: 60,
    overlapSummary: "Moderate-high overlap for message schema definitions; Stricture adds provenance and drift governance.",
    oneLineDifference: "AsyncAPI defines event interfaces; Stricture verifies event payload lineage and policy compliance over time.",
    bestAt: [
      "Documenting event-driven contracts",
      "Broker/channel/message schema modeling",
      "Async ecosystem tooling"
    ],
    overlapAreas: [
      "Field schema and message contract definitions",
      "Service/channel dependency awareness",
      "Breaking change concerns"
    ],
    strictureAdds: [
      "Field-level source and enrichment annotations on event payloads",
      "As-of/version freshness checks for external providers",
      "Policy-based warn/block deploy gating",
      "Escalation routing tied to event fields"
    ],
    toolAdds: [
      "Native async channel and protocol modeling",
      "Message contract docs and generator tooling",
      "Event domain communication standards"
    ],
    whenUseBoth: [
      "Use AsyncAPI as canonical async contract format and Stricture for provenance/governance enforcement.",
      "Use overlays so payload metadata serves both AsyncAPI extensions and Stricture checks without redundant annotations.",
      "Keep AsyncAPI as system-of-record for channel/message structure and Stricture for lineage drift risk and escalation outcomes."
    ],
    whenUseToolOnly: [
      "You only need event contract documentation and not lineage policy gates.",
      "Event consumers are internal and can handle drift manually."
    ],
    adoptionSteps: [
      "Annotate top critical event payload fields with source provenance.",
      "Map channel IDs to Stricture source-system IDs.",
      "Run warn mode for one release cycle across event producers.",
      "Enable block mode for high-severity producer/consumer drift."
    ],
    overlaySnippet: `components:\n  messages:\n    ShipmentUpdated:\n      payload:\n        type: object\n        properties:\n          eta:\n            type: string\n            x-stricture-lineage:\n              from:\n                - system: CarrierGateway\n                  version: 2026.01\n              aliases:\n                asyncapi.extension: x-origin-system`,
    matrix: [
      { criterion: "Async protocol/channel modeling", stricture: "No", tool: "Yes", note: "AsyncAPI is primary async contract model." },
      { criterion: "Field-level lineage", stricture: "Yes", tool: "Partial", note: "AsyncAPI extensions possible; Stricture defines semantics/enforcement." },
      { criterion: "Release drift gate", stricture: "Yes", tool: "No", note: "Stricture enforces drift policy outcomes." },
      { criterion: "Escalation metadata", stricture: "Yes", tool: "No", note: "Stricture links failures to owner chains." },
      { criterion: "Schema docs ecosystem", stricture: "Partial", tool: "Yes", note: "AsyncAPI has broader async docs/generator support." },
      { criterion: "External provider freshness checks", stricture: "Yes", tool: "No", note: "Stricture tracks as-of/version assertions." },
      { criterion: "Multi-repo enforcement", stricture: "Yes", tool: "Partial", note: "Stricture compares producer/consumer expectations." },
      { criterion: "Override policy controls", stricture: "Yes", tool: "No", note: "Stricture supports expiring governance overrides." }
    ],
    diagrams: [
      {
        title: "Event Contract and Lineage Overlay",
        description: "AsyncAPI channel/message definitions with Stricture lineage cards attached to payload fields."
      },
      {
        title: "Producer-Consumer Drift Gate",
        description: "Producer change triggers Stricture check before deployment to prevent consumer breakage."
      }
    ],
    references: [
      { label: "AsyncAPI Docs", href: "https://www.asyncapi.com/docs" }
    ]
  },
  {
    slug: "spectral",
    tool: "Spectral",
    shortName: "Spectral",
    category: "Linting Engine",
    overlapScore: 35,
    overlapSummary: "Low-moderate overlap in linting behavior; Stricture focuses on lineage-aware drift policies across systems.",
    oneLineDifference: "Spectral lints static API documents; Stricture lints and validates lineage, ownership, and cross-system drift semantics.",
    bestAt: [
      "Fast rule-based linting for OpenAPI/AsyncAPI",
      "Custom style and governance rule packs",
      "Simple CI integration for spec hygiene"
    ],
    overlapAreas: [
      "Policy rule evaluation",
      "Spec quality checks",
      "CI integration"
    ],
    strictureAdds: [
      "Runtime-aware lineage validation",
      "Source-system version/as-of drift checks",
      "Escalation contact routing and override lifecycle",
      "Cross-repo multi-service reconciliation"
    ],
    toolAdds: [
      "Mature static linting rule ecosystem",
      "Good authoring feedback loops for schema docs",
      "Low-friction adoption for spec conventions"
    ],
    whenUseBoth: [
      "Use Spectral for fast static style/consistency linting and Stricture for deeper cross-system lineage drift enforcement.",
      "Use overlays so field metadata you add for Stricture can also satisfy Spectral extension rules, avoiding duplicate policy work.",
      "Keep Spectral as system-of-record for style conventions and Stricture as system-of-record for provenance and release-risk gates."
    ],
    whenUseToolOnly: [
      "You only need static document linting and not source provenance or operational escalation workflows.",
      "No cross-repo dependency drift checks are required."
    ],
    adoptionSteps: [
      "Continue existing Spectral rules for style and consistency.",
      "Add Stricture checks for field provenance on critical payloads.",
      "Align Spectral extension rules with Stricture overlay keys.",
      "Use Spectral as pre-commit and Stricture as release gate."
    ],
    overlaySnippet: `extends: ["spectral:oas"]\nrules:\n  stricture-lineage-required:\n    given: "$.components.schemas.*.properties[*]"\n    then:\n      field: "x-stricture-lineage"\n      function: truthy`,
    matrix: [
      { criterion: "Static spec linting", stricture: "Partial", tool: "Yes", note: "Spectral is stronger static lint engine." },
      { criterion: "Field provenance semantics", stricture: "Yes", tool: "No", note: "Stricture defines and validates lineage annotations." },
      { criterion: "Cross-system drift detection", stricture: "Yes", tool: "No", note: "Stricture inspects multi-service lineage and version drift." },
      { criterion: "Override + expiry workflow", stricture: "Yes", tool: "No", note: "Stricture tracks governance override lifecycle." },
      { criterion: "Escalation chain output", stricture: "Yes", tool: "No", note: "Stricture ties findings to responders." },
      { criterion: "Rule ecosystem maturity", stricture: "Partial", tool: "Yes", note: "Spectral has broad existing rule packs." },
      { criterion: "CI performance for syntax/style", stricture: "Partial", tool: "Yes", note: "Spectral excels for quick schema checks." },
      { criterion: "Multi-repo lineage truth", stricture: "Yes", tool: "No", note: "Out of scope for Spectral." }
    ],
    diagrams: [
      {
        title: "Two-Stage Lint Pipeline",
        description: "Stage 1 Spectral style lint, Stage 2 Stricture lineage drift gate before deploy."
      },
      {
        title: "Rule Scope Contrast",
        description: "Matrix view of static syntax checks vs lineage/ownership policy checks."
      }
    ],
    references: [
      { label: "Stoplight Spectral", href: "https://docs.stoplight.io/docs/spectral/674b27b261c3c-overview" }
    ]
  },
  {
    slug: "buf-protobuf",
    tool: "Buf + Protobuf Breaking Checks",
    shortName: "Buf/Proto",
    category: "IDL Contract Toolchain",
    overlapScore: 40,
    overlapSummary: "Moderate overlap in contract compatibility checks for protobuf systems.",
    oneLineDifference: "Buf enforces protobuf compatibility; Stricture enforces lineage, source ownership, and cross-system drift semantics.",
    bestAt: [
      "Schema evolution governance for protobuf",
      "Breaking change checks and linting",
      "Registry and module workflows for proto APIs"
    ],
    overlapAreas: [
      "Breaking change prevention",
      "CI gating",
      "Schema governance"
    ],
    strictureAdds: [
      "Field source provenance including external providers",
      "API/output drift checks beyond IDL compatibility",
      "Escalation chain metadata",
      "Version/as-of checks across non-protobuf dependencies"
    ],
    toolAdds: [
      "Strong protobuf semantics and compatibility guarantees",
      "Developer workflow for proto modules and registries",
      "Generated artifact consistency"
    ],
    whenUseBoth: [
      "Use Buf for protobuf evolution correctness and Stricture for provenance + dependency drift correctness.",
      "Use overlays so proto options and Stricture annotations map to shared lineage IDs without repeating metadata.",
      "Keep Buf as system-of-record for protobuf compatibility and Stricture as system-of-record for source traceability and escalation policy."
    ],
    whenUseToolOnly: [
      "Your ecosystem is exclusively protobuf and you do not need source provenance governance.",
      "No need to trace data origins across external/internal provider boundaries."
    ],
    adoptionSteps: [
      "Retain existing Buf breaking checks in CI.",
      "Annotate critical protobuf response/output fields with Stricture lineage metadata.",
      "Bridge proto package names to Stricture source IDs through overlay aliases.",
      "Activate Stricture block mode for high-impact source/version drift only."
    ],
    overlaySnippet: `message PaymentStatus {\n  string status = 1 [(stricture.field) = "payment.status"];\n  string source_system = 2 [(stricture.source) = "PaymentsCore@2026.02"];\n  // overlay aliases map this to Buf-compatible custom options and Stricture keys\n}`,
    matrix: [
      { criterion: "Protobuf compatibility checks", stricture: "Partial", tool: "Yes", note: "Buf is purpose-built for protobuf break detection." },
      { criterion: "Field provenance across systems", stricture: "Yes", tool: "No", note: "Stricture tracks source and enrichment provenance." },
      { criterion: "CI policy gating", stricture: "Yes", tool: "Yes", note: "Both support CI gates with different focus." },
      { criterion: "External dependency freshness", stricture: "Yes", tool: "No", note: "Stricture checks as-of/version drift for providers." },
      { criterion: "Escalation contact outputs", stricture: "Yes", tool: "No", note: "Stricture outputs responder chains." },
      { criterion: "Schema registry workflow", stricture: "No", tool: "Yes", note: "Buf registry tooling is out of Stricture scope." },
      { criterion: "Cross-protocol coverage", stricture: "Yes", tool: "No", note: "Stricture spans REST/events/proto with same lineage model." },
      { criterion: "Override workflow", stricture: "Yes", tool: "Partial", note: "Stricture has explicit time-bounded override comments." }
    ],
    diagrams: [
      {
        title: "Proto Compatibility + Lineage Gate",
        description: "Buf passes compatibility, then Stricture validates provenance and source-version drift before release."
      },
      {
        title: "Shared Annotation Map",
        description: "Show proto options mapped through overlay aliases to Stricture canonical fields."
      }
    ],
    references: [
      { label: "Buf Docs", href: "https://buf.build/docs" },
      { label: "Protocol Buffers Docs", href: "https://protobuf.dev/" }
    ]
  },
  {
    slug: "datahub-atlas",
    tool: "DataHub / Apache Atlas",
    shortName: "Catalog Tools",
    category: "Metadata Catalog & Governance",
    overlapScore: 30,
    overlapSummary: "Low-moderate overlap in governance metadata; Stricture is focused on deploy-time API drift enforcement.",
    oneLineDifference: "Catalog platforms curate metadata at scale; Stricture enforces field-level API lineage and release gates in engineering workflows.",
    bestAt: [
      "Enterprise metadata catalog and discovery",
      "Governance workflows and stewardship",
      "Global lineage and glossary search"
    ],
    overlapAreas: [
      "Lineage visualization",
      "Ownership metadata",
      "Data governance context"
    ],
    strictureAdds: [
      "Pre-deploy drift policy decisions in CI/CD",
      "Field-level API output source checks",
      "Time-bounded override workflows",
      "On-call escalation routing from failing fields"
    ],
    toolAdds: [
      "Enterprise catalog UX and search",
      "Broader metadata domain model",
      "Stewardship/governance operating model"
    ],
    whenUseBoth: [
      "Use catalog tools for broad metadata governance and discovery, and Stricture for execution-time enforcement in delivery pipelines.",
      "Use overlays so lineage identities are reused between catalog entities and Stricture annotations with no redundant authoring.",
      "Keep DataHub/Atlas as system-of-record for enterprise metadata and Stricture as system-of-record for release gate policy and incident escalation."
    ],
    whenUseToolOnly: [
      "Your priority is metadata discovery and catalog governance, not API release gating.",
      "No requirement for field-level CI drift blocking or override controls."
    ],
    adoptionSteps: [
      "Map catalog asset IDs to Stricture source-system IDs.",
      "Start Stricture on a small set of customer-facing APIs.",
      "Publish Stricture outputs into catalog metadata for discoverability.",
      "Automate escalation links between Stricture findings and catalog owners."
    ],
    overlaySnippet: `x-stricture-lineage:\n  field: catalog_item.price\n  from:\n    - system: PricingCore\n      version: 2026.02\n  aliases:\n    datahub.urn: "urn:li:dataset:(urn:li:dataPlatform:hive,pricing,PROD)"\n    atlas.guid: "8f7b2..."`,
    matrix: [
      { criterion: "Metadata catalog and search", stricture: "No", tool: "Yes", note: "Catalog platforms dominate this area." },
      { criterion: "Field-level API drift gate", stricture: "Yes", tool: "No", note: "Stricture targets release-time enforcement." },
      { criterion: "Lineage visualization", stricture: "Partial", tool: "Yes", note: "Catalog tools provide richer global lineage UX." },
      { criterion: "Escalation workflow", stricture: "Yes", tool: "Partial", note: "Catalogs may store owners; Stricture operationalizes responder chains." },
      { criterion: "Override lifecycle", stricture: "Yes", tool: "No", note: "Stricture supports expiry and reason metadata for exceptions." },
      { criterion: "Cross-repo contract verification", stricture: "Yes", tool: "Partial", note: "Catalog tools are not usually release gate engines." },
      { criterion: "Enterprise governance model", stricture: "Partial", tool: "Yes", note: "Catalogs provide broader stewardship and policy workflows." },
      { criterion: "CI-first developer workflow", stricture: "Yes", tool: "Partial", note: "Stricture is designed to fail/passing builds." }
    ],
    diagrams: [
      {
        title: "Catalog + Gate Architecture",
        description: "Catalog is long-lived metadata plane; Stricture sits in CI/CD for pre-deploy enforcement and feeds back results."
      },
      {
        title: "Owner Escalation Loop",
        description: "Failing field links to Stricture escalation chain and corresponding catalog ownership record."
      }
    ],
    references: [
      { label: "DataHub Docs", href: "https://docs.datahub.com/" },
      { label: "Apache Atlas Docs", href: "https://atlas.apache.org/" }
    ]
  }
];

export const comparisonBySlug = Object.fromEntries(comparisons.map((entry) => [entry.slug, entry])) as Record<string, ComparisonEntry>;

export const highOverlapComparisons = comparisons.filter((entry) => entry.overlapScore >= 40);
