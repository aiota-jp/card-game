const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright');
const fs=require('fs'),path=require('path'),{spawn}=require('child_process'),readline=require('readline');
const root=path.resolve(__dirname,'..'),aliases=JSON.parse(fs.readFileSync(path.join(root,'frontend/asset-map.json'))),pending=[];
const server=spawn(path.join(root,'build/ui-test.exe'),['-test.run=^TestUIBridge$'],{env:{...process.env,CARDLOUNGE_UI_TEST:'1'},stdio:['pipe','pipe','inherit']});
readline.createInterface({input:server.stdout}).on('line',line=>{if(line.startsWith('{')){const p=pending.shift(),v=JSON.parse(line);if(p)v.error?p.reject(Error(v.error)):p.resolve(v.result)}});
const rpc=(controller,method,args=[])=>new Promise((resolve,reject)=>{pending.push({resolve,reject});server.stdin.write(JSON.stringify({controller,method,args})+'\n')});
const cast=['misaki','ayaka','rina','shizue','kathy'];
(async()=>{let browser;try{
 browser=await chromium.launch({channel:'chrome',headless:true,args:['--autoplay-policy=no-user-gesture-required']});
 const page=await browser.newPage({viewport:{width:1024,height:700}}),errors=[],missing=[],played=[],results=[];
 page.on('pageerror',e=>errors.push(e.message));
 await page.exposeFunction('gameBridge',rpc);await page.exposeFunction('voiceEvent',url=>played.push(url));
 await page.addInitScript(()=>{if(window===window.top){localStorage.setItem('cardlounge.music.enabled','false');localStorage.setItem('cardlounge.lounge.rate','1.794')}const A=window.Audio;window.Audio=function(...args){const a=new A(...args);a.addEventListener('playing',()=>window.voiceEvent(a.src));return a}});
 await page.route('http://fivecast.test/**',async route=>{
  const name=new URL(route.request().url()).pathname.slice(1)||'index.html',target=path.join(root,'frontend',aliases[name]||name);
  if(!fs.existsSync(target)){missing.push(name);return route.fulfill({status:404,body:''})}
  let body=fs.readFileSync(target);if(name==='index.html')body=body.toString().replace('<script type="module"',`<script>window.go={main:{}};for(const [c,ms]of Object.entries({PokerApp:['Snapshot','OpenGame','Act','Draw','NextHand','Reset'],BlackjackApp:['Snapshot','Deal','Act','NextHand','Reset'],BaccaratApp:['Snapshot','OpenGame','SetRole','Deal','Act','NextHand','Reset']})){window.go.main[c]={};for(const m of ms)window.go.main[c][m]=(...a)=>window.gameBridge(c,m,a)}</script><script type="module"`);
  await route.fulfill({body,contentType:({'.html':'text/html; charset=utf-8','.js':'application/javascript','.css':'text/css','.png':'image/png','.wav':'audio/wav'})[path.extname(target)]||'application/json'});
 });
 await page.goto('http://fivecast.test/');await page.locator('#suite-title-enter').click();
 if(await page.evaluate(()=>localStorage.getItem('cardlounge.lounge.rate'))!=='1')throw Error('Old default migration failed');
 for(const width of [1024,640,390]){
  await page.setViewportSize({width,height:700});
  for(const cid of cast){await page.locator(`[data-host="${cid}"]`).click();await page.locator('#host-image').evaluate(img=>img.decode());if(await page.evaluate(()=>document.documentElement.scrollWidth>innerWidth))throw Error('Home overflow '+width);}
  await page.screenshot({path:path.join(root,`docs/five-cast-home-${width}.png`),fullPage:true});
 }
 await page.setViewportSize({width:1280,height:850});
 for(const cid of cast){
  await page.locator(`[data-host="${cid}"]`).click();
  await page.locator('#greeting-box.speaking').waitFor();
  const line=await page.locator('#greeting-box').getAttribute('data-line');if(!line.startsWith(cid+'-'))throw Error('Wrong greeting '+cid);
  for(const game of ['poker','blackjack','baccarat']){
   if(['shizue','kathy'].includes(cid)&&game!=='baccarat')await rpc(game==='poker'?'PokerApp':'BlackjackApp','Reset');
   await page.locator(`[data-game="${game}"]`).click();const frame=page.frameLocator('#game-frame');
   if(game==='poker')await frame.locator('#variant-texas').click();
   if(game==='baccarat')await frame.locator('#variant-punto').click();
   await frame.locator('#home-start').waitFor();
   await frame.locator(`body[data-character="${cid}"]`).waitFor();
   const imgs=await frame.locator('body').evaluate(async(_,character)=>{const {opponents}=await import('./opponents.js');const p=opponents[character];for(const src of [p.full,p.portrait,p.reward]){const img=new Image();img.src=src;await img.decode()}return p},cid);
   await frame.locator('#home-start').click();
   await frame.locator('#dealer-stage.speaking').waitFor();
   await frame.locator('.dealer-portrait').evaluate(img=>img.decode());
   if(!['misaki','ayaka','rina'].includes(cid)){
    if(game==='poker'){
     await frame.locator('[data-action="fold"]').click();await frame.locator('#busy').waitFor({state:'hidden'});
    }else{
     await frame.locator('[data-action="deal"]').click();await frame.locator('#busy').waitFor({state:'hidden'});
     const ctrl=game==='blackjack'?'BlackjackApp':'BaccaratApp';
     for(let i=0;i<10;i++){const s=await rpc(ctrl,'Snapshot');if(s.done)break;await frame.locator(`[data-action="${game==='blackjack'?'stand':s.phase==='dealt'?'continue':'stand'}"]`).click();await frame.locator('#busy').waitFor({state:'hidden'})}
     if(!(await rpc(ctrl,'Snapshot')).done)throw Error('Failed settlement '+game+' '+cid);
    }
   }
   if(cid==='shizue'||cid==='kathy')await page.screenshot({path:path.join(root,`docs/five-cast-${game}-${cid}.png`),fullPage:true});
   const voiceLine=await frame.locator('#dealer-stage').getAttribute('data-line');
   results.push({character:cid,game,images:imgs,voiceLine});
   await page.locator('#back').click();await page.locator('#hub').waitFor({state:'visible'});
  }
 }
 const moduleChecks=await page.evaluate(async cast=>{
  const {snapshot,applyBackup,validateBackup}=await import('./suite-tools.js');
  for(const cid of cast){localStorage.setItem('cardlounge.lounge.character',cid);const raw=JSON.stringify(snapshot());validateBackup(raw);applyBackup(raw);if(localStorage.getItem('cardlounge.lounge.character')!==cid)throw Error('Backup lost '+cid)}
  return true;
 },cast);
 if(errors.length||missing.length||!played.length)throw Error(JSON.stringify({errors,missing,played}));
 const report={results,layouts:[1024,640,390],backups:moduleChecks,oldDefaultMigrated:true,audioPlaying:played.length,errors,missing};
 fs.writeFileSync(path.join(root,'docs/five-cast-ui-check.json'),JSON.stringify(report,null,2));console.log(JSON.stringify(report));
}finally{if(browser)await browser.close();server.kill()}})().catch(e=>{console.error(e);server.kill();process.exitCode=1});
