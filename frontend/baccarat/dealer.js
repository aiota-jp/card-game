import {voiceLines as shizueLines} from './shizue-voice-lines.js';
import {voiceLines as shizueReward} from './shizue-reward-lines.js';
import {voiceLines as kathyLines} from './kathy-voice-lines.js';
import {voiceLines as kathyReward} from './kathy-reward-lines.js';
import {voiceLines as misakiReward} from './misaki-reward-lines.js';
import {voiceLines as ayakaReward} from './ayaka-reward-lines.js';
import {voiceLines as rinaReward} from './rina-reward-lines.js';
import {voiceLines as rinaLines} from './rina-voice-lines.js';
import {voiceLines as ayakaLines} from './ayaka-voice-lines.js';
import {voiceLines} from './voice-lines.js';
export {voiceLines};
const libraries={misaki:[...voiceLines,...misakiReward],ayaka:[...ayakaLines,...ayakaReward],rina:[...rinaLines,...rinaReward],shizue:[...shizueLines,...shizueReward],kathy:[...kathyLines,...kathyReward]};
for(const id of Object.keys(libraries))libraries[id]=libraries[id].filter(x=>!/私|うち|負けちゃ|やられ|こっち|こちら/.test(x.text));
export const lines=Object.fromEntries([...new Set(voiceLines.map(x=>x.kind))].map(kind=>[kind,voiceLines.filter(x=>x.kind===kind).map(x=>x.text)]));

