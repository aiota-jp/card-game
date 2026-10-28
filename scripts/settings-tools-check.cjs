const {chromium}=require(process.env.PLAYWRIGHT_MODULE),fs=require('fs'),path=require('path');
const root=path.resolve(__dirname,'..'),stage=path.join(root,'docs'),aliases=JSON.parse(fs.readFileSync(path.join(root,'frontend/asset-map.json')));
(async()=>{const browser=await chromium.launch({channel:'chrome',headless:true,args:['--autoplay-policy=no-user-gesture-required']});try{
const page=await browser.newPage({viewport:{width:1024,height:700}}),errors=[];page.on('pageerror',e=>errors.push(e.message));
await page.addInitScript(()=>{window.savedBackup='';window.exitCalls=0;window.go={main:{SystemApp:{SaveBackup:async raw=>{window.savedBackup=raw;return 'test.json'},LoadBackup:async()=>window.savedBackup,Exit:async()=>{window.exitCalls++}}}}});
await page.route('http://settings.test/**',async route=>{const n=new URL(route.request().url()).pathname.slice(1)||'index.html',p=path.join(root,'frontend',aliases[n]||n);if(!fs.existsSync(p))return route.fulfill({status:404,body:''});await route.fulfill({body:fs.readFileSync(p),contentType:({'.js':'application/javascript','.css':'text/css','.html':'text/html; charset=utf-8','.png':'image/png','.wav':'audio/wav'})[path.extname(p)]||'application/json'})});
await page.goto('http://settings.test/');await page.locator('#suite-title-enter').click();await page.locator('#hub').waitFor();
await page.screenshot({path:stage+'/card-home.png'});
await page.locator('#settings-open').click();await page.locator('#setting-music-volume').fill('17');await page.locator('#setting-voice').uncheck();await page.locator('#settings-save').click();
if(await page.evaluate(()=>localStorage.getItem('cardlounge.music.volume'))!=='17')throw Error('settings save');
await page.screenshot({path:stage+'/card-settings.png'});
await page.locator('#backup-save').click();await page.waitForFunction(()=>window.savedBackup.length>0);
await page.locator('#setting-music-volume').fill('65');await page.locator('#settings-save').click();await page.locator('#backup-restore').click();await page.locator('#restore-cancel').click();
if(await page.evaluate(()=>localStorage.getItem('cardlounge.music.volume'))!=='65')throw Error('cancel mutated settings');
await page.locator('#backup-restore').click();await page.locator('#restore-confirm').click();
if(await page.evaluate(()=>localStorage.getItem('cardlounge.music.volume'))!=='17')throw Error('roundtrip');
if(!await page.evaluate(()=>localStorage.getItem('cardlounge.backup.before-restore')))throw Error('missing safety copy');
const rejected=await page.evaluate(async()=>{const m=await import('/suite-tools.js');const before=JSON.stringify(m.snapshot());let rejected=0;for(const raw of ['not json',JSON.stringify({format:'Other',settings:{}}),JSON.stringify({format:'CardGameLoungeBackup/1',settings:{'cardlounge.lounge.volume':'200'}})]){try{m.applyBackup(raw)}catch{rejected++}}return rejected===3&&before===JSON.stringify(m.snapshot())});if(!rejected)throw Error('invalid backup accepted');
await page.locator('#settings-close').click();await page.locator('#exit-open').click();await page.locator('#exit-cancel').click();if(await page.evaluate(()=>window.exitCalls)!==0)throw Error('exit cancel');await page.locator('#exit-open').click();await page.locator('#exit-confirm').click();if(await page.evaluate(()=>window.exitCalls)!==1)throw Error('exit callback');await page.locator('#exit-cancel').click();
for(const size of [{width:880,height:600},{width:390,height:700}]){await page.setViewportSize(size);await page.locator('#settings-open').click();const b=await page.locator('#settings-dialog').boundingBox();if(b.x<0||b.x+b.width>size.width+1||b.y<0||b.y+b.height>size.height+1)throw Error('dialog layout');await page.screenshot({path:stage+`/card-settings-${size.width}.png`});await page.locator('#settings-close').click()}
if(errors.length)throw Error(errors.join('\n'));console.log('CARD SETTINGS / BACKUP / RESTORE / EXIT: PASS');
}finally{await browser.close()}})().catch(e=>{console.error(e);process.exitCode=1});
