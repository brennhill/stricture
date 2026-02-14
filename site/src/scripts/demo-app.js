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
  applyScenario: document.querySelector("#apply-scenario"),
  toggleEdges: document.querySelector("#toggle-edges"),
};

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
  const candidates = [
    { id: "enum_changed", label: "Payments enum drift" },
    { id: "type_changed", label: "Orders type drift" },
    { id: "external_as_of_stale", label: "External as-of stale" },
    { id: "annotation_missing", label: "Missing annotation" },
  ];
  const presets = candidates
    .map((item) => {
      const fields = snapshot.fieldsByMutation?.[item.id] || [];
      return fields.length ? { ...item, fieldId: fields[0] } : null;
    })
    .filter(Boolean);
  if (selectors.presetScenario) {
    setOptions(
      selectors.presetScenario,
      presets.map((p) => p.id),
      (id) => presets.find((p) => p.id === id)?.label || id,
    );
    if (presets.length) {
      selectors.presetScenario.value = presets[0].id;
      updateNarrative(presets[0].id);
    }
  }
  return presets;
}

function updateNarrative(presetId) {
  if (!selectors.scenarioNarrative) {
    return;
  }
  const narratives = {
    enum_changed: "Payments enum drift: PSP added a new status value without bumping contract. Downstream billing and notifications may misclassify payments until code updates ship.",
    type_changed: "Orders type drift: fulfillment switched quantity from int to string for partial units. Legacy consumers treat it as numeric and will fail parsing.",
    external_as_of_stale: "External as-of stale: vendor shipping ETA feed is older than allowed freshness window, so SLAs and alerts rely on outdated data.",
    annotation_missing: "Missing annotation: a new field shipped without lineage metadata; Stricture blocks because provenance and owners are unknown.",
  };
  selectors.scenarioNarrative.textContent = narratives[presetId] || "Choose a preset to see the drift story and impact.";
}

