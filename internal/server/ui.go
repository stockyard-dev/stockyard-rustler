package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Rustler — Stockyard</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:wght@400;700&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;--leather:#a0845c;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--green:#5ba86e;--red:#c0392b;--gold:#d4a843;--font-serif:'Libre Baskerville',Georgia,serif;--font-mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}a{color:var(--rust-light);text-decoration:none}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}.hdr-left{display:flex;align-items:center;gap:1rem}.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream)}.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;border:1px solid var(--green);color:var(--green);letter-spacing:1px;text-transform:uppercase}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:1rem;margin-bottom:2rem}.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;display:block}.card-lbl{font-family:var(--font-mono);font-size:.58rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}th{background:var(--bg3);padding:.4rem .8rem;text-align:left;color:#c4a87a;font-weight:400;font-size:.62rem;letter-spacing:1px;text-transform:uppercase}td{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim)}tr:hover td{background:var(--bg2)}.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer}.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.pill{display:inline-block;font-family:var(--font-mono);font-size:.55rem;padding:.1rem .4rem;border-radius:2px;text-transform:uppercase}.pill-running{background:#1a2a3a;color:#4a90d9}.pill-completed{background:#1a3a2a;color:var(--green)}.pill-failed{background:#2a1a1a;color:var(--red)}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}input{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none}input:focus{border-color:var(--leather)}.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}.field{display:flex;flex-direction:column;gap:.3rem}
</style></head><body>
<div class="hdr"><div class="hdr-left">
<svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
<span class="hdr-brand">Stockyard</span><span class="hdr-title">Rustler</span></div>
<div><span class="badge">Free</span></div></div>
<div class="main">
<div class="cards">
  <div class="card"><span class="card-val" id="s-scans">—</span><span class="card-lbl">Scans</span></div>
  <div class="card"><span class="card-val" id="s-broken">—</span><span class="card-lbl">Broken Found</span></div>
</div>
<div class="section"><div class="section-title">Scan a URL</div>
<div class="row">
  <div class="field"><span class="lbl">URL</span><input id="c-url" placeholder="https://example.com" style="width:300px"></div>
  <button class="btn btn-rust" onclick="startScan()">Scan</button>
</div><div id="c-result"></div></div>
<div class="section"><div class="section-title">Scans</div>
<table><thead><tr><th>URL</th><th>Status</th><th>Links</th><th>Broken</th><th>SSL</th><th></th></tr></thead>
<tbody id="scans-body"></tbody></table></div>
</div>
<script>
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-scans').textContent=s.scans||0;document.getElementById('s-broken').textContent=s.total_broken_found||0;}catch(e){}
  try{const d=await(await fetch('/api/scans')).json();const ss=d.scans||[];const tb=document.getElementById('scans-body');
  if(!ss.length){tb.innerHTML='<tr><td colspan="6" class="empty">No scans yet.</td></tr>';return;}
  tb.innerHTML=ss.map(s=>'<tr><td style="font-size:.68rem;color:var(--cream)">'+esc(s.url)+'</td><td><span class="pill pill-'+s.status+'">'+s.status+'</span></td><td>'+s.total_links+'</td><td style="color:'+(s.broken_links?'var(--red)':'var(--green)')+'">'+s.broken_links+'</td><td>'+(s.ssl_issues||0)+'</td><td><button class="btn btn-sm" onclick="del(\''+s.id+'\')">Del</button></td></tr>').join('');}catch(e){}
}
async function startScan(){const url=document.getElementById('c-url').value.trim();if(!url)return;const r=await fetch('/api/scans',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url})});const d=await r.json();if(r.ok){document.getElementById('c-result').innerHTML='<span style="color:var(--green);font-family:var(--font-mono);font-size:.75rem">Scanning...</span>';document.getElementById('c-url').value='';refresh();}else{document.getElementById('c-result').innerHTML='<span style="color:var(--red)">'+esc(d.error)+'</span>';}}
async function del(id){await fetch('/api/scans/'+id,{method:'DELETE'});refresh();}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
refresh();setInterval(refresh,3000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
