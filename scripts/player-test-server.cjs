// Isolated, loopback-only test harness. Production source and user settings are untouched.
const http=require('http'),fs=require('fs'),path=require('path'),{spawn}=require('child_process'),readline=require('readline'),crypto=require('crypto');
const root=path.resolve(__dirname,'..'),front=path.join(root,'frontend'),aliases=JSON.parse(fs.readFileSync(path.join(front,'asset-map.json'))),queue=[],events=[],settled=new Map(),failures=[];
const bridge=spawn(path.join(root,'build/player-test-bridge.exe'),['-test.run=^TestUIBridge$'],{cwd:root,env:{...process.env,CARDLOUNGE_UI_TEST:'1'},stdio:['pipe','pipe','inherit'],windowsHide:true});
readline.createInterface({input:bridge.stdout}).on('line',line=>{if(!line.startsWith('{'))return;const q=queue.shift();if(!q)return;const value=JSON.parse(line);value.error?q.reject(Error(value.error)):q.resolve(value.result)});
const rpc=(controller,method,args)=>new Promise((resolve,reject)=>{queue.push({resolve,reject});bridge.stdin.write(JSON.stringify({controller,method,args})+'\n')});
let session='';const generations=new Map();
function capture(req,view){
 if(req.method==='OpenGame')session=req.controller+':'+view.variant+':'+(req.controller==='BaccaratApp'?req.args[1]:'');
 if(req.controller==='BlackjackApp')session='BlackjackApp:blackjack:';
 if(req.method==='Reset')generations.set(session,(generations.get(session)||0)+1);
 const state={variant:view.variant,hand:view.hand,phase:view.phase,stage:view.stage,done:view.done,role:view.role,stacks:view.stacks,chips:view.chips,pot:view.pot,bet:view.bet,profit:view.profit,outcome:view.outcome,result:view.result,cleared:view.cleared,canNext:view.canNext};
 const event={time:new Date().toISOString(),controller:req.controller,method:req.method,args:req.args,ui:req.ui,state};events.push(event);
 if(view.done)settled.set(session+':'+(generations.get(session)||0)+':'+view.hand,{...state,controller:req.controller,session});
 fs.writeFileSync(path.join(root,'docs/player-test-events.json'),JSON.stringify({events,settled:[...settled.values()],failures},null,2));
}
const injection=`<script>window.go={main:{}};for(const[c,ms]of Object.entries({PokerApp:['Snapshot','OpenGame','Act','Draw','NextHand','Reset'],BlackjackApp:['Snapshot','Deal','Act','NextHand','Reset'],BaccaratApp:['Snapshot','OpenGame','SetRole','Deal','Act','NextHand','Reset']})){window.go.main[c]={};for(const m of ms)window.go.main[c][m]=async(...args)=>{const d=document.querySelector('#game-frame')?.contentDocument;const text=id=>d?.getElementById(id)?.textContent?.trim();const ui={handName:text('hand-name'),equity:text('share'),hint:text('action-hint'),playerTotal:text('player-total'),dealerTotal:text('dealer-total'),chips:text('chips'),role:text('role-note'),visiblePlayerCards:[...(d?.querySelectorAll('#player-cards .card:not(.back):not(.empty)')||[])].map(e=>e.getAttribute('aria-label'))};const r=await fetch('/__bridge',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({controller:c,method:m,args,ui})});const v=await r.json();if(!r.ok)throw Error(v.error);return v;}}</script>`;
const server=http.createServer(async(req,res)=>{try{
 const url=new URL(req.url,'http://127.0.0.1');
 if(url.pathname==='/__stats'){res.setHeader('Content-Type','application/json');res.end(JSON.stringify({events:events.length,hands:[...settled.values()],failures}));return}
 if(url.pathname==='/__bridge'&&req.method==='POST'){
  let raw='';for await(const b of req)raw+=b;if(raw.length>100000)throw Error('request too large');const q=JSON.parse(raw);if(!['PokerApp','BlackjackApp','BaccaratApp'].includes(q.controller))throw Error('unsupported test controller');
  const view=await rpc(q.controller,q.method,q.args);capture(q,view);res.setHeader('Content-Type','application/json');res.end(JSON.stringify(view));return;
 }
 const n=decodeURIComponent(url.pathname.slice(1))||'index.html';const target=path.resolve(front,aliases[n]||n);if(!target.startsWith(front+path.sep))throw Error('invalid path');
 if(!fs.existsSync(target)||!fs.statSync(target).isFile()){failures.push({path:n,status:404});res.writeHead(404);res.end();return}
 let data=fs.readFileSync(target);if(n==='index.html')data=Buffer.from(data.toString().replace('<script type="module"',injection+'<script type="module"'));
 const types={'.html':'text/html; charset=utf-8','.js':'application/javascript','.css':'text/css','.png':'image/png','.wav':'audio/wav','.json':'application/json'};
 res.setHeader('Content-Type',types[path.extname(target)]||'application/octet-stream');res.setHeader('Cache-Control','no-store');res.end(data);
}catch(e){res.writeHead(500,{'Content-Type':'application/json'});res.end(JSON.stringify({error:e.message}));failures.push({path:req.url,error:e.message})}});
server.listen(0,'127.0.0.1',()=>{const port=server.address().port;fs.writeFileSync(path.join(root,'build/player-test-port.json'),JSON.stringify({port,pid:process.pid,productionEXEHash:crypto.createHash('sha256').update(fs.readFileSync(path.join(root,'card-game-lounge.exe'))).digest('hex')}));console.log('PLAYER TEST http://127.0.0.1:'+port)});
process.on('SIGINT',()=>{bridge.kill();server.close();process.exit(0)});process.on('exit',()=>bridge.kill());
