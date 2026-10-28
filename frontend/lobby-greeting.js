import {lobbyLines} from './lobby-lines.js';
const read=(key,fallback)=>{try{return localStorage.getItem('cardlounge.lounge.'+key)??fallback}catch{return fallback}};
const save=(key,value)=>{try{localStorage.setItem('cardlounge.lounge.'+key,String(value))}catch{}};
export function chooseGreeting(character,previous,random=Math.random){
 const lines=lobbyLines.filter(line=>line.character===character),choices=lines.filter(line=>line.id!==previous);
 return (choices.length?choices:lines)[Math.min((choices.length||lines.length)-1,Math.floor(Math.max(0,random())*(choices.length||lines.length)))];
}
export function createLobbyGreeting({audioFactory=path=>new Audio(path),onSpeaking=()=>{}}={}){
 const $=id=>document.getElementById(id),last=new Map();let startupPending=false,startupCharacter,selected,current,audio,generation=0,enabled=true,volume=30,rate=1.0,expressive=true;
 function refresh(){enabled=read('voice','true')==='true';volume=Math.max(0,Math.min(100,Number(read('volume','30'))||0));const saved=Number(read('rate','1.0'));rate=Math.round((Number.isFinite(saved)?Math.max(.8,Math.min(2.5,saved)):1.0)*10000)/10000;expressive=read('expression','true')==='true';$('greeting-voice').textContent=enabled?'音声 ON':'音声 OFF';$('greeting-voice').setAttribute('aria-pressed',String(enabled));$('greeting-volume').value=volume;$('greeting-volume-value').textContent=volume+'%';$('greeting-rate').textContent=rate+'倍';}
 function stop(){startupPending=false;generation++;if(audio){audio.pause();audio.removeAttribute('src');audio.load();audio=null}$('greeting-box').classList.remove('speaking');onSpeaking(false);$('greeting-state').textContent='選び直すと別のお誘いも聞けます';}
 function play(){stop();refresh();if(!current||!enabled||!volume||document.hidden)return;const token=generation;
  try{const player=audioFactory(current.files[expressive?'expressive':'calm'].path);audio=player;player.volume=volume/100;player.playbackRate=rate;player.preservesPitch=true;
   player.onplaying=()=>{if(token===generation){$('greeting-box').classList.add('speaking');onSpeaking(true);$('greeting-state').textContent='ご挨拶中…';}};
   player.onended=()=>{if(token===generation){audio=null;$('greeting-box').classList.remove('speaking');onSpeaking(false);$('greeting-state').textContent='お好きなゲームをどうぞ';}};
   const failed=()=>{if(token===generation){audio=null;$('greeting-box').classList.remove('speaking');onSpeaking(false);$('greeting-state').textContent='再生できませんでした。「もう一度聞く」で再生できます';}};
   player.onerror=failed;Promise.resolve(player.play()).catch(failed);
  }catch{$('greeting-state').textContent='音声を再生できませんでした';}
 }
 function select(character,{speak=false}={}){stop();selected=character;refresh();current=speak?chooseGreeting(character,last.get(character)):lobbyLines.find(x=>x.id===last.get(character))||lobbyLines.find(x=>x.character===character);if(!current)return;if(speak)last.set(character,current.id);$('greeting-box').dataset.line=current.id;$('greeting-name').textContent=character.toUpperCase()+'からのご挨拶';$('greeting-text').textContent=current.text;if(speak)play();}
 $('greeting-repeat').onclick=play;$('greeting-stop').onclick=stop;$('greeting-voice').onclick=()=>{enabled=!enabled;save('voice',enabled);refresh();if(enabled)play();else stop()};
 $('greeting-volume').oninput=e=>{volume=Number(e.target.value);save('volume',volume);$('greeting-volume-value').textContent=volume+'%';if(audio)audio.volume=volume/100;if(!volume)stop()};
 function startWhenVisible(){if(startupPending&&!document.hidden){const character=startupCharacter;startupPending=false;select(character,{speak:true})}}
 function start(character){select(character);startupCharacter=character;startupPending=true;startWhenVisible()}
 document.addEventListener('visibilitychange',()=>{if(document.hidden){if(!startupPending)stop()}else startWhenVisible()});window.addEventListener('pagehide',stop);refresh();
 return {select,stop,start};
}
