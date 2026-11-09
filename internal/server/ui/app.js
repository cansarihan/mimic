"use strict";

const $ = (id) => document.getElementById(id);
const token = new URLSearchParams(location.search).get("token") || "";
const api = (p) => fetch(token ? p + (p.includes("?") ? "&" : "?") + "token=" + encodeURIComponent(token) : p).then((r) => r.json());
const esc = (s) => (s || "").replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
const base = (p) => (p || "").split("/").pop();

let findingsCache = null;
let functionsCache = [];

function route(page) {
  document.querySelectorAll(".nav").forEach((b) => b.classList.toggle("active", b.dataset.page === page));
  document.querySelectorAll(".page").forEach((p) => p.classList.toggle("hidden", p.id !== "page-" + page));
  $("crumb").textContent = page.charAt(0).toUpperCase() + page.slice(1);
  if (page === "findings") loadFindings();
  if (page === "rules") loadRules();
  if (page === "siblings") loadSiblings();
}

async function loadOverview() {
  const st = await api("/api/v1/status");
  $("target").textContent = (st.dir || "") + " · mimic " + (st.version || "");
  $("side-foot").textContent = (st.platform || "") + "\n" + (st.go_version || "");
  $("t-functions").textContent = st.functions || 0;
  $("t-findings").textContent = st.findings || 0;
  $("t-rules").textContent = st.rules || 0;

  const rep = await api("/api/v1/findings");
  findingsCache = rep;
  $("t-errors").textContent = (rep.by_severity && rep.by_severity.error) || 0;
  $("t-warnings").textContent = (rep.by_severity && rep.by_severity.warning) || 0;

  const counts = rep.counts || {};
  const max = Math.max(1, ...Object.values(counts));
  const bars = $("by-rule");
  bars.innerHTML = Object.keys(counts).sort((a, b) => counts[b] - counts[a]).map((rule) =>
    '<div class="bar-row"><span>' + esc(rule) + '</span><div class="bar-track"><div class="bar-fill" style="width:' +
    Math.round((counts[rule] / max) * 100) + '%"></div></div><span>' + counts[rule] + '</span></div>').join("") ||
    '<div class="muted" style="padding:6px 2px">no findings</div>';

  const recent = $("recent");
  recent.innerHTML = (rep.findings || []).slice(0, 8).map((f) =>
    '<li><span class="sev ' + f.severity + '">' + f.severity + '</span>' + esc(f.func) + ' — ' + esc(f.rule) + '</li>').join("") ||
    '<li class="muted">nothing to show</li>';
}

async function loadFindings() {
  if (!findingsCache) findingsCache = await api("/api/v1/findings");
  renderFindings();
}

function renderFindings() {
  const sev = $("sev-filter").value;
  const rows = (findingsCache.findings || []).filter((f) => !sev || f.severity === sev);
  $("findings-count").textContent = rows.length + " finding(s)";
  $("findings-body").innerHTML = rows.map((f) =>
    '<tr><td><span class="sev ' + f.severity + '">' + f.severity + '</span></td>' +
    '<td><span class="tag">' + esc(f.rule) + '</span></td>' +
    '<td>' + esc(f.func) + '</td>' +
    '<td class="muted">' + esc(base(f.file)) + ':' + f.line + '</td>' +
    '<td>' + esc(f.message) + '</td></tr>').join("") ||
    '<tr><td colspan="5" class="muted" style="padding:20px;text-align:center">no findings</td></tr>';
}

async function loadRules() {
  const rules = await api("/api/v1/rules");
  $("rules-cards").innerHTML = rules.map((r) =>
    '<div class="rule-card"><h3>' + esc(r.name) + '</h3><p>' + esc(r.describe) + '</p></div>').join("");
}

function init() {
  document.querySelectorAll(".nav").forEach((b) => b.addEventListener("click", () => route(b.dataset.page)));
  $("sev-filter").addEventListener("change", renderFindings);
  loadOverview();
}

init();
