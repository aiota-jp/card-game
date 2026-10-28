// One music player spans the title and lobby; individual games own their music.
const read=(key,fallback)=>{try{return localStorage.getItem('cardlounge.music.'+key)??fallback}catch{return fallback}};
export function createLobbyMusic({audioFactory=src=>new Audio(src)}={}){
 const player=audioFactory('lounge-bgm.wav?v=20261028'),buttons=['title-music-toggle','hub-music-toggle'].map(id=>document.getElementById(id)),slider=document.getElementById('title-music-volume');
 player.loop=true;player.preload='auto';
 let enabled=true,volume=25,active=true,ducked=false,blocked=false,attempt=0,playing=false;
 function level(){player.volume=enabled?volume/100*(ducked?.22:1):0;document.body.dataset.lobbyMusicDucked=String(ducked)}
 function controls(){for(const b of buttons){b.textContent=enabled?(blocked?'♫ BGMを再生':'♫ BGM ON'):'♪ BGM OFF';b.setAttribute('aria-pressed',String(enabled))}slider.value=volume;document.getElementById('title-music-volume-value').textContent=volume+'%'}
 function stop(){attempt++;player.pause();playing=false;document.body.dataset.lobbyMusicPlaying='false'}
 async function play(){
  if(!enabled||!volume||!active||document.hidden){stop();return}
  if(playing&&!player.paused)return;
  const token=++attempt;
  try{await player.play();if(token!==attempt)return;playing=true;blocked=false;document.body.dataset.lobbyMusicPlaying='true'}
  catch{if(token!==attempt)return;playing=false;blocked=true;document.body.dataset.lobbyMusicPlaying='false'}
  controls();
 }
 function refresh(){enabled=read('enabled','true')==='true';volume=Math.max(0,Math.min(100,Number(read('volume','25'))||0));level();controls();play()}
 function save(){try{localStorage.setItem('cardlounge.music.enabled',String(enabled));localStorage.setItem('cardlounge.music.volume',String(volume))}catch{}}
 for(const b of buttons)b.onclick=()=>{if(blocked&&enabled){blocked=false;play()}else{enabled=!enabled;save();level();controls();play()}};
 slider.oninput=()=>{volume=Number(slider.value);save();level();controls();play()};
 const unlock=()=>{if(active&&enabled)play()};
 document.addEventListener('pointerdown',unlock,{capture:true});document.addEventListener('keydown',unlock,{capture:true});
 document.addEventListener('visibilitychange',()=>{if(document.hidden)stop();else refresh()});
 window.addEventListener('pagehide',stop);
 refresh();
 return {refresh,setActive(value){active=Boolean(value);if(active)refresh();else stop()},setDucked(value){ducked=Boolean(value);level()},stop};
}
