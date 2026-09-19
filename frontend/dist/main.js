const SVG = "http://www.w3.org/2000/svg";

const PROGRAMMA_INIZIALE = `# Incremento binario
inizio: q0
finali: qf
# stato letto scrive mossa nuovo
q0 0 0 R q0
q0 1 1 R q0
q0 _ _ L q1
q1 1 0 L q1
q1 0 1 N qf
q1 _ 1 N qf`;

const LARGHEZZA_CELLA = 38;
const MAX_ETICHETTE = 3;

const el = (id) => document.getElementById(id);

const nodi = {
  sorgente: el("sorgente"),
  ingresso: el("ingresso"),
  errore: el("errore"),
  esito: el("esito"),
  tavola: el("tavola-corpo"),
  stato: el("m-stato"),
  pos: el("m-pos"),
  passi: el("m-passi"),
  nastroTesto: el("m-nastro"),
  nastro: el("nastro"),
  indici: el("indici"),
  grafo: el("grafo"),
  legenda: el("legenda"),
  registro: el("registro"),
  velocita: el("velocita"),
};

let programma = null;
let istantanea = null;
let animazione = null;
let raggio = 12;

function api() {
  return window.go && window.go.main && window.go.main.App;
}

async function attendiWails() {
  for (let i = 0; i < 200 && !api(); i++) {
    await new Promise((r) => setTimeout(r, 25));
  }
  return !!api();
}

function raggioDaLarghezza() {
  const larghezza = nodi.nastro.clientWidth || 600;
  const celle = Math.floor(larghezza / LARGHEZZA_CELLA);
  return Math.max(3, Math.floor((celle - 1) / 2));
}

function abilita(attiva) {
  for (const id of ["btn-passo", "btn-indietro", "btn-avvia", "btn-corri"]) {
    el(id).disabled = !attiva;
  }
}

async function carica() {
  fermaAnimazione();
  const p = await api().Load(nodi.sorgente.value, nodi.ingresso.value);
  if (p.error) {
    programma = null;
    istantanea = null;
    nodi.errore.textContent = p.error;
    nodi.tavola.innerHTML = "";
    nodi.grafo.innerHTML = "";
    nodi.legenda.textContent = "";
    abilita(false);
    return;
  }
  nodi.errore.textContent = "";
  programma = p;
  nodi.legenda.textContent = p.legend;
  disegnaTavola();
  abilita(true);
  raggio = raggioDaLarghezza();
  applica(await api().SetRadius(raggio));
}

async function azzera() {
  fermaAnimazione();
  if (!programma) return carica();
  applica(await api().Reset(nodi.ingresso.value));
}

async function passo() {
  if (!programma) return;
  applica(await api().Step());
}

async function indietro() {
  fermaAnimazione();
  if (!programma) return;
  applica(await api().Back());
}

async function corri() {
  fermaAnimazione();
  if (!programma) return;
  applica(await api().Run(0));
}

function intervallo() {
  return 1100 - Number(nodi.velocita.value) * 100;
}

function fermaAnimazione() {
  if (animazione) clearTimeout(animazione);
  animazione = null;
  el("btn-avvia").textContent = "Avvia";
}

async function battito() {
  if (!animazione) return;
  await passo();
  if (!animazione) return;
  if (istantanea && istantanea.halted) return fermaAnimazione();
  animazione = setTimeout(battito, intervallo());
}

function avvia() {
  if (animazione) return fermaAnimazione();
  if (!programma || (istantanea && istantanea.halted)) return;
  animazione = setTimeout(battito, 0);
  el("btn-avvia").textContent = "Pausa";
}

function applica(s) {
  istantanea = s;
  if (s.error) nodi.errore.textContent = s.error;

  nodi.stato.textContent = s.state || "—";
  nodi.pos.textContent = s.pos;
  nodi.passi.textContent = s.steps;
  nodi.nastroTesto.textContent = s.content || "—";

  nodi.esito.textContent = s.halted ? "ferma: " + s.reason : (s.steps ? "in esecuzione" : "pronta");
  nodi.esito.className = "esito" + (s.halted ? (s.final ? " ferma-ok" : " ferma-ko") : "");

  disegnaNastro(s);
  evidenziaTavola(s.next);
  disegnaGrafo(s);
  disegnaRegistro(s);

  el("btn-passo").disabled = s.halted;
  el("btn-corri").disabled = s.halted;
  el("btn-avvia").disabled = s.halted;
  el("btn-indietro").disabled = s.steps === 0;
  if (s.halted) fermaAnimazione();
}

function disegnaNastro(s) {
  const celle = document.createDocumentFragment();
  const indici = document.createDocumentFragment();
  for (const c of s.cells) {
    const d = document.createElement("div");
    d.className = "cella" + (c.head ? " sotto-testina" : "") + (c.sym === "_" ? " bianca" : "");
    d.textContent = c.sym;
    celle.appendChild(d);

    const i = document.createElement("div");
    i.className = "indice";
    i.textContent = c.pos;
    indici.appendChild(i);
  }
  nodi.nastro.replaceChildren(celle);
  nodi.indici.replaceChildren(indici);
}

