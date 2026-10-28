const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright'),fs=require('fs'),path=require('path'),{spawn}=require('child_process'),readline=require('readline');
const root=path.resolve(__dirname,'..'),aliases=JSON.parse(fs.readFileSync(path.join(root,'frontend/asset-map.json'))),pending=[],baseline=process.argv.includes('--baseline');
const server=spawn(path.join(root,'build/ui-test.exe'),['-test.run=^TestUIBridge$'],{env:{...process.env,CARDLOUNGE_UI_TEST:'1'},stdio:['pipe','pipe','inherit']});
readline.createInterface({input:server.stdout}).on('line',line=>{if(line.startsWith('{')){const p=pending.shift(),v=JSON.parse(line);if(p)v.error?p.reject(Error(v.error)):p.resolve(v.result)}});
const rpc=(controller,method,args=[])=>new Promise((resolve,reject)=>{pending.push({resolve,reject});server.stdin.write(JSON.stringify({controller,method,args})+'\n')});
let delayNext,failNext;const counts={};
(async()=>{let browser;try{
 browser=await chromium.launch({channel:'chrome',headless:true,args:['--autoplay-policy=no-user-gesture-required']});const page=await browser.newPage({viewport:{width:1280,height:850}}),errors=[],missing=[],checks=[];page.on('pageerror',e=>errors.push(e.message));page.on('dialog',d=>d.accept());
 await page.exposeFunction('gameBridge',async(c,m,a)=>{const key=c+'.'+m;counts[key]=(counts[key]||0)+1;if(delayNext===key){delayNext=null;await new Promise(r=>setTimeout(r,500))}if(failNext===key){failNext=null;throw Error('AUDIT simulated temporary connection error')}return rpc(c,m,a)});
 await page.addInitScript(()=>{localStorage.setItem('cardlounge.lounge.voice','false');localStorage.setItem('cardlounge.music.enabled','false')});
 await page.route('http://audit.test/**',async route=>{const n=new URL(route.request().url()).pathname.slice(1)||'index.html',p=path.join(root,'frontend',aliases[n]||n);if(!fs.existsSync(p)){missing.push(n);return route.fulfill({status:404,body:''})}let body=fs.readFileSync(p);if(n==='index.html')body=body.toString().replace('<script type="module"',`<script>window.go={main:{}};for(const[c,ms]of Object.entries({PokerApp:['Snapshot','OpenGame','Act','Draw','NextHand','Reset'],BlackjackApp:['Snapshot','Deal','Act','NextHand','Reset'],BaccaratApp:['Snapshot','OpenGame','SetRole','Deal','Act','NextHand','Reset']})){window.go.main[c]={};for(const m of ms)window.go.main[c][m]=(...a)=>window.gameBridge(c,m,a)}</script><script type="module"`);await route.fulfill({body,contentType:({'.html':'text/html; charset=utf-8','.js':'application/javascript','.css':'text/css','.png':'image/png','.wav':'audio/wav'})[path.extname(p)]||'application/json'})});
 await page.goto('http://audit.test/');await page.locator('#suite-title-enter').click();
 let frame;async function hub(){if(await page.locator('#game-frame').count()){await frame.locator('#busy').waitFor({state:'hidden'});await page.locator('#back').click();await page.locator('#hub').waitFor({state:'visible'})}}
 async function open(game,cid='misaki',variant){await hub();await page.locator(`[data-host="${cid}"]`).click();await page.locator(`[data-game="${game}"]`).click();frame=page.frameLocator('#game-frame');await frame.locator(`body[data-character="${cid}"]`).waitFor();if(variant)await frame.locator('#variant-'+variant).click();await frame.locator('#home-start').waitFor()}
 async function run(name,fn){try{await fn();checks.push({name,passed:true})}catch(e){checks.push({name,passed:false,error:e.message});console.log('FAILED: '+name+' '+e.message)}finally{await hub()}}
 await run('backup rejects JavaScript-only numeric literals',async()=>{
  const bad=await page.evaluate(async()=>{const {allowedPreference}=await import('./suite-tools.js');return ['null','0x10','+25','025','0b10'].filter(v=>allowedPreference('cardlounge.music.volume',v))});if(bad.length)throw Error('Accepted '+bad.join(','));
 });
 await run('baccarat character switch opens its own role/session',async()=>{
  await rpc('BaccaratApp','OpenGame',['banque','shizue']);await rpc('BaccaratApp','Deal',[20,'player']);for(let i=0;i<8;i++){const s=await rpc('BaccaratApp','Snapshot');if(s.done)break;await rpc('BaccaratApp','Act',[s.phase==='dealt'?'continue':'stand'])}
  await open('baccarat','shizue','banque');if(!await frame.locator('#bank-role').isDisabled())throw Error('Prepared prior session missing');
  await frame.locator('.character-switch:visible').click();await frame.locator('body[data-character="kathy"]').waitFor();await frame.locator('#busy').waitFor({state:'hidden'});
  const s=await rpc('BaccaratApp','Snapshot');if(await frame.locator('#bank-role').isDisabled()||s.seats[1].id!=='kathy'||s.hand!==0)throw Error('Previous SHIZUE session/role still displayed');
  await frame.locator('#bank-role').selectOption('banker');await frame.locator('#home-start').click();await frame.locator('#game-screen').waitFor({state:'visible'});if((await rpc('BaccaratApp','Snapshot')).role!=='banker')throw Error('New host banker role not applied');
 });
 for(const [game,ctrl,method,action]of [['poker','PokerApp','Act','fold'],['blackjack','BlackjackApp','Deal','deal'],['baccarat','BaccaratApp','Deal','deal']]){
  await run(game+' home cannot interrupt an in-flight action',async()=>{
   await rpc(ctrl,'Reset');await open(game,'misaki',game==='poker'?'texas':game==='baccarat'?'punto':undefined);await frame.locator('#home-start').click();delayNext=ctrl+'.'+method;
   await frame.locator(`[data-action="${action}"]`).click();if(!await frame.locator('#home-button').isDisabled())throw Error('Home remains enabled while request is pending');await frame.locator('#busy').waitFor({state:'hidden'});
  });
 }
 for(const [game,ctrl,method,action]of [['poker','PokerApp','Act','fold'],['blackjack','BlackjackApp','Deal','deal'],['baccarat','BaccaratApp','Deal','deal']]){
  await run(game+' recovers after a rejected action',async()=>{
   await rpc(ctrl,'Reset');await open(game,'misaki',game==='poker'?'texas':game==='baccarat'?'punto':undefined);await frame.locator('#home-start').click();failNext=ctrl+'.'+method;
   await frame.locator(`[data-action="${action}"]`).click();await frame.locator('#error').waitFor({state:'visible'});await frame.locator('#busy').waitFor({state:'hidden'});if(await frame.locator(`[data-action="${action}"]`).isDisabled())throw Error('Action remains disabled after rejection');
   await frame.locator(`[data-action="${action}"]`).click();await frame.locator('#busy').waitFor({state:'hidden'});if(await frame.locator('#error').isVisible())throw Error('Retry did not clear error');
  });
 }
 await run('double-clicking Deal sends one request',async()=>{
  await rpc('BlackjackApp','Reset');await open('blackjack');await frame.locator('#home-start').click();const initial=counts['BlackjackApp.Deal']||0;delayNext='BlackjackApp.Deal';await frame.locator('[data-action="deal"]').evaluate(b=>{b.click();b.click()});await frame.locator('#busy').waitFor({state:'hidden'});if(counts['BlackjackApp.Deal']!==initial+1)throw Error('Duplicate deal sent');
 });
 const report={checks,passed:checks.filter(x=>x.passed).length,failed:checks.filter(x=>!x.passed).length,errors,missing};fs.writeFileSync(path.join(root,baseline?'docs/audit-ui-before.json':'docs/audit-ui-check.json'),JSON.stringify(report,null,2));console.log(JSON.stringify(report));if(!baseline&&(report.failed||errors.length||missing.length))throw Error('Audit regression failed');
}finally{if(browser)await browser.close();server.kill()}})().catch(e=>{console.error(e);server.kill();process.exitCode=1});