// Shuffle bags use every line in a category before refilling; avoid boundary repeats.
export function createLinePicker(random=Math.random,library=voiceLines){
 const bags=new Map(),last=new Map();
 return kind=>{
  if(!library.some(x=>x.kind===kind))kind='welcome';
  let bag=bags.get(kind);
  if(!bag?.length){bag=library.filter(x=>x.kind===kind);for(let i=bag.length-1;i>0;i--){const j=Math.floor(random()*(i+1));[bag[i],bag[j]]=[bag[j],bag[i]]}if(bag.at(-1)?.id===last.get(kind)&&bag.length>1)[bag[0],bag[bag.length-1]]=[bag.at(-1),bag[0]];bags.set(kind,bag)}
  const line=bag.pop();last.set(kind,line.id);return line;
 };
}
export function normalizeVoiceRate(value){const n=Number(value);return value===null||value===''||!Number.isFinite(n)?1.0:Math.round(Math.max(.8,Math.min(2.5,n))*1000)/1000;}
export function createDealer({onSpeaking=()=>{},audioFactory=path=>new Audio(path),character='misaki'}={}){
 const $=id=>document.getElementById(id),streaks=new Map();let pick=createLinePicker(Math.random,libraries[character]||voiceLines);let enabled=true,volume=30,rate=1.0,expressive=true,spoken,generation=0,timer,audio,lastAmbient=0,raiseHand='';
 try{const savedRate=localStorage.getItem('cardlounge.lounge.rate');const migrate=!localStorage.getItem('cardlounge.lounge.rate-default-v3');const oldDefault=Number(savedRate)===1.38||(!localStorage.getItem('cardlounge.lounge.rate-default-v2')&&Number(savedRate)===1.15);rate=normalizeVoiceRate(migrate&&(savedRate===null||oldDefault)?null:savedRate);localStorage.setItem('cardlounge.lounge.rate-default-v3','1');localStorage.setItem('cardlounge.lounge.rate',String(rate));localStorage.setItem('cardlounge.lounge.rate-default-v2','1');enabled=(localStorage.getItem('cardlounge.lounge.voice')??'true')==='true';const savedVolume=localStorage.getItem('cardlounge.lounge.volume');const migrateVolume=!localStorage.getItem('cardlounge.lounge.volume-default-v2');volume=savedVolume===null||(migrateVolume&&Number(savedVolume)===65)?30:Math.max(0,Math.min(100,Number(savedVolume)||0));localStorage.setItem('cardlounge.lounge.volume',String(volume));localStorage.setItem('cardlounge.lounge.volume-default-v2','1');expressive=(localStorage.getItem('cardlounge.lounge.expression')??'true')==='true'}catch{}
 const credit='AivisSpeech：美咲＝凛音エル／あやか＝まお／リナ＝みちのくあいり／静江＝まい／キャシー＝コハク。音声は内蔵されています。';
 function finish(id){if(id!==generation)return;$('dealer-stage').classList.remove('speaking');onSpeaking(false);audio=null}
 function stop(){generation++;clearTimeout(timer);if(audio){audio.pause();audio=null}$('dealer-stage').classList.remove('speaking');onSpeaking(false)}
 function controls(){$('dealer-voice').textContent=enabled?'音声 ON':'音声 OFF';$('dealer-voice').setAttribute('aria-pressed',String(enabled));$('dealer-rate').value=rate;$('dealer-rate-value').textContent=String(rate)+'倍';$('dealer-volume').value=volume;$('dealer-volume-value').textContent=volume+'%';$('dealer-voice-note').textContent=credit;$('dealer-expression').textContent=expressive?'抑揚 ON':'抑揚 OFF';$('dealer-expression').setAttribute('aria-pressed',String(expressive))}
 function speak(delay=0){
  stop();if(!spoken||!enabled||volume===0||document.hidden)return;
  const id=generation;
  timer=setTimeout(()=>{
   if(id!==generation||document.hidden)return;
   try{
    const player=audioFactory(spoken.files[expressive?'expressive':'calm'].path);audio=player;player.volume=volume/100;player.playbackRate=rate;player.preservesPitch=true;if('webkitPreservesPitch' in player)player.webkitPreservesPitch=true;
    player.onplaying=()=>{if(id===generation){$('dealer-stage').classList.add('speaking');onSpeaking(true)}};
    player.onended=()=>finish(id);
    const failed=()=>{if(id!==generation)return;finish(id);$('dealer-voice-note').textContent='音声を再生できませんでした。「もう一度聞く」で再生できます。 '+credit};
    player.onerror=failed;Promise.resolve(player.play()).catch(failed);
   }catch{finish(id)}
  },delay);
 }
 function say(kind,mood=kind,delay=80){spoken=pick(kind);$('dealer-line').textContent=spoken.text;$('dealer-stage').dataset.mood=mood;$('dealer-stage').dataset.line=spoken.id;speak(delay);lastAmbient=Date.now()}
 function save(){try{localStorage.setItem('cardlounge.lounge.rate',String(rate));localStorage.setItem('cardlounge.lounge.voice',String(enabled));localStorage.setItem('cardlounge.lounge.volume',String(volume));localStorage.setItem('cardlounge.lounge.expression',String(expressive))}catch{}}
 $('dealer-voice').onclick=()=>{enabled=!enabled;save();controls();if(enabled)speak();else stop()};$('dealer-repeat').onclick=()=>speak();
 $('dealer-volume').oninput=e=>{volume=Number(e.target.value);save();controls();if(audio)audio.volume=volume/100;if(volume===0)stop()};
 $('dealer-rate').oninput=e=>{rate=normalizeVoiceRate(e.target.value);save();controls();if(audio)audio.playbackRate=rate};
 $('dealer-expression').onclick=()=>{expressive=!expressive;save();controls();speak()};
 $('dealer-settings-open').onclick=()=>$('dealer-settings').showModal();$('dealer-settings-close').onclick=()=>$('dealer-settings').close();controls();
 document.addEventListener('visibilitychange',()=>{if(document.hidden)stop()});window.addEventListener('pagehide',stop);
 return {stop,greet(){say('welcome');return spoken.text},reward(){say('reward','win',80)},setCharacter(id){stop();pick=createLinePicker(Math.random,libraries[id]||voiceLines);spoken=pick('welcome');$('dealer-line').textContent=spoken.text;controls()},reset(state){streaks.delete(state.variant);raiseHand='';say('welcome')},open(state){say(state.done?(state.outcome||'tie'):'welcome',state.done?(state.outcome||'tie'):'welcome')},update(previous,next,{action=''}={}){
  if(next.done&&(!previous?.done||previous.hand!==next.hand)){
   const key=next.variant||'texas',record=streaks.get(key)||{outcome:'',count:0};record.count=record.outcome===next.outcome?record.count+1:1;record.outcome=next.outcome;streaks.set(key,record);
   const kind=action==='fold'?'fold':record.count>=2&&next.outcome==='win'?'streakWin':record.count>=2&&next.outcome==='loss'?'streakLoss':next.outcome||'tie';say(kind,next.outcome||'tie',180);
  }else if(!previous||next.hand!==previous.hand){raiseHand='';say('deal')}
  else if(next.drawing&&!previous.drawing)say('draw');
  else if(next.stageIndex!==previous.stageIndex&&next.stageIndex===next.stages.length-1)say('reveal');
  else if(action==='raise'&&raiseHand!==`${next.variant}:${next.hand}`&&!audio&&Date.now()-lastAmbient>4000){raiseHand=`${next.variant}:${next.hand}`;say('raise')}
 }};
}