function disegnaTavola() {
  const corpo = document.createDocumentFragment();
  for (const r of programma.rules) {
    const tr = document.createElement("tr");
    for (const v of [r.from, r.read, r.write, r.move, r.to]) {
      const td = document.createElement("td");
      td.textContent = v;
      tr.appendChild(td);
    }
    corpo.appendChild(tr);
  }
  nodi.tavola.replaceChildren(corpo);
}

function evidenziaTavola(next) {
  const righe = nodi.tavola.children;
  for (let i = 0; i < righe.length; i++) {
    righe[i].className = i === next ? "prossima" : "";
  }
}

function disegnaRegistro(s) {
  if (!s.last) {
    nodi.registro.textContent = s.steps ? "" : "nessun passo eseguito";
    return;
  }
  const righe = nodi.registro.textContent.split("\n").filter(Boolean);
  righe.unshift(s.last);
  nodi.registro.textContent = righe.slice(0, 6).join("\n");
}

function crea(nome, attributi, classe) {
  const e = document.createElementNS(SVG, nome);
  for (const k in attributi) e.setAttribute(k, attributi[k]);
  if (classe) e.setAttribute("class", classe);
  return e;
}

function testo(x, y, contenuto, classe) {
  const t = crea("text", { x, y }, classe);
  t.textContent = contenuto;
  return t;
}

function punta(svg, x, y, angolo, attivo) {
  const l = 8, a = 0.42;
  const p = [
    [x, y],
    [x - l * Math.cos(angolo - a), y - l * Math.sin(angolo - a)],
    [x - l * Math.cos(angolo + a), y - l * Math.sin(angolo + a)],
  ].map((c) => c.join(",")).join(" ");
  svg.appendChild(crea("polygon", { points: p }, "marcatore" + (attivo ? " attivo" : "")));
}

function etichette(edge) {
  if (edge.labels.length <= MAX_ETICHETTE) return edge.labels;
  const resto = edge.labels.length - MAX_ETICHETTE;
  return edge.labels.slice(0, MAX_ETICHETTE).concat("+" + resto + " altre");
}

function disegnaGrafo(s) {
  if (!programma || !programma.graph) return;
  const g = programma.graph;
  const W = nodi.grafo.clientWidth || 640;

  if (g.synthetic) return disegnaRegistro2(s, W);

  const R = g.nodes.length <= 4 ? 30 : 24;
  const haCappi = g.edges.some((e) => e.loop);

  // Un cappio si alza di circa 1.45 raggi sopra il nodo, e sopra il cappio
  // stanno le sue etichette: il margine superiore deve prevederli, altrimenti
  // il nodo in cima al cerchio si porta le etichette fuori dal riquadro.
  const ingombroCappio = haCappi ? R * 1.45 + (MAX_ETICHETTE + 1) * 12 + 10 : 0;
  const margineLati = R + 34;
  const margineSotto = R + 30;
  const margineSopra = R + 24 + ingombroCappio;
  const H = Math.max(250, Math.min(380, W * 0.5)) + ingombroCappio;

  const svg = crea("svg", { viewBox: `0 0 ${W} ${H}`, width: "100%", height: H });
  const punti = {};
  for (const n of g.nodes) {
    punti[n.name] = {
      x: margineLati + n.x * (W - 2 * margineLati),
      y: margineSopra + n.y * (H - margineSopra - margineSotto),
    };
  }

  const regolaProssima = s.next >= 0 ? programma.rules[s.next] : null;
  const arcoAttivo = (e) =>
    regolaProssima && e.from === regolaProssima.from && e.to === regolaProssima.to;

  for (const e of g.edges) {
    const attivo = arcoAttivo(e);
    const cl = attivo ? " attivo" : "";
    const a = punti[e.from];

    if (e.loop) {
      const cx = a.x, cy = a.y - R;
      const d = `M ${cx - R * 0.6} ${cy} C ${cx - R * 1.5} ${cy - R * 1.9}, ${cx + R * 1.5} ${cy - R * 1.9}, ${cx + R * 0.6} ${cy}`;
      svg.appendChild(crea("path", { d }, "arco" + cl));
      punta(svg, cx + R * 0.6, cy, 1.1, attivo);
      // Le etichette si impilano verso l'alto partendo dalla sommità
      // dell'arco, così l'ultima resta sempre staccata dal disegno.
      const righe = etichette(e);
      const cima = cy - R * 1.45 - 9 - (righe.length - 1) * 12;
      righe.forEach((t, i) => {
        svg.appendChild(testo(cx, cima + i * 12, t, "arco-etichetta" + cl));
      });
      continue;
    }

    const b = punti[e.to];
    const dx = b.x - a.x, dy = b.y - a.y;
    const lung = Math.hypot(dx, dy) || 1;
    const ux = dx / lung, uy = dy / lung;
    const scostamento = lung * 0.16;
    const cx = (a.x + b.x) / 2 - uy * scostamento;
    const cy = (a.y + b.y) / 2 + ux * scostamento;

    const p0 = { x: a.x + ux * R * 0.75 - uy * R * 0.5, y: a.y + uy * R * 0.75 + ux * R * 0.5 };
    const p2 = { x: b.x - ux * R * 0.95 - uy * R * 0.4, y: b.y - uy * R * 0.95 + ux * R * 0.4 };

    svg.appendChild(crea("path", { d: `M ${p0.x} ${p0.y} Q ${cx} ${cy} ${p2.x} ${p2.y}` }, "arco" + cl));
    punta(svg, p2.x, p2.y, Math.atan2(p2.y - cy, p2.x - cx), attivo);

    const mx = (p0.x + 2 * cx + p2.x) / 4;
    const my = (p0.y + 2 * cy + p2.y) / 4;
    etichette(e).forEach((t, i) => {
      svg.appendChild(testo(mx, my - 4 + i * 12, t, "arco-etichetta" + cl));
    });
  }

  for (const n of g.nodes) {
    const p = punti[n.name];
    const attivo = n.name === s.state;
    const cl = attivo ? " attivo" : "";
    if (n.final) {
      svg.appendChild(crea("circle", { cx: p.x, cy: p.y, r: R + 4 }, "nodo-cerchio" + cl));
    }
    svg.appendChild(crea("circle", { cx: p.x, cy: p.y, r: R }, "nodo-cerchio" + cl));
    svg.appendChild(testo(p.x, p.y, n.name, "nodo-testo" + cl));
    if (n.start) {
      svg.appendChild(crea("path", { d: `M ${p.x - R - 26} ${p.y} L ${p.x - R - 6} ${p.y}` }, "arco" + cl));
      punta(svg, p.x - R - 6, p.y, 0, attivo);
    }
  }

  nodi.grafo.replaceChildren(svg);
}