function renderGate(summary) {
  if (!selectors.gateBanner || !summary) {
    return;
  }
  selectors.gateBanner.classList.remove("gate-ok", "gate-warn", "gate-block");
  const cls = summary.gate === "BLOCK" ? "gate-block" : summary.findingCount > 0 ? "gate-warn" : "gate-ok";
  selectors.gateBanner.classList.add(cls);
  selectors.gateBanner.textContent = `Run #${summary.runCount} | ${summary.gate} | findings=${summary.findingCount} | mode=${summary.mode}`;
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
  const gatePhrase = summary.gate === "BLOCK" ? "blocked deploy" : "passed gate";
  const detail = `policy=${summary.mode}/${snapshot.policy.failOn}, high=${summary.blockedCount}, warn=${summary.warningCount}`;
  selectors.runSummaryText.textContent = `Run #${summary.runCount} ${gatePhrase}: ${summary.findingCount} findings (${detail}). Stricture flags severity from the mutation scenario and applies policy thresholds; high-severity findings block when policy=block and severity ≥ failOn.`;
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

  renderList(
    selectors.findings,
    snapshot.findings,
    "No findings yet.",
    {
      className: (finding) => `list-item sev-${finding.severity}`,
      html: (finding) => `
        <h3>${finding.changeType} (${finding.severity})</h3>
        <p>${finding.summary}</p>
        <p class="item-meta">service=${finding.serviceId} field=${finding.fieldId}</p>
        <p class="item-meta">Fix: ${finding.remediation}</p>
      `,
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
  selectors.applyScenario?.addEventListener("click", () => applyPreset().catch(showError));
  selectors.presetScenario?.addEventListener("change", (event) => {
    const target = event.target;
    updateNarrative(target?.value);
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
    selectors.mutationService.value = selectors.mutationService.options[0].value;
  }
  updateNarrative(id);
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

function renderGraph(snapshot) {
  if (!selectors.topologyGraph) {
    return;
  }
  const container = selectors.topologyGraph;
  container.innerHTML = "";
  const label = document.createElement("div");
  label.className = "graph-label";
  label.textContent = "Animated arrows show field flow direction";
  container.appendChild(label);
  const svgNS = "http://www.w3.org/2000/svg";
  const { width: boxWidth } = container.getBoundingClientRect();
  const width = Math.max(boxWidth || 640, 1100);
  const height = Math.max(420, Math.ceil((snapshot.services?.length || 8) / 4) * 120);
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
  const positions = computeLayeredLayout(nodes, edges, width, height);

  const { activeFields, sourceServices, flowNodes, flowEdges, focusFieldSeverity } = computeImpacts(snapshot);

  const edgeColor = (status) => {
    if (status === "blocked") return "var(--danger)";
    if (status === "warning") return "var(--warn)";
    if (status === "healthy") return "var(--ok)";
    return "var(--line)";
  };

  edges.forEach((edge) => {
    const from = positions.get(edge.from);
    const to = positions.get(edge.to);
    if (!from || !to) return;
    const curve = document.createElementNS(svgNS, "path");
    const midX = (from.x + to.x) / 2;
    const midY = (from.y + to.y) / 2 - 18;
    const d = `M ${from.x} ${from.y} Q ${midX} ${midY} ${to.x} ${to.y}`;
    curve.setAttribute("d", d);
    const isFlow = flowEdges.has(edge.id);
    const isActiveField = activeFields.has(edge.fieldId);
    const color = isActiveField ? "var(--accent-2)" : isFlow ? "#2563eb" : edgeColor(edge.status);
    curve.setAttribute("stroke", color);
    const classes = [`graph-edge`, `edge-${edge.status}`, "flow"];
    if (!isFlow) classes.push("dimmed");
    if (isActiveField) classes.push("focus");
    curve.setAttribute("class", classes.join(" "));
    const title = document.createElementNS(svgNS, "title");
    title.textContent = `${edge.from} → ${edge.to} • ${edge.fieldId} • ${edge.status}`;
    curve.appendChild(title);
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
    const classes = ["graph-node", statusClass];
    if (isolated) classes.push("isolated");
    if (!inFlow) classes.push("dimmed");
    if (isSource) classes.push("source", "focus");
    g.setAttribute("class", classes.join(" "));
    const circle = document.createElementNS(svgNS, "circle");
    circle.setAttribute("cx", pos.x);
    circle.setAttribute("cy", pos.y);
    circle.setAttribute("r", inFlow ? "26" : "18");
    g.appendChild(circle);
    const label = document.createElementNS(svgNS, "text");
    label.setAttribute("x", pos.x);
    label.setAttribute("y", pos.y + 34);
    label.textContent = node.name;
    const title = document.createElementNS(svgNS, "title");
    title.textContent = `${node.name} (${node.domain}) • owner=${node.owner}`;
    g.appendChild(title);
    g.appendChild(label);
    svg.appendChild(g);
  });

  container.appendChild(svg);

  if (selectors.flowPathSummary) {
    selectors.flowPathSummary.textContent = summarizeFlow(nodes, edges, focusFieldSeverity, activeFields);
  }
}

function computeLayeredLayout(nodes, edges, width, height) {
  const positions = new Map();
  const incoming = new Map();
  const outgoing = new Map();
  nodes.forEach((n) => {
    incoming.set(n.id, 0);
    outgoing.set(n.id, 0);
  });
  edges.forEach((e) => {
    incoming.set(e.to, (incoming.get(e.to) || 0) + 1);
    outgoing.set(e.from, (outgoing.get(e.from) || 0) + 1);
  });
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
  const colWidth = Math.max((width - 40) / columnCount, 140);
  layers.forEach((layerNodeIds, colIdx) => {
    const rowCount = layerNodeIds.length;
    const rowGap = Math.max((height - 80) / (rowCount || 1), 70);
    layerNodeIds.forEach((nodeId, rowIdx) => {
      const x = 20 + colWidth * colIdx + colWidth / 2;
      const y = 50 + rowGap * rowIdx;
      positions.set(nodeId, { x, y });
      const node = nodes.find((n) => n.id === nodeId);
      if (node) {
        node.edgeDegree = (incoming.get(nodeId) || 0) + (outgoing.get(nodeId) || 0);
      }
    });
  });
  return positions;
}

function summarizeFlow(nodes, edges, severity, activeFields) {
  if (!nodes.length) return "No services loaded.";
  const incoming = new Map();
  edges.forEach((e) => incoming.set(e.to, (incoming.get(e.to) || 0) + 1));
  let start = nodes.find((n) => !incoming.get(n.id)) || nodes[0];
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
  const sevText = severity ? ` | severity=${severity}` : "";
  return `Sample flow: ${path.join(" → ")}${sevText}`;
}

function computeImpacts(snapshot) {
  const sevRank = { high: 3, medium: 2, low: 1, info: 0 };
  const findings = snapshot.findings || [];
  const sorted = [...findings].sort((a, b) => (sevRank[b.severity] || 0) - (sevRank[a.severity] || 0));
  let focus = sorted[0];
  if (!focus && snapshot.mutations?.length) {
    const last = snapshot.mutations[snapshot.mutations.length - 1];
    focus = { fieldId: last.fieldId, serviceId: last.serviceId, severity: "info" };
  }
  const activeFields = new Set(focus ? [focus.fieldId] : []);
  const sourceServices = new Set(focus && focus.serviceId ? [focus.serviceId] : []);
  const flowEdges = new Set();
  const flowNodes = new Set();
  (snapshot.edges || []).forEach((e) => {
    if (activeFields.has(e.fieldId)) {
      flowEdges.add(e.id);
      flowNodes.add(e.from);
      flowNodes.add(e.to);
    }
  });
  return {
    activeFields,
    sourceServices,
    flowNodes,
    flowEdges,
    focusFieldSeverity: focus?.severity || null,
  };
}
