import {createLobbyMusic} from './lobby-music.js';
import {createSuiteTools} from './suite-tools.js?v=20261028-audit';
import {startSuiteTitle} from './suite-title.js';
import {createLobbyGreeting} from './lobby-greeting.js';
const $=id=>document.getElementById(id),lobbyMusic=createLobbyMusic(),greeting=createLobbyGreeting({onSpeaking:playing=>lobbyMusic.setDucked(playing)});
const games={poker:{name:'ポーカー',api:'PokerApp'},blackjack:{name:'ブラックジャック',api:'BlackjackApp'},baccarat:{name:'バカラ',api:'BaccaratApp'}};
let frame,current,lastButton,statusTimer;
function notify(text){clearTimeout(statusTimer);$('status').textContent=text;$('status').hidden=false;statusTimer=setTimeout(()=>$('status').hidden=true,6000)}
function host(id,speak=false){if(!['misaki','ayaka','rina','shizue','kathy'].includes(id))id='misaki';$('host-image').src=`poker/assets/${id}-full.png`;$('host-image').alt=id.toUpperCase()+'の全身画像';$('host-name').textContent=id.toUpperCase();document.querySelectorAll('[data-host]').forEach(b=>b.setAttribute('aria-pressed',String(b.dataset.host===id)));try{localStorage.setItem('cardlounge.lounge.character',id)}catch{}greeting.select(id,{speak})}
try{if(!localStorage.getItem('cardlounge.lounge.aivis-default-v1')){const old=localStorage.getItem('cardlounge.lounge.rate');if(old===null||Number(old)===1.794)localStorage.setItem('cardlounge.lounge.rate','1');localStorage.setItem('cardlounge.lounge.aivis-default-v1','1')}}catch{}
host('misaki');startSuiteTitle(()=>greeting.start('misaki'));
for(const b of document.querySelectorAll('[data-host]'))b.onclick=()=>host(b.dataset.host,true);
function leave(){if(!frame)return;try{if(!frame.contentWindow.loungeCanLeave?.()){notify('配札・精算が終わるまで少しお待ちください');return}frame.contentWindow.loungeDispose?.()}catch{}frame.remove();frame=null;current=null;lobbyMusic.setActive(true);$('suite-tools').hidden=false;$('workspace').hidden=true;$('playing').hidden=true;$('hub').hidden=false;$('suite-tag').hidden=false;try{host(localStorage.getItem('cardlounge.lounge.character'))}catch{}window.scrollTo(0,0);(lastButton||$('hub-title')).focus()}
function open(id,button){if(!games[id]||frame)return;const api=games[id];if(!window.go?.main?.[api.api]){notify('card-game-lounge.exeから起動してください');return}greeting.stop();lobbyMusic.setActive(false);$('suite-tools').hidden=true;lastButton=button;current=id;$('current-game').textContent=api.name;$('hub').hidden=true;$('suite-tag').hidden=true;$('playing').hidden=false;$('workspace').hidden=false;frame=document.createElement('iframe');frame.title=api.name+'のゲーム画面';frame.id='game-frame';frame.src=`${id}/game.html?v=20261028-audit`;frame.addEventListener('load',()=>{try{if(typeof frame?.contentWindow.loungeCanLeave!=='function'){notify('ゲーム画面を読み込めませんでした。ゲーム一覧に戻って再度お試しください');if(frame)frame.contentWindow.loungeCanLeave=()=>true}}catch{notify('画面の読み込みに失敗しました')}});$('workspace').append(frame);window.scrollTo(0,0);$('back').focus()}
for(const b of document.querySelectorAll('[data-game]'))b.onclick=()=>open(b.dataset.game,b);
$('back').onclick=leave;$('brand').onclick=e=>{e.preventDefault();if(current)leave();else $('hub-title').focus()};

const creditsDialog=$('credits-dialog');
$('credits-open').addEventListener('click',()=>{greeting.stop();creditsDialog.showModal();});
$('credits-close').addEventListener('click',()=>creditsDialog.close());

createSuiteTools({stop:()=>greeting.stop(),refresh:()=>{host(localStorage.getItem('cardlounge.lounge.character'));lobbyMusic.refresh()},notify});