function disegnaRegistro2(s, W) {
  const H = 200;
  const svg = crea("svg", { viewBox: `0 0 ${W} ${H}`, width: "100%", height: H });
  const cx = W / 2, cy = H / 2, R = 62;
  const finale = programma.graph.nodes.some((n) => n.name === s.state && n.final);
  if (finale) svg.appendChild(crea("circle", { cx, cy, r: R + 6 }, "nodo-cerchio attivo"));
  svg.appendChild(crea("circle", { cx, cy, r: R }, "nodo-cerchio attivo"));
  svg.appendChild(testo(cx, cy, s.state, "nodo-testo nodo-grande attivo"));
  const r = s.next >= 0 ? programma.rules[s.next] : null;
  if (r) {
    svg.appendChild(testo(cx, cy + R + 24, r.label + " → " + r.to, "arco-etichetta attivo"));
  }
  svg.appendChild(testo(cx, 20, programma.graph.nodes.length + " stati: mostrato il solo registro", "arco-etichetta"));
  nodi.grafo.replaceChildren(svg);
}

let attesaRidimensionamento = null;
function ridimensiona() {
  if (attesaRidimensionamento) clearTimeout(attesaRidimensionamento);
  attesaRidimensionamento = setTimeout(async () => {
    if (!programma) return;
    const nuovo = raggioDaLarghezza();
    if (nuovo !== raggio) {
      raggio = nuovo;
      applica(await api().SetRadius(raggio));
    } else if (istantanea) {
      disegnaGrafo(istantanea);
    }
  }, 120);
}

el("btn-carica").addEventListener("click", carica);
el("btn-azzera").addEventListener("click", azzera);
el("btn-passo").addEventListener("click", () => { fermaAnimazione(); passo(); });
el("btn-indietro").addEventListener("click", indietro);
el("btn-avvia").addEventListener("click", avvia);
el("btn-corri").addEventListener("click", corri);
nodi.sorgente.addEventListener("input", () => { nodi.errore.textContent = ""; });
window.addEventListener("resize", ridimensiona);

document.addEventListener("keydown", (ev) => {
  if (ev.target.tagName === "TEXTAREA" || ev.target.tagName === "INPUT") return;
  if (ev.key === "ArrowRight") { ev.preventDefault(); fermaAnimazione(); passo(); }
  if (ev.key === "ArrowLeft") { ev.preventDefault(); indietro(); }
  if (ev.key === " ") { ev.preventDefault(); avvia(); }
});

(async () => {
  nodi.sorgente.value = PROGRAMMA_INIZIALE;
  abilita(false);
  if (!(await attendiWails())) {
    nodi.errore.textContent = "motore non raggiungibile: avvia con wails dev o wails build";
    return;
  }
  await carica();
})();
