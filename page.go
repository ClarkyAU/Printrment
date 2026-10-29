package main

const page = `<!doctype html><html lang=en><meta charset=utf-8>
<meta name=viewport content="width=device-width,initial-scale=1">
<title>Printrment {{VERSION}}</title>
<link rel=icon href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'%3E%3Crect width='32' height='32' rx='7' fill='%232f6fed'/%3E%3Ccircle cx='16' cy='16' r='10' fill='%23f2f5fa'/%3E%3Ccircle cx='16' cy='16' r='6.4' fill='%23121824'/%3E%3Ccircle cx='25' cy='7' r='2.3' fill='%23ff5a5a'/%3E%3C/svg%3E">
<script>try{var t=localStorage.getItem("bc_theme");if(t)document.documentElement.dataset.theme=t}catch(e){}</script>
<style>
@font-face{font-family:"Exo 2";font-weight:400;font-display:swap;src:url(/fonts/exo-2-latin-400-normal.woff2) format("woff2")}
@font-face{font-family:"Exo 2";font-weight:600;font-display:swap;src:url(/fonts/exo-2-latin-600-normal.woff2) format("woff2")}
@font-face{font-family:"Exo 2";font-weight:700;font-display:swap;src:url(/fonts/exo-2-latin-700-normal.woff2) format("woff2")}
@font-face{font-family:"Space Mono";font-weight:400;font-display:swap;src:url(/fonts/space-mono-latin-400-normal.woff2) format("woff2")}
@font-face{font-family:"Space Mono";font-weight:700;font-display:swap;src:url(/fonts/space-mono-latin-700-normal.woff2) format("woff2")}
@font-face{font-family:"Silkscreen";font-weight:400;font-display:swap;src:url(/fonts/silkscreen-latin-400-normal.woff2) format("woff2")}
/* ---------------- themes: every colour and shape comes from these tokens */
:root,[data-theme=default]{color-scheme:dark;
 --bg:#0b0d10;--bg-img:none;--panel:#13161b;--card:#171b21;--line:#252b34;--line2:#313844;
 --tx:#dde2e8;--mut:#8b94a1;--dim:#5f6875;
 --acc:#2f6fed;--acc-h:#4c84f5;--acc-tx:#fff;--prog:#4c84f5;
 --ok:#3ecf6e;--warn:#f5b544;--err:#ff6b6b;
 --btn:#1d222a;--btn-h:#252b35;--btn-b:#313844;--btn-tx:#dde2e8;
 --on:#20304d;--on-b:#35558f;--on-tx:#cfe0ff;
 --inp:#0e1115;--tile:#0f1216;--toast:#1b2028;--veil:#05070acc;--chip:#0b0d10cc;
 --hms:#2a2112;--hms-b:#5a4520;--hms-tx:#f6d38f;--danger-b:#5a2a2e;--danger-tx:#ffb3b3;--link:#9cc0ff;
 --r:10px;--rb:7px;--shadow:none;--bshadow:none;
 --font:"Segoe UI",system-ui,-apple-system,sans-serif;--label:var(--font);--mono:Consolas,ui-monospace,monospace;
 --label-tt:uppercase;--label-ls:.7px;--label-fs:11.5px;--btn-tt:none;--btn-font:var(--font)}

/* Clarky3D - the webstore: #121212 ground, cream type, #FF8000, hard offset shadows, square corners */
[data-theme=webstore]{color-scheme:dark;
 --bg:#121212;--bg-img:repeating-linear-gradient(0deg,rgba(255,255,255,.018) 0 1px,transparent 1px 4px);
 --panel:#161616;--card:#1A1A1A;--line:#383633;--line2:#4a4742;
 --tx:#F5F0E6;--mut:#A8A298;--dim:#767169;
 --acc:#FF8000;--acc-h:#FF9326;--acc-tx:#121212;--prog:#FF8000;
 --ok:#7ad67f;--warn:#ffc24d;--err:#ff5f57;
 --btn:#1A1A1A;--btn-h:#232220;--btn-b:#4a4742;--btn-tx:#F5F0E6;
 --on:#FF8000;--on-b:#FF8000;--on-tx:#121212;
 --inp:#121212;--tile:#121212;--toast:#1A1A1A;--veil:#000000cc;--chip:#121212e6;
 --hms:#2a1d0c;--hms-b:#FF8000;--hms-tx:#ffd3a3;--danger-b:#ff5f57;--danger-tx:#ff8c86;--link:#FF9326;
 --r:0px;--rb:0px;--shadow:6px 6px 0 #000;--bshadow:3px 3px 0 #000;
 --font:"Exo 2","Segoe UI",system-ui,sans-serif;--label:"Silkscreen","Space Mono",ui-monospace,monospace;--mono:"Space Mono",ui-monospace,Consolas,monospace;
 --label-tt:uppercase;--label-ls:.14em;--label-fs:12px;--btn-tt:uppercase;--btn-font:"Space Mono",ui-monospace,Consolas,monospace}

/* MakerWorld-style: light, airy, Bambu green, soft rounded cards */
[data-theme=makerworld]{color-scheme:light;
 --bg:#f4f5f7;--bg-img:none;--panel:#ffffff;--card:#ffffff;--line:#e6e8eb;--line2:#d5d9de;
 --tx:#1f2329;--mut:#5d636e;--dim:#8a919c;
 --acc:#00ae42;--acc-h:#009a3a;--acc-tx:#fff;--prog:#00ae42;
 --ok:#00a03c;--warn:#d97706;--err:#e5484d;
 --btn:#f2f3f5;--btn-h:#e7e9ec;--btn-b:#f2f3f5;--btn-tx:#1f2329;
 --on:#e3f6ea;--on-b:#00ae42;--on-tx:#00782d;
 --inp:#ffffff;--tile:#f7f8fa;--toast:#ffffff;--veil:#1f232973;--chip:#ffffffe6;
 --hms:#fff6e5;--hms-b:#f5c06b;--hms-tx:#8a5300;--danger-b:#f3c1c2;--danger-tx:#d4363b;--link:#00873a;
 --r:14px;--rb:8px;--shadow:0 1px 2px rgba(16,24,40,.05),0 1px 3px rgba(16,24,40,.06);--bshadow:none;
 --font:"Inter","Segoe UI",system-ui,-apple-system,sans-serif;--label:var(--font);--mono:Consolas,ui-monospace,monospace;
 --label-tt:none;--label-ls:0;--label-fs:13px;--btn-tt:none;--btn-font:var(--font)}

*{box-sizing:border-box}
html,body{margin:0;height:100%;background:var(--bg);background-image:var(--bg-img);color:var(--tx);font:13.5px/1.4 var(--font)}
body{display:flex;flex-direction:column;overflow:hidden}
button{font:inherit;font-family:var(--btn-font);text-transform:var(--btn-tt);color:var(--btn-tx);background:var(--btn);border:1px solid var(--btn-b);border-radius:var(--rb);padding:6px 11px;cursor:pointer;white-space:nowrap;box-shadow:var(--bshadow)}
[data-theme=webstore] button{font-size:12px;letter-spacing:.06em}
button:hover:not(:disabled){background:var(--btn-h)}
button:active:not(:disabled){transform:translate(1px,1px)}
button:disabled{opacity:.45;cursor:default}
button:focus-visible{outline:2px solid var(--acc);outline-offset:2px}
button.pri{background:var(--acc);border-color:var(--acc);color:var(--acc-tx)}button.pri:hover:not(:disabled){background:var(--acc-h)}
[data-theme=webstore] button.pri{font-weight:700}
button.danger{border-color:var(--danger-b);color:var(--danger-tx)}button.danger.armed{background:var(--err);border-color:var(--err);color:#fff}
button.ghost{background:transparent;border-color:transparent;color:var(--mut);box-shadow:none}button.ghost:hover{color:var(--tx);background:var(--btn)}
button.on{background:var(--on);border-color:var(--on-b);color:var(--on-tx)}
input,select{font:inherit;color:var(--tx);background:var(--inp);border:1px solid var(--line2);border-radius:var(--rb);padding:7px 9px;width:100%}
input:focus,select:focus{outline:2px solid var(--acc);outline-offset:-1px}
a{color:var(--link)}
.mut{color:var(--mut)}.dim{color:var(--dim)}.ok{color:var(--ok)}.warn{color:var(--warn)}.err{color:var(--err)}
.dot{display:inline-block;width:8px;height:8px;border-radius:50%;background:var(--dim);flex:none}
[data-theme=webstore] .dot{border-radius:0}
.dot.ok{background:var(--ok)}.dot.warn{background:var(--warn)}.dot.err{background:var(--err)}.dot.acc{background:var(--acc)}
.dot.pulse{animation:pulse 1.4s infinite}@keyframes pulse{50%{opacity:.35}}

#top{height:48px;display:flex;align-items:center;gap:10px;padding:0 10px 0 14px;background:var(--panel);border-bottom:1px solid var(--line);flex:none}
#brand{display:flex;align-items:center;gap:9px;font-weight:700;letter-spacing:.2px;margin-right:6px;white-space:nowrap}
#brand .lens{width:20px;height:20px;border-radius:5px;background:var(--acc);display:grid;place-items:center}
#brand .lens:after{content:"";width:9px;height:9px;border-radius:50%;background:#121824;box-shadow:0 0 0 3px #f2f5fa}
#brand .pix{display:none;height:24px;width:auto}
[data-theme=webstore] #brand .lens{display:none}[data-theme=webstore] #brand .pix{display:block}
[data-theme=webstore] #brand b{font-size:16px}
#tabs{display:flex;gap:6px;flex:1;min-width:0;overflow-x:auto;scrollbar-width:none}
.tab{display:flex;align-items:center;gap:7px;background:transparent;border-color:transparent;color:var(--mut);padding:6px 10px;box-shadow:none}
.tab.sel{background:var(--btn);border-color:var(--btn-b);color:var(--tx);box-shadow:var(--bshadow)}
[data-theme=webstore] .tab.sel{border-color:var(--acc);color:var(--acc)}
.tab small{color:var(--mut);font-variant-numeric:tabular-nums}
#topr{display:flex;gap:6px}

#main{flex:1;min-height:0;display:flex}
#stage{flex:1;min-width:0;position:relative;background:#000;display:flex;align-items:center;justify-content:center;overflow:hidden}
#v{max-width:100%;max-height:100%;display:none;transition:filter .3s}
#stage.paused #v{filter:grayscale(.7) brightness(.45)}
#msg{position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;text-align:center;padding:24px;color:#9aa3ae}
#msg b{color:#eef1f4;font-size:15px;font-weight:600}
#msg p{margin:0;max-width:520px}
#pausebox{position:absolute;inset:0;display:none;flex-direction:column;align-items:center;justify-content:center;gap:12px;color:#eef1f4;text-align:center}
#stage.paused #pausebox{display:flex}
#pausebox b{font-size:16px;display:flex;align-items:center;gap:8px}
#pausebox b:before{content:"";width:12px;height:14px;border-left:4px solid currentColor;border-right:4px solid currentColor}
#pausebox span{color:#c2c8d0;max-width:420px;font-size:12.5px}
.chip{position:absolute;display:flex;align-items:center;gap:7px;background:var(--chip);color:var(--tx);border:1px solid var(--line);padding:5px 10px;border-radius:999px;font-size:12.5px;backdrop-filter:blur(6px);box-shadow:var(--bshadow)}
[data-theme=webstore] .chip{border-radius:0;font-family:var(--mono);text-transform:uppercase;font-size:11px;letter-spacing:.08em}
#badge{top:12px;left:12px}
#rec{top:12px;right:12px;display:none}
#ov{position:absolute;left:0;right:0;bottom:0;padding:34px 16px 13px;background:linear-gradient(transparent,#000000d0);display:none;color:#f2f4f6}
#ov .row{display:flex;justify-content:space-between;gap:12px;align-items:baseline}
#ov .file{font-weight:600;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
#ov .meta{color:#d0d5dc;font-variant-numeric:tabular-nums;white-space:nowrap}
.bar{height:5px;border-radius:9px;background:#8882;overflow:hidden;margin-top:8px}.bar i{display:block;height:100%;background:var(--prog);border-radius:9px;transition:width .6s}
#ov .bar{background:#ffffff2e}
[data-theme=webstore] .bar,[data-theme=webstore] .bar i{border-radius:0;height:6px}
.bar.warn i{background:var(--warn)}
#cbtns{position:absolute;top:8px;right:8px;display:none;gap:4px;opacity:0;transition:opacity .2s;z-index:5}
body.compact #stage:hover #cbtns{opacity:1}
#cbtns .ib{background:rgba(10,12,15,.6);color:#dfe4ea;height:30px}
#cbtns .ib.on{background:var(--acc);color:var(--acc-tx)}

#side{width:344px;flex:none;background:var(--panel);border-left:1px solid var(--line);overflow-y:auto;padding:12px 14px 14px 12px;display:flex;flex-direction:column;gap:12px}
[data-theme=makerworld] #side{background:var(--bg)}
.card{background:var(--card);border:1px solid var(--line);border-radius:var(--r);padding:12px 13px;box-shadow:var(--shadow)}
.card h3{margin:0 0 10px;font-family:var(--label);font-size:var(--label-fs);font-weight:600;letter-spacing:var(--label-ls);text-transform:var(--label-tt);color:var(--mut);display:flex;justify-content:space-between;align-items:center;gap:8px}
[data-theme=makerworld] .card h3{color:var(--tx);font-weight:650}
.card h3 .hs{font-family:var(--font);text-transform:none;letter-spacing:0;font-weight:400;font-size:12px}
.kv{display:grid;grid-template-columns:auto 1fr;gap:6px 14px;font-variant-numeric:tabular-nums}
.kv span:nth-child(odd){color:var(--mut)}.kv span:nth-child(even){text-align:right}
.big{font-size:26px;font-weight:700;font-variant-numeric:tabular-nums;line-height:1.1}
[data-theme=webstore] .big{color:var(--acc)}
.temps{display:grid;grid-template-columns:repeat(auto-fit,minmax(88px,1fr));gap:8px}
.temp{background:var(--tile);border:1px solid var(--line);border-radius:calc(var(--r) * .7);padding:8px 10px}
.temp small{color:var(--mut);display:block;font-size:11.5px}.temp b{font-size:17px;font-weight:650;font-variant-numeric:tabular-nums}.temp em{font-style:normal;color:var(--dim);font-size:11.5px;display:block}
.btns{display:flex;gap:6px;flex-wrap:wrap}.btns>button{flex:1}
.seg{display:flex;border:1px solid var(--line2);border-radius:var(--rb);overflow:hidden;box-shadow:var(--bshadow)}
.seg button{flex:1;border:0;border-radius:0;background:transparent;color:var(--mut);padding:6px 4px;font-size:12.5px;box-shadow:none}
[data-theme=webstore] .seg button{font-size:11px}
.seg button+button{border-left:1px solid var(--line2)}.seg button.on{background:var(--on);color:var(--on-tx)}
[data-theme=webstore] .seg button.on{font-weight:700}
.lbl{font-size:12px;color:var(--mut);margin:12px 0 6px}
.note{font-size:12px;color:var(--mut);margin-top:8px;line-height:1.45}.note.err{color:var(--err)}.note.ok{color:var(--ok)}
.ltabs{display:flex;gap:4px}
.ltabs button{padding:2px 8px;font-size:11.5px;box-shadow:none;background:transparent;border-color:transparent;color:var(--mut);font-family:var(--font);text-transform:none;letter-spacing:0}
.ltabs button.on{background:var(--btn);border-color:var(--btn-b);color:var(--tx)}
.ltabs button i{font-style:normal;background:var(--err);color:#fff;border-radius:9px;padding:0 5px;margin-left:4px;font-size:10.5px}
.alerts{display:flex;flex-direction:column;max-height:260px;overflow-y:auto}
.al{display:grid;grid-template-columns:10px 1fr auto;gap:8px;align-items:baseline;padding:6px 0;border-top:1px solid var(--line)}
.al:first-child{border-top:0}.al .dot{transform:translateY(-1px)}.al time{color:var(--dim);font-size:11.5px;white-space:nowrap}
.hms{background:var(--hms);border:1px solid var(--hms-b);color:var(--hms-tx);border-radius:calc(var(--r) * .7);padding:8px 10px;margin-top:10px;font-size:12.5px}
.hms .mut{color:inherit;opacity:.8}
.sw{display:flex;align-items:center;justify-content:space-between;gap:10px;cursor:pointer;padding:3px 0}
.sw input{display:none}.sw i{width:34px;height:20px;border-radius:20px;background:var(--line2);position:relative;flex:none;transition:.2s}
.sw i:after{content:"";position:absolute;top:3px;left:3px;width:14px;height:14px;border-radius:50%;background:#fff;transition:.2s;box-shadow:0 1px 2px #0003}
.sw input:checked+i{background:var(--acc)}.sw input:checked+i:after{left:17px}
[data-theme=webstore] .sw i,[data-theme=webstore] .sw i:after{border-radius:0}

#bar{height:42px;flex:none;display:flex;align-items:center;gap:8px;padding:0 10px 0 14px;background:var(--panel);border-top:1px solid var(--line)}
#st{flex:1;min-width:0;color:var(--mut);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;display:flex;align-items:center;gap:8px}

#toasts{position:fixed;top:100px;right:360px;display:flex;flex-direction:column;gap:8px;z-index:30;width:320px;pointer-events:none}
.toast{pointer-events:auto;background:var(--toast);color:var(--tx);border:1px solid var(--line2);border-top:3px solid var(--acc);border-radius:var(--r);padding:10px 12px;box-shadow:0 10px 30px #0006;display:grid;grid-template-columns:1fr auto;gap:4px 10px;animation:tin .25s ease-out}
[data-theme=webstore] .toast{box-shadow:6px 6px 0 #000}
[data-theme=makerworld] .toast{box-shadow:0 10px 30px #1f232926}
.toast.success{border-top-color:var(--ok)}.toast.warn{border-top-color:var(--warn)}.toast.error{border-top-color:var(--err)}
.toast small{color:var(--mut)}.toast .x{grid-row:1/3;grid-column:2;align-self:start;padding:0 4px;background:none;border:0;color:var(--mut);box-shadow:none}
.toast .act{grid-column:1;justify-self:start;margin-top:4px;padding:3px 8px;font-size:12px}
@keyframes tin{from{transform:translateY(-6px);opacity:0}}

#set{position:fixed;inset:0;background:var(--veil);display:none;z-index:40;align-items:center;justify-content:center;padding:20px}
#set.open{display:flex}
#setbox{width:min(840px,100%);height:min(640px,100%);background:var(--panel);border:1px solid var(--line2);border-radius:calc(var(--r) * 1.4);display:flex;overflow:hidden;box-shadow:0 20px 60px #0008}
[data-theme=webstore] #setbox{box-shadow:10px 10px 0 #000;border:2px solid var(--acc)}
#setnav{width:200px;flex:none;border-right:1px solid var(--line);padding:14px 10px;display:flex;flex-direction:column;gap:2px;background:var(--panel)}
#setnav h2{font-size:15px;margin:2px 8px 12px;font-family:var(--label);text-transform:var(--label-tt);letter-spacing:var(--label-ls)}
#setnav button{text-align:left;background:transparent;border-color:transparent;color:var(--mut);box-shadow:none}
#setnav button.sel{background:var(--btn);color:var(--tx);border-color:var(--btn-b)}
[data-theme=webstore] #setnav button.sel{border-color:var(--acc);color:var(--acc)}
#setnav .grow{flex:1}
#setbody{flex:1;overflow-y:auto;padding:20px 22px;background:var(--panel)}
#setbody h4{margin:0 0 4px;font-size:16px}
#setbody .sub{color:var(--mut);margin:0 0 16px}
.field{display:flex;flex-direction:column;gap:5px;margin-bottom:12px}.field>span{font-size:12px;color:var(--mut)}
.grid2{display:grid;grid-template-columns:1fr 1fr;gap:0 12px}
.plist{display:flex;flex-direction:column;gap:6px;margin-bottom:16px}
.pitem{display:flex;align-items:center;gap:10px;background:var(--card);border:1px solid var(--line);border-radius:calc(var(--r) * .8);padding:9px 11px}
.pitem div{flex:1;min-width:0}.pitem small{display:block;color:var(--mut)}
.pwrap{display:flex;flex-direction:column;gap:0}
.pcheck{border:1px solid var(--line);border-top:0;border-radius:0 0 calc(var(--r) * .8) calc(var(--r) * .8);margin:-2px 0 0;padding:8px 11px 10px;background:var(--tile);display:flex;flex-direction:column;gap:6px}
.pcheck small{font-size:11px;margin-top:2px}
.ck{display:grid;grid-template-columns:10px 170px minmax(0,1fr);gap:8px;align-items:baseline;font-size:12.5px;line-height:1.4}
.ck i{width:8px;height:8px;border-radius:50%;background:var(--dim);display:block;transform:translateY(1px)}
.ck.ok i{background:var(--ok)}.ck.warn i{background:var(--warn)}.ck.bad i{background:var(--err)}
.ck b{font-weight:600;color:var(--tx)}.ck span{color:var(--mut)}
.ck.bad span,.ck.warn span{color:var(--tx)}
@media (max-width:640px){.ck{grid-template-columns:10px minmax(0,1fr)}.ck span{grid-column:2}}
.found{display:flex;flex-direction:column;gap:6px;margin:8px 0 14px}.found button{display:flex;justify-content:space-between;text-align:left}
.box{background:var(--card);border:1px solid var(--line);border-radius:calc(var(--r) * .8);padding:12px 14px;margin-bottom:12px}
.srow{display:flex;align-items:center;gap:12px;padding:9px 0;border-top:1px solid var(--line)}.srow:first-child{border-top:0;padding-top:2px}.srow .st{flex:1;min-width:0}.srow .st b{display:block;font-weight:600}.srow .st small{color:var(--mut);font-size:12px}.srow .btns button{padding:4px 9px;font-size:12px}
.url{font-family:var(--mono);font-size:14px;color:var(--link);user-select:all}
.pin{font-family:var(--mono);font-size:22px;letter-spacing:5px}
pre{background:var(--inp);border:1px solid var(--line);border-radius:8px;padding:10px;font-size:11.5px;max-height:330px;overflow:auto;white-space:pre-wrap;word-break:break-all;margin:8px 0 0;font-family:var(--mono)}
.themes{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}
.th{text-align:left;padding:0;overflow:hidden;display:flex;flex-direction:column;background:var(--card);border:2px solid var(--line);border-radius:12px;box-shadow:none;text-transform:none;font-family:var(--font);letter-spacing:0}
.th.on{border-color:var(--acc);background:var(--card);color:var(--tx)}
.th .pv{height:96px;display:grid;grid-template-columns:1fr 34%;gap:6px;padding:8px}
.th .pv i{display:block}
.th .pv .a{background:#000;border-radius:inherit}
.th .pv .b{display:flex;flex-direction:column;gap:5px}
.th .pv .b i{height:16px}
.th .nm{padding:9px 11px 11px;display:flex;flex-direction:column;gap:2px}
.th .nm b{font-size:13.5px}.th .nm small{color:var(--mut);font-size:11.5px;white-space:normal}


/* ---- own title bar (native window) */
#winbtns{display:none;align-self:stretch;margin-right:-10px}
.native #winbtns{display:flex}
#winbtns button{width:46px;height:100%;border:0;border-radius:0;background:transparent;box-shadow:none;color:var(--mut);display:grid;place-items:center;padding:0}
#winbtns button:hover{background:var(--btn-h);color:var(--tx)}
#winbtns button.x:hover{background:#e81123;color:#fff}
#winbtns svg{width:11px;height:11px;stroke:currentColor;fill:none;stroke-width:1.1;shape-rendering:crispEdges}
[data-theme=webstore] #winbtns button:hover{background:var(--acc);color:var(--acc-tx)}
[data-theme=webstore] #winbtns button.x:hover{background:#ff5f57;color:#121212}
.native #top{user-select:none}
.native #bquit{display:none}
.rz{position:fixed;z-index:60;display:none}
.native:not(.maxed) .rz{display:block}
.rz.t{top:0;left:8px;right:8px;height:4px;cursor:ns-resize}.rz.b{bottom:0;left:8px;right:8px;height:5px;cursor:ns-resize}
.rz.l{left:0;top:8px;bottom:8px;width:5px;cursor:ew-resize}.rz.r{right:0;top:8px;bottom:8px;width:5px;cursor:ew-resize}
.rz.tl{top:0;left:0;width:9px;height:9px;cursor:nwse-resize}.rz.br{bottom:0;right:0;width:10px;height:10px;cursor:nwse-resize}
.rz.tr{top:0;right:0;width:9px;height:9px;cursor:nesw-resize}.rz.bl{bottom:0;left:0;width:10px;height:10px;cursor:nesw-resize}

/* ---- AMS (compact: one row per unit) */
.amsr{display:grid;grid-template-columns:30px repeat(4,minmax(0,1fr));gap:4px;align-items:stretch}
.amsr+.amsr{margin-top:6px}
.amsl{display:flex;flex-direction:column;justify-content:center;line-height:1.15;min-width:0}
.amsl b{font-size:12px;font-weight:700;white-space:nowrap}
.amsl small{font-size:10px;color:var(--mut);white-space:nowrap;font-variant-numeric:tabular-nums}
.chipf{position:relative;display:flex;flex-direction:column;justify-content:center;gap:4px;padding:5px 5px;min-width:0;background:var(--tile);border:1px solid var(--line);border-radius:calc(var(--r) * .55)}
.chipf.on{border-color:var(--acc);box-shadow:0 0 0 1px var(--acc) inset}
.chipf.empty{border-style:dashed;background:transparent}
.chipf .l1{display:flex;align-items:center;gap:4px;min-width:0}
.chipf .sw8{width:10px;height:10px;border-radius:50%;flex:none;border:1px solid #8886}
[data-theme=webstore] .chipf .sw8{border-radius:0}
.chipf .ty{font-size:11px;font-weight:650;letter-spacing:-.1px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.chipf.empty .ty{color:var(--dim);font-weight:400}
.chipf .l2{display:flex;align-items:center;gap:4px}
.chipf .rm{flex:1;height:3px;background:#8883;border-radius:3px;overflow:hidden}
.chipf .rm i{display:block;height:100%;background:var(--mut)}
.chipf .rm.low i{background:var(--warn)}
.chipf .pc{font-size:9.5px;color:var(--mut);font-variant-numeric:tabular-nums;min-width:20px;text-align:right}
.chipf.on .ty{color:var(--acc)}
[data-theme=webstore] .chipf .rm,[data-theme=webstore] .chipf .rm i{border-radius:0}
.amsnote{font-size:11px;color:var(--mut);margin-top:7px}
/* ---- controls */
.cbig{display:grid;grid-template-columns:1fr 1fr;gap:8px}
.cbig button{height:42px;display:flex;align-items:center;justify-content:center;gap:8px;font-weight:600}
.cbig svg,.lrow svg,.spd svg{flex:none}
.cbig svg{width:16px;height:16px;fill:currentColor}
.spd{display:grid;grid-template-columns:repeat(4,1fr);gap:6px}
.spd button{display:flex;flex-direction:column;align-items:center;gap:5px;padding:8px 2px 7px;font-size:11.5px;color:var(--mut)}
.spd button svg{width:22px;height:14px}
.spd button svg rect{fill:var(--line2)}
.spd button svg rect.f{fill:currentColor}
.spd button.on{color:var(--on-tx);background:var(--on);border-color:var(--on-b)}
.spd button.on svg rect.f{fill:var(--acc)}
[data-theme=webstore] .spd button.on svg rect.f{fill:var(--acc-tx)}
[data-theme=webstore] .spd button{font-size:10.5px}
.lrow{display:flex;align-items:center;gap:10px;margin-top:12px;padding:9px 11px;background:var(--tile);border:1px solid var(--line);border-radius:calc(var(--r) * .7)}
.lrow .ic{width:30px;height:30px;border-radius:50%;display:grid;place-items:center;background:var(--btn);color:var(--mut)}
.lrow.lit .ic{background:#ffd76a33;color:#f5b524}
[data-theme=webstore] .lrow .ic{border-radius:0}
[data-theme=webstore] .lrow.lit .ic{background:var(--acc);color:var(--acc-tx)}
.lrow .ic svg{width:17px;height:17px;fill:currentColor}
.lrow .tx{flex:1}.lrow .tx b{display:block;font-weight:600}.lrow .tx small{color:var(--mut)}
.chead{display:flex;justify-content:space-between;align-items:baseline;margin:14px 0 7px}
.chead span{font-size:12px;color:var(--mut)}.chead b{font-size:12px;font-weight:600}
[data-theme=makerworld] .cbig button.pri{border-radius:999px}
[data-theme=makerworld] .cbig button{border-radius:999px}

/* ================= compact layout ================= */
html,body{font-size:12.5px}
button{padding:4px 9px}
#top{height:38px;gap:8px;padding-left:10px}
#brand{gap:7px;font-size:12.5px}#brand .lens{width:17px;height:17px}#brand .lens:after{width:7px;height:7px;box-shadow:0 0 0 2.5px #f2f5fa}#brand .pix{height:20px}
[data-theme=webstore] #brand b{font-size:14px}
.tab{padding:3px 8px;gap:6px}
#topr{gap:2px}#topr button{padding:4px 8px}
#winbtns button{width:42px}
#bar{height:34px;gap:6px}#bar button{padding:3px 9px}
#side{width:300px;padding:10px 10px 10px 9px;gap:8px}
.card{padding:9px 12px 11px}
.card h3{margin:0 0 8px;font-size:10.5px;display:flex;align-items:center;gap:6px;cursor:pointer;user-select:none;min-height:18px}
[data-theme=makerworld] .card h3{font-size:12px}
.card h3 .ht{flex:none}
.card h3 .sum,.card h3 .xtra{flex:1;min-width:0;text-align:right;font-family:var(--font);text-transform:none;letter-spacing:0;font-weight:400;font-size:11px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.card h3:after{content:"";flex:none;width:6px;height:6px;border-right:1.5px solid var(--dim);border-bottom:1.5px solid var(--dim);transform:rotate(45deg) translate(-1px,-1px);transition:transform .15s}
.card.col h3:after{transform:rotate(-45deg)}
.card:not(.col) .sum{display:none}.card.col .xtra{display:none}
.card.col{padding-bottom:7px}.card.col h3{margin:0}
.card.col>*:not(h3){display:none!important}
.card h3:hover .ht{color:var(--tx)}
.prow{display:flex;align-items:baseline;gap:8px}
.big{font-size:20px}.big.sm{font-size:15px}
.pst{color:var(--mut);font-size:11.5px}
.peta{margin-left:auto;color:var(--mut);font-size:11.5px;font-variant-numeric:tabular-nums;white-space:nowrap}
.bar{margin-top:5px;height:4px}
.pmeta{display:flex;justify-content:space-between;gap:10px;margin-top:5px;font-size:11.5px;color:var(--mut);font-variant-numeric:tabular-nums}
.pmeta .pf{color:var(--tx);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;min-width:0}
.hms{margin-top:6px;padding:4px 7px;font-size:11.5px;display:flex;flex-wrap:wrap;gap:3px 6px;align-items:center}
.hms .code{font-family:var(--mono);font-size:10.5px}
.temps{grid-template-columns:repeat(4,minmax(0,1fr));gap:4px}
.temp{padding:4px 6px;display:flex;flex-direction:column;line-height:1.2}
.temp small{font-size:10px}.temp b{font-size:13.5px}.temp em{display:inline;font-size:10px}
.temp em{margin-left:2px;color:var(--dim)}
.temp .tv{white-space:nowrap}
.lastjob{display:flex;align-items:center;gap:8px;margin-top:8px;font-size:11.5px;color:var(--mut)}
.lastjob b{color:var(--tx)}.lastjob.ok b{color:var(--ok)}.lastjob.bad b{color:var(--err)}
.lastjob button{margin-left:auto;padding:2px 9px;font-size:11px}
.guide{margin-top:16px;border:1px solid var(--line);border-radius:10px;padding:10px 14px;font-size:13px;line-height:1.5;max-width:600px}.guide summary{cursor:pointer;font-weight:600}.guide ol,.guide ul{padding-left:20px;margin:6px 0 12px}.guide li{margin:0 0 8px}.guide>b{display:block;margin-top:6px}
.about{max-width:560px;font-size:13px;line-height:1.5}
.about h5{margin:16px 0 4px;font-size:11px;text-transform:uppercase;letter-spacing:.06em;color:var(--mut)}
.about p{margin:4px 0}.about ul{margin:4px 0;padding-left:18px}.about li{margin:0 0 8px}
.abhead{display:flex;align-items:center;gap:12px;margin-bottom:10px}
.ablogo{width:44px;height:44px;border-radius:var(--r);background:var(--acc);color:#fff;display:grid;place-items:center}
.ablogo svg{width:24px;height:24px;fill:currentColor}
/* no visible scroll bars anywhere - everything still scrolls with the wheel, touch and keys */
*{scrollbar-width:none}*::-webkit-scrollbar{display:none;width:0;height:0}
.pstage{font-size:11.5px;color:var(--acc);margin-top:4px}
.fans{margin-top:6px}.fans small{font-size:9.5px;white-space:nowrap;letter-spacing:0}
.wifiw{display:inline-flex;vertical-align:middle;margin-left:6px}
.wifi{width:16px;height:16px;fill:none;stroke:var(--dim);stroke-width:2.4;stroke-linecap:round}
.wifi circle{fill:var(--dim);stroke:none}
.wifi .on{stroke:var(--ok)}.wifi circle.on{fill:var(--ok)}
.wifi.l1 .on{stroke:var(--warn)}.wifi.l1 circle.on{fill:var(--warn)}
.wifi.l0 .on{stroke:var(--err)}.wifi.l0 circle.on{fill:var(--err)}
.card .more.sub{margin-top:6px;opacity:.85}
.facts{display:grid;grid-template-columns:auto minmax(0,1fr);gap:3px 10px;margin:8px 0 0;font-size:11.5px;line-height:1.35}
.facts dt{color:var(--mut)}.facts dd{margin:0;color:var(--tx);overflow-wrap:anywhere}.facts dd.warn{color:var(--warn)}
.more-facts{border-top:1px solid var(--line);padding-top:8px}
.card .more{margin-top:8px;width:100%;padding:4px 8px;font-size:11.5px}
.upd{font-size:10px;font-weight:700;color:var(--acc);border:1px solid var(--acc);border-radius:99px;padding:0 6px;margin-left:6px}
.mods{margin-top:8px;font-size:11px;border-top:1px solid var(--line);padding-top:6px}
.mods .mh,.mods .mr{display:grid;grid-template-columns:minmax(0,1.2fr) minmax(0,1fr) minmax(0,1fr);gap:6px;padding:2px 0}
.mods .mh{color:var(--dim);font-size:10px;text-transform:uppercase;letter-spacing:.04em}
.mods span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.mods .mono{font-family:var(--mono);font-size:10.5px}.mods .sel{user-select:all}
.amsr{display:grid;grid-template-columns:16px repeat(4,minmax(0,1fr));gap:3px;align-items:center}
.amsr+.amsr{margin-top:3px}
.amsl{font-size:10.5px;font-weight:700;color:var(--mut);text-align:center}
.chip2{position:relative;display:flex;align-items:center;gap:4px;height:22px;padding:0 5px;background:var(--tile);border:1px solid var(--line);border-radius:calc(var(--r) * .5);overflow:hidden;min-width:0}
.chip2.on{border-color:var(--acc);box-shadow:0 0 0 1px var(--acc) inset}
.chip2.empty{border-style:dashed;background:transparent;justify-content:center;color:var(--dim)}
.chip2 .sw8{width:9px;height:9px;border-radius:50%;flex:none;border:1px solid #8886}
.chip2.empty .sw8{display:none}
[data-theme=webstore] .chip2 .sw8{border-radius:0}
.chip2 .ty{font-size:11px;font-weight:650;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;min-width:0}
.chip2.low .ty{color:var(--warn)}
.chip2.on .ty{color:var(--acc)}
.chip2 .pc{margin-left:auto;font-size:9.5px;color:var(--mut);font-variant-numeric:tabular-nums;flex:none}
.chip2.low .pc{color:var(--warn)}
.chip2 .rmb{position:absolute;left:0;bottom:0;height:3px;background:var(--mut);opacity:.7}
.chip2.low .rmb{background:var(--warn);opacity:1}.chip2.on .rmb{background:var(--acc)}
.card:empty,body.remote #bfiles{display:none}
.crow{display:grid;grid-template-columns:1fr 1fr 40px;gap:8px}.crow.few{display:flex}.crow.few>button{flex:1}.crow.few>.lbtn.solo{gap:8px}
.crow button{height:36px;display:flex;align-items:center;justify-content:center;gap:6px;font-weight:600;padding:0 6px}
.crow svg{width:14px;height:14px;fill:currentColor;flex:none}
.crow .lbtn svg{width:17px;height:17px}
.crow .lbtn.on{background:#ffd76a33;border-color:#f5b52488;color:#f5b524}
[data-theme=webstore] .crow .lbtn.on{background:var(--acc);border-color:var(--acc);color:var(--acc-tx)}
[data-theme=makerworld] .crow button{border-radius:999px}
.spd{grid-template-columns:repeat(4,1fr);gap:6px;margin-top:8px}
.spd button{flex-direction:column;justify-content:center;gap:3px;padding:5px 2px;font-size:11px;height:40px}
.spd button svg{width:16px;height:10px}
.ctl-l{display:flex;justify-content:space-between;font-size:11px;color:var(--mut);margin:10px 0 0}
[data-theme=webstore] .spd button{font-size:9.5px;letter-spacing:.02em}
.note{margin-top:5px;font-size:11px}
#balert{position:relative;padding:4px 7px}#balert svg{width:15px;height:15px;fill:currentColor;display:block}
#balert .badge{position:absolute;top:1px;right:1px;min-width:15px;height:15px;border-radius:8px;background:var(--err);color:#fff;font-size:9.5px;font-style:normal;font-weight:700;line-height:15px;text-align:center;padding:0 3px}
[data-theme=webstore] #balert .badge{border-radius:0}
#btl.rec{color:var(--err);border-color:var(--err)}#btl{display:flex;align-items:center;gap:6px}
#pop{display:none;position:fixed;z-index:35;width:320px;max-height:min(60vh,460px);overflow:auto;background:var(--panel);color:var(--tx);border:1px solid var(--line2);border-radius:var(--r);padding:10px 12px;box-shadow:0 12px 32px #0007}
[data-theme=webstore] #pop{box-shadow:6px 6px 0 #000;border-color:var(--acc)}
[data-theme=makerworld] #pop{box-shadow:0 12px 32px #1f232933}
#pop.open{display:block}

#pop .ph{display:flex;justify-content:space-between;align-items:center;gap:8px;margin-bottom:6px}
#pop .ph .ghost{padding:2px 6px;font-size:11px}
#pop .alerts{max-height:none}
.al{padding:5px 0;font-size:12px}
#toasts{top:84px;right:308px;width:300px}
.toast{padding:8px 10px}
#ov{padding:24px 14px 10px}
.chip{padding:3px 8px;font-size:11.5px}
#badge,#rec{top:9px}#badge{left:9px}#rec{right:9px}

/* tidy icon buttons */
.ib{display:inline-flex;align-items:center;justify-content:center;gap:5px;height:28px;min-width:30px;padding:0 7px;background:transparent;border-color:transparent;box-shadow:none;color:var(--mut)}
.ib:hover:not(:disabled){background:var(--btn);color:var(--tx)}
.ib svg{width:16px;height:16px;fill:currentColor;flex:none}
.ib.on{background:var(--on);border-color:var(--on-b);color:var(--on-tx)}
.ib span{font-size:11.5px;font-weight:600}
#topr{display:flex;gap:2px;align-items:center}
#topr .ib{height:30px;min-width:32px}
#bar .tools{display:flex;gap:2px;align-items:center;padding-left:6px;border-left:1px solid var(--line)}
#btl .dot{margin-right:1px}#btl span{font-variant-numeric:tabular-nums}
#btl.rec{color:var(--err);border-color:transparent;background:transparent}
#balert{padding:0 7px}
[data-theme=webstore] .ib{box-shadow:none}
[data-theme=webstore] .ib.on{box-shadow:var(--bshadow)}

/* ---- camera toolbar on the video */
#stage #bar{position:absolute;left:0;right:0;bottom:0;height:38px;padding:0 8px 0 12px;gap:6px;display:flex;align-items:center;
 background:linear-gradient(transparent,rgba(0,0,0,.72) 45%);border:0;color:#e9edf2;z-index:4}
#stage #bar #st{color:#c9cfd7}
#stage #bar .tools{border-left:0;padding-left:0;gap:2px;background:rgba(10,12,15,.55);border:1px solid rgba(255,255,255,.08);border-radius:999px;padding:2px 4px;backdrop-filter:blur(6px)}
[data-theme=webstore] #stage #bar .tools{border-radius:0;border-color:#383633;background:rgba(18,18,18,.8)}
#stage #bar .ib{color:#dfe4ea;height:28px;border-radius:999px}
#stage #bar .ib:hover:not(:disabled){background:rgba(255,255,255,.14);color:#fff}
#stage #bar .ib.on{background:var(--acc);color:var(--acc-tx);border-color:var(--acc)}
#stage #bar #btl.rec{color:#ff6b6b;background:transparent}
[data-theme=webstore] #stage #bar .ib{border-radius:0}
#stage #bar #bretry{height:26px;padding:0 10px}
#stage #ov{bottom:38px;padding-bottom:6px}
body.compact #stage #bar{opacity:0;transition:opacity .2s}
body.compact #stage:hover #bar{opacity:1}

.ib.folder svg{width:17px;height:17px}

/* ---- printer files panel */
#files{position:fixed;inset:0;background:var(--veil);display:none;z-index:41;align-items:center;justify-content:center;padding:18px}
#files.open{display:flex}
#fbox{width:min(980px,100%);height:min(660px,100%);background:var(--panel);border:1px solid var(--line2);border-radius:calc(var(--r) * 1.4);display:flex;flex-direction:column;overflow:hidden;box-shadow:0 20px 60px #0008}
[data-theme=webstore] #fbox{box-shadow:10px 10px 0 #000;border:2px solid var(--acc)}
#fhead{display:flex;align-items:center;gap:8px;padding:10px 12px;border-bottom:1px solid var(--line)}
#fhead h2{margin:0 8px 0 2px;font-size:14px;font-family:var(--label);text-transform:var(--label-tt);letter-spacing:var(--label-ls)}
#fcrumb{flex:1;display:flex;gap:2px;align-items:center;min-width:0;overflow:hidden;color:var(--mut)}
#fcrumb button{padding:2px 6px;background:transparent;border-color:transparent;box-shadow:none;color:var(--mut);font-family:var(--font);text-transform:none;letter-spacing:0}
#fcrumb button:last-child{color:var(--tx);font-weight:600}
#fbody{flex:1;min-height:0;display:grid;grid-template-columns:minmax(0,1fr) 380px}
#flist{overflow-y:auto;border-right:1px solid var(--line);padding:6px}
.frow{display:grid;grid-template-columns:22px minmax(0,1fr) 70px 96px;gap:8px;align-items:center;width:100%;text-align:left;padding:7px 8px;background:transparent;border:1px solid transparent;border-radius:calc(var(--r) * .6);box-shadow:none;color:var(--tx);font-family:var(--font);text-transform:none;letter-spacing:0}
.frow:hover{background:var(--btn)}
.frow.sel,.frow.sel:hover:not(:disabled){background:var(--tile);border-color:var(--acc);color:var(--tx)}.frow.sel .nm{font-weight:650}.frow.sel svg{color:var(--acc);opacity:1}
.frow svg{width:18px;height:18px;fill:currentColor;opacity:.75}
.frow .nm{white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.frow .sz,.frow .dt{color:var(--mut);font-size:11.5px;text-align:right;font-variant-numeric:tabular-nums}
.frow.dim{opacity:.55}
#fdet{overflow-y:auto;overflow-x:hidden;padding:14px;min-width:0}
.fthumb{width:100%;aspect-ratio:1/1;background:#0003;border:1px solid var(--line);border-radius:calc(var(--r) * .8);display:grid;place-items:center;overflow:hidden;color:var(--dim)}
.fthumb img{width:100%;height:100%;object-fit:contain}
.fname{font-weight:700;margin:10px 0 2px;word-break:break-word}
.plates{display:flex;gap:5px;flex-wrap:wrap;margin:10px 0}
.plates button{padding:4px 10px}
.pinfo{display:flex;gap:14px;color:var(--mut);font-size:12px;margin-bottom:10px}
.fhint{font-size:11.5px;color:var(--mut);margin:2px 0 4px}
.fil{display:grid;grid-template-columns:84px minmax(0,1fr);gap:8px;align-items:center;margin-top:8px}
.fneed{display:flex;align-items:center;gap:6px;min-width:0}
.fneed .mut{font-size:11px}
.fil .sw8{width:14px;height:14px;border-radius:50%;border:1px solid #8886;flex:none}
.spick{display:flex;flex-wrap:wrap;gap:3px}
.sp{display:flex;flex-direction:column;align-items:center;gap:1px;width:40px;padding:4px 0 3px;background:var(--tile);border:1px solid var(--line);border-radius:calc(var(--r) * .6);box-shadow:none;font-family:var(--font);text-transform:none;letter-spacing:0}
.sp .sw8{width:18px;height:18px;border-radius:50%;border:1px solid #8888}
.sp small{font-size:10px;color:var(--mut);font-weight:600}
.sp em{font-style:normal;font-size:9px;color:var(--dim);max-width:36px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.sp.other{border-style:dashed}
.sp.other em{color:var(--warn)}
.sp:hover:not(:disabled){background:var(--btn-h)}
.sp.on,.sp.on:hover:not(:disabled){background:var(--tile);border-color:var(--acc);box-shadow:0 0 0 1px var(--acc) inset;color:var(--tx)}
.sp.on small{color:var(--acc)}
[data-theme=webstore] .sp .sw8{border-radius:0}
.fsel{font-size:11.5px;color:var(--mut);margin:3px 0 0 92px}
.fsel b{color:var(--tx)}
[data-theme=webstore] .fil .sw8{border-radius:0}
.fil select{padding:4px 6px;font-size:12px}
.fopts{margin-top:12px;border-top:1px solid var(--line);padding-top:10px}
.fopts .sw{padding:4px 0}
.fgo{margin-top:12px}
.fgo button{width:100%;min-height:38px;height:auto;padding:8px 10px;font-weight:700;white-space:normal;line-height:1.25;overflow-wrap:anywhere}
.pmode{display:flex;align-items:center;gap:6px;margin-top:8px;font-size:11.5px;color:var(--mut)}
.pmode.bad{color:var(--hms-tx)}
.locked{background:var(--hms);border:1px solid var(--hms-b);color:var(--hms-tx);border-radius:calc(var(--r) * .7);padding:10px 12px;font-size:12.5px;line-height:1.45}
.locked p{margin:4px 0 6px;color:var(--tx);opacity:.85}
.locked ol{margin:0 0 8px;padding-left:18px;color:var(--tx)}
.locked button{padding:4px 12px}
.fwarn{font-size:11.5px;color:var(--mut);margin-top:8px;line-height:1.45}
@media (max-width:820px){#fbody{grid-template-columns:1fr;grid-template-rows:45% 55%}#flist{border-right:0;border-bottom:1px solid var(--line)}.frow{grid-template-columns:22px minmax(0,1fr) 64px}.frow .dt{display:none}}
body.compact #top,body.compact #side{display:none}
body.fs #top,body.fs #side{display:none}body.fs .rz{display:none!important}
body.fs #stage #bar,body.fs #badge,body.fs #ov,body.fs #rec{transition:opacity .35s}
body.fs.idle #stage #bar,body.fs.idle #badge,body.fs.idle #ov,body.fs.idle #rec{opacity:0;pointer-events:none}
body.fs.idle,body.fs.idle *{cursor:none!important}
body.compact #cbtns{display:flex}
body.compact .rz{display:none}
body.compact #badge{top:8px;left:8px}
body.compact #toasts{right:10px;top:10px;width:260px}

body.remote .local{display:none!important}
@media (max-width:820px){
 body{overflow:auto}
 #main{flex-direction:column;min-height:auto}
 #stage{flex:none;aspect-ratio:16/9;min-height:200px}
 #side{width:auto;border-left:0;border-top:1px solid var(--line);overflow:visible}
 #toasts{right:10px;left:10px;width:auto;top:56px}
 #stage #bar #st{display:none}
 #brand b{display:none}
 #bcompact{display:none}

 #setbox{flex-direction:column;height:100%}#setnav{width:auto;flex-direction:row;overflow-x:auto;border-right:0;border-bottom:1px solid var(--line)}
 #setnav h2,#setnav .grow{display:none}.grid2,.themes{grid-template-columns:1fr}
}
</style>
<body>
<div id=top>
 <div id=brand><i class=lens></i><svg class=pix viewBox="0 0 17 33" shape-rendering="crispEdges" aria-hidden="true"><path fill="#FF8000" d="M11 13 L4 13 L4 15 L11 15ZM0 21 L6 21 L6 19 L0 19ZM0 24 L6 24 L6 22 L0 22ZM0 27 L6 27 L6 25 L0 25ZM1 18 L17 18 L17 16 L1 16ZM1 30 L17 30 L17 28 L1 28ZM4 33 L17 33 L17 31 L4 31Z"/><path fill="#F5F0E6" d="M8 0 L8 1 L10 1 L10 2 L8 2 L8 3 L10 3 L10 4 L6 4 L6 7 L16 7 L16 4 L12 4 L12 3 L14 3 L14 2 L12 2 L12 1 L14 1 L14 0ZM8 10 L9 10 L9 11 L10 11 L10 12 L12 12 L12 11 L13 11 L13 10 L14 10 L14 9 L15 9 L15 8 L7 8 L7 9 L8 9Z"/></svg><b>Printrment</b></div>
 <div id=tabs></div>
 <div id=topr>
  <button class="ib local" id=bcompact title="Compact view (small, always on top)"></button>
  <button class="ib folder" id=bfiles title="Files on the printer"></button>
  <button class="ib" id=balert title="Alerts and activity"></button>
  <button class="ib local" id=bset title="Settings"></button>
 </div>
 <div id=winbtns>
  <button id=wmin title="Minimise"><svg viewBox="0 0 11 11"><path d="M0 5.5h11"/></svg></button>
  <button id=wmax title="Maximise"><svg viewBox="0 0 11 11"><rect x=".5" y=".5" width="10" height="10"/></svg></button>
  <button id=wclose class=x title="Close"><svg viewBox="0 0 11 11"><path d="M.5.5l10 10M10.5.5l-10 10"/></svg></button>
 </div>
</div>
<div id=main>
 <div id=stage>
  <img id=v alt="">
  <div id=msg></div>
  <div id=pausebox><b>Camera paused</b><span>Printrment has disconnected from the camera. Status, alerts and controls keep working.</span><button class=pri id=bresume2>Resume camera</button></div>
  <div id=badge class=chip style="display:none"></div>
  <div id=rec class=chip><span class="dot err pulse"></span><span>REC</span></div>
  <div id=ov></div>
  <div id=cbtns><button class=ib id=cpin title="Keep on top of other windows"></button><button id=cexit>Exit compact</button></div>
  <div id=bar>
   <div id=st></div>
   <button id=bretry style="display:none">Retry now</button>
   <div class=tools>
    <button class=ib id=btl title="Timelapse"></button>
    <button class=ib id=bcam title="Pause camera"></button>
    <button class=ib id=bsnap title="Save a snapshot"></button>
    <button class=ib id=bfs title="Fullscreen"></button>
   </div>
   <button class="ghost local" id=bquit>Quit</button>
  </div>
 </div>
 <div id=side>
  <div class=card id=cprint></div>
  <div class=card id=cams></div>
  <div class=card id=ctemp></div>
  <div class=card id=cctl></div>
 </div>
</div>
<div id=toasts></div>
<div id=pop></div>
<div id=files><div id=fbox>
 <div id=fhead><h2>Printer files</h2><div id=fcrumb></div><button class=ib id=frefresh title="Refresh"></button><button id=fclose>Close</button></div>
 <div id=fbody><div id=flist></div><div id=fdet></div></div>
</div></div>
<div class="rz t" data-e=t></div><div class="rz b" data-e=b></div><div class="rz l" data-e=l></div><div class="rz r" data-e=r></div>
<div class="rz tl" data-e=tl></div><div class="rz tr" data-e=tr></div><div class="rz bl" data-e=bl></div><div class="rz br" data-e=br></div>
<input type=file id=impfile accept=".json,application/json" hidden>
<div id=set><div id=setbox><div id=setnav></div><div id=setbody></div></div></div>
<script>
"use strict";
const $=s=>document.querySelector(s);
function h(tag,props){const e=document.createElement(tag);if(props)for(const k in props){const v=props[k];if(v==null||v===false)continue;
 if(k==="class")e.className=v;else if(k==="text")e.textContent=v;else if(k.slice(0,2)==="on")e[k]=v;else if(k==="checked"||k==="disabled"||k==="value")e[k]=v;else e.setAttribute(k,v)}
 for(let i=2;i<arguments.length;i++){const c=arguments[i];(Array.isArray(c)?c:[c]).forEach(x=>{if(x==null||x===false)return;if(Array.isArray(x)){x.forEach(y=>{if(y!=null&&y!==false)e.append(y.nodeType?y:document.createTextNode(String(y)))});return}e.append(x.nodeType?x:document.createTextNode(String(x)))})}return e}
// This run's secret token: the app window gets it directly, an Edge fallback
// window in the address fragment (never sent over the network). Phones use
// their sign-in cookie instead and have no token.
const T=(()=>{let t=window.bcToken||"";const m=/[#&]t=([0-9a-f]{16,})/.exec(location.hash);
 if(m){t=m[1];try{history.replaceState(null,"",location.pathname)}catch(e){}}
 try{if(t)sessionStorage.setItem("bc_t",t);else t=sessionStorage.getItem("bc_t")||""}catch(e){}return t})();
const TQ=T?"&t="+T:"";
const HDR={"X-Printrment":T||"1"};
async function api(u,body){const r=await fetch(u,{method:body===undefined?"GET":"POST",headers:Object.assign({"Content-Type":"application/json"},HDR),body:body===undefined?undefined:JSON.stringify(body)});
 if(r.status===401){location.reload();throw new Error("signed out")}
 const ct=r.headers.get("content-type")||"";if(!r.ok){throw new Error((await r.text())||r.statusText)}return ct.includes("json")?r.json():r.text()}
function patch(el,key,build){if(el._k===key)return;el._k=key;el.replaceChildren.apply(el,[].concat(build()).flat().filter(x=>x!=null&&x!==false))}
const NAMES={IDLE:"Idle",PREPARE:"Preparing",RUNNING:"Printing",PAUSE:"Paused",FINISH:"Finished",FAILED:"Failed",SLICING:"Slicing"};
const SPEEDS=["Silent","Standard","Sport","Ludicrous"];
function svg(html,cls){const s=document.createElementNS("http://www.w3.org/2000/svg","svg");s.setAttribute("viewBox","0 0 24 24");if(cls)s.setAttribute("class",cls);s.innerHTML=html;return s}
const IC={pause:'<rect x="6" y="5" width="4" height="14" rx="1"/><rect x="14" y="5" width="4" height="14" rx="1"/>',
 play:'<path d="M8 5.5v13a1 1 0 0 0 1.5.86l10.2-6.5a1 1 0 0 0 0-1.72L9.5 4.64A1 1 0 0 0 8 5.5z"/>',
 stop:'<rect x="6" y="6" width="12" height="12" rx="1.5"/>',
 gear:'<path d="M19.4 13a7.6 7.6 0 0 0 0-2l2.1-1.6a.5.5 0 0 0 .1-.6l-2-3.5a.5.5 0 0 0-.6-.2l-2.5 1a7.4 7.4 0 0 0-1.7-1l-.4-2.6a.5.5 0 0 0-.5-.4h-4a.5.5 0 0 0-.5.4l-.4 2.6a7.4 7.4 0 0 0-1.7 1l-2.5-1a.5.5 0 0 0-.6.2l-2 3.5a.5.5 0 0 0 .1.6L4.6 11a7.6 7.6 0 0 0 0 2l-2.1 1.6a.5.5 0 0 0-.1.6l2 3.5c.1.2.4.3.6.2l2.5-1c.5.4 1.1.7 1.7 1l.4 2.6c0 .2.3.4.5.4h4c.2 0 .5-.2.5-.4l.4-2.6c.6-.3 1.2-.6 1.7-1l2.5 1c.2.1.5 0 .6-.2l2-3.5a.5.5 0 0 0-.1-.6zM12 15.5a3.5 3.5 0 1 1 0-7 3.5 3.5 0 0 1 0 7z"/>',
 pin:'<path d="M15.5 3.5l5 5-2.2.7-3.6 3.6.6 4.4-1.6 1.6-3.6-3.6L5 20.3 3.7 19l5.1-5.1-3.6-3.6 1.6-1.6 4.4.6 3.6-3.6z"/>',
 pip:'<path d="M3 5h18a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1zm1 2v10h16V7zm8 4h7v5h-7z"/>',
 snap:'<path d="M9 4h6l1.5 2H20a1 1 0 0 1 1 1v11a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h3.5zm3 4a4.5 4.5 0 1 0 0 9 4.5 4.5 0 0 0 0-9zm0 2a2.5 2.5 0 1 1 0 5 2.5 2.5 0 0 1 0-5z"/>',
 full:'<path d="M3 3h7v2H5v5H3zm11 0h7v7h-2V5h-5zM3 14h2v5h5v2H3zm16 0h2v7h-7v-2h5z"/>',
 tl:'<path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 2a8 8 0 1 1 0 16 8 8 0 0 1 0-16zm-1 3v5.4l4.3 2.6 1-1.7-3.3-2V7z"/>',
 campause:'<path d="M3 6h12a1 1 0 0 1 1 1v3.2l4.4-3a.5.5 0 0 1 .8.4v8.8a.5.5 0 0 1-.8.4l-4.4-3V17a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1zm3.5 3v6h2V9zm4 0v6h2V9z"/>',
 camplay:'<path d="M3 6h12a1 1 0 0 1 1 1v3.2l4.4-3a.5.5 0 0 1 .8.4v8.8a.5.5 0 0 1-.8.4l-4.4-3V17a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1zm4.5 2.8v6.4l5-3.2z"/>',
 folder:'<path d="M3 5a2 2 0 0 1 2-2h4.2l2 2H19a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>',
 file3mf:'<path d="M6 2h8l5 5v13a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm7 1.5V8h4.5zM8 12l4-2 4 2v4l-4 2-4-2zm4 1.1L9.8 12 12 11l2.2 1z"/>',
 filex:'<path d="M6 2h8l5 5v13a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm7 1.5V8h4.5z"/>',
 refresh:'<path d="M17.7 6.3A8 8 0 1 0 20 12h-2a6 6 0 1 1-1.8-4.3L13 11h7V4z"/>',
 bell:'<path d="M12 22a2.5 2.5 0 0 0 2.45-2h-4.9A2.5 2.5 0 0 0 12 22zm7-6V11a7 7 0 0 0-5.5-6.84V3.5a1.5 1.5 0 0 0-3 0v.66A7 7 0 0 0 5 11v5l-1.7 1.7a.75.75 0 0 0 .53 1.3h16.34a.75.75 0 0 0 .53-1.3z"/>',
 bulb:'<path d="M12 2a7 7 0 0 0-4 12.74V17a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1v-2.26A7 7 0 0 0 12 2zM9 20h6v1a1 1 0 0 1-1 1h-4a1 1 0 0 1-1-1v-1z"/>'};
function bars(n){const s=document.createElementNS("http://www.w3.org/2000/svg","svg");s.setAttribute("viewBox","0 0 22 14");
 for(let i=0;i<4;i++){const r=document.createElementNS("http://www.w3.org/2000/svg","rect");r.setAttribute("x",i*6);r.setAttribute("y",10-i*3.2);r.setAttribute("width",4);r.setAttribute("height",4+i*3.2);r.setAttribute("rx",1);if(i<n)r.setAttribute("class","f");s.append(r)}return s}
function swatch(t){const d=h("span",{class:"sw8"});if(t.colors&&t.colors.length>1){d.style.background="conic-gradient("+t.colors.map((c,i)=>c+" "+(i*100/t.colors.length)+"% "+((i+1)*100/t.colors.length)+"%").join(",")+")"}else d.style.background=t.color||"transparent";return d}
function slotTile(t,label){
 if(t.empty)return h("div",{class:"chipf empty",title:label+": empty"},h("span",{class:"l1"},h("span",{class:"sw8",style:"border-style:dashed"}),h("span",{class:"ty",text:"–"})),h("span",{class:"l2"},h("span",{class:"rm"}),h("span",{class:"pc",text:""})));
 const known=t.remain>=0,low=known&&t.remain<15;
 return h("div",{class:"chipf"+(t.active?" on":""),title:label+": "+(t.brand||t.type)+(known?" · "+t.remain+"% left":" · amount unknown")+(t.active?" · in use":"")},
  h("span",{class:"l1"},swatch(t),h("span",{class:"ty",text:t.type||"?"})),
  h("span",{class:"l2"},h("span",{class:"rm"+(low?" low":"")},h("i",{style:"width:"+(known?t.remain:0)+"%"})),h("span",{class:"pc",text:known?t.remain+"%":"–"})))}
const THEMES=[["default","Default","The original dark look.",["#0b0d10","#171b21","#2f6fed","#313844"]],
 ["webstore","Clarky3D","Your webstore: dark, cream and orange, pixel logo.",["#121212","#1A1A1A","#FF8000","#F5F0E6"]],
 ["makerworld","MakerWorld","Light and clean with Bambu green.",["#f4f5f7","#ffffff","#00ae42","#e6e8eb"]]];
function fmtMin(m){if(!(m>0))return "";const hh=Math.floor(m/60),mm=Math.round(m%60);return hh?hh+"h "+String(mm).padStart(2,"0")+"m":mm+"m"}
function fmtClock(ms){const d=new Date(ms),n=new Date();let s=d.toLocaleTimeString([],{hour:"numeric",minute:"2-digit"});if(d.toDateString()!==n.toDateString())s=d.toLocaleDateString([],{weekday:"short"})+" "+s;return s}
function ago(ms){const s=(Date.now()-ms)/1000;if(s<60)return "just now";if(s<3600)return Math.floor(s/60)+"m ago";if(s<86400)return Math.floor(s/3600)+"h ago";return new Date(ms).toLocaleDateString()}
function stateDot(p){const s=p.status||{};if(!s.connected)return "dot";if(s.state==="RUNNING")return "dot acc pulse";if(s.state==="PAUSE")return "dot warn";if(s.state==="FAILED"||(s.hms&&s.hms.length)||s.print_error)return "dot err";return "dot ok"}

let S=null,active=null,lastEv=-1,remote=false,events=[],cpin=true,fsMode=false,compact=false,armStop=0,cmdNote={},busy={},streamFor=null,logTab="alerts",unseen=0;
let settingsCache=null,setTab="printers",setData=null,editP=null,scanRes=null,scanning=false,rawText="",saveNote="";
try{cpin=localStorage.getItem("bc_cpin")!=="0"}catch(e){}
let fwOpen=false;try{fwOpen=localStorage.getItem("bc_fw")==="1"}catch(e){}
let pinfoOpen=false;try{pinfoOpen=localStorage.getItem("bc_pinfo")==="1"}catch(e){}
function wifiDbm(w){const n=parseInt(w,10);return isNaN(n)?0:n}
function wifiIcon(w){const n=wifiDbm(w),lv=!n?0:n>=-55?3:n>=-65?2:n>=-75?1:0;
 const NS="http://www.w3.org/2000/svg",sv=document.createElementNS(NS,"svg");sv.setAttribute("viewBox","0 0 24 24");sv.setAttribute("class","wifi l"+lv);
 [["M2 9a15 15 0 0 1 20 0",3],["M5 12.5a10.5 10.5 0 0 1 14 0",2],["M8.5 16a5.5 5.5 0 0 1 7 0",1]].forEach(a=>{const p=document.createElementNS(NS,"path");p.setAttribute("d",a[0]);p.setAttribute("class",lv>=a[1]?"on":"");sv.append(p)});
 const c=document.createElementNS(NS,"circle");c.setAttribute("cx","12");c.setAttribute("cy","19.5");c.setAttribute("r","1.6");c.setAttribute("class","on");sv.append(c);
 const w2=h("span",{class:"wifiw",title:"Wi-Fi "+wifiText(w)});w2.append(sv);return w2}
function wifiText(w){const n=wifiDbm(w);if(!n)return w;return n+" dBm ("+(n>=-60?"strong":n>=-70?"OK":"weak")+")"}
const P=()=>S&&S.printers.find(p=>p.id===active);

function applyTheme(t){t=t||"default";document.documentElement.dataset.theme=t;try{localStorage.setItem("bc_theme",t)}catch(e){}
}
applyTheme(document.documentElement.dataset.theme);

function setStream(){const v=$("#v");if(!active){v.removeAttribute("src");streamFor=null;return}
 if(streamFor===active)return;streamFor=active;v.src="/stream?id="+encodeURIComponent(active)+"&n="+Date.now()+TQ}
$("#v").onerror=()=>{streamFor=null;setTimeout(setStream,3000)};

async function poll(){
 try{const j=await api("/api/state?since="+lastEv);j.printers=j.printers||[];j.events=j.events||[];const first=lastEv<0;S=j;remote=j.remote;document.body.classList.toggle("remote",remote);
  if(j.theme&&j.theme!==document.documentElement.dataset.theme)applyTheme(j.theme);
  if(!active||!j.printers.find(p=>p.id===active)){let a=null;try{a=localStorage.getItem("bc_active")}catch(e){}
   if(!a||!j.printers.find(p=>p.id===a))a=j.active;if(!a||!j.printers.find(p=>p.id===a))a=j.printers[0]?j.printers[0].id:null;active=a}
  setStream();
  const evs=j.events||[];if(first)evs.reverse();
  evs.forEach(e=>{if(first){events.push(e);return}events.unshift(e);if(e.kind==="alert"){toast(e);if(!(pop==="alerts"&&logTab==="alerts"))unseen++}});
  events=events.slice(0,120);lastEv=j.last_event;
  render();if(F.open&&!F.infoLoading)renderFilesSoft();
  if(first&&!remote&&!j.printers.length)openSettings("printers",{});
 }catch(e){$("#st").textContent=(e instanceof TypeError&&S)?"Display problem: "+e.message:"Printrment isn't responding - it may have been closed.";$("#st")._k=null;if(window.console)console.error(e)}
}

function selectPrinter(id){active=id;try{localStorage.setItem("bc_active",id)}catch(e){};armStop=0;if(!remote)api("/api/active",{id:id}).catch(()=>{});setStream();render()}

function render(){
 if(!S)return;
 const p=P(),st=p?p.status:{},cam=p?p.cam:{status:""},tl=p?p.timelapse:{};
 patch($("#tabs"),JSON.stringify(S.printers.map(x=>[x.id,x.name,stateDot(x),x.status.state,Math.floor(x.status.percent||0)]).concat([active,remote])),()=>[
  S.printers.map(x=>{const s=x.status,showPct=s.state==="RUNNING"||s.state==="PAUSE";
   return h("button",{class:"tab"+(x.id===active?" sel":""),onclick:()=>selectPrinter(x.id)},h("span",{class:stateDot(x)}),x.name,showPct?h("small",{text:Math.floor(s.percent)+"%"}):null)}),
  remote?null:h("button",{class:"tab",title:"Add a printer",onclick:()=>openSettings("printers",{})},"+ Add")]);

 const live=cam.status==="Live",paused=!!cam.paused,showImg=live||(paused&&cam.has_frame);
 $("#stage").classList.toggle("paused",paused&&cam.has_frame);
 $("#v").style.display=showImg?"block":"none";
 const msg=$("#msg");msg.style.display=showImg?"none":"flex";
 if(!p){patch(msg,"none"+remote,()=>remote?[h("b",{text:"No printers set up"}),h("p",{text:"Add one in Printrment on your PC."})]:
   [h("b",{text:"Add your printer to get started"}),h("p",{text:"Printrment will scan your network for it. You'll need the printer's access code from Settings > Network on the printer."}),h("button",{class:"pri",onclick:()=>openSettings("printers",{})},"Add printer")])}
 else if(paused&&!cam.has_frame){patch(msg,"paused",()=>[h("b",{text:"Camera paused"}),h("p",{text:"Status, alerts and controls keep working."}),h("button",{class:"pri",onclick:()=>camPause(false)},"Resume camera")])}
 else if(!showImg){patch(msg,cam.status,()=>{const s=cam.status||"";const conn=!s||s==="Idle"||s.indexOf("Connecting")===0;
   const hint=s.includes("closed")?"On the printer, check that LAN Mode Liveview is on. The camera connects by itself once the port opens.":s.includes("reach")?"Check the printer is on and on the same network as this computer.":s.includes("access code")?"Re-check the access code on the printer's screen and update it in Settings.":"";
   return [h("b",{text:conn?"Connecting to the camera...":"Camera not available"}),conn?null:h("p",{text:s}),hint?h("p",{style:"opacity:.75",text:hint}):null]})}

 const badge=$("#badge");
 if(p&&st.connected&&st.state&&(compact||fsMode||document.fullscreenElement)){badge.style.display="flex";patch(badge,st.state+stateDot(p),()=>[h("span",{class:stateDot(p)}),(NAMES[st.state]||st.state)])}else badge.style.display="none";
 $("#rec").style.display=tl.recording?"flex":"none";

 const ov=$("#ov"),activeJob=st.state==="RUNNING"||st.state==="PAUSE"||st.state==="PREPARE";
 ov.style.display=p&&st.connected&&activeJob&&showImg&&!paused&&(compact||fsMode||document.fullscreenElement)?"block":"none";
 if(activeJob){const eta=st.remaining>0?" · done ~"+fmtClock(Date.now()+st.remaining*60000):"";
  patch(ov,JSON.stringify([st.file,Math.floor(st.percent),st.layer,st.total_layers,Math.round(st.remaining),st.state]),()=>[
   h("div",{class:"row"},h("div",{class:"file",text:st.file||"Print"}),h("div",{class:"meta",text:(st.total_layers?"Layer "+st.layer+"/"+st.total_layers+" · ":"")+Math.floor(st.percent)+"%"+(st.remaining>0?" · "+fmtMin(st.remaining)+" left":"")+eta})),
   h("div",{class:"bar"+(st.state==="PAUSE"?" warn":"")},h("i",{style:"width:"+Math.min(100,st.percent||0)+"%"}))])}

 const hd=(title,sum,xtra)=>h("h3",null,h("span",{class:"ht",text:title}),h("span",{class:"sum",text:sum||""}),xtra?h("span",{class:"xtra"},xtra):null);
 const name=NAMES[st.state]||st.state||"";
 patch($("#cprint"),JSON.stringify([active,st.connected,st.err,st.state,st.file,Math.floor(st.percent),st.layer,st.total_layers,Math.round(st.remaining),st.hms,st.print_error,st.has,st.mode,st.info&&st.info.stage,st.last_job]),()=>{
  const sum=!p?"":!st.connected?"offline":activeJob?Math.floor(st.percent)+"%"+(st.remaining>0?" · "+fmtMin(st.remaining)+" left":""):name;
  const out=[hd("Print",sum,p&&st.connected?h("span",{class:"ok",title:"Live status from the printer",text:"● live"}):null)];
  if(!p)return out.concat(h("div",{class:"mut",text:"No printer selected."}));
  if(!st.connected){out.push(h("div",{class:"mut",text:st.err&&st.err!=="not set up"?"Status feed: "+st.err:"Connecting to the printer's status feed..."}));
   if(st.err&&st.err.includes("access code"))out.push(h("div",{class:"note",text:"Check the access code in Settings."}));return out}
  if(!st.has)return out.concat(h("div",{class:"mut",text:"Waiting for the first status update..."}));
  if(activeJob){
   out.push(h("div",{class:"prow"},h("span",{class:"big",text:Math.floor(st.percent)+"%"}),h("span",{class:"pst",text:name}),
     h("span",{class:"peta",text:st.remaining>0?fmtMin(st.remaining)+" · "+fmtClock(Date.now()+st.remaining*60000):""})),
    h("div",{class:"bar"+(st.state==="PAUSE"?" warn":"")},h("i",{style:"width:"+Math.min(100,st.percent)+"%"})),
    h("div",{class:"pmeta"},h("span",{class:"pf",title:st.file||"",text:st.file||"-"}),st.total_layers?h("span",{text:"Layer "+st.layer+"/"+st.total_layers}):null));
   if(st.info&&st.info.stage)out.push(h("div",{class:"pstage",text:"Now: "+st.info.stage}));
  if(!activeJob&&st.last_job){const j=st.last_job,ok=j.result==="finished";
   out.push(h("div",{class:"lastjob "+(ok?"ok":j.result==="failed"?"bad":"")},
    h("b",{text:ok?"Finished":j.result==="failed"?"Failed":"Stopped"}),
    h("span",{text:[j.seconds?fmtDur(j.seconds):"",j.layers?j.layers+" layers":"",new Date(j.ended*1000).toLocaleTimeString([],{hour:"numeric",minute:"2-digit"})].filter(Boolean).join(" · ")}),
    j.snapshot&&!remote?h("button",{title:"Show the picture taken when it ended",onclick:()=>api("/api/openfolder",{which:"file",path:j.snapshot}).catch(()=>{})},"Picture"):null))}
  }else{out.push(h("div",{class:"prow"},h("span",{class:"big sm",text:name||"Unknown"})),st.file?h("div",{class:"pmeta"},h("span",{class:"pf",title:st.file,text:(st.state==="FINISH"?"Last: ":st.state==="FAILED"?"Stopped: ":"")+st.file})):null)}
  if(st.print_error)out.push(h("div",{class:"hms",title:"Check the printer's screen for details"},"Print error "+st.print_error));
  if(st.hms&&st.hms.length)out.push(h("div",{class:"hms",title:"Look these codes up on wiki.bambulab.com, or check the printer's screen"},"Alert ",st.hms.map(c=>h("span",{class:"code",text:c.replace("HMS_","")}))));
  return out});

 patch($("#cams"),JSON.stringify([active,st.connected,st.ams,st.external,st.state]),()=>{
  // hide AMS units with no slots and spool holders with nothing on them
  const units=(st.ams||[]).filter(u=>(u.trays||[]).length),ext=(st.external||[]).filter(t=>!t.empty);
  let inUse="";const hum=[];
  units.forEach(u=>{const tag=u.id>=128?"HT":(u.id<26?String.fromCharCode(65+u.id):"?");
   if(u.humidity_pct>=0)hum.push(tag+" "+Math.round(u.humidity_pct)+"%");else if(u.humidity>0)hum.push(tag+" lvl"+u.humidity);
   (u.trays||[]).forEach(t=>{if(t.active)inUse=(tag==="HT"?"HT":tag+(t.slot+1))+" "+(t.brand||t.type)})});
  ext.forEach(t=>{if(t.active)inUse="Ext "+(t.brand||t.type)});
  const printing=["RUNNING","PAUSE","PREPARE"].includes(st.state);
  const out=[hd("AMS",inUse&&printing?"Feeding "+inUse:"",hum.length?h("span",{class:"mut",title:"Humidity inside each AMS",text:"Humidity "+hum.join(" · ")}):null)];
  if(!st.connected)return out.concat(h("div",{class:"mut",text:"Shown once the status feed connects."}));
  if(!units.length&&!ext.length)return out.concat(h("div",{class:"mut",text:"No AMS reported by this printer."}));
  units.forEach(u=>{const tag=u.id>=128?"HT":(u.id<26?String.fromCharCode(65+u.id):"?");
   out.push(h("div",{class:"amsr",title:u.name+(u.humidity_pct>=0?" · humidity "+Math.round(u.humidity_pct)+"%":u.humidity>0?" · humidity level "+u.humidity+"/5":"")+(u.temp>=0?" · "+Math.round(u.temp)+"°C":"")},
    h("b",{class:"amsl",text:tag}),(u.trays||[]).slice(0,4).map(t=>chip(t,tag==="HT"?"HT":tag+(t.slot+1)))))});
  if(ext.length)out.push(h("div",{class:"amsr"},h("b",{class:"amsl",text:"Ext"}),ext.map((t,i)=>chip(t,(st.external||[]).length>1?"External "+(t.slot+1):"External"))));
  return out});

 patch($("#ctemp"),JSON.stringify([active,st.connected,st.nozzles,st.bed,st.chamber,st.info,st.mode,pinfoOpen,fwOpen,p&&p.ip]),()=>{
  const nz=st.nozzles||[],inf=st.info||{};
  const t=(label,x)=>h("div",{class:"temp",title:label+(x&&x.target>0?" · target "+Math.round(x.target)+"°":"")},h("small",{text:label}),h("span",{class:"tv"},h("b",{text:x?Math.round(x.cur)+"°":"–"}),x&&x.target>0?h("em",{text:"/"+Math.round(x.target)}):null));
  const sum=st.connected?[nz[0]?Math.round(nz[0].cur)+"°":"",st.bed?Math.round(st.bed.cur)+"°":""].filter(Boolean).join(" / "):"";
  const upd=inf.updates&&inf.updates.length;
  const xtra=[inf.wifi?wifiIcon(inf.wifi):null,upd?h("span",{class:"upd",title:"Firmware update available"},"Update"):null].filter(Boolean);
  const out=[hd("Printer",sum,xtra.length?xtra:null)];
  if(!st.connected)return out.concat(h("div",{class:"mut",text:"Shown once the status feed connects."}));
  out.push(h("div",{class:"temps"},nz.length?nz.map((x,i)=>t(nz.length>1?"Nozzle "+(i+1):"Nozzle",x)):t("Nozzle",null),t("Bed",st.bed),st.chamber?t("Chamber",st.chamber):null));
  if(inf.fans&&inf.fans.length)out.push(h("div",{class:"temps fans"},inf.fans.map(x=>h("div",{class:"temp",title:x.name+" fan"},h("small",{text:x.name+" fan"}),h("span",{class:"tv"},h("b",{text:x.pct+"%"}))))));
  const f=[];const row=(k,v,cls)=>{if(v)f.push(h("dt",{text:k}),h("dd",{class:cls||"",text:v}))};
  (inf.nozzles||[]).forEach(n=>row((inf.nozzles.length>1?n.label:"Nozzle"),n.diameter+" mm · "+n.type));
  row("Door",inf.door==="open"?"Open":inf.door==="closed"?"Closed":"",inf.door==="open"?"warn":"");
  row("Storage",{ok:"Card OK",missing:"No storage card",problem:"Storage card problem"}[inf.storage],inf.storage&&inf.storage!=="ok"?"warn":"");
  if(f.length)out.push(h("dl",{class:"facts"},f));
  if(upd)out.push(h("div",{class:"note",style:"margin-top:8px"},h("b",null,"Firmware update available"),inf.updates.map(u=>h("div",{class:"mut",text:u})),h("div",{class:"mut",text:"Printrment only tells you - install it from the printer's screen or Bambu Handy."})));
  out.push(h("button",{class:"more",onclick:()=>{pinfoOpen=!pinfoOpen;try{localStorage.setItem("bc_pinfo",pinfoOpen?"1":"0")}catch(e){};render()}},pinfoOpen?"Less info":"More info"));
  if(pinfoOpen){
   const g=[];const r2=(k,v)=>{if(v)g.push(h("dt",{text:k}),h("dd",{text:v}))};
   r2("Model",inf.model);r2("Serial number",inf.serial||st.serial);
   if(inf.wifi)r2("Wi-Fi signal",wifiText(inf.wifi));
   if(!remote)r2("IP address",p.ip);
   r2("Developer Mode",st.mode==="developer"?"On":st.mode==="locked"?"Off":"Not reported");
   (inf.settings||[]).forEach(x=>r2(x.k,x.v));
   out.push(h("dl",{class:"facts more-facts"},g));
   const mods=inf.modules||[];
   if(inf.firmware||mods.length){
    out.push(h("button",{class:"more sub",onclick:()=>{fwOpen=!fwOpen;try{localStorage.setItem("bc_fw",fwOpen?"1":"0")}catch(e){};render()}},(fwOpen?"Hide":"Show")+" firmware"+(inf.firmware?" ("+inf.firmware+")":"")));
    if(fwOpen&&mods.length)out.push(h("div",{class:"mods"},h("div",{class:"mh"},h("span",null,"Part"),h("span",null,"Firmware"),h("span",null,"Serial")),
     mods.map(m=>h("div",{class:"mr",title:m.hardware?"Hardware "+m.hardware:""},h("span",{text:m.name}),h("span",{class:"mono",text:m.firmware||"–"}),h("span",{class:"mono sel",title:m.serial||"",text:m.serial||"–"})))));
   }
  }
  return out});

 const canCtl=!remote||S.remote_controls;
 patch($("#cctl"),JSON.stringify([active,st.connected,st.state,st.speed,st.light,armStop>Date.now(),cmdNote[active],busy[active],canCtl,st.refused,st.mode]),()=>{
  const dis=!st.connected||!!busy[active],job=st.state==="RUNNING"||st.state==="PAUSE"||st.state==="PREPARE",armed=armStop>Date.now();
  const lit=st.light==="on",n=cmdNote[active];
  const sum=(st.speed?SPEEDS[st.speed-1]:"")+(st.light?" · light "+(lit?"on":"off"):"");
  if(!canCtl)return [];
  // Hide what this printer won't accept: pause/resume/stop need Developer Mode,
  // and anything it has already refused.
  const locked=st.mode==="locked",ref=st.refused||[],dev=st.mode==="developer";
  const noPause=!dev&&(locked||ref.includes("pause")||ref.includes("resume")),noStop=!dev&&(locked||ref.includes("stop"));
  const noSpeed=!dev&&ref.includes("print_speed"),noLight=!dev&&ref.includes("ledctrl");
  const row=[noPause?null:st.state==="PAUSE"?h("button",{class:"pri",disabled:dis,onclick:()=>cmd("resume")},svg(IC.play),"Resume"):h("button",{disabled:dis||st.state!=="RUNNING",onclick:()=>cmd("pause")},svg(IC.pause),"Pause"),
    noStop?null:h("button",{class:"danger"+(armed?" armed":""),disabled:dis||!job,onclick:stopClick},svg(IC.stop),armed?"Confirm":"Stop"),
    noLight?null:h("button",{class:"lbtn"+(lit?" on":"")+(noPause&&noStop?" solo":""),disabled:dis,title:"Chamber light "+(lit?"on - click to turn off":"off - click to turn on"),onclick:()=>cmd("light",lit?"off":"on")},svg(IC.bulb),noPause&&noStop?(lit?"Chamber light on":"Chamber light off"):null)].filter(Boolean);
  return [hd("Controls",sum,busy[active]?h("span",{class:"mut",text:"sending..."}):null),
   row.length?h("div",{class:"crow"+(row.length<3?" few":"")},...row):null,
   noSpeed?null:h("div",{class:"ctl-l"},h("span",null,"Print speed"),h("span",{text:st.speed?SPEEDS[st.speed-1]:""})),
   noSpeed?null:h("div",{class:"spd",title:"Print speed"},SPEEDS.map((s,i)=>h("button",{class:st.speed===i+1?"on":"",disabled:dis||!job,title:s+" speed",onclick:()=>cmd("speed",String(i+1))},bars(i+1),s))),

   n?h("div",{class:"note "+(n.ok?"ok":"err"),text:n.text}):null]});

 // timelapse + alerts live behind buttons
 const btl=$("#btl");btl.classList.toggle("rec",!!tl.recording);
 btl.title=tl.recording?"Timelapse recording - "+tl.frames+" frames":tl.rendering?"Saving timelapse...":"Timelapse";
 patch(btl,JSON.stringify([tl.recording,tl.frames,tl.rendering]),()=>tl.recording?[h("span",{class:"dot err pulse"}),h("span",{text:"REC "+tl.frames})]:[svg(IC.tl),tl.rendering?h("span",{text:"Saving"}):null]);
 const bal=$("#balert");patch(bal,String(unseen),()=>[svg(IC.bell),unseen?h("i",{class:"badge",text:unseen>9?"9+":unseen}):null]);
 renderPop();
 patch($("#st"),JSON.stringify([cam.status,Math.round(cam.src_fps),Math.round(cam.out_fps),p&&p.name,paused]),()=>{
  if(!p)return [h("span",{text:"No printer"})];
  const s=cam.status||"Idle";
  return [h("span",{class:"dot "+(live?"ok":paused?"":s.includes("retrying")||s.includes("closed")?"warn":"")}),h("span",{text:live?"Camera live · printer sends "+Math.round(cam.src_fps)+" fps · showing "+Math.round(cam.out_fps)+" fps":paused?"Camera paused - not connected to the printer's camera":s})]});
 $("#bretry").style.display=p&&!paused&&(cam.status||"").includes("retrying")?"":"none";
 const bc=$("#bcam");bc.disabled=!p||(remote&&!S.remote_controls);bc.title=paused?"Resume camera":"Pause camera";bc.classList.toggle("on",paused);patch(bc,String(paused),()=>[svg(paused?IC.camplay:IC.campause),paused?h("span",{text:"Resume"}):null]);
 $("#bsnap").disabled=!live;
 $("#bcompact").classList.toggle("on",compact);$("#cpin").classList.toggle("on",cpin);$("#cpin").title=cpin?"Staying on top - click to let other windows cover it":"Keep on top of other windows";$("#bfs").title=fsMode||document.fullscreenElement?"Exit fullscreen (Esc)":"Fullscreen";
}

function chip(t,label){
 if(t.empty)return h("div",{class:"chip2 empty",title:label+": empty"},h("span",{class:"sw8"}),h("span",{class:"ty",text:"–"}));
 const known=t.remain>=0,low=known&&t.remain<15;
 return h("div",{class:"chip2"+(t.active?" on":"")+(low?" low":""),title:label+": "+(t.brand||t.type)+(known?" · "+t.remain+"% left":" · amount unknown")+(t.active?" · in use":"")},
  swatch(t),h("span",{class:"ty",text:t.type||"?"}),known?h("i",{class:"rmb",style:"width:"+t.remain+"%"}):null)}

// ---- popovers (alerts, timelapse)
let pop=null;
function openPop(which){pop=pop===which?null:which;if(pop==="alerts"){unseen=0}render()}
document.addEventListener("mousedown",e=>{if(!pop)return;if(e.target.closest("#pop,#balert,#btl"))return;pop=null;render()});
function renderPop(){const el=$("#pop");if(!pop||!S){el.className="";el.replaceChildren();el._k=null;return}
 const p=P(),tl=p?p.timelapse:{},cam=p?p.cam:{},canCtl=!remote||S.remote_controls;
 el.className="open "+pop;
 const anchor=$(pop==="alerts"?"#balert":"#btl").getBoundingClientRect(),pw=Math.min(pop==="alerts"?330:290,window.innerWidth-16);
 el.style.width=pw+"px";el.style.left=Math.max(8,Math.min(window.innerWidth-pw-8,anchor.right-pw))+"px";
 if(pop==="alerts"){el.style.top=(anchor.bottom+6)+"px";el.style.bottom="auto"}else{el.style.bottom=(window.innerHeight-anchor.top+6)+"px";el.style.top="auto"}
 if(pop==="alerts"){
  const mine=events.filter(a=>(a.printer===active||!a.printer)&&(logTab==="alerts"?a.kind==="alert":a.kind!=="alert")).slice(0,25);
  patch(el,JSON.stringify(["a",active,logTab,mine.map(a=>a.id),Math.floor(Date.now()/60000)]),()=>[
   h("div",{class:"ph"},h("span",{class:"ltabs"},h("button",{class:logTab==="alerts"?"on":"",onclick:()=>{logTab="alerts";render()}},"Alerts"),h("button",{class:logTab==="activity"?"on":"",onclick:()=>{logTab="activity";render()}},"Activity")),
    mine.length?h("button",{class:"ghost",onclick:()=>{events=events.filter(a=>!(a.printer===active&&(logTab==="alerts"?a.kind==="alert":a.kind!=="alert")));render()}},"Clear"):null),
   mine.length?h("div",{class:"alerts"},mine.map(a=>h("div",{class:"al"},h("span",{class:"dot "+({success:"ok",warn:"warn",error:"err",info:"acc"}[a.level]||"")}),h("span",{text:a.text}),h("time",{text:ago(a.time)})))):
    h("div",{class:"mut",style:"padding:6px 0",text:logTab==="alerts"?"No alerts. Failed or paused prints, printer errors and finished prints show up here.":"Nothing yet. Prints starting, timelapses and things you do show up here."})])
 }else{
  patch(el,JSON.stringify(["t",active,tl,canCtl,remote,cam.paused,settingsCache&&settingsCache.timelapse_auto]),()=>{
   const out=[h("div",{class:"ph"},h("b",null,"Timelapse"),tl.recording?h("span",{class:"err",text:"● Recording"}):null)];
   if(tl.recording)out.push(h("div",{class:"mut",text:tl.frames+" frames · started "+fmtClock(tl.since)}));
   else out.push(h("div",{class:"mut",text:tl.rendering?"Saving video...":"Not recording."}));
   if(tl.recording&&cam.paused)out.push(h("div",{class:"note",text:"No frames are taken while the camera is paused."}));
   if(canCtl)out.push(h("div",{class:"btns",style:"margin-top:8px"},tl.recording?h("button",{onclick:()=>tlAct("stop")},"Stop & save"):h("button",{class:"pri",disabled:!p,onclick:()=>tlAct("start")},"Start recording")));
   if(!remote&&settingsCache)out.push(h("label",{class:"sw",style:"margin-top:8px"},h("span",null,"Record every print automatically"),h("input",{type:"checkbox",checked:settingsCache.timelapse_auto,onchange:e=>saveSetting({timelapse_auto:e.target.checked})}),h("i")));
   if(tl.last_file&&!remote)out.push(h("div",{class:"note"},"Last: ",h("a",{href:"#",onclick:e=>{e.preventDefault();api("/api/openfolder",{which:"file",path:tl.last_file})},text:tl.last_file.split(/[\\/]/).pop()})));
   if(!remote)out.push(h("div",{class:"note"},h("a",{href:"#",onclick:e=>{e.preventDefault();pop=null;openSettings("files")},text:"Timelapse settings"})));
   return out})
 }}

// ---- collapsible cards (remembered)
let collapsed={};try{collapsed=JSON.parse(localStorage.getItem("bc_col")||"{}")}catch(e){}
document.querySelectorAll("#side .card").forEach(c=>{if(collapsed[c.id])c.classList.add("col")});
$("#side").addEventListener("click",e=>{const hd=e.target.closest(".card h3");if(!hd||e.target.closest("button,input,a"))return;
 const c=hd.parentElement;c.classList.toggle("col");collapsed[c.id]=c.classList.contains("col");try{localStorage.setItem("bc_col",JSON.stringify(collapsed))}catch(x){}});
$("#balert").onclick=()=>openPop("alerts");
$("#bset").replaceChildren(svg(IC.gear));$("#bfiles").replaceChildren(svg(IC.folder));$("#cpin").replaceChildren(svg(IC.pin));$("#bcompact").replaceChildren(svg(IC.pip));
$("#bsnap").replaceChildren(svg(IC.snap));$("#bfs").replaceChildren(svg(IC.full));$("#btl").onclick=()=>openPop("tl");

// ---------------- printer files
let F={open:false,path:"/",list:null,err:"",loading:false,sel:null,info:null,infoErr:"",infoLoading:false,plate:1,map:{},opts:{use_ams:true,bed_level:true,flow_cali:false,timelapse:false},arm:0,note:null,busy:false};
function fmtSize(n){return n>=1048576?(n/1048576).toFixed(1)+" MB":n>=1024?Math.round(n/1024)+" KB":n+" B"}
function fmtDur(s){return fmtMin(s/60)||"<1m"}
function openFiles(){if(!P()||remote)return;F.open=true;$("#files").classList.add("open");loadDir(F.path)}
function closeFiles(){F.open=false;$("#files").classList.remove("open")}
$("#bfiles").onclick=openFiles;$("#fclose").onclick=closeFiles;$("#files").onclick=e=>{if(e.target.id==="files")closeFiles()};
$("#frefresh").onclick=()=>loadDir(F.path);
async function loadDir(path){const p=P();if(!p)return;F.path=path;F.loading=true;F.err="";renderFiles();
 try{const r=await api("/api/files?id="+encodeURIComponent(p.id)+"&path="+encodeURIComponent(path));if(r.ok){F.list=r.entries}else{F.err=r.reason;F.list=[]}}catch(e){F.err=String(e.message||e);F.list=[]}
 F.loading=false;renderFiles()}
function joinPath(a,b){return (a.endsWith("/")?a:a+"/")+b}
async function selectFile(path){const p=P();F.sel=path;F.info=null;F.infoErr="";F.infoLoading=true;F.arm=0;F.note=null;renderFiles();
 try{const r=await api("/api/fileinfo?id="+encodeURIComponent(p.id)+"&path="+encodeURIComponent(path));if(F.sel!==path)return;
  if(r.ok){F.info=r.info;F.plate=(r.info.plates[0]||{}).index||1;autoMap()}else F.infoErr=r.reason}catch(e){F.infoErr=String(e.message||e)}
 F.infoLoading=false;renderFiles()}
// every loaded slot, with the index the printer uses for it
function loadedSlots(){const st=(P()||{}).status||{},out=[];
 (st.ams||[]).forEach(u=>{const tag=u.id>=128?"HT":(u.id<26?String.fromCharCode(65+u.id):"?");(u.trays||[]).forEach(t=>{if(!t.empty)out.push({g:u.id>=128?u.id:u.id*4+t.slot,label:(tag==="HT"?"HT":tag+(t.slot+1)),t:t})})});
 const ex=(st.external||[]).filter(t=>!t.empty);ex.forEach(t=>{out.push({g:254+t.slot,label:(ex.length>1?"Ext "+(t.slot+1):"Ext"),t:t})});return out}
function hexRGB(c){c=(c||"#000000").replace("#","");return [0,2,4].map(i=>parseInt(c.substr(i,2),16)||0)}
function cdist(a,b){const x=hexRGB(a),y=hexRGB(b);return Math.sqrt((x[0]-y[0])**2+(x[1]-y[1])**2+(x[2]-y[2])**2)}
// pick the best loaded slot for each filament: same type first, then closest colour
function autoMap(){const pl=(F.info&&F.info.plates||[]).find(x=>x.index===F.plate);F.map={};if(!pl)return;const slots=loadedSlots(),used={};
 pl.filaments.forEach(f=>{let best=null,bs=1e9;slots.forEach(s=>{let sc=cdist(f.color,s.t.color);if((s.t.type||"").toUpperCase()!==(f.type||"").toUpperCase())sc+=600;if(used[s.g])sc+=120;if(sc<bs){bs=sc;best=s}});
  if(best){F.map[f.id]=best.g;used[best.g]=1}})}
async function startPrint(){const p=P(),pl=F.info.plates.find(x=>x.index===F.plate);if(!pl)return;
 if(F.arm<Date.now()){F.arm=Date.now()+5000;renderFiles();setTimeout(renderFiles,5100);return}
 F.arm=0;F.busy=true;F.note=null;renderFiles();
 try{const r=await api("/api/startprint",{id:p.id,path:F.sel,plate:pl.index,mapping:pl.filaments.map(f=>F.map[f.id]),use_ams:F.opts.use_ams,bed_level:F.opts.bed_level,flow_cali:F.opts.flow_cali,timelapse:F.opts.timelapse});
  F.note={ok:r.ok,text:r.ok?"Print started. You can close this panel.":(r.reason||"The printer didn't accept the print.")}}catch(e){F.note={ok:false,text:String(e.message||e)}}
 F.busy=false;renderFiles();poll()}
// Thumbnails are fetched like every other API call (with our header) and
// shown from memory, so no API route needs an exception for <img> tags.
const thumbs=new Map();
function thumbImg(u){const c=thumbs.get(u);if(c)return h("img",{src:c,alt:""});
 if(c===undefined){thumbs.set(u,null);fetch(u,{headers:HDR}).then(r=>r.ok?r.blob():null).then(b=>{if(!b){thumbs.set(u,"");renderFiles();return}
  if(thumbs.size>60){for(const [k,v] of thumbs){if(v)URL.revokeObjectURL(v);thumbs.delete(k);if(thumbs.size<=40)break}}
  thumbs.set(u,URL.createObjectURL(b));renderFiles()}).catch(()=>{thumbs.set(u,"")})}
 return h("span",null,c===undefined||c===null?"Loading preview...":"No preview")}
// What the printer reports about Developer Mode (information only).
function modeLine(st){
 if(st.mode==="developer")return h("div",{class:"pmode ok"},h("span",{class:"dot ok"}),"Developer Mode on");
 if(st.mode==="locked")return h("div",{class:"pmode",title:"Most controls still work. Some commands, like starting a print, may be refused - Printrment will tell you if one is."},h("span",{class:"dot"}),"Developer Mode off");
 return null}
function renderFiles(){if(!F.open)return;const p=P();if(!p){closeFiles();return}
 const parts=F.path.split("/").filter(Boolean);
 $("#fcrumb").replaceChildren(h("button",{onclick:()=>loadDir("/")},p.name),...parts.map((x,i)=>[h("span",null,"/"),h("button",{onclick:()=>loadDir("/"+parts.slice(0,i+1).join("/"))},x)]).flat());
 $("#frefresh").replaceChildren(svg(IC.refresh));
 const L=$("#flist");
 if(F.loading&&!F.list)L.replaceChildren(h("div",{class:"mut",style:"padding:12px"},"Reading the printer's storage..."));
 else if(F.err)L.replaceChildren(h("div",{class:"note err",style:"padding:12px"},F.err));
 else{const rows=[];if(F.path!=="/")rows.push(h("button",{class:"frow",onclick:()=>loadDir("/"+parts.slice(0,-1).join("/"))},svg(IC.folder),h("span",{class:"nm"},".. (up)"),h("span"),h("span")));
  (F.list||[]).forEach(e=>{const fp=joinPath(F.path,e.name),is3mf=/\.3mf$/i.test(e.name);
   rows.push(h("button",{class:"frow"+(F.sel===fp?" sel":"")+(!e.dir&&!is3mf?" dim":""),title:e.name,onclick:()=>e.dir?loadDir(fp):selectFile(fp)},svg(e.dir?IC.folder:is3mf?IC.file3mf:IC.filex),h("span",{class:"nm",text:e.name}),h("span",{class:"sz",text:e.dir?"":fmtSize(e.size)}),h("span",{class:"dt",text:e.time})))});
  if(!rows.length)rows.push(h("div",{class:"mut",style:"padding:12px"},"This folder is empty."));
  L.replaceChildren(...rows)}
 const D=$("#fdet"),st=p.status||{};
 // printers that refuse print starts without Developer Mode: browsing only, no start controls
 const noStart=st.mode!=="developer"&&(st.mode==="locked"||(st.refused||[]).includes("project_file"));
 if(!F.sel){D.replaceChildren(h("div",{class:"mut"},noStart?"Pick a print file (.3mf) to see its plates.":"Pick a print file (.3mf) to see its plates and start it."));return}
 if(F.infoLoading){D.replaceChildren(h("div",{class:"mut"},"Opening the file..."));return}
 if(F.infoErr){D.replaceChildren(h("div",{class:"note err"},F.infoErr));return}
 const inf=F.info,pl=(inf.plates||[]).find(x=>x.index===F.plate),out=[];
 out.push(h("div",{class:"fthumb"},pl&&pl.has_thumb?thumbImg("/api/filethumb?id="+encodeURIComponent(p.id)+"&path="+encodeURIComponent(F.sel)+"&plate="+pl.index):h("span",null,"No preview")),
  h("div",{class:"fname",text:inf.name}),h("div",{class:"mut",style:"font-size:11.5px",text:fmtSize(inf.size)}));
 if(inf.note){out.push(h("div",{class:"note"},inf.note));D.replaceChildren(...out);return}
 if(inf.plates.length>1)out.push(h("div",{class:"plates"},inf.plates.map(x=>h("button",{class:x.index===F.plate?"on":"",onclick:()=>{F.plate=x.index;autoMap();F.arm=0;F.note=null;renderFiles()}},"Plate "+x.index))));
 if(pl){out.push(h("div",{class:"pinfo"},pl.prediction?h("span",null,"⏱ "+fmtDur(pl.prediction)):null,pl.weight?h("span",null,Math.round(pl.weight*10)/10+" g"):null,h("span",null,pl.filaments.length+" filament"+(pl.filaments.length===1?"":"s"))));
  if(noStart){pl.filaments.forEach(f=>out.push(h("div",{class:"fil"},h("div",{class:"fneed"},h("span",{class:"sw8",style:"background:"+(f.color||"transparent")}),h("b",{text:f.type}),h("span",{class:"mut",text:f.used_g?Math.round(f.used_g)+" g":""})))));D.replaceChildren(...out.filter(Boolean));return}
  const slots=loadedSlots();
  if(F.opts.use_ams)out.push(h("div",{class:"fhint",text:"Pick the AMS slot for each filament:"}));
  pl.filaments.forEach(f=>{
   const m=slots.find(s=>s.g===F.map[f.id]),mismatch=m&&(m.t.type||"").toUpperCase()!==(f.type||"").toUpperCase();
   // one colour swatch per loaded slot - the selected one is outlined
   const pick=!F.opts.use_ams?h("span",{class:"mut",style:"font-size:11.5px",text:"AMS off - uses the loaded spool"}):
    !slots.length?h("span",{class:"mut",text:"No filament loaded"}):
    h("div",{class:"spick"},slots.map(s=>{const same=(s.t.type||"").toUpperCase()===(f.type||"").toUpperCase();
     return h("button",{class:"sp"+(F.map[f.id]===s.g?" on":"")+(same?"":" other"),title:s.label+": "+(s.t.brand||s.t.type)+(s.t.remain>=0?" · "+s.t.remain+"% left":"")+(same?"":" (different type)"),
      onclick:()=>{F.map[f.id]=s.g;F.arm=0;renderFiles()}},swatch(s.t),h("small",{text:s.label}),h("em",{text:s.t.type||"?"}))}));
   out.push(h("div",{class:"fil"},h("div",{class:"fneed"},h("span",{class:"sw8",style:"background:"+(f.color||"transparent")}),h("b",{text:f.type}),h("span",{class:"mut",text:f.used_g?Math.round(f.used_g)+" g":""})),pick));
   if(m)out.push(h("div",{class:"fsel"},"→ ",h("b",{text:m.label}),h("span",{text:" "+(m.t.brand||m.t.type)+(m.t.remain>=0?" · "+m.t.remain+"% left":"")})));
   if(mismatch)out.push(h("div",{class:"note err",text:"Slot "+m.label+" has "+m.t.type+", but this filament is "+f.type+"."}))});
  const o=(k,label)=>h("label",{class:"sw"},h("span",null,label),h("input",{type:"checkbox",checked:!!F.opts[k],onchange:e=>{F.opts[k]=e.target.checked;F.arm=0;renderFiles()}}),h("i"));
  out.push(h("div",{class:"fopts"},o("use_ams","Use AMS"),o("bed_level","Bed levelling"),o("flow_cali","Flow calibration"),o("timelapse","Printer's own timelapse")));
  const busyP=["RUNNING","PAUSE","PREPARE","SLICING"].includes(st.state),armed=F.arm>Date.now();
  const why=!st.connected?"Not connected to the printer's status feed.":busyP?"The printer is busy - wait for the current print to finish.":(F.opts.use_ams&&pl.filaments.some(f=>F.map[f.id]==null))?"Choose a slot for every filament.":"";
  out.push(h("div",{class:"fgo"},h("button",{class:armed?"danger armed":"pri",disabled:!!why||F.busy,onclick:startPrint},F.busy?"Sending...":armed?"Click again to start":"Start print")),
   why?h("div",{class:"fwarn",text:why}):h("div",{class:"fwarn",text:"Check the build plate is clear first."}),
   F.note?h("div",{class:"note "+(F.note.ok?"ok":"err"),text:F.note.text}):null)}
 D.replaceChildren(...out.filter(Boolean))}

let fSoftKey="";function renderFilesSoft(){const st=(P()||{}).status||{};const k=JSON.stringify([st.state,st.connected,st.ams,st.external,st.refused,st.mode]);if(k!==fSoftKey){fSoftKey=k;if(!document.activeElement||document.activeElement.tagName!=="SELECT")renderFiles()}}

async function camPause(v){const p=P();if(!p)return;try{await api("/api/campause",{id:p.id,paused:v})}catch(e){toast({level:"error",text:String(e.message||e)})}poll()}
$("#bcam").onclick=()=>{const p=P();if(p)camPause(!p.cam.paused)};
$("#bresume2").onclick=()=>camPause(false);

async function cmd(c,v){const id=active;busy[id]=true;cmdNote[id]=null;render();
 try{const r=await api("/api/cmd",{id:id,cmd:c,value:v||""});cmdNote[id]={ok:r.ok,text:r.ok?({pause:"Print paused.",resume:"Print resumed.",stop:"Print stopped.",speed:"Speed set to "+SPEEDS[(+v)-1]+".",light:"Light "+v+"."}[c]):(r.reason||"The printer refused the command.")}}
 catch(e){cmdNote[id]={ok:false,text:String(e.message||e)}}
 busy[id]=false;armStop=0;render();setTimeout(()=>{if(cmdNote[id]&&cmdNote[id].ok){cmdNote[id]=null;render()}},5000)}
function stopClick(){if(armStop>Date.now()){armStop=0;cmd("stop")}else{armStop=Date.now()+4000;render();setTimeout(render,4100)}}
async function tlAct(a){try{const r=await api("/api/timelapse",{id:active,action:a});if(!r.ok)toast({level:"error",text:r.reason})}catch(e){toast({level:"error",text:String(e.message||e)})}poll()}

function toast(e,action){const t=h("div",{class:"toast "+(e.level||"info")},h("div",null,h("b",{text:e.text})),h("button",{class:"x",onclick:()=>t.remove()},"✕"),
  e.name?h("small",{text:e.name}):null,action?h("button",{class:"act",onclick:()=>{action.fn();t.remove()}},action.label):null);
 $("#toasts").prepend(t);setTimeout(()=>t.remove(),e.level==="error"?20000:e.level==="warn"?12000:7000)}

$("#bsnap").onclick=async()=>{const p=P();if(!p)return;
 if(remote){const a=h("a",{href:"/snapshot.jpg?id="+encodeURIComponent(p.id),download:"snapshot.jpg"});document.body.append(a);a.click();a.remove();return}
 try{const r=await api("/api/snapshot?id="+encodeURIComponent(p.id),{});if(r.ok)toast({level:"success",text:"Snapshot saved",name:r.path.split(/[\\/]/).pop()},{label:"Show in folder",fn:()=>api("/api/openfolder",{which:"file",path:r.path})});else toast({level:"error",text:r.reason})}catch(e){toast({level:"error",text:String(e.message||e)})}};
$("#bretry").onclick=()=>api("/api/retry?id="+encodeURIComponent(active)).catch(()=>{});
function setFull(on){
 if(typeof window.bcWin==="function"){
  // from compact view, the app leaves compact and goes fullscreen in one step
  if(on&&compact){compact=false;document.body.classList.remove("compact")}
  fsMode=on;document.body.classList.toggle("fs",on);window.bcWin(on?"fs_on":"fs_off").catch(()=>{});$("#ov")._k=null;wake();render();return}
 if(on)$("#stage").requestFullscreen().catch(()=>{});else if(document.fullscreenElement)document.exitFullscreen()}
$("#bfs").onclick=()=>setFull(!(fsMode||document.fullscreenElement));
// in fullscreen, tuck everything away after a moment without the mouse moving
let idleT=null;
function wake(){document.body.classList.remove("idle");clearTimeout(idleT);
 if(fsMode||document.fullscreenElement)idleT=setTimeout(()=>{if(!document.querySelector("#bar:hover,#pop.open"))document.body.classList.add("idle");else wake()},2500)}
["mousemove","mousedown","keydown","wheel"].forEach(ev=>document.addEventListener(ev,wake,{passive:true}));
document.addEventListener("fullscreenchange",()=>{if(!native)document.body.classList.toggle("fs",!!document.fullscreenElement);wake()});
document.addEventListener("fullscreenchange",()=>{$("#ov")._k=null;render()});
$("#stage").ondblclick=e=>{if(e.target.closest("button,#bar"))return;if(compact)setCompact(false);else setFull(!(fsMode||document.fullscreenElement))};
$("#bquit").onclick=async()=>{try{await api("/api/quit",{})}catch(e){}document.body.innerHTML="<p style='padding:30px'>Printrment has closed. You can close this window.</p>";setTimeout(()=>window.close(),300)};
function winMode(){if(!remote)api("/api/window",{ontop:compact&&cpin,compact:compact}).catch(()=>{})}
function setCompact(c){if(c&&fsMode)setFull(false);compact=c;document.body.classList.toggle("compact",c);winMode();render()}
$("#bcompact").onclick=()=>setCompact(!compact);$("#cexit").onclick=()=>setCompact(false);
$("#cpin").onclick=()=>{cpin=!cpin;try{localStorage.setItem("bc_cpin",cpin?"1":"0")}catch(e){};winMode();render()};
$("#bset").onclick=()=>openSettings();

// own title bar when running in Printrment's window
const native=typeof window.bcWin==="function"||window.bcNative===true;
let maxed=false;
function win(a){if(!native||typeof window.bcWin!=="function")return;window.bcWin(a).then(r=>{if(r){maxed=!!r.max;document.body.classList.toggle("maxed",maxed);$("#wmax").title=maxed?"Restore":"Maximise"}}).catch(()=>{})}
if(native){document.body.classList.add("native");
 $("#top").addEventListener("mousedown",e=>{if(e.button!==0||e.target.closest("button,input,select,a"))return;e.preventDefault();win("drag")});
 $("#top").addEventListener("dblclick",e=>{if(e.target.closest("button,input,select,a"))return;win("max")});
 $("#wmin").onclick=()=>win("min");$("#wmax").onclick=()=>win("max");$("#wclose").onclick=()=>win("close");
 document.querySelectorAll(".rz").forEach(el=>el.addEventListener("mousedown",e=>{if(e.button!==0)return;e.preventDefault();win(el.dataset.e)}));
 window.addEventListener("resize",()=>win("state"));setTimeout(()=>win("state"),300);
}else{
 // Edge fallback window: tell the app when this window goes away
 window.addEventListener("pagehide",()=>{try{if(!remote)navigator.sendBeacon("/api/closing?x=1"+TQ)}catch(e){}});
}
document.addEventListener("keydown",e=>{if(e.key==="Escape"){if(fsMode)setFull(false);else if(F.open)closeFiles();else if(pop){pop=null;render()}else if($("#set").classList.contains("open"))closeSettings();else if(compact)setCompact(false)}});

// Settings > Printers > Check
const checkRes={};
async function runCheck(id){checkRes[id]={loading:true,items:(checkRes[id]||{}).items,at:(checkRes[id]||{}).at};renderSettings();
 try{const r=await api("/api/printercheck?id="+encodeURIComponent(id));checkRes[id]={items:r.items||[],at:Date.now()}}
 catch(e){checkRes[id]={err:String(e.message||e),at:Date.now()}}
 if(settingsOpen())renderSettings()}
function settingsOpen(){return $("#set").classList.contains("open")}
async function loadSettings(){if(remote)return;try{setData=await api("/api/settings");settingsCache=setData.settings;if(settingsCache)settingsCache.printers=settingsCache.printers||[]}catch(e){}}
async function saveSetting(part,extra){const s=Object.assign({},settingsCache,part,extra||{});try{setData=await api("/api/settings",s);settingsCache=setData.settings;saveNote="Saved."}catch(e){saveNote="Couldn't save: "+e.message}if($("#set").classList.contains("open"))renderSettings();render()}
async function openSettings(tab,edit){if(remote)return;await loadSettings();if(tab)setTab=tab;
 if(edit!==undefined){editP=Object.assign({name:"",ip:"",code:"",serial:"",mode:"auto"},edit);if(!edit.id)scan()}
 $("#set").classList.add("open");renderSettings();
 // run the printer check automatically (once per opening)
 if(settingsCache)settingsCache.printers.forEach(p=>{if(!(checkRes[p.id]||{}).loading)runCheck(p.id)})}
function closeSettings(){$("#set").classList.remove("open");editP=null;scanRes=null;saveNote=""}
$("#set").onclick=e=>{if(e.target.id==="set")closeSettings()};
async function scan(){scanning=true;scanRes=null;renderSettings();try{scanRes=await api("/api/scan")}catch(e){scanRes={printers:[],checked:0}}scanning=false;renderSettings()}
function sw(label,key,note){return h("div",{style:"margin-bottom:10px"},h("label",{class:"sw"},h("span",null,label),h("input",{type:"checkbox",checked:!!settingsCache[key],onchange:e=>{const o={};o[key]=e.target.checked;saveSetting(o)}}),h("i")),note?h("div",{class:"note",style:"margin-top:2px",text:note}):null)}
function field(label,input,note){return h("label",{class:"field"},h("span",null,label),input,note?h("small",{class:"dim",text:note}):null)}
function opt(v,label,sel){return h("option",{value:v,selected:sel?"selected":null},label)}
function srow(label,key,note,btns){return h("div",{class:"srow"},h("div",{class:"st"},h("b",null,label),note?h("small",{text:note}):null),h("span",{class:"btns"},btns),
 h("label",{class:"sw"},h("input",{type:"checkbox",checked:!!settingsCache[key],onchange:e=>{const o={};o[key]=e.target.checked;saveSetting(o)}}),h("i")))}
function test(kind,label){return h("button",{onclick:()=>api("/api/testnotify",{kind:kind})},label)}

function renderSettings(){
 const tabs=[["printers","Printers"],["look","Appearance"],["notify","Alerts & sounds"],["files","Recording & files"],["phone","Remote access"],["app","App"],["diag","Diagnostics"],["about","About"]];
 $("#setnav").replaceChildren(h("h2",null,"Settings"),...tabs.map(t=>h("button",{class:setTab===t[0]?"sel":"",onclick:()=>{setTab=t[0];editP=null;saveNote="";if(t[0]==="diag")loadRaw();renderSettings()}},t[1])),h("div",{class:"grow"}),h("button",{onclick:closeSettings},"Close"));
 const b=$("#setbody"),s=settingsCache;if(!s){b.replaceChildren(h("div",{class:"mut"},"Loading..."));return}
 const out=[];
 if(setTab==="printers"){
  if(editP){const e=editP;
   out.push(h("h4",null,e.id?"Edit printer":"Add a printer"),h("p",{class:"sub"},"Find the IP and access code on the printer: Settings > Network."));
   out.push(h("div",{style:"display:flex;align-items:center;gap:10px;margin-bottom:6px"},h("button",{disabled:scanning,onclick:scan},scanning?"Scanning...":"Scan network"),h("span",{class:"mut",text:scanning?"Looking for printers on your network (about 6 seconds)...":scanRes?(scanRes.printers.length?"Found "+scanRes.printers.length+" - click one to use it.":"None found (checked "+scanRes.checked+" addresses). Type the IP instead."):""})));
   if(scanRes&&scanRes.printers.length)out.push(h("div",{class:"found"},scanRes.printers.map(f=>h("button",{onclick:()=>{e.ip=f.ip;if(!e.name&&f.name&&f.name!=="Possible printer")e.name=f.name;renderSettings()}},h("span",{text:f.name||f.model||"Bambu printer"}),h("span",{class:"mut",text:f.ip})))));
   out.push(h("div",{class:"grid2"},field("Name",h("input",{value:e.name,placeholder:"e.g. X2D",oninput:x=>e.name=x.target.value})),field("IP address",h("input",{value:e.ip,placeholder:"192.168.1.50",oninput:x=>e.ip=x.target.value.trim()}))),
    h("div",{class:"grid2"},field("Access code",h("input",{value:e.code,placeholder:e.id?"saved - type to change":"8 characters",autocomplete:"off",oninput:x=>e.code=x.target.value.trim()})),
     field("Serial number",h("input",{value:e.serial,placeholder:"detected automatically",oninput:x=>e.serial=x.target.value.trim()}),"Leave blank unless detection fails.")),
    field("Camera connection",h("select",{onchange:x=>e.mode=x.target.value},opt("auto","Automatic (RTSP - X1 / H2 / X2 series)",e.mode!=="jpeg"),opt("jpeg","JPEG port 6000 (P1 / A1 series only)",e.mode==="jpeg"))),
    e.id?h("div",{style:"margin:2px 0 10px"},h("label",{class:"sw"},h("span",null,"Trust the printer's new certificate"),h("input",{type:"checkbox",checked:!!e.trust_new,onchange:x=>e.trust_new=x.target.checked}),h("i")),
     h("div",{class:"note",style:"margin-top:2px"},"Printrment remembers each printer's security certificate and refuses to connect if it changes, so nothing else on your network can pretend to be the printer. Tick this only after you've reset, replaced or updated the printer.")):null,
    h("div",{class:"btns",style:"max-width:340px"},h("button",{onclick:()=>{editP=null;scanRes=null;renderSettings()}},"Cancel"),h("button",{class:"pri",onclick:savePrinter},e.id?"Save":"Add printer")),
    e.err?h("div",{class:"note err",text:e.err}):null);
  }else{
   out.push(h("h4",null,"Printers"),h("p",{class:"sub"},"Printrment keeps a status connection to each printer, so you get alerts for all of them. The camera only connects for the printer you're looking at, or while recording."));
   out.push(h("div",{class:"plist"},s.printers.length?s.printers.map(p=>{const live=S&&S.printers.find(x=>x.id===p.id);const st=live?live.status:{};
    const ck=checkRes[p.id];
    return h("div",{class:"pwrap"},h("div",{class:"pitem"},h("span",{class:live?stateDot(live):"dot"}),h("div",null,h("b",{text:p.name}),h("small",{text:p.ip+(p.serial?" · "+p.serial:"")+(st.connected?" · connected":st.err?" · "+st.err:"")+(st.mode==="developer"?" · Developer Mode on":st.mode==="locked"?" · Developer Mode off":"")})),
     h("button",{disabled:!!(ck&&ck.loading),title:"Check LAN Only / Developer Mode, camera, storage and connection",onclick:()=>runCheck(p.id)},ck&&ck.loading?"Checking...":"Check"),
     h("button",{onclick:()=>{editP=Object.assign({},p);renderSettings()}},"Edit"),
     h("button",{class:"danger",onclick:async ev=>{const bt=ev.currentTarget;if(bt.dataset.armed){await api("/api/printer/delete",{id:p.id});await loadSettings();await poll();renderSettings()}else{bt.dataset.armed=1;bt.textContent="Confirm remove";bt.classList.add("armed")}}},"Remove")),
     ck&&(ck.items||ck.err)?h("div",{class:"pcheck"},ck.err?h("div",{class:"ck bad"},h("i"),h("b",null,"Check failed"),h("span",{text:ck.err})):
      ck.items.map(it=>h("div",{class:"ck "+it.state},h("i"),h("b",{text:it.name}),h("span",{text:it.detail}))),
      h("small",{class:"mut",text:"Checked "+new Date(ck.at).toLocaleTimeString()+". This only reads from the printer - it doesn't send it any commands."})):null)}):h("div",{class:"mut"},"No printers yet.")),
    h("button",{class:"pri",onclick:()=>{editP={name:"",ip:"",code:"",serial:"",mode:"auto"};renderSettings();scan()}},"+ Add printer"));
  }
 }else if(setTab==="look"){
  out.push(h("h4",null,"Appearance"),h("p",{class:"sub"},"Pick a theme. Remote access in a browser uses the same one."),
   h("div",{class:"themes"},THEMES.map(t=>{const c=t[3];return h("button",{class:"th"+(s.theme===t[0]?" on":""),onclick:()=>{applyTheme(t[0]);saveSetting({theme:t[0]})}},
    h("div",{class:"pv",style:"background:"+c[0]+";border-radius:"+(t[0]==="webstore"?"0":"8px")},h("i",{class:"a"}),h("span",{class:"b"},h("i",{style:"background:"+c[1]+";border:1px solid "+c[3]}),h("i",{style:"background:"+c[2]}),h("i",{style:"background:"+c[1]+";border:1px solid "+c[3]}))),
    h("span",{class:"nm"},h("b",{text:t[1]+(s.theme===t[0]?"  ✓":"")}),h("small",{text:t[2]})))})));
 }else if(setTab==="notify"){
  out.push(h("h4",null,"Alerts & sounds"),h("p",{class:"sub"},"Alerts are things that need you: a print that failed, was stopped, or was paused by the printer (filament runout, spaghetti detection...), printer error codes, and finished prints. Routine things like a print starting or a timelapse saving only go in the Activity list."),
   h("div",{class:"box"},
    srow("Sound for problems","sound_problems","Failed or stopped prints, printer pauses and error codes.",[test("warn","Test pause"),test("error","Test failure")]),
    srow("Sound when a print finishes","sound_finished","A short chime.",[test("finished","Test")]),
    srow("Flash the taskbar button for alerts","notify_flash","Only when the window isn't in front.",[])),
   h("div",{class:"note",text:"Sounds play from Printrment itself, so you'll hear them even with the window minimised or closed to the tray. Pausing or stopping a print from Printrment doesn't sound an alarm."}));
 }else if(setTab==="files"){
  const sd=h("input",{value:s.snapshot_dir}),td=h("input",{value:s.timelapse_dir}),iv=h("input",{type:"number",min:2,max:600,value:s.timelapse_interval,style:"max-width:120px"});
  out.push(h("h4",null,"Recording & files"),h("p",{class:"sub"},"Where snapshots and timelapse videos are saved."),
   field("Snapshot folder",h("div",{style:"display:flex;gap:6px"},sd,h("button",{onclick:ev=>{ev.preventDefault();api("/api/openfolder",{which:"snapshots"})}},"Open"))),
   field("Timelapse folder",h("div",{style:"display:flex;gap:6px"},td,h("button",{onclick:ev=>{ev.preventDefault();api("/api/openfolder",{which:"timelapses"})}},"Open"))),
   h("div",{class:"box"},sw("Record a timelapse of every print","timelapse_auto","Starts when a print starts and saves an .mp4 when it finishes. The camera stays connected while recording, even with this window closed."),
    field("Take a frame",h("select",{onchange:e=>saveSetting({timelapse_mode:e.target.value})},opt("layer","Every layer (smooth, like Bambu's own timelapse)",s.timelapse_mode==="layer"),opt("interval","Every few seconds",s.timelapse_mode==="interval"))),
    s.timelapse_mode==="interval"?field("Seconds between frames",iv):null),
   h("button",{class:"pri",onclick:()=>saveSetting({snapshot_dir:sd.value,timelapse_dir:td.value,timelapse_interval:parseInt(iv.value)||10})},"Save"),
   saveNote?h("span",{class:"mut",style:"margin-left:10px",text:saveNote}):null);
 }else if(setTab==="phone"){
  const port=h("input",{type:"number",value:s.remote_port,style:"max-width:120px"});
  out.push(h("h4",null,"Remote access"),h("p",{class:"sub"},"A web version of Printrment you can open in the browser on a phone, tablet or another computer on your home network. It's protected by a PIN and isn't reachable from the internet."),
   h("div",{class:"box"},sw("Allow viewing from other devices on my network","remote",setData.remote_error||null),sw("Allow controls (pause, stop, speed, light) from other devices","remote_controls")));
  if(s.remote&&setData.remote_running)out.push(h("div",{class:"box"},h("div",{class:"mut"},"In a browser on the other device, open:"),(setData.remote_urls||[]).map(u=>h("div",{class:"url",text:u})),
    h("div",{class:"mut",style:"margin-top:10px"},"PIN"),h("div",{class:"pin",text:s.remote_pin}),
    h("div",{class:"btns",style:"margin-top:10px;max-width:380px"},h("button",{class:"pri",onclick:openRemote},"Open on this PC"),h("button",{onclick:()=>saveSetting({},{new_pin:true})},"New PIN (signs everyone out)")),
    h("div",{class:"mut",style:"margin-top:6px;font-size:12px",text:"Open on this PC shows the web version in your normal browser, so you can check the page and PIN. It doesn't test Wi-Fi or the firewall - use a phone for that."}),
    remoteNote?h("div",{class:"note "+(remoteNote.ok?"ok":"err"),text:remoteNote.text}):null,
    h("div",{class:"note",text:LIN()?"Fedora's firewall usually allows this already. If other devices can't connect, allow the port: sudo firewall-cmd --permanent --add-port="+(s.remote_port||8788)+"/tcp && sudo firewall-cmd --reload":"The first time, Windows may ask whether to allow Printrment on private networks. Allow it, or other devices can't connect."})));
  out.push(field("Port",h("div",{style:"display:flex;gap:6px"},port,h("button",{onclick:ev=>{ev.preventDefault();saveSetting({remote_port:parseInt(port.value)||8788})}},"Save"))));
  const st=(t,d)=>h("li",null,h("b",{text:t}),h("div",{class:"mut",text:d}));
  out.push(h("details",{class:"guide",open:guideOpen==null?!s.remote:guideOpen,ontoggle:e=>{guideOpen=e.target.open}},h("summary",null,"How to set up remote access"),
   h("ol",null,
    st("Turn it on and try it here","Switch on \"Allow viewing from other devices on my network\" above. An address and a PIN appear. Press \"Open on this PC\" to check the page and PIN in your browser first."),
    LIN()?st("Allow it through the firewall","Fedora's firewall (firewalld) is on, but Fedora Workstation already lets other devices on your network reach ports like this one, so usually there's nothing to do. If they can't connect, run in a terminal: sudo firewall-cmd --permanent --add-port="+(s.remote_port||8788)+"/tcp && sudo firewall-cmd --reload"):
    st("Allow it through Windows Firewall","The first time, Windows asks whether Printrment may use the network. Tick Private networks and choose Allow. If you missed it: Start > Windows Security > Firewall & network protection > Allow an app through firewall > tick Private next to Printrment."),
    st("Put the other device on the same network","Remote access only works on your home network - for a phone that means Wi-Fi, not mobile data. Guest Wi-Fi networks usually block devices from seeing each other, so use the main one."),
    st("Open the address in a browser","On the other device, type the address shown above into the browser (Safari, Chrome, Edge...) exactly as shown, including the :"+(s.remote_port||8788)+" at the end. If more than one address is listed, try the first; the others are for other network connections on this PC."),
    st("Enter the PIN","That browser stays signed in for 90 days. Tip: on a phone, use Add to Home Screen so it opens like an app."),
    st("Controls (optional)","Out of the box, other devices can only watch. Switch on \"Allow controls\" above to let them use the light, speed and other controls too.")),
   h("b",null,"If it doesn't connect"),
   h("ul",null,
    h("li",null,"Printrment must be running on this PC, and the PC awake. With the window closed it only keeps running if \"Closing the window hides Printrment in the tray\" is on (Settings > App)."),
    h("li",null,"Check the firewall step above, and that the other device is on the same network."),
    h("li",null,"The address can change if your router gives this PC a new IP address. Check it here again."),
    h("li",null,"5 wrong PINs from one device lock it out for 15 minutes."),
    h("li",null,"Away from home it won't work, on purpose, so it's never exposed to the internet. A private VPN such as Tailscale on both devices does work.")),
   h("b",null,"Keeping it safe"),
   h("ul",null,
    h("li",null,"Anyone on your Wi-Fi who knows the PIN can view the camera, so don't share it."),
    h("li",null,"\"New PIN\" signs every other device out straight away."),
    h("li",null,"Turn remote access off when you don't need it."))));
 }else if(setTab==="app"){
  out.push(h("h4",null,"App"),h("p",{class:"sub"},"How Printrment starts and closes."),
   h("div",{class:"box"},sw("Closing the window hides Printrment in the tray","tray","Off (default): the close button quits Printrment completely. On: it keeps running in the tray by the clock, so alerts and timelapses carry on - right-click the tray icon and choose Quit to close it fully."+(LIN()?" Fedora's GNOME desktop only shows tray icons with the AppIndicator extension: sudo dnf install gnome-shell-extension-appindicator, then sign out and back in. KDE shows them already.":"")),sw(LIN()?"Start when I sign in":"Start with Windows","autostart","Starts quietly in the tray when you sign in. If you move "+(LIN()?"Printrment":"Printrment.exe")+", turn this off and on again.")),
   h("h4",{style:"margin-top:18px"},"Back up your settings"),
   h("p",{class:"sub"},"Saves your printers, theme and options to a file, to set up another PC or after a reinstall. Access codes, the remote access PIN and printer certificates are left out - you enter the codes again after importing."),
   h("div",{style:"display:flex;gap:8px;flex-wrap:wrap"},h("button",{onclick:exportSettings},"Export settings"),h("button",{onclick:()=>$("#impfile").click()},"Import settings...")),
   backupNote?h("div",{class:"note "+(backupNote.ok?"ok":"err"),style:"margin-top:8px",text:backupNote.text}):null,
   h("div",{class:"mut",style:"margin-top:14px"},"Version "+(S?S.version:"")),native?null:h("button",{style:"margin-top:12px",onclick:()=>$("#bquit").click()},"Quit Printrment"));
 }else if(setTab==="about"){
  const li3=(n,d,l)=>h("li",null,h("b",{text:n}),h("div",{text:d}),h("div",{class:"mut",style:"font-size:11px",text:l}));
  out.push(h("div",{class:"about"},
   h("div",{class:"abhead"},h("div",{class:"ablogo"},svg(IC.snap)),h("div",null,h("h4",{style:"margin:0"},"Printrment"),h("div",{class:"mut",text:"Version "+(S?S.version:"")}))),
   h("p",null,"A small "+(LIN()?"Linux":"Windows")+" app to watch and look after your Bambu Lab printer's camera, status, AMS, files and timelapses without running Bambu's own software."),
   h("h5",null,"Created by"),
   h("p",null,h("b",null,"Clarky3D"),", with ",h("b",null,"Claude"),", Anthropic's AI assistant."),
   h("h5",null,"Third-party software"),
   h("p",{class:"mut"},"Printrment is built from the open-source and freely licensed parts below. Thank you to everyone who makes them."),
   h("ul",null,
    li3("FFmpeg "+(LIN()?"7.0.2 (static build by John Van Sickle)":"7.1"),"Video decoding, snapshots and timelapse rendering. Runs as a separate helper program built into this app.","GPL v3 - (c) the FFmpeg developers - ffmpeg.org"),
    li3("x264 (inside FFmpeg)","H.264 video encoder for timelapse files.","GPL v2 or later - (c) VideoLAN - videolan.org"),
    LIN()?li3("Your web browser","Shows this window: a Chrome, Chromium, Edge or Brave app window if one is installed (including Flatpak), otherwise a tab in your default browser.",""):li3("Microsoft Edge WebView2","Shows this window. Uses the WebView2 Runtime already installed with Windows.","Microsoft WebView2 licence - (c) Microsoft"),
    li3("go-webview2","Go bindings that host WebView2"+(LIN()?" (Windows version only).":"."),"MIT - (c) John Chadwick, Serge Zaitsev"),
    li3("go-winloader","Loads the WebView2 loader"+(LIN()?" (Windows version only).":"."),"ISC - (c) John Chadwick"),
    li3("fyne.io/systray","The tray icon and menu.","Apache 2.0 - (c) Fyne.io and the systray authors"),
    li3("godbus/dbus",LIN()?"Talks to the desktop: tray icon and showing files in the file manager.":"Linux desktop support library (included in the code, unused on Windows).","BSD 2-clause - (c) Georg Reinke, Google"),
    li3("golang.org/x/sys","Operating system calls.","BSD 3-clause - (c) The Go Authors"),
    li3("Go","The programming language and standard library this app is written in.","BSD 3-clause - (c) The Go Authors - go.dev")),
   h("h5",null,"Fonts"),
   h("ul",null,
    li3("Exo 2","Main text.","SIL Open Font Licence 1.1 - (c) Natanael Gama"),
    li3("Space Mono","Numbers and code.","SIL Open Font Licence 1.1 - (c) Colophon Foundry"),
    li3("Silkscreen","Pixel-style labels.","SIL Open Font Licence 1.1 - (c) Jason Kottke")),
   h("h5",null,"Printer protocols"),
   h("p",{class:"mut"},"Printrment talks to your printer the same way other local tools do: MQTT over TLS for status and commands, RTSPS for the camera, FTPS for files. These are standard protocols; the Bambu-specific message formats were worked out from the printer's own behaviour and public community documentation."),
   h("h5",null,"Licences"),
   h("p",{class:"mut"},"FFmpeg and x264 are free software under the GNU General Public License. You can get their source code, and the licence texts, from ffmpeg.org and videolan.org. The other components are used under the licences named above; full texts are available from each project. Printrment itself is provided as is, with no warranty."),
   h("h5",null,"Your data"),
   h("p",{class:"mut"},LIN()?"Settings, printer details and logs are stored in your home folder (~/.config/Printrment). Access codes and PINs are encrypted with a key kept separately (~/.local/share/Printrment), readable only by you. Nothing is uploaded anywhere.":"Settings, printer details and logs are stored on this PC in your Windows user profile. Access codes and PINs are encrypted with Windows. Nothing is uploaded anywhere."),
   h("h5",null,"Good to know"),
   h("p",{class:"mut"},"Printrment is an independent project and is not made by, or connected to, Bambu Lab. Bambu Lab and its product names belong to their owners. It talks to your printer over your local network only; nothing is sent to the internet.")))
 }else if(setTab==="diag"){
  out.push(h("h4",null,"Diagnostics"),h("p",{class:"sub"},"The raw status the selected printer is sending. Useful if a value shows wrong."),
   h("button",{onclick:loadRaw},"Refresh"),h("pre",{text:rawText||"Loading..."}));
 }
 b.replaceChildren(...out.filter(Boolean));
}
const LIN=()=>!!(S&&S.os==="linux");
let backupNote=null,guideOpen=null,remoteNote=null;
async function openRemote(){try{const r=await api("/api/openremote",{});remoteNote=r.ok?{ok:true,text:"Opened in your browser."}:{ok:false,text:r.reason}}catch(e){remoteNote={ok:false,text:String(e.message||e)}}renderSettings();setTimeout(()=>{remoteNote=null;if(setTab==="phone")renderSettings()},4000)}
async function exportSettings(){try{const r=await fetch("/api/export",{headers:HDR});if(!r.ok)throw new Error(await r.text());
  const u=URL.createObjectURL(await r.blob()),a=h("a",{href:u,download:"printrment-settings.json"});document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(u),2000);
  backupNote={ok:true,text:"Settings exported."}}catch(e){backupNote={ok:false,text:String(e.message||e)}}renderSettings()}
async function importSettings(file){try{const txt=await file.text();let j;try{j=JSON.parse(txt)}catch(e){throw new Error("That file isn't readable.")}
  const r=await api("/api/import",j);if(!r.ok)throw new Error(r.reason);
  await loadSettings();await poll();applyTheme(settingsCache.theme);
  backupNote={ok:true,text:"Imported: "+r.added+" printer"+(r.added===1?"":"s")+" added, "+r.updated+" updated."+(r.need_code&&r.need_code.length?" Enter the access code for "+r.need_code.join(", ")+" in Printers > Edit.":"")}}
 catch(e){backupNote={ok:false,text:String(e.message||e)}}renderSettings()}
async function loadRaw(){try{rawText=await api("/api/raw?id="+encodeURIComponent(active||""));if(typeof rawText!=="string")rawText=JSON.stringify(rawText,null,2)}catch(e){rawText=String(e)}if(setTab==="diag")renderSettings()}
async function savePrinter(){const e=editP;e.err="";if(!e.ip||(!e.code&&!e.id)){e.err="Enter the IP address and access code.";renderSettings();return}
 try{const r=await api("/api/printer",{id:e.id||"",name:e.name||"",ip:e.ip,code:e.code||"",serial:e.serial||"",mode:e.mode||"auto",trust_new:!!e.trust_new});
  if(!r.ok){e.err=r.reason;renderSettings();return}const wasNew=!e.id;editP=null;scanRes=null;await loadSettings();await poll();if(wasNew)selectPrinter(r.id);renderSettings();
  if(wasNew&&S&&S.printers.length===1)closeSettings()}catch(x){e.err=String(x.message||x);renderSettings()}}

$("#impfile").onchange=e=>{const f=e.target.files[0];e.target.value="";if(f)importSettings(f)};
if(!T&&location.port==="8787"){document.body.replaceChildren(h("div",{style:"margin:auto;max-width:420px;padding:24px;text-align:center;font:15px system-ui;color:#c9ced6"},
  h("b",null,"Open Printrment from its own window"),h("p",null,"For security, this page only works inside the Printrment window. Start Printrment from the Start menu, the desktop or the tray icon.")));document.body.style.display="flex";document.body.style.background="#0b0d10"}
else loadSettings().then(()=>{poll();setInterval(poll,1000)});
</script>`
