const state = {
  sessionId: "",
  snapshot: null,
};

const selectors = {
  gateBanner: document.querySelector("#gate-banner"),
  topology: document.querySelector("#topology"),
  topologyGraph: document.querySelector("#topology-graph"),
  edgeList: document.querySelector("#edge-list"),
  flowPathSummary: document.querySelector("#flow-path-summary"),
  findings: document.querySelector("#findings"),
  overrides: document.querySelector("#overrides"),
  escalation: document.querySelector("#escalation"),
  runSummaryText: document.querySelector("#run-summary-text"),
  controlStats: document.querySelector("#control-stats"),
  scenarioNarrative: document.querySelector("#scenario-narrative"),
  presetScenario: document.querySelector("#preset-scenario"),
  mutationType: document.querySelector("#mutation-type"),
  mutationService: document.querySelector("#mutation-service"),
  mutationField: document.querySelector("#mutation-field"),
  policyMode: document.querySelector("#policy-mode"),
  policyFailOn: document.querySelector("#policy-fail-on"),
  overrideField: document.querySelector("#override-field"),
  overrideChange: document.querySelector("#override-change"),
  overrideExpires: document.querySelector("#override-expires"),
  overrideReason: document.querySelector("#override-reason"),
  overrideTicket: document.querySelector("#override-ticket"),
  escalationService: document.querySelector("#escalation-service"),
  applyMutation: document.querySelector("#apply-mutation"),
  updatePolicy: document.querySelector("#update-policy"),
  addOverride: document.querySelector("#add-override"),
  runStricture: document.querySelector("#run-stricture"),
  resetSession: document.querySelector("#reset-session"),
  loadEscalation: document.querySelector("#load-escalation"),
  toggleEdges: document.querySelector("#toggle-edges"),
};

const humanChange = {
  annotation_missing: "annotation missing",
  external_as_of_rollback: "external snapshot stale",
  external_as_of_stale: "external snapshot stale",
  field_removed: "field removed",
  merge_strategy_changed: "merge behavior changed",
  source_contract_ref_changed: "contract reference changed",
  source_removed: "source mapping removed",
  source_version_changed: "source version changed",
  type_changed: "type changed",
  enum_changed: "enum changed",
};

function serviceById(snapshot) {
  return new Map((snapshot.services || []).map((service) => [service.id, service]));
}

function normalizeServiceToken(value) {
  return String(value || "").trim().toLowerCase().replace(/[^a-z0-9]/g, "");
}

function resolveServiceId(snapshot, token) {
  const want = normalizeServiceToken(token);
  if (!want) return "";
  const services = snapshot.services || [];
  for (const service of services) {
    if (normalizeServiceToken(service.id) === want) return service.id;
    if (normalizeServiceToken(service.name) === want) return service.id;
  }
  return "";
}

function serviceName(snapshot, serviceId) {
  const service = serviceById(snapshot).get(serviceId);
  return service?.name || serviceId || "unknown service";
}

function serviceOwner(snapshot, serviceId) {
  const service = serviceById(snapshot).get(serviceId);
  return service?.owner || "unknown owner";
}

function latestMutation(snapshot) {
  if (!snapshot?.mutations?.length) {
    return null;
  }
  return snapshot.mutations[snapshot.mutations.length - 1];
}

function listServiceNames(snapshot, ids, limit = 3) {
  const names = [...new Set((ids || []).map((id) => serviceName(snapshot, id)).filter(Boolean))];
  if (!names.length) return "none";
  if (names.length <= limit) return names.join(", ");
  return `${names.slice(0, limit).join(", ")} +${names.length - limit} more`;
}

function inferSourceServices(snapshot, findings, latest) {
  const sourceServices = new Set();
  (findings || []).forEach((finding) => {
    if (finding?.source?.serviceId) {
      sourceServices.add(finding.source.serviceId);
      return;
    }
    if (finding?.source?.service) {
      const sourceId = resolveServiceId(snapshot, finding.source.service);
      if (sourceId) {
        sourceServices.add(sourceId);
        return;
      }
    }
    const summary = String(finding.summary || "");
    const sourceMatch = /source\s+[a-z_]+\|([A-Za-z0-9_-]+)\./i.exec(summary);
    if (sourceMatch?.[1]) {
      const sourceId = resolveServiceId(snapshot, sourceMatch[1]);
      if (sourceId) {
        sourceServices.add(sourceId);
        return;
      }
    }
    const upstreamDriven = (
      finding.changeType === "source_contract_ref_changed" ||
      finding.changeType === "source_removed" ||
      finding.changeType === "source_provider_changed" ||
      finding.changeType === "source_upstream_system_changed" ||
      finding.changeType === "external_as_of_rollback" ||
      finding.changeType === "external_as_of_changed" ||
      finding.changeType === "external_as_of_advanced"
    );
    if (upstreamDriven && finding.fieldId) {
      const fieldEdges = (snapshot.edges || []).filter((edge) => edge.fieldId === finding.fieldId);
      const external = fieldEdges.find((edge) => {
        const from = serviceById(snapshot).get(edge.from);
        return from?.kind === "external";
      });
      if (external?.from) {
        sourceServices.add(external.from);
      }
    }
  });
  if (!sourceServices.size && latest?.serviceId) {
    sourceServices.add(latest.serviceId);
  }
  return sourceServices;
}

