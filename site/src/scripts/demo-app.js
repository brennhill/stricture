const state = {
  sessionId: "",
  snapshot: null,
};

const selectors = {
  gateBanner: document.querySelector("#gate-banner"),
  topology: document.querySelector("#topology"),
  edgeList: document.querySelector("#edge-list"),
  findings: document.querySelector("#findings"),
  overrides: document.querySelector("#overrides"),
  escalation: document.querySelector("#escalation"),
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

function renderGate(summary) {
  if (!selectors.gateBanner || !summary) {
    return;
  }
  selectors.gateBanner.classList.remove("gate-ok", "gate-warn", "gate-block");
  const cls = summary.gate === "BLOCK" ? "gate-block" : summary.findingCount > 0 ? "gate-warn" : "gate-ok";
  selectors.gateBanner.classList.add(cls);
  selectors.gateBanner.textContent = `Run #${summary.runCount} | ${summary.gate} | findings=${summary.findingCount} | mode=${summary.mode}`;
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
      item.textContent = `${edge.from} -> ${edge.to} | ${edge.fieldId} | status=${edge.status}`;
      selectors.edgeList.appendChild(item);
    });
  }
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
  renderTopology(snapshot);

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
  selectors.mutationType?.addEventListener("change", () => {
    if (state.snapshot) {
      refreshMutationFieldOptions(state.snapshot);
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

bindEvents();
bootstrap().catch(showError);
