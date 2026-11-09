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

let siblingState = { ref: "", matches: [] };

async function loadSiblings() {
  if (functionsCache.length === 0) {
    functionsCache = await api("/api/v1/functions");
    const sel = $("func-select");
    sel.innerHTML = functionsCache.map((f) => '<option value="' + esc(f.name) + '">' + esc(f.name) + '</option>').join("");
  }
  runSiblings();
}

async function runSiblings() {
  const ref = $("func-select").value;
  if (!ref) return;
  const min = (parseInt($("min").value, 10) / 100).toFixed(2);
  const rep = await api("/api/v1/like?ref=" + encodeURIComponent(ref) + "&min=" + min);
  siblingState = { ref: rep.reference || ref, matches: rep.matches || [] };
  renderSiblingList();
  drawConstellation();
}

function renderSiblingList() {
  const ol = $("sibling-list");
  if (siblingState.matches.length === 0) {
    ol.innerHTML = '<li class="muted" style="padding:20px;text-align:center">no siblings above the threshold</li>';
    return;
  }
  ol.innerHTML = siblingState.matches.map((m) =>
    '<li><span>' + esc(m.func) + '</span><span>' + Math.round(m.similarity * 100) + '%</span>' +
    '<span class="muted" style="grid-column:1/3;font-size:10px">' + esc(base(m.file)) + ':' + m.line + '</span>' +
    '<div class="simbar"><i style="width:' + Math.round(m.similarity * 100) + '%"></i></div></li>').join("");
}

function drawConstellation() {
  const canvas = $("constellation");
  const ctx = canvas.getContext("2d");
  const dpr = Math.max(1, window.devicePixelRatio || 1);
  const r = canvas.getBoundingClientRect();
  canvas.width = r.width * dpr; canvas.height = r.height * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  const W = r.width, H = r.height, cx = W / 2, cy = H / 2;
  ctx.clearRect(0, 0, W, H);

  // orbit rings
  for (let i = 1; i <= 3; i++) {
    ctx.beginPath(); ctx.arc(cx, cy, (Math.min(W, H) / 2 - 30) * (i / 3), 0, Math.PI * 2);
    ctx.strokeStyle = "rgba(167,139,250,0.08)"; ctx.lineWidth = 1; ctx.stroke();
  }

  const maxR = Math.min(W, H) / 2 - 36;
  siblingState.matches.forEach((m, i) => {
    const angle = (i / Math.max(1, siblingState.matches.length)) * Math.PI * 2 - Math.PI / 2;
    const dist = maxR * (1 - m.similarity) + 24;
    const x = cx + Math.cos(angle) * dist, y = cy + Math.sin(angle) * dist;
    const hot = m.similarity;
    const color = "rgb(" + Math.round(167 + hot * 65) + "," + Math.round(139 - hot * 20) + "," + Math.round(250) + ")";
    ctx.beginPath(); ctx.moveTo(cx, cy); ctx.lineTo(x, y);
    ctx.strokeStyle = "rgba(167,139,250," + (0.15 + hot * 0.5) + ")"; ctx.lineWidth = 1 + hot * 1.5; ctx.stroke();
    ctx.beginPath(); ctx.arc(x, y, 5 + hot * 5, 0, Math.PI * 2);
    ctx.fillStyle = color; ctx.shadowColor = color; ctx.shadowBlur = 10; ctx.fill(); ctx.shadowBlur = 0;
    ctx.fillStyle = "#e8e4f0"; ctx.font = "11px ui-monospace, monospace"; ctx.textAlign = "center";
    ctx.fillText(m.func, x, y - 12 - hot * 4);
  });

  // reference node at the centre
  ctx.beginPath(); ctx.arc(cx, cy, 9, 0, Math.PI * 2);
  ctx.fillStyle = "#e879f9"; ctx.shadowColor = "#e879f9"; ctx.shadowBlur = 16; ctx.fill(); ctx.shadowBlur = 0;
  ctx.fillStyle = "#e8e4f0"; ctx.font = "bold 12px ui-monospace, monospace"; ctx.textAlign = "center";
  ctx.fillText(siblingState.ref, cx, cy + 26);
}

function init() {
  document.querySelectorAll(".nav").forEach((b) => b.addEventListener("click", () => route(b.dataset.page)));
  $("sev-filter").addEventListener("change", renderFindings);
  $("func-select").addEventListener("change", runSiblings);
  $("min").addEventListener("input", () => { $("min-val").textContent = $("min").value + "%"; });
  $("min").addEventListener("change", runSiblings);
  window.addEventListener("resize", () => { if (!$("page-siblings").classList.contains("hidden")) drawConstellation(); });
  loadOverview();
}

init();
