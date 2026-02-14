const state = {
  sessionId: "",
  snapshot: null,
};

const selectors = {
  gateBanner: document.querySelector("#gate-banner"),
  topology: document.querySelector("#topology"),
  topologyGraph: document.querySelector("#topology-graph"),
  edgeList: document.querySelector("#edge-list"),
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

bindEvents();
bootstrap().catch(showError);

function renderGraph(snapshot) {
  if (!selectors.topologyGraph) {
    return;
  }
  const container = selectors.topologyGraph;
  container.innerHTML = "";
  const svgNS = "http://www.w3.org/2000/svg";
  const { width: boxWidth } = container.getBoundingClientRect();
  const width = boxWidth || 640;
  const height = 320;
  const svg = document.createElementNS(svgNS, "svg");
  svg.setAttribute("viewBox", `0 0 ${width} ${height}`);

  const nodes = snapshot.services || [];
  const edges = snapshot.edges || [];
  const centerX = width / 2;
  const centerY = height / 2;
  const radius = Math.max(Math.min(width, height) / 2 - 60, 120);

  const positions = new Map();
  nodes.forEach((node, idx) => {
    const angle = (2 * Math.PI * idx) / Math.max(nodes.length, 1) - Math.PI / 2;
    const x = centerX + radius * Math.cos(angle);
    const y = centerY + radius * Math.sin(angle);
    positions.set(node.id, { x, y });
  });

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
    const line = document.createElementNS(svgNS, "line");
    line.setAttribute("x1", from.x);
    line.setAttribute("y1", from.y);
    line.setAttribute("x2", to.x);
    line.setAttribute("y2", to.y);
    line.setAttribute("class", `graph-edge edge-${edge.status}`);
    line.setAttribute("stroke", edgeColor(edge.status));
    const title = document.createElementNS(svgNS, "title");
    title.textContent = `${edge.from} → ${edge.to} • ${edge.fieldId} • ${edge.status}`;
    line.appendChild(title);
    svg.appendChild(line);
  });

  nodes.forEach((node) => {
    const pos = positions.get(node.id);
    if (!pos) return;
    const g = document.createElementNS(svgNS, "g");
    g.setAttribute("class", `graph-node ${nodeStatusForService(node.id, snapshot.findings)} `);
    const circle = document.createElementNS(svgNS, "circle");
    circle.setAttribute("cx", pos.x);
    circle.setAttribute("cy", pos.y);
    circle.setAttribute("r", "22");
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
}
