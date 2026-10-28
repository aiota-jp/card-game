const {chromium}=require(process.env.PLAYWRIGHT_MODULE||'playwright'),fs=require('fs'),path=require('path'),{spawn}=require('child_process'),readline=require('readline');
const root=path.resolve(__dirname,'..'),aliases=JSON.parse(fs.readFileSync(path.join(root,'frontend/asset-map.json'))),pending=[];
const server=spawn(path.join(root,'build/ui-test.exe'),['-test.run=^TestUIBridge$'],{env:{...process.env,CARDLOUNGE_UI_TEST:'1'},stdio:['pipe','pipe','inherit']});
readline.createInterface({input:server.stdout}).on('line',line=>{if(line.startsWith('{')){const p=pending.shift(),v=JSON.parse(line);if(p)v.error?p.reject(Error(v.error)):p.resolve(v.result)}});
const rpc=(controller,method,args=[])=>new Promise((resolve,reject)=>{pending.push({resolve,reject});server.stdin.write(JSON.stringify({controller,method,args})+'\n')});
(async()=>{let browser;try{
 browser=await chromium.launch({channel:'chrome',headless:true,args:['--autoplay-policy=no-user-gesture-required']});
 const page=await browser.newPage({viewport:{width:1366,height:768}}),errors=[],missing=[];page.on('pageerror',e=>errors.push(e.message));
 await page.exposeFunction('gameBridge',rpc);
 await page.addInitScript(()=>{window.trackedAudio=[];window.played=[];const A=window.Audio;window.Audio=function(...args){const a=new A(...args);window.trackedAudio.push(a);a.addEventListener('playing',()=>window.played.push(a.src));return a}});
 await page.route('http://topmusic.test/**',async route=>{const n=new URL(route.request().url()).pathname.slice(1)||'index.html',p=path.join(root,'frontend',aliases[n]||n);if(!fs.existsSync(p)){missing.push(n);return route.fulfill({status:404,body:''})}let body=fs.readFileSync(p);if(n==='index.html')body=body.toString().replace('<script type="module"',`<script>window.go={main:{}};for(const[c,ms]of Object.entries({PokerApp:['Snapshot','OpenGame','Act','Draw','NextHand','Reset'],BlackjackApp:['Snapshot','Deal','Act','NextHand','Reset'],BaccaratApp:['Snapshot','OpenGame','SetRole','Deal','Act','NextHand','Reset']})){window.go.main[c]={};for(const m of ms)window.go.main[c][m]=(...a)=>window.gameBridge(c,m,a)}</script><script type="module"`);await route.fulfill({body,contentType:({'.html':'text/html; charset=utf-8','.js':'application/javascript','.css':'text/css','.png':'image/png','.wav':'audio/wav'})[path.extname(p)]||'application/json'})});
 await page.goto('http://topmusic.test/');await page.locator('#suite-title-image').evaluate(x=>x.decode());await page.locator('#suite-title.counting').waitFor();const started=Date.now();
 await page.waitForFunction(()=>document.body.dataset.lobbyMusicPlaying==='true');
 const title=await page.evaluate(()=>{const a=trackedAudio.find(a=>a.src.includes('lounge-bgm.wav'));return {image:document.getElementById('suite-title-image').getAttribute('src'),alt:document.getElementById('suite-title-image').alt,loop:a.loop,volume:a.volume,playing:!a.paused}});
 if(title.image!=='lounge-title-five.png?v=20261028-kathy'||!title.loop||!title.playing||title.volume!==.25||!title.alt.includes('5人'))throw Error('Wrong title/BGM '+JSON.stringify(title));
 if(await page.evaluate(()=>played.some(x=>x.includes('/cast-voice/'))))throw Error('Voice started over title');
 await page.screenshot({path:path.join(root,'docs/top-five-bgm-1366.png')});
 await page.locator('#suite-title').waitFor({state:'hidden'});const durationMs=Date.now()-started;if(durationMs<4500||durationMs>6500)throw Error('Wrong duration '+durationMs);
 await page.locator('#greeting-box.speaking').waitFor();
 let volume=await page.evaluate(()=>trackedAudio.find(a=>a.src.includes('lounge-bgm.wav')).volume);if(Math.abs(volume-.055)>.001)throw Error('Greeting did not duck music '+volume);
 await page.locator('#greeting-stop').click();volume=await page.evaluate(()=>trackedAudio.find(a=>a.src.includes('lounge-bgm.wav')).volume);if(volume!==.25)throw Error('Music did not recover');
 await page.locator('#settings-open').click();await page.locator('#setting-music-volume').fill('40');await page.locator('#settings-save').click();await page.locator('#settings-close').click();
 if(await page.evaluate(()=>trackedAudio.find(a=>a.src.includes('lounge-bgm.wav')).volume)!==.4)throw Error('Settings volume not applied');
 await page.locator('[data-game="blackjack"]').click();const frame=page.frameLocator('#game-frame');await frame.locator('#home-start').waitFor();
 if(!await page.evaluate(()=>trackedAudio.find(a=>a.src.includes('lounge-bgm.wav')).paused))throw Error('Lobby music overlapped game');
 await frame.locator('#home-start').click();await page.locator('#back').click();await page.waitForFunction(()=>document.body.dataset.lobbyMusicPlaying==='true');
 await page.locator('#hub-music-toggle').click();if(await page.evaluate(()=>localStorage.getItem('cardlounge.music.enabled'))!=='false')throw Error('Mute not saved');
 await page.reload();await page.locator('#suite-title-image').evaluate(x=>x.decode());if(!await page.evaluate(()=>trackedAudio.find(a=>a.src.includes('lounge-bgm.wav')).paused))throw Error('Saved mute ignored');
 await page.locator('#title-music-toggle').click();await page.waitForFunction(()=>document.body.dataset.lobbyMusicPlaying==='true');
 for(const size of [{width:1920,height:1040},{width:1024,height:700},{width:880,height:600},{width:390,height:700},{width:640,height:360}]){
  await page.setViewportSize(size);await page.reload();await page.locator('#suite-title-image').evaluate(x=>x.decode());
  for(const id of ['suite-title-enter','title-music-toggle','title-music-volume']){const b=await page.locator('#'+id).boundingBox();if(!b||b.x<0||b.y<0||b.x+b.width>size.width+1||b.y+b.height>size.height+1)throw Error('Control offscreen '+id+' '+JSON.stringify(size))}
  await page.screenshot({path:path.join(root,`docs/top-five-bgm-${size.width}x${size.height}.png`)});
 }
 if(errors.length||missing.length)throw Error(JSON.stringify({errors,missing}));
 const report={fivePersonTitle:title,autoAdvanceMs:durationMs,titleMusicPlays:true,loop:true,greetingDucking:true,settingsVolume:true,gameMusicDoesNotOverlap:true,lobbyMusicResumes:true,mutePersists:true,titleToggleWorks:true,responsive:true,errors,missing};fs.writeFileSync(path.join(root,'docs/top-bgm-check.json'),JSON.stringify(report,null,2));console.log(JSON.stringify(report));
}finally{if(browser)await browser.close();server.kill()}})().catch(e=>{console.error(e);server.kill();process.exitCode=1});
