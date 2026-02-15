const state = {
  sessionId: "",
  snapshot: null,
  mutationPending: false,
};

const selectors = {
  gateBanner: document.querySelector("#gate-banner"),
  topology: document.querySelector("#topology"),
  topologyGraph: document.querySelector("#topology-graph"),
  edgeList: document.querySelector("#edge-list"),
  flowPathSummary: document.querySelector("#flow-path-summary"),
  findings: document.querySelector("#findings"),
  runSummaryText: document.querySelector("#run-summary-text"),
  controlStats: document.querySelector("#control-stats"),
  scenarioNarrative: document.querySelector("#scenario-narrative"),
  presetScenario: document.querySelector("#preset-scenario"),
  mutationType: document.querySelector("#mutation-type"),
  mutationService: document.querySelector("#mutation-service"),
  mutationField: document.querySelector("#mutation-field"),
  policyMode: document.querySelector("#policy-mode"),
  policyFailOn: document.querySelector("#policy-fail-on"),
  policyCriticalService: document.querySelector("#policy-critical-service"),
  policyPromotionsHardBlock: document.querySelector("#policy-promotions-hard-block"),
  policyHardBlockReason: document.querySelector("#policy-hard-block-reason"),
  overrideTarget: document.querySelector("#override-target"),
  overrideReason: document.querySelector("#override-reason"),
  applyMutation: document.querySelector("#apply-mutation"),
  updatePolicy: document.querySelector("#update-policy"),
  addOverride: document.querySelector("#add-override"),
  runStricture: document.querySelector("#run-stricture"),
  resetSession: document.querySelector("#reset-session"),
  toggleEdges: document.querySelector("#toggle-edges"),
};

const OVERRIDE_DAYS = 7;

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

const mutationTypeLabels = {
  annotation_missing: "Annotation Missing",
  enum_changed: "Enum values updated",
  external_as_of_stale: "External As-Of Stale",
  field_removed: "Field Removed",
  source_version_changed: "Source Version Changed",
  type_changed: "Field type changed",
};

function fieldLabel(fieldId) {
  return String(fieldId || "")
    .replace(/^response_/, "response.")
    .replace(/_/g, ".");
}

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

function parseFindingDelta(finding) {
  const summary = String(finding.summary || "");
  const fromTo = /from\s+([^\s,]+)\s+to\s+([^\s,]+)/i.exec(summary);

  if (finding.changeType === "merge_strategy_changed" && fromTo) {
    const meanings = {
      custom: "custom logic combines sources with service-specific rules",
      single_source: "one source is selected and others are ignored",
      priority: "sources are evaluated in priority order",
      first_non_null: "first non-null source wins",
      union: "values from multiple sources are combined",
    };
    return {
      whatChanged: `Merge strategy changed from ${fromTo[1]} (${meanings[fromTo[1]] || "previous merge behavior"}) to ${fromTo[2]} (${meanings[fromTo[2]] || "new merge behavior"}).`,
      shortDelta: `${fromTo[1]} -> ${fromTo[2]}`,
    };
  }
  if (finding.changeType === "source_version_changed" && fromTo) {
    return { whatChanged: `Source version changed from ${fromTo[1]} to ${fromTo[2]}.`, shortDelta: `${fromTo[1]} -> ${fromTo[2]}` };
  }
  if (finding.changeType === "external_as_of_rollback" && fromTo) {
    return { whatChanged: `External snapshot date moved from ${fromTo[1]} to ${fromTo[2]}.`, shortDelta: `${fromTo[1]} -> ${fromTo[2]}` };
  }
  if (finding.changeType === "field_removed") {
    return { whatChanged: "Field was removed from the lineage contract.", shortDelta: "field removed" };
  }
  if (finding.changeType === "source_removed") {
    return { whatChanged: "A required upstream source mapping was removed.", shortDelta: "source removed" };
  }
  if (finding.changeType === "source_contract_ref_changed") {
    return { whatChanged: "Upstream contract reference changed.", shortDelta: "contract ref changed" };
  }
  return {
    whatChanged: summary ? `${summary}.` : "Contract behavior changed.",
    shortDelta: humanChange[finding.changeType] || finding.changeType,
  };
}