function findingSummaryText(snapshot, finding) {
  void snapshot;
  return finding.summary || "";
}

async function request(path, method = "GET", body) {
  const response = await fetch(path, {
    method,
    headers: body ? { "content-type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!response.ok) {
    const details = await response.text();
    throw new Error(`${method} ${path} failed: ${response.status} ${details}`);
  }
  return response.json();
}

function setOptions(element, values, formatter) {
  if (!element) {
    return;
  }
  element.innerHTML = "";
  values.forEach((value) => {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = formatter ? formatter(value) : value;
    element.appendChild(option);
  });
}

function refreshMutationFieldOptions(snapshot) {
  if (!selectors.mutationType || !selectors.mutationField) {
    return;
  }
  const mutationType = selectors.mutationType.value;
  const fieldIDs = snapshot.fieldsByMutation?.[mutationType] || [];
  setOptions(selectors.mutationField, fieldIDs, (id) => id);
}

function updateControlStats(snapshot) {
  if (!selectors.controlStats || !selectors.mutationType) {
    return;
  }
  const mutationType = selectors.mutationType.value;
  const fieldCount = snapshot.fieldsByMutation?.[mutationType]?.length || 0;
  const serviceCount = snapshot.services?.length || 0;
  selectors.controlStats.textContent = `${serviceCount} services • ${fieldCount} fields for this mutation`;
}

function buildPresets(snapshot) {
  const nonePreset = { id: "__none", label: "No scenario (baseline)" };
  const candidates = [
    { id: "enum_changed", label: "Payments enum drift" },
    { id: "type_changed", label: "Orders type drift" },
    { id: "external_as_of_stale", label: "External as-of stale" },
    { id: "annotation_missing", label: "Missing annotation" },
    { id: "numeric_widen", label: "Numeric widening: uint8→uint16" },
  ];
  const presets = [nonePreset, ...candidates
    .map((item) => {
      const fields = snapshot.fieldsByMutation?.[item.id] || [];
      return fields.length ? { ...item, fieldId: fields[0] } : null;
    })
    .filter(Boolean)];
  if (selectors.presetScenario) {
    const current = selectors.presetScenario.value;
    setOptions(
      selectors.presetScenario,
      presets.map((p) => p.id),
      (id) => presets.find((p) => p.id === id)?.label || id,
    );
    if (presets.length) {
      if (current && presets.find((p) => p.id === current)) {
        selectors.presetScenario.value = current;
      } else {
        selectors.presetScenario.value = nonePreset.id;
      }
      updateNarrative(selectors.presetScenario.value);
    }
  }
  return presets;
}

function updateNarrative(presetId) {
  if (!selectors.scenarioNarrative) {
    return;
  }
  const narratives = {
    __none: "No mutation scenario selected. This shows the baseline topology with no active drift findings.",
    enum_changed: "Payments enum drift: a producer changed payment status handling without a coordinated contract update. Downstream services can misclassify payment state until contracts and consumers align.",
    type_changed: "Orders type drift: fulfillment switched quantity from int to string for partial units. Legacy consumers treat it as numeric and will fail parsing.",
    external_as_of_stale: "External as-of stale: vendor shipping ETA feed is older than allowed freshness window, so SLAs and alerts rely on outdated data.",
    annotation_missing: "Missing annotation: a new field shipped without lineage metadata; Stricture blocks because provenance and owners are unknown.",
    numeric_widen: "Numeric widening: Go producer widened uint8 to uint16; JS consumer coerces to Number and forwards large values; downstream Go consumer overflows. Guard size limits and bump contracts.",
  };
  selectors.scenarioNarrative.textContent = narratives[presetId] || "Choose a preset to see the drift story and impact.";
}

function updateNarrativeFromSnapshot(snapshot) {
  if (!selectors.scenarioNarrative) return;
  const findings = snapshot.findings || [];
  if (!findings.length) return;
  const mutation = latestMutation(snapshot);
  const sourceServices = inferSourceServices(snapshot, findings, mutation);
  const sourceID = sourceServices.size ? [...sourceServices][0] : mutation?.serviceId;
  const top = findings[0];
  const sourceNameText = sourceID ? serviceName(snapshot, sourceID) : "Unknown source";
  const impactNameText = serviceName(snapshot, top.serviceId);
  selectors.scenarioNarrative.textContent = `${sourceNameText} introduced ${humanChange[top.changeType] || top.changeType} on ${top.fieldId}. Stricture flags impact in ${impactNameText}.`;
}

function renderGate(summary) {
  if (!selectors.gateBanner || !summary) {
    return;
  }
  selectors.gateBanner.classList.remove("gate-ok", "gate-warn", "gate-block");
  const cls = summary.gate === "BLOCK" ? "gate-block" : summary.findingCount > 0 ? "gate-warn" : "gate-ok";
  selectors.gateBanner.classList.add(cls);
  const gateText = summary.gate === "BLOCK" ? "Deploy blocked" : "Deploy allowed";
  selectors.gateBanner.textContent = `Run #${summary.runCount} | ${gateText} | ${summary.findingCount} finding(s)`;
}

function renderRunSummary(snapshot) {
  if (!selectors.runSummaryText) {
    return;
  }
  const summary = snapshot.runSummary;
  if (!summary || summary.runCount === 0) {
    selectors.runSummaryText.textContent = "Awaiting first run. Apply a mutation then rerun to see how Stricture flags drift.";
    return;
  }
  const findingWord = summary.findingCount === 1 ? "finding" : "findings";
  const gateText = summary.gate === "BLOCK" ? "Deploy blocked." : "Deploy allowed.";
  const policyText = summary.mode === "block"
    ? `Policy is block at ${snapshot.policy.failOn}+ severity.`
    : `Policy is warn (never blocks), threshold is ${snapshot.policy.failOn}+ for escalation.`;
  const mutation = latestMutation(snapshot);
  const sources = inferSourceServices(snapshot, snapshot.findings || [], mutation);
  const sourceText = sources.size
    ? `Changed source: ${listServiceNames(snapshot, [...sources], 2)}.`
    : mutation ? `Changed source: ${serviceName(snapshot, mutation.serviceId)}.` : "";
  const impactedText = summary.findingCount
    ? `Flagged services: ${listServiceNames(snapshot, snapshot.findings.map((finding) => finding.serviceId))}.`
    : "No services were flagged.";
  selectors.runSummaryText.textContent = `Run #${summary.runCount}: ${summary.findingCount} ${findingWord} (${summary.blockedCount} high, ${summary.warningCount} warning). ${gateText} ${policyText} ${sourceText} ${impactedText}`;
}

function nodeStatusForService(serviceId, findings) {
  let blocked = false;
  let warning = false;
  findings.forEach((item) => {
    if (item.serviceId !== serviceId) {
      return;
    }
    if (item.severity === "high") {
      blocked = true;
      return;
    }
    warning = true;
  });
  if (blocked) {
    return "node-block";
  }
  if (warning) {
    return "node-warn";
  }
  return "node-ok";
}

function renderTopology(snapshot) {
  if (!selectors.topology) {
    return;
  }
  selectors.topology.innerHTML = "";
  snapshot.services.forEach((service) => {
    const card = document.createElement("article");
    card.className = `node ${nodeStatusForService(service.id, snapshot.findings)}`;
    card.innerHTML = `
      <h3>${service.name}</h3>
      <p>domain=${service.domain} owner=${service.owner}</p>
      <p>flows=${service.flowCount} kind=${service.kind}</p>
      <span class="node-badge">${service.escalation}</span>
    `;
    selectors.topology.appendChild(card);
  });

  if (selectors.edgeList) {
    selectors.edgeList.innerHTML = "";
    snapshot.edges.forEach((edge) => {
      const item = document.createElement("article");
      item.className = `edge edge-${edge.status}`;
      const label = document.createElement("div");
      label.className = "edge-label";
      label.textContent = `${edge.from} → ${edge.to}`;
      const meta = document.createElement("div");
      meta.className = "edge-meta";
      meta.textContent = `field=${edge.fieldId} • status=${edge.status}`;
      item.appendChild(label);
      item.appendChild(meta);
      selectors.edgeList.appendChild(item);
    });
  }

  renderGraph(snapshot);
}

function renderList(element, rows, emptyText, mapper) {
  if (!element) {
    return;
  }
  element.innerHTML = "";
  if (!rows.length) {
    const empty = document.createElement("div");
    empty.className = "list-item";
    empty.textContent = emptyText;
    element.appendChild(empty);
    return;
  }
  rows.forEach((item) => {
    const node = document.createElement("article");
    node.className = mapper.className ? mapper.className(item) : "list-item";
    node.innerHTML = mapper.html(item);
    element.appendChild(node);
  });
}

function render(snapshot) {
  state.snapshot = snapshot;

  renderGate(snapshot.runSummary);
  renderRunSummary(snapshot);
  renderTopology(snapshot);
  updateControlStats(snapshot);
  updateNarrativeFromSnapshot(snapshot);

  renderList(
    selectors.findings,
    snapshot.findings,
    "No findings yet.",
    {
      className: (finding) => `list-item sev-${finding.severity}`,
      html: (finding) => {
        const mutation = latestMutation(snapshot);
        const sources = inferSourceServices(snapshot, [finding], mutation);
        const source = sources.size ? serviceName(snapshot, [...sources][0]) : mutation ? serviceName(snapshot, mutation.serviceId) : "unknown";
        const impacted = serviceName(snapshot, finding.serviceId);
        const owner = serviceOwner(snapshot, finding.serviceId);
        const flowServices = (snapshot.edges || [])
          .filter((edge) => edge.fieldId === finding.fieldId)
          .flatMap((edge) => [edge.from, edge.to]);
        const blastRadius = listServiceNames(snapshot, [...flowServices, finding.serviceId], 4);
        return `
        <h3>${humanChange[finding.changeType] || finding.changeType} — ${finding.fieldId} (${finding.severity.toUpperCase()}) in ${impacted}</h3>
        <p>${findingSummaryText(snapshot, finding)}</p>
        <p class="item-meta chips">
          <span class="chip">Cause: ${source}</span>
          <span class="chip">Impact: ${impacted}</span>
          <span class="chip">Blast radius: ${blastRadius}</span>
          <span class="chip">Owner: ${owner}</span>
        </p>
        ${finding.validation ? `<p class="item-meta">Validation: ${finding.validation}</p>` : ""}
        ${finding.suggestion ? `<p class="item-meta">Suggestion: ${finding.suggestion}</p>` : ""}
        <p class="item-meta">Remediation: ${finding.remediation}</p>
      `;
      },
    },
  );

  renderList(
    selectors.overrides,
    snapshot.overrides,
    "No overrides active.",
    {
      html: (entry) => `
        <h3>${entry.fieldId} (${entry.changeType})</h3>
        <p>${entry.reason}</p>
        <p class="item-meta">expires=${entry.expires} ticket=${entry.ticket || "n/a"}</p>
      `,
    },
  );

  if (selectors.mutationType && Array.isArray(snapshot.mutationTypes)) {
    const selected = selectors.mutationType.value;
    setOptions(selectors.mutationType, snapshot.mutationTypes, (id) => id);
    if (snapshot.mutationTypes.includes(selected)) {
      selectors.mutationType.value = selected;
    }
  }

  const serviceIDs = snapshot.services.map((service) => service.id);
  setOptions(selectors.mutationService, serviceIDs, (id) => id);
  setOptions(selectors.escalationService, serviceIDs, (id) => id);

  refreshMutationFieldOptions(snapshot);
  buildPresets(snapshot);

  if (selectors.policyMode) {
    selectors.policyMode.value = snapshot.policy.mode;
  }
  if (selectors.policyFailOn) {
    selectors.policyFailOn.value = snapshot.policy.failOn;
  }
}

async function bootstrap() {
  const created = await request("/api/session", "POST", {});
  state.sessionId = created.sessionId;
  render(created.snapshot);
}

async function run() {
  if (!state.sessionId) {
    return;
  }
  const response = await request(`/api/session/${state.sessionId}/run`, "POST", {});
  render(response.snapshot);
  revealFindings();
}

async function mutate() {
  if (!state.sessionId) {
    return;
  }
  const payload = {
    type: selectors.mutationType?.value,
    serviceId: selectors.mutationService?.value,
    fieldId: selectors.mutationField?.value,
  };
  const response = await request(`/api/session/${state.sessionId}/mutations`, "POST", payload);
  render(response.snapshot);
  revealFindings();
}

async function updatePolicy() {
  if (!state.sessionId) {
    return;
  }
  const payload = {
    mode: selectors.policyMode?.value,
    failOn: selectors.policyFailOn?.value,
  };
  const response = await request(`/api/session/${state.sessionId}/policy`, "POST", payload);
  render(response.snapshot);
}

async function addOverride() {
  if (!state.sessionId) {
    return;
  }
  const payload = {
    fieldId: selectors.overrideField?.value.trim(),
    changeType: selectors.overrideChange?.value.trim() || "*",
    expires: selectors.overrideExpires?.value,
    reason: selectors.overrideReason?.value.trim(),
    ticket: selectors.overrideTicket?.value.trim(),
  };
  const response = await request(`/api/session/${state.sessionId}/override`, "POST", payload);
  render(response.snapshot);
}

function contactsToText(contacts) {
  if (!Array.isArray(contacts) || contacts.length === 0) {
    return "n/a";
  }
  return contacts
    .map((entry) => entry.channel || entry.name || "unknown")
    .filter(Boolean)
    .join(", ");
}

async function loadEscalation() {
  if (!state.sessionId) {
    return;
  }
  const serviceId = selectors.escalationService?.value;
  const response = await request(`/api/session/${state.sessionId}/escalation?serviceId=${encodeURIComponent(serviceId)}`);

  renderList(
    selectors.escalation,
    response.chain || [],
    "No chain available.",
    {
      html: (entry) => `
        <h3>depth=${entry.depth} ${entry.system_id}</h3>
        <p>${entry.reason}</p>
        <p class="item-meta">owner=${entry.owner || "n/a"} contacts=${contactsToText(entry.contacts)}</p>
      `,
    },
  );
}

function bindEvents() {
  selectors.applyMutation?.addEventListener("click", () => mutate().catch(showError));
  selectors.updatePolicy?.addEventListener("click", () => updatePolicy().catch(showError));
  selectors.addOverride?.addEventListener("click", () => addOverride().catch(showError));
  selectors.runStricture?.addEventListener("click", () => run().catch(showError));
  selectors.resetSession?.addEventListener("click", () => bootstrap().catch(showError));
  selectors.loadEscalation?.addEventListener("click", () => loadEscalation().catch(showError));
  selectors.presetScenario?.addEventListener("change", (event) => {
    const target = event.target;
    updateNarrative(target?.value);
    applyPreset().catch(showError);
  });
  selectors.toggleEdges?.addEventListener("click", () => toggleEdgeList());
  selectors.mutationType?.addEventListener("change", () => {
    if (state.snapshot) {
      refreshMutationFieldOptions(state.snapshot);
      updateControlStats(state.snapshot);
    }
  });
}

function showError(error) {
  console.error(error);
  if (!selectors.gateBanner) {
    return;
  }
  selectors.gateBanner.classList.remove("gate-ok", "gate-warn", "gate-block");
  selectors.gateBanner.classList.add("gate-block");
  selectors.gateBanner.textContent = `Error: ${error.message}`;
}

function revealFindings() {
  if (!selectors.findings) {
    return;
  }
  selectors.findings.scrollIntoView({ behavior: "smooth", block: "start" });
  selectors.findings.classList.add("pulse-once");
  window.setTimeout(() => {
    selectors.findings?.classList.remove("pulse-once");
  }, 900);
}

async function applyPreset() {
  if (!state.snapshot || !selectors.presetScenario) {
    return;
  }
  const id = selectors.presetScenario.value;
  if (id === "__none") {
    updateNarrative(id);
    await bootstrap();
    return;
  }
  const fields = state.snapshot.fieldsByMutation?.[id] || [];
  const fieldId = fields[0];
  if (!fieldId) {
    throw new Error("No fields available for this preset.");
  }
  if (selectors.mutationType) {
    selectors.mutationType.value = id;
  }
  refreshMutationFieldOptions(state.snapshot);
  if (selectors.mutationField) {
    selectors.mutationField.value = fieldId;
  }
  if (selectors.mutationService && selectors.mutationService.options.length) {
    const candidateSource = (state.snapshot.edges || []).find((edge) => edge.fieldId === fieldId)?.from;
    if (candidateSource && [...selectors.mutationService.options].some((option) => option.value === candidateSource)) {
      selectors.mutationService.value = candidateSource;
    } else {
      selectors.mutationService.value = selectors.mutationService.options[0].value;
    }
  }
  updateNarrative(id);
  await bootstrap(); // reset session so findings don’t pile up
  await mutate();
  await run();
}

function toggleEdgeList() {
  if (!selectors.edgeList || !selectors.toggleEdges) return;
  const isCollapsed = selectors.edgeList.classList.toggle("is-collapsed");
  selectors.toggleEdges.textContent = isCollapsed ? "Show details" : "Hide details";
}

bindEvents();
bootstrap().catch(showError);
scheduleResizeRender();

function renderGraph(snapshot) {
  if (!selectors.topologyGraph) {
    return;
  }
  const container = selectors.topologyGraph;
  container.innerHTML = "";
  const svgNS = "http://www.w3.org/2000/svg";
  const { width: boxWidth } = container.getBoundingClientRect();
  const width = Math.max(boxWidth || 640, 820);
  const height = Math.max(420, Math.ceil((snapshot.services?.length || 8) / 3) * 120);
  const svg = document.createElementNS(svgNS, "svg");
  svg.setAttribute("viewBox", `0 0 ${width} ${height}`);

  const defs = document.createElementNS(svgNS, "defs");
  const marker = document.createElementNS(svgNS, "marker");
  marker.setAttribute("id", "arrowhead");
  marker.setAttribute("markerWidth", "8");
  marker.setAttribute("markerHeight", "8");
  marker.setAttribute("refX", "4");
  marker.setAttribute("refY", "3.5");
  marker.setAttribute("orient", "auto");
  const arrowPath = document.createElementNS(svgNS, "path");
  arrowPath.setAttribute("d", "M0,0 L8,3.5 L0,7 z");
  arrowPath.setAttribute("fill", "currentColor");
  marker.appendChild(arrowPath);
  defs.appendChild(marker);
  svg.appendChild(defs);

  const nodes = snapshot.services || [];
  const edges = snapshot.edges || [];
  const longestLabel = nodes.reduce((max, node) => {
    const label = typeof node?.name === "string" ? node.name : "";
    return Math.max(max, label.length);
  }, 0);
  const sideLabelPadding = Math.min(Math.max(56, Math.ceil(longestLabel * 3.8)), 190);
  const {
    activeFields,
    sourceServices,
    impactedServices,
    failedServices,
    flowNodes,
    flowEdges,
    affectedServices,
    focusFieldSeverity,
  } = computeImpacts(snapshot);
  const positions = normalizePositions(
    computeLayeredLayout(nodes, edges, width, height, { affectedServices, sourceServices, impactedServices, flowNodes }),
    width,
    height,
    {
      left: sideLabelPadding,
      right: sideLabelPadding,
      top: 44,
      bottom: 72,
    },
  );

  const edgeColor = (status) => {
    if (status === "blocked") return "var(--danger)";
    if (status === "warning") return "var(--warn)";
    if (status === "healthy") return "var(--ok)";
    return "var(--line)";
  };

  const nodeRadius = (id) => {
    if (sourceServices.has(id) || impactedServices.has(id)) return 28;
    if (flowNodes.has(id)) return 24;
    return 16;
  };

  edges.forEach((edge) => {
    const from = positions.get(edge.from);
    const to = positions.get(edge.to);
    if (!from || !to) return;
    const fromR = nodeRadius(edge.from);
    const toR = nodeRadius(edge.to);
    const dx = to.x - from.x;
    const dy = to.y - from.y;
    const distance = Math.max(Math.hypot(dx, dy), 1);
    const ux = dx / distance;
    const uy = dy / distance;
    const startX = from.x + ux * fromR;
    const startY = from.y + uy * fromR;
    const endX = to.x - ux * toR;
    const endY = to.y - uy * toR;
    const curve = document.createElementNS(svgNS, "path");
    const midX = (startX + endX) / 2;
    const midY = (startY + endY) / 2 - 24;
    const d = `M ${startX} ${startY} Q ${midX} ${midY} ${endX} ${endY}`;
    curve.setAttribute("d", d);
    const isFlow = flowEdges.has(edge.id);
    const isActiveField = activeFields.has(edge.fieldId);
    const isSourceImpactEdge =
      isFlow && (sourceServices.has(edge.from) || sourceServices.has(edge.to)) &&
      (impactedServices.has(edge.from) || impactedServices.has(edge.to) || isActiveField);
    const color = !isFlow ? "var(--line)" : isSourceImpactEdge ? "var(--warn)" : edgeColor(edge.status);
    curve.setAttribute("stroke", color);
    const classes = [`graph-edge`, `edge-${edge.status}`, "flow"];
    if (!isFlow) classes.push("dimmed");
    if (isSourceImpactEdge) classes.push("edge-affected");
    curve.setAttribute("class", classes.join(" "));
    const title = document.createElementNS(svgNS, "title");
    title.textContent = `${edge.from} → ${edge.to} • ${edge.fieldId} • ${edge.status}`;
    curve.appendChild(title);
    curve.dataset.edgeId = edge.id;
    curve.dataset.from = edge.from;
    curve.dataset.to = edge.to;
    svg.appendChild(curve);
  });

  nodes.forEach((node) => {
    const pos = positions.get(node.id);
    if (!pos) return;
    const g = document.createElementNS(svgNS, "g");
    const statusClass = nodeStatusForService(node.id, snapshot.findings);
    const isolated = (node.edgeDegree || 0) === 0;
    const inFlow = flowNodes.has(node.id);
    const isSource = sourceServices.has(node.id);
    const isImpacted = impactedServices.has(node.id);
    const isFailing = failedServices.has(node.id);
    const classes = ["graph-node", statusClass];
    if (isolated) classes.push("isolated");
    if (!inFlow && !isSource && !isImpacted) classes.push("dimmed");
    if (isSource) classes.push("source");
    if (isImpacted) classes.push("impacted");
    if (isFailing) classes.push("failing");
    g.setAttribute("class", classes.join(" "));
    const circle = document.createElementNS(svgNS, "circle");
    circle.setAttribute("cx", pos.x);
    circle.setAttribute("cy", pos.y);
    circle.setAttribute("r", isSource || isImpacted ? "28" : inFlow ? "24" : "16");
    g.appendChild(circle);
    const label = document.createElementNS(svgNS, "text");
    label.setAttribute("x", pos.x);
    label.setAttribute("y", pos.y + 37);
    if (isSource || isImpacted) {
      label.setAttribute("font-weight", "700");
    }
    label.textContent = node.name;
    const title = document.createElementNS(svgNS, "title");
    title.textContent = `${node.name} (${node.domain}) • owner=${node.owner}`;
    g.appendChild(title);
    g.appendChild(label);
    g.dataset.nodeId = node.id;
    svg.appendChild(g);
  });

  container.appendChild(svg);

  if (selectors.flowPathSummary) {
    selectors.flowPathSummary.textContent = summarizeFlow(
      snapshot,
      nodes,
      edges,
      sourceServices,
      impactedServices,
      focusFieldSeverity,
      activeFields,
    );
  }

  addGraphInteractions(svg);
}

function scheduleResizeRender() {
  let timer = null;
  const handler = () => {
    if (!state.snapshot) return;
    window.clearTimeout(timer);
    timer = window.setTimeout(() => renderGraph(state.snapshot), 120);
  };
  window.addEventListener("resize", handler);
}

function computeLayeredLayout(nodes, edges, width, height, focus = {}) {
  const positions = new Map();
  const incoming = new Map();
  const outgoing = new Map();
  const focusedNodes = new Set(
    (focus.affectedServices && focus.affectedServices.size
      ? [...focus.affectedServices]
      : [...(focus.flowNodes || []), ...(focus.sourceServices || [])]),
  );
  nodes.forEach((n) => {
    incoming.set(n.id, 0);
    outgoing.set(n.id, 0);
  });
  edges.forEach((e) => {
    incoming.set(e.to, (incoming.get(e.to) || 0) + 1);
    outgoing.set(e.from, (outgoing.get(e.from) || 0) + 1);
  });
  const topStart = 56;
  const topEnd = Math.max(topStart + 120, Math.floor(height * 0.6));
  const bottomStart = Math.max(topEnd + 20, Math.floor(height * 0.74));
  const bottomEnd = Math.max(bottomStart + 30, height - 40);

  const placeInBand = (ids, x, startY, endY) => {
    if (!ids.length) return;
    if (ids.length === 1) {
      positions.set(ids[0], { x, y: (startY + endY) / 2 });
      return;
    }
    const gap = (endY - startY) / (ids.length - 1);
    ids.forEach((nodeID, idx) => {
      positions.set(nodeID, { x, y: startY + gap * idx });
    });
  };

  if (focusedNodes.size > 0) {
    const sourceSet = new Set([...(focus.sourceServices || [])]);
    const impactedSet = new Set([...(focus.impactedServices || [])].filter((id) => !sourceSet.has(id)));
    const keySet = new Set([...focusedNodes, ...sourceSet, ...impactedSet]);
    const allNodeIDs = nodes.map((n) => n.id);
    const source = allNodeIDs.filter((id) => sourceSet.has(id));
    const impacted = allNodeIDs.filter((id) => impactedSet.has(id));
    const bridge = allNodeIDs.filter((id) => keySet.has(id) && !sourceSet.has(id) && !impactedSet.has(id));
    const background = allNodeIDs.filter((id) => !keySet.has(id));

    if (!source.length && bridge.length) {
      source.push(bridge.shift());
    }
    if (!impacted.length && bridge.length > 1) {
      impacted.push(bridge.pop());
    }

    const laneCount = 4;
    const laneWidth = Math.max((width - 40) / laneCount, 120);
    const laneX = (idx) => 20 + laneWidth * idx + laneWidth / 2;
    const bridgeLeft = [];
    const bridgeRight = [];
    bridge.forEach((id, idx) => {
      if (idx % 2 === 0) {
        bridgeLeft.push(id);
      } else {
        bridgeRight.push(id);
      }
    });

    const depth = new Map();
    const roots = nodes.filter((n) => (incoming.get(n.id) || 0) === 0);
    const queue = roots.map((node) => node.id);
    roots.forEach((node) => depth.set(node.id, 0));
    while (queue.length) {
      const current = queue.shift();
      const currentDepth = depth.get(current) || 0;
      edges
        .filter((edge) => edge.from === current)
        .forEach((edge) => {
          if (depth.has(edge.to)) return;
          depth.set(edge.to, currentDepth + 1);
          queue.push(edge.to);
        });
    }

    let maxDepth = 0;
    depth.forEach((value) => {
      if (value > maxDepth) {
        maxDepth = value;
      }
    });
    if (maxDepth < 1) maxDepth = 1;

    const backgroundByLane = Array.from({ length: laneCount }, () => []);
    background.forEach((id, idx) => {
      const d = depth.has(id) ? (depth.get(id) || 0) : (idx % laneCount);
      const lane = Math.max(0, Math.min(Math.round((d / maxDepth) * (laneCount - 1)), laneCount - 1));
      backgroundByLane[lane].push(id);
    });

    // Keep cause -> impact left-to-right on top row.
    placeInBand(source, laneX(0), topStart, topEnd);
    placeInBand(bridgeLeft, laneX(1), topStart, topEnd);
    placeInBand(bridgeRight, laneX(2), topStart, topEnd);
    placeInBand(impacted, laneX(3), topStart, topEnd);

    // Distribute unaffected services by topology depth to preserve ecosystem legibility.
    backgroundByLane.forEach((ids, lane) => {
      placeInBand(ids, laneX(lane), bottomStart, bottomEnd);
    });

    nodes.forEach((node) => {
      node.edgeDegree = (incoming.get(node.id) || 0) + (outgoing.get(node.id) || 0);
    });
    return positions;
  }

  const roots = nodes.filter((n) => (incoming.get(n.id) || 0) === 0);
  const layers = [];
  const queue = [...roots];
  const depth = new Map();
  queue.forEach((r) => depth.set(r.id, 0));

  while (queue.length) {
    const current = queue.shift();
    const d = depth.get(current.id) || 0;
    layers[d] = layers[d] || [];
    layers[d].push(current.id);
    edges
      .filter((e) => e.from === current.id)
      .forEach((e) => {
        if (!depth.has(e.to)) {
          depth.set(e.to, d + 1);
          const nextNode = nodes.find((n) => n.id === e.to);
          if (nextNode) {
            queue.push(nextNode);
          }
        }
      });
  }

  nodes.forEach((n) => {
    if (!depth.has(n.id)) {
      depth.set(n.id, layers.length);
      layers[layers.length] = layers[layers.length] || [];
      layers[layers.length - 1].push(n.id);
    }
  });

  const columnCount = layers.length || 1;
  const slotCount = Math.max(columnCount, 4);
  const slotWidth = Math.max((width - 40) / slotCount, 120);
  const slotX = (slotIndex) => 20 + slotWidth * slotIndex + slotWidth / 2;

  const placeInBandAcrossSlots = (ids, slots, startY, endY) => {
    if (!ids.length || !slots.length) return;
    const buckets = slots.map(() => []);
    ids.forEach((id, idx) => {
      buckets[idx % slots.length].push(id);
    });
    buckets.forEach((bucket, idx) => {
      if (!bucket.length) return;
      placeInBand(bucket, slotX(slots[idx]), startY, endY);
    });
  };

  layers.forEach((layerNodeIds, colIdx) => {
    let slotStart = Math.floor((colIdx * slotCount) / columnCount);
    let slotEnd = Math.floor(((colIdx + 1) * slotCount) / columnCount) - 1;
    if (slotEnd < slotStart) slotEnd = slotStart;
    if (colIdx === columnCount - 1) slotEnd = slotCount - 1;
    slotStart = Math.max(0, Math.min(slotStart, slotCount - 1));
    slotEnd = Math.max(slotStart, Math.min(slotEnd, slotCount - 1));
    const slots = [];
    for (let s = slotStart; s <= slotEnd; s++) {
      slots.push(s);
    }

    if (focusedNodes.size === 0) {
      placeInBandAcrossSlots(layerNodeIds, slots, 50, height - 40);
      return;
    }

    const key = layerNodeIds.filter((id) => focusedNodes.has(id));
    const nonKey = layerNodeIds.filter((id) => !focusedNodes.has(id));
    placeInBandAcrossSlots(key, slots, topStart, topEnd);
    // Keep non-key/non-affected services compressed in the lower band.
    placeInBandAcrossSlots(nonKey, slots, bottomStart, bottomEnd);
  });

  nodes.forEach((node) => {
    node.edgeDegree = (incoming.get(node.id) || 0) + (outgoing.get(node.id) || 0);
  });
  return positions;
}

function normalizePositions(positions, width, height, padding = {}) {
  const left = Math.max(0, padding.left ?? 30);
  const right = Math.max(0, padding.right ?? 30);
  const top = Math.max(0, padding.top ?? 30);
  const bottom = Math.max(0, padding.bottom ?? 30);
  let minX = Infinity;
  let maxX = -Infinity;
  let minY = Infinity;
  let maxY = -Infinity;

  positions.forEach((pos) => {
    minX = Math.min(minX, pos.x);
    maxX = Math.max(maxX, pos.x);
    minY = Math.min(minY, pos.y);
    maxY = Math.max(maxY, pos.y);
  });

  if (!isFinite(minX) || !isFinite(minY)) {
    return positions;
  }

  const spanX = Math.max(maxX - minX, 1);
  const spanY = Math.max(maxY - minY, 1);
  const drawWidth = Math.max(width - left - right, 1);
  const drawHeight = Math.max(height - top - bottom, 1);
  // Scale each axis to fit the available draw area, shrinking when needed.
  const scaleX = drawWidth / spanX;
  const scaleY = drawHeight / spanY;

  const normalized = new Map();
  positions.forEach((pos, key) => {
    normalized.set(key, {
      x: (pos.x - minX) * scaleX + left,
      y: (pos.y - minY) * scaleY + top,
    });
  });

  return normalized;
}

function addGraphInteractions(svg) {
  // Hover zoom/focus intentionally disabled to keep topology stable.
  void svg;
}

function summarizeFlow(snapshot, nodes, edges, sourceServices, impactedServices, severity, activeFields) {
  if (!nodes.length) return "No services loaded.";
  const incoming = new Map();
  edges.forEach((e) => incoming.set(e.to, (incoming.get(e.to) || 0) + 1));
  const sourceID = sourceServices && sourceServices.size ? [...sourceServices][0] : null;
  let start = (sourceID && nodes.find((n) => n.id === sourceID)) || nodes.find((n) => !incoming.get(n.id)) || nodes[0];
  const path = [start.name];
  const seen = new Set();
  let current = start.id;
  while (true) {
    seen.add(current);
    const nextEdge = edges.find((e) => e.from === current && !seen.has(e.to) && (!activeFields || activeFields.size === 0 || activeFields.has(e.fieldId)));
    if (!nextEdge) break;
    const nextNode = nodes.find((n) => n.id === nextEdge.to);
    if (!nextNode) break;
    path.push(nextNode.name);
    current = nextNode.id;
    if (seen.has(current)) break;
  }
  const sevText = severity ? ` (${severity} severity)` : "";
  const sourceText = sourceID ? `Source changed: ${serviceName(snapshot, sourceID)}.` : "";
  const impactText = impactedServices?.size
    ? `Impacted services: ${listServiceNames(snapshot, [...impactedServices], 4)}.`
    : "No impacted services.";
  return `${sourceText} ${impactText} Flow: ${path.join(" -> ")}${sevText}`;
}

function computeImpacts(snapshot) {
  const sevRank = { high: 3, medium: 2, low: 1, info: 0 };
  const findings = snapshot.findings || [];
  const sorted = [...findings].sort((a, b) => (sevRank[b.severity] || 0) - (sevRank[a.severity] || 0));
  const latest = latestMutation(snapshot);
  let focus = sorted[0];
  if (!focus && latest) {
    focus = { fieldId: latest.fieldId, serviceId: latest.serviceId, severity: "info" };
  }
  const activeFields = new Set(
    findings
      .map((item) => item.fieldId)
      .filter((value) => typeof value === "string" && value.length > 0),
  );
  if (activeFields.size === 0 && focus?.fieldId) {
    activeFields.add(focus.fieldId);
  }
  const sourceServices = inferSourceServices(snapshot, findings, latest || focus);
  const impactedServices = new Set(
    findings
      .map((item) => item.serviceId)
      .filter((value) => typeof value === "string" && value.length > 0),
  );
  if (impactedServices.size === 0 && focus?.serviceId) {
    impactedServices.add(focus.serviceId);
  }
  const failedServices = new Set(
    findings
      .filter((item) => item.severity === "high")
      .map((item) => item.serviceId)
      .filter((value) => typeof value === "string" && value.length > 0),
  );
  const affectedServices = new Set(
    findings
      .map((item) => item.serviceId)
      .filter((value) => typeof value === "string" && value.length > 0),
  );
  if (affectedServices.size === 0 && focus?.serviceId) {
    affectedServices.add(focus.serviceId);
  }
  const flowEdges = new Set();
  const flowNodes = new Set();
  (snapshot.edges || []).forEach((e) => {
    if (activeFields.has(e.fieldId)) {
      flowEdges.add(e.id);
      flowNodes.add(e.from);
      flowNodes.add(e.to);
      affectedServices.add(e.from);
      affectedServices.add(e.to);
    }
  });
  sourceServices.forEach((serviceId) => affectedServices.add(serviceId));
  impactedServices.forEach((serviceId) => affectedServices.add(serviceId));
  return {
    activeFields,
    sourceServices,
    impactedServices,
    failedServices,
    affectedServices,
    flowNodes,
    flowEdges,
    focusFieldSeverity: focus?.severity || null,
  };
}