function findingCauseImpact(snapshot, finding, mutation) {
  const inferredSources = inferSourceServices(snapshot, [finding], mutation);
  const sourceId = finding.source?.serviceId
    || (inferredSources.size ? [...inferredSources][0] : "")
    || mutation?.serviceId
    || "";
  const impactId = finding.impact?.serviceId || finding.serviceId || "";
  return {
    sourceName: sourceId ? serviceName(snapshot, sourceId) : "unknown source",
    impactName: impactId ? serviceName(snapshot, impactId) : "unknown impacted service",
  };
}

function findingStory(snapshot, finding, mutation) {
  const { sourceName, impactName } = findingCauseImpact(snapshot, finding, mutation);
  const delta = parseFindingDelta(finding);
  return `${sourceName} changed ${fieldLabel(finding.fieldId)}. ${delta.whatChanged} Stricture flags ${impactName} as impacted by this drift.`;
}

function conciseLeftNarrative(snapshot, finding, mutation) {
  if (finding?.changeType === "enum_changed") {
    return "PromotionConfig added a promotion type and not all downstream services were properly updated. Stricture flags the problematic enum drift and impacted services.";
  }
  return findingStory(snapshot, finding, mutation);
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

function buildFieldMutationMap(snapshot) {
  const map = new Map();
  Object.entries(snapshot.fieldsByMutation || {}).forEach(([mutationType, fields]) => {
    (fields || []).forEach((fieldId) => {
      if (!map.has(fieldId)) {
        map.set(fieldId, []);
      }
      map.get(fieldId).push(mutationType);
    });
  });
  return map;
}

function fieldsForService(snapshot, serviceId, fieldMutationMap) {
  const edges = snapshot.edges || [];
  const fields = new Set();
  edges.forEach((edge) => {
    if (edge.from === serviceId && fieldMutationMap.has(edge.fieldId)) {
      fields.add(edge.fieldId);
    }
  });
  return [...fields];
}

function syncMutationControls(snapshot, preferred = {}) {
  if (!selectors.mutationService || !selectors.mutationField || !selectors.mutationType) {
    return;
  }
  const services = snapshot.services || [];
  const serviceIDs = services.map((service) => service.id);
  const existingService = preferred.serviceId ?? selectors.mutationService.value;
  setOptions(selectors.mutationService, serviceIDs, (id) => id);
  const selectedService = serviceIDs.includes(existingService) ? existingService : (serviceIDs[0] || "");
  selectors.mutationService.value = selectedService;

  const fieldMutationMap = buildFieldMutationMap(snapshot);
  const fieldIDs = fieldsForService(snapshot, selectedService, fieldMutationMap);
  const existingField = preferred.fieldId ?? selectors.mutationField.value;
  setOptions(selectors.mutationField, fieldIDs, (id) => id);
  const selectedField = fieldIDs.includes(existingField) ? existingField : (fieldIDs[0] || "");
  if (selectedField) {
    selectors.mutationField.value = selectedField;
  }

  const allowedTypes = selectedField ? (fieldMutationMap.get(selectedField) || []) : [];
  const existingType = preferred.mutationType ?? selectors.mutationType.value;
  setOptions(selectors.mutationType, allowedTypes, (id) => mutationTypeLabels[id] || id);
  const selectedType = allowedTypes.includes(existingType) ? existingType : (allowedTypes[0] || "");
  if (selectedType) {
    selectors.mutationType.value = selectedType;
  }

  if (selectors.applyMutation) {
    selectors.applyMutation.disabled = !(selectedService && selectedField && selectedType);
  }

  return {
    selectedService,
    selectedField,
    selectedType,
    serviceFieldCount: fieldIDs.length,
    fieldTypeCount: allowedTypes.length,
  };
}

function updateControlStats(snapshot) {
  if (!selectors.controlStats || !selectors.mutationService || !selectors.mutationField) {
    return;
  }
  const fieldMutationMap = buildFieldMutationMap(snapshot);
  const serviceId = selectors.mutationService.value;
  const fieldId = selectors.mutationField.value;
  const serviceFieldCount = fieldsForService(snapshot, serviceId, fieldMutationMap).length;
  const fieldTypeCount = (fieldMutationMap.get(fieldId) || []).length;
  const serviceLabel = serviceName(snapshot, serviceId);
  const stagedCount = snapshot.mutations?.length || 0;
  selectors.controlStats.textContent = `${serviceLabel} • ${serviceFieldCount} fields • ${fieldTypeCount} available changes for selected field • ${stagedCount} staged change${stagedCount === 1 ? "" : "s"}`;
}

function currentTopFinding(snapshot) {
  if (!snapshot?.findings?.length) {
    return null;
  }
  return snapshot.findings[0];
}

function datePlusDays(days) {
  const next = new Date();
  next.setDate(next.getDate() + days);
  return next.toISOString().slice(0, 10);
}

function updateOverrideControls(snapshot) {
  const top = currentTopFinding(snapshot);
  if (selectors.overrideTarget) {
    if (!top) {
      selectors.overrideTarget.textContent = "No active finding to suppress yet. Run Stricture after a change.";
    } else {
      const change = humanChange[top.changeType] || top.changeType;
      selectors.overrideTarget.textContent = `Will suppress: ${change} on ${fieldLabel(top.fieldId)} for ${OVERRIDE_DAYS} days.`;
    }
  }
  if (selectors.addOverride) {
    selectors.addOverride.disabled = !top;
  }
}

function updatePolicyControls(snapshot) {
  if (!selectors.updatePolicy) {
    return;
  }
  const stagedCount = snapshot.mutations?.length || 0;
  const enabled = stagedCount > 0;
  selectors.updatePolicy.disabled = !enabled;
  selectors.updatePolicy.title = enabled
    ? "Apply current policy settings to staged changes."
    : "Stage at least one change before applying policy.";
}

function relatedServiceIdsForFinding(snapshot, finding) {
  const related = new Set();
  if (finding?.serviceId) {
    related.add(finding.serviceId);
  }
  if (finding?.source?.serviceId) {
    related.add(finding.source.serviceId);
  }
  if (finding?.impact?.serviceId) {
    related.add(finding.impact.serviceId);
  }
  (snapshot.edges || []).forEach((edge) => {
    if (edge.fieldId !== finding.fieldId) return;
    related.add(edge.from);
    related.add(edge.to);
  });
  return [...related].filter(Boolean);
}

async function renderFindingEscalation(snapshot, findingId, button, container) {
  const finding = (snapshot.findings || []).find((row) => row.id === findingId);
  if (!finding) {
    container.textContent = "Escalation details unavailable for this finding.";
    return;
  }

  const currentlyHidden = container.classList.contains("is-collapsed");
  if (!currentlyHidden && container.dataset.loaded === "true") {
    container.classList.add("is-collapsed");
    button.textContent = "Show escalation chain";
    return;
  }
  if (container.dataset.loaded === "true") {
    container.classList.remove("is-collapsed");
    button.textContent = "Hide escalation chain";
    return;
  }

  button.disabled = true;
  button.textContent = "Loading escalation...";

  const serviceIDs = relatedServiceIdsForFinding(snapshot, finding);
  const rows = await Promise.all(serviceIDs.map(async (serviceId) => {
    try {
      const payload = await request(`/api/session/${state.sessionId}/escalation?serviceId=${encodeURIComponent(serviceId)}`);
      return { serviceId, chain: payload.chain || [] };
    } catch {
      return { serviceId, chain: [], failed: true };
    }
  }));

  const services = serviceById(snapshot);
  const cards = rows.map((row) => {
    const service = services.get(row.serviceId);
    const serviceNameLabel = service?.name || row.serviceId;
    const owner = service?.owner || "unknown owner";
    const primary = service?.escalation || "n/a";
    const chainText = row.failed
      ? "Chain lookup failed."
      : row.chain.length
        ? row.chain.map((step) => `L${step.depth}: ${step.system_id} (${contactsToText(step.contacts)})`).join(" -> ")
        : "No additional chain configured.";
    return `
      <article class="finding-escalation-item">
        <h4>${serviceNameLabel}</h4>
        <p class="item-meta">Primary on-call: ${primary}</p>
        <p class="item-meta">Owner: ${owner}</p>
        <p class="item-meta">Chain: ${chainText}</p>
      </article>
    `;
  }).join("");

  container.innerHTML = `
    <p class="item-meta finding-escalation-title">Escalation contacts for services on this affected flow:</p>
    <div class="finding-escalation-grid">${cards}</div>
  `;
  container.dataset.loaded = "true";
  container.classList.remove("is-collapsed");
  button.disabled = false;
  button.textContent = "Hide escalation chain";
}

function bindFindingEscalationButtons(snapshot) {
  if (!selectors.findings) {
    return;
  }
  selectors.findings.querySelectorAll(".finding-escalation-toggle").forEach((button) => {
    button.addEventListener("click", () => {
      const encoded = button.getAttribute("data-finding-id") || "";
      const findingID = decodeURIComponent(encoded);
      const container = selectors.findings.querySelector(`[data-escalation-for="${encoded}"]`);
      if (!container) {
        return;
      }
      renderFindingEscalation(snapshot, findingID, button, container).catch(showError);
    });
  });
}

function buildPresets(snapshot) {
  const nonePreset = { id: "__none", label: "No scenario (baseline)" };
  const candidates = [
    { id: "enum_changed", label: "Payments promotion drift" },
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
    enum_changed: "Payments promotion drift: PromotionsConfig introduced a new promotion type. PromotionsApplication forwards it, but payment services still only handle older enum values, so promotion-applied checkouts can fail.",
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
  const mutation = latestMutation(snapshot);
  if (!findings.length) {
    if (snapshot.runSummary?.runCount > 0 && mutation) {
      selectors.scenarioNarrative.textContent = `No downstream impact found for ${fieldLabel(mutation.fieldId)}. Stricture tracked the mutation but did not flag a finding under current policy.`;
    }
    return;
  }
  const top = findings[0];
  selectors.scenarioNarrative.textContent = conciseLeftNarrative(snapshot, top, mutation);
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
    selectors.runSummaryText.textContent = "Awaiting first run. Apply a change to run Stricture instantly and see impact.";
    return;
  }
  const mutation = latestMutation(snapshot);
  if (!summary.findingCount) {
    selectors.runSummaryText.textContent = `Run #${summary.runCount}: no downstream-impact findings. Deploy allowed. Stricture tracked the change but found no affected downstream service.`;
    return;
  }
  const top = snapshot.findings[0];
  const delta = parseFindingDelta(top);
  const { sourceName, impactName } = findingCauseImpact(snapshot, top, mutation);
  const gateText = summary.gate === "BLOCK" ? "Deploy blocked." : "Deploy allowed.";
  const policyConfig = snapshot.policy?.hardBlockPromotionsDrift
    ? `Policy: ${snapshot.policy.mode}/${snapshot.policy.failOn}+ with promotions hard-block on ${serviceName(snapshot, snapshot.policy.criticalServiceId)}.`
    : `Policy: ${snapshot.policy.mode}/${snapshot.policy.failOn}+.`;
  const stagedText = ` Staged changes: ${snapshot.mutations?.length || 0}.`;
  const policyText = summary.policyRationale ? ` ${summary.policyRationale}` : "";
  selectors.runSummaryText.textContent = `Run #${summary.runCount}: ${summary.findingCount} finding (${summary.blockedCount} blocking, ${summary.warningCount} warning). ${gateText} ${policyConfig}${stagedText}${policyText} Top issue: ${humanChange[top.changeType] || top.changeType} on ${fieldLabel(top.fieldId)} (${delta.shortDelta}). Cause: ${sourceName}. Impact: ${impactName}.`;
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

  renderList(
    selectors.findings,
    snapshot.findings,
    "No findings yet.",
    {
      className: (finding) => `list-item sev-${finding.severity}`,
      html: (finding) => {
        const mutation = latestMutation(snapshot);
        const { sourceName: source, impactName: impacted } = findingCauseImpact(snapshot, finding, mutation);
        const delta = parseFindingDelta(finding);
        const owner = serviceOwner(snapshot, finding.serviceId);
        const flowServices = (snapshot.edges || [])
          .filter((edge) => edge.fieldId === finding.fieldId)
          .flatMap((edge) => [edge.from, edge.to]);
        const blastRadius = listServiceNames(snapshot, [...flowServices, finding.serviceId], 4);
        const encodedFindingID = encodeURIComponent(finding.id);
        return `
        <h3>${humanChange[finding.changeType] || finding.changeType} — ${fieldLabel(finding.fieldId)} (${finding.severity.toUpperCase()}) in ${impacted}</h3>
        <p><strong>What changed:</strong> ${delta.whatChanged}</p>
        <p><strong>Why flagged:</strong> ${source} changed the producer contract/path and ${impacted} is downstream on this flow.</p>
        <p class="item-meta chips">
          <span class="chip">Cause: ${source}</span>
          <span class="chip">Impact: ${impacted}</span>
          <span class="chip">Blast radius: ${blastRadius}</span>
          <span class="chip">Owner: ${owner}</span>
        </p>
        ${finding.validation ? `<p class="item-meta">Validation: ${finding.validation}</p>` : ""}
        ${finding.suggestion ? `<p class="item-meta">Suggestion: ${finding.suggestion}</p>` : ""}
        ${finding.policyRationale ? `<p class="item-meta"><strong>Policy:</strong> Hard block (${finding.policyRationale})</p>` : ""}
        <p class="item-meta">Remediation: ${finding.remediation}</p>
        <button type="button" class="button button-secondary button-ghost finding-escalation-toggle" data-finding-id="${encodedFindingID}">Show escalation chain</button>
        <div class="finding-escalation is-collapsed" data-escalation-for="${encodedFindingID}"></div>
      `;
      },
    },
  );

  syncMutationControls(snapshot);
  updateControlStats(snapshot);
  updatePolicyControls(snapshot);
  updateOverrideControls(snapshot);
  bindFindingEscalationButtons(snapshot);
  buildPresets(snapshot);
  updateNarrativeFromSnapshot(snapshot);

  if (selectors.policyMode) {
    selectors.policyMode.value = snapshot.policy.mode;
  }
  if (selectors.policyFailOn) {
    selectors.policyFailOn.value = snapshot.policy.failOn;
  }
  if (selectors.policyCriticalService) {
    const serviceIDs = snapshot.services.map((service) => service.id);
    setOptions(selectors.policyCriticalService, serviceIDs, (id) => serviceName(snapshot, id));
    if (serviceIDs.includes(snapshot.policy.criticalServiceId)) {
      selectors.policyCriticalService.value = snapshot.policy.criticalServiceId;
    }
  }
  if (selectors.policyPromotionsHardBlock) {
    selectors.policyPromotionsHardBlock.checked = !!snapshot.policy.hardBlockPromotionsDrift;
  }
  if (selectors.policyHardBlockReason) {
    selectors.policyHardBlockReason.value = snapshot.policy.hardBlockReason || "";
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
  if (!state.sessionId || state.mutationPending) {
    return;
  }
  state.mutationPending = true;
  if (selectors.applyMutation) {
    selectors.applyMutation.disabled = true;
  }
  const payload = {
    type: selectors.mutationType?.value,
    serviceId: selectors.mutationService?.value,
    fieldId: selectors.mutationField?.value,
  };
  try {
    await request(`/api/session/${state.sessionId}/mutations`, "POST", payload);
    const rerun = await request(`/api/session/${state.sessionId}/run`, "POST", {});
    render(rerun.snapshot);
    revealFindings();
  } finally {
    state.mutationPending = false;
    if (state.snapshot) {
      syncMutationControls(state.snapshot);
      updateControlStats(state.snapshot);
    } else if (selectors.applyMutation) {
      selectors.applyMutation.disabled = false;
    }
  }
}

async function updatePolicy() {
  if (!state.sessionId) {
    return;
  }
  if ((state.snapshot?.mutations?.length || 0) === 0) {
    return;
  }
  const payload = {
    mode: selectors.policyMode?.value,
    failOn: selectors.policyFailOn?.value,
    criticalServiceId: selectors.policyCriticalService?.value,
    hardBlockPromotionsDrift: selectors.policyPromotionsHardBlock?.checked,
    hardBlockReason: selectors.policyHardBlockReason?.value.trim(),
  };
  const response = await request(`/api/session/${state.sessionId}/policy`, "POST", payload);
  render(response.snapshot);
}

async function addOverride() {
  if (!state.sessionId) {
    return;
  }
  const top = currentTopFinding(state.snapshot);
  if (!top) {
    throw new Error("No active finding available to suppress.");
  }
  const reasonInput = selectors.overrideReason?.value.trim() || "temporary mitigation window";
  const payload = {
    fieldId: top.fieldId,
    changeType: top.changeType,
    expires: datePlusDays(OVERRIDE_DAYS),
    reason: reasonInput,
  };
  const response = await request(`/api/session/${state.sessionId}/override`, "POST", payload);
  if (selectors.overrideReason) {
    selectors.overrideReason.value = "";
  }
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

function bindEvents() {
  selectors.applyMutation?.addEventListener("click", () => mutate().catch(showError));
  selectors.updatePolicy?.addEventListener("click", () => updatePolicy().catch(showError));
  selectors.addOverride?.addEventListener("click", () => addOverride().catch(showError));
  selectors.runStricture?.addEventListener("click", () => run().catch(showError));
  selectors.resetSession?.addEventListener("click", () => bootstrap().catch(showError));
  selectors.presetScenario?.addEventListener("change", (event) => {
    const target = event.target;
    updateNarrative(target?.value);
    applyPreset().catch(showError);
  });
  selectors.toggleEdges?.addEventListener("click", () => toggleEdgeList());
  selectors.mutationService?.addEventListener("change", () => {
    if (state.snapshot) {
      syncMutationControls(state.snapshot, { serviceId: selectors.mutationService?.value });
      updateControlStats(state.snapshot);
    }
  });
  selectors.mutationField?.addEventListener("change", () => {
    if (state.snapshot) {
      syncMutationControls(state.snapshot, {
        serviceId: selectors.mutationService?.value,
        fieldId: selectors.mutationField?.value,
      });
      updateControlStats(state.snapshot);
    }
  });
  selectors.mutationType?.addEventListener("change", () => {
    if (state.snapshot) {
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
  const candidateSource = (state.snapshot.edges || []).find((edge) => edge.fieldId === fieldId)?.from;
  updateNarrative(id);
  await bootstrap(); // reset session so findings don’t pile up
  syncMutationControls(state.snapshot, {
    serviceId: candidateSource,
    fieldId,
    mutationType: id,
  });
  updateControlStats(state.snapshot);
  await mutate();
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
    pathNodes,
    pathEdges,
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
    const isSourceImpactEdge = pathEdges.has(edge.id) ||
      (isFlow && (sourceServices.has(edge.from) || sourceServices.has(edge.to)) &&
      (impactedServices.has(edge.from) || impactedServices.has(edge.to) || isActiveField));
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
    const inPath = pathNodes.has(node.id);
    const isSource = sourceServices.has(node.id);
    const isImpacted = impactedServices.has(node.id);
    const isTransit = inPath && !isSource && !isImpacted;
    const isContributor = inFlow && !inPath && !isSource && !isImpacted;
    const isFailing = failedServices.has(node.id);
    const classes = ["graph-node", statusClass];
    if (isolated) classes.push("isolated");
    if (!inFlow && !isSource && !isImpacted) classes.push("dimmed");
    if (isSource) classes.push("source");
    if (isImpacted) classes.push("impacted");
    if (isTransit) classes.push("transit");
    if (isContributor) classes.push("contributor");
    if (isFailing) classes.push("failing");
    g.setAttribute("class", classes.join(" "));
    const circle = document.createElementNS(svgNS, "circle");
    circle.setAttribute("cx", pos.x);
    circle.setAttribute("cy", pos.y);
    circle.setAttribute("r", isSource || isImpacted ? "28" : isTransit ? "24" : isContributor ? "20" : inFlow ? "22" : "16");
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
  const activeFieldEdges = [];
  (snapshot.edges || []).forEach((e) => {
    if (activeFields.has(e.fieldId)) {
      activeFieldEdges.push(e);
      flowEdges.add(e.id);
      flowNodes.add(e.from);
      flowNodes.add(e.to);
      affectedServices.add(e.from);
      affectedServices.add(e.to);
    }
  });
  const outgoing = new Map();
  activeFieldEdges.forEach((edge) => {
    const list = outgoing.get(edge.from) || [];
    list.push(edge);
    outgoing.set(edge.from, list);
  });

  const pathNodes = new Set();
  const pathEdges = new Set();
  const shortestPath = (start, goal) => {
    if (!start || !goal) return null;
    if (start === goal) return { nodes: [start], edges: [] };
    const queue = [start];
    const seen = new Set([start]);
    const prevNode = new Map();
    const prevEdge = new Map();
    while (queue.length) {
      const current = queue.shift();
      const nextEdges = outgoing.get(current) || [];
      for (const edge of nextEdges) {
        if (seen.has(edge.to)) continue;
        seen.add(edge.to);
        prevNode.set(edge.to, current);
        prevEdge.set(edge.to, edge.id);
        if (edge.to === goal) {
          const nodes = [];
          const edges = [];
          let node = goal;
          nodes.push(node);
          while (prevNode.has(node)) {
            edges.push(prevEdge.get(node));
            node = prevNode.get(node);
            nodes.push(node);
          }
          nodes.reverse();
          edges.reverse();
          return { nodes, edges };
        }
        queue.push(edge.to);
      }
    }
    return null;
  };

  sourceServices.forEach((sourceID) => {
    impactedServices.forEach((impactID) => {
      const path = shortestPath(sourceID, impactID);
      if (!path) return;
      path.nodes.forEach((id) => pathNodes.add(id));
      path.edges.forEach((id) => pathEdges.add(id));
    });
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
    pathNodes,
    pathEdges,
    focusFieldSeverity: focus?.severity || null,
  };
}
