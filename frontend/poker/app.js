import {startTitleScreen} from './intro.js';
import {isMatchVictory} from './rewards.js';
import {getOpponent,refreshOpponentUI,setupOpponentSwitch} from './opponents.js';
import {createDealer} from './dealer.js';
import {createMusic} from './music.js';
import './responsive.js';
import {createSound, transitionEffects} from './audio.js';
import {createTutorial} from './tutorial.js';
import {createVariantSelector, variants, variantMeta} from './variants.js';

const $ = (id) => document.getElementById(id);
const sound = createSound();
const music = createMusic();
const dealer = createDealer({character:getOpponent().id,onSpeaking:value=>music.setDucked(value)});
function musicControls(){$('music-toggle').textContent=music.enabled?'BGM ON':'BGM OFF';$('music-toggle').setAttribute('aria-pressed',String(music.enabled));$('music-volume').value=music.volume;$('music-value').textContent=music.volume+'%';}
$('music-toggle').onclick=()=>{music.setEnabled(!music.enabled);musicControls()};
$('music-volume').oninput=e=>{music.setVolume(e.target.value);musicControls()};
musicControls();
let state;
let working = false;
let rewardTimer;
function clearReward(){clearTimeout(rewardTimer);$('reward-screen').hidden=true;document.body.classList.remove('is-reward')}
function showReward(result,person){
 if(state!==result||$('game-screen').hidden)return;
 dealer.stop();$('variant-screen').hidden=true;$('start-screen').hidden=true;$('game-screen').hidden=true;$('reward-screen').hidden=false;
 document.body.classList.add('is-home','is-reward');document.body.classList.remove('is-catalog');
 const image=$('reward-image');image.src=person.reward||person.full;image.alt=person.name+'の夏服・完全勝利お祝い画像';
 $('reward-opponent').textContent=person.name+'からのお祝い';$('reward-result').textContent=result.result+' · 相手の持ち点 0';
 $('reward-next').textContent=result.canNext?'次のハンドへ →':'新しいセッション →';
 dealer.reward();$('reward-line').textContent=$('dealer-line').textContent;
 window.scrollTo(0,0);$('reward-title').focus();
}
function returnToTable(){clearReward();dealer.stop();$('game-screen').hidden=false;document.body.classList.remove('is-home','is-catalog');window.scrollTo(0,0);$('game-screen').focus()}

let selected=variantMeta('texas');
let stageNames=selected.stages;
const sessions=new Map();
const drawSelection=new Set();
const percent = (value) => `${value.toFixed(1)}%`;
let tutorialCompleted = false;
const tutorial = createTutorial({cardElement, sound, onClose({completed, origin}) {
 tutorialCompleted ||= completed;
 updateHome();
 if (origin === 'home') $(completed ? 'home-start' : 'home-tutorial').focus();
}});

function updateHome() {
 try { tutorialCompleted ||= localStorage.getItem(`poker.tutorial.completed.v1.${selected.id}`) === '1'; } catch { /* Session state still works. */ }
 $('home-tutorial-badge').textContent = tutorialCompleted ? '練習済み ✓ · いつでも復習できます' : 'はじめての方へ · 約2分';
 $('home-play-title').textContent = state ? 'ゲームに戻る' : 'ゲーム開始';
 $('home-play-kicker').textContent = state ? 'つづきからプレイ' : '準備ができたら';
 $('home-play-description').textContent = state ? '現在のセッションをそのまま再開。' : 'MISAKIとの1対1のテーブルへ。';
 $('home-play-note').textContent = state ? '手札とチップはそのまま保持しています' : '各1,000チップ · 勝率ガイド付き';
 $('home-session-note').textContent = state ? '開始画面に戻っても対戦はリセットされません' : tutorialCompleted ? '練習完了！ ゲーム開始から対戦を楽しめます' : '初めての方はチュートリアルがおすすめです';
 refreshOpponentUI();
}
function showHome() {
 if(working)return;
 clearReward();
 dealer.stop();
 $('variant-screen').hidden = true; document.body.classList.remove('is-catalog');
 $('game-screen').hidden = true; $('start-screen').hidden = false;
 document.body.classList.add('is-home'); updateHome(); window.scrollTo(0, 0); $('home-title').focus();
}
function showVariants() {
 if(working)return;
 clearReward();
 dealer.stop();
 $('game-screen').hidden = true; $('start-screen').hidden = true; $('variant-screen').hidden = false;
 document.body.classList.add('is-home', 'is-catalog');
 $('catalog-resume-note').hidden = sessions.size===0;refreshOpponentUI();
 for(const v of variants)document.querySelector(`#variant-${v.id} .variant-action`).textContent=sessions.has(v.id)?'つづきのセッションへ →':'選択して進む →';
 window.scrollTo(0, 0); $(`variant-${selected.id}`).focus();
}
function chooseVariant(id){
 selected=variantMeta(id);stageNames=selected.stages;state=sessions.get(id);tutorialCompleted=false;drawSelection.clear();
 document.body.dataset.variant=id;
 document.querySelector('.variant-breadcrumb b').textContent=selected.name;
 $('home-title').textContent=selected.name+'を楽しもう。';
 document.querySelector('.home-description').textContent=selected.description;
 document.querySelector('.home-tags').replaceChildren(...[selected.count,selected.kind,'MISAKIと1対1'].map(text=>{const el=document.createElement('span');el.textContent=text;return el}));
 const basics=document.querySelector('.home-basics ol');basics.replaceChildren(...selected.guides.map((text,i)=>{const li=document.createElement('li');const n=document.createElement('span');n.textContent=`0${i+1}`;const div=document.createElement('div');const b=document.createElement('b');b.textContent=text;div.append(b);li.append(n,div);return li}));
 document.querySelector('#help-dialog h2').textContent=selected.name+'のルール';
 const helpP=document.querySelectorAll('#help-dialog>p');helpP[0].textContent=selected.summary;helpP[1].textContent='練習用の固定リミット：前半10・後半20ずつ増額し各ラウンド合計40まで。短いスタックでは双方が払える額を上限とします。';
 const list=document.querySelector('#help-dialog>ol');list.replaceChildren(...selected.guides.map(text=>{const li=document.createElement('li');li.textContent=text;return li}));
 document.querySelector('.stage-track').replaceChildren(...stageNames.flatMap((text,i)=>{const span=document.createElement('span');span.textContent=`0${i+1} ${text}`;if(i===stageNames.length-1)return[span];return[span,document.createElement('i')]}));
 document.querySelector('.felt-mark').firstChild.textContent=selected.english+' ';
 document.querySelector('.felt-mark small').textContent='HEADS-UP · PRACTICE TABLE';
 showHome();
}
async function enterGame() {
 clearReward();
 if (working) return;
 $('home-start').disabled = true;
 $('home-tutorial').disabled = true;
 $('choose-variant').disabled = true;
 try {
  const first = !state;
  if (!await perform('OpenGame',selected.id)) return;
  $('variant-screen').hidden = true; $('start-screen').hidden = true; $('game-screen').hidden = false;
  document.body.classList.remove('is-home', 'is-catalog'); window.scrollTo(0, 0); $('game-screen').focus();
  dealer.open(state);
  if (first) sound.play('deal', 0.05);
 } finally { $('home-start').disabled = false; $('home-tutorial').disabled = false; $('choose-variant').disabled = false; }
}

function soundControls() {
 for (const id of ['sound-toggle', 'tutorial-sound']) {
  $(id).textContent = sound.enabled ? '♫ 音 ON' : '♪ 音 OFF';
  $(id).setAttribute('aria-pressed', String(sound.enabled));
 }
 $('sound-volume').value = sound.volume;
 $('sound-volume').setAttribute('aria-valuetext', `${sound.volume}%`);
 $('volume-value').textContent = `${sound.volume}%`;
}
function toggleSound() { sound.setEnabled(!sound.enabled); soundControls(); if (sound.enabled) sound.play('click'); }
document.addEventListener('pointerdown', () => {sound.unlock();music.unlock()}, {capture: true});
document.addEventListener('keydown', event => { if (event.key === 'Enter' || event.key === ' ') {sound.unlock();music.unlock();} }, {capture: true});
document.addEventListener('click', event => {
 const button = event.target.closest('button');
 if (!button || button.disabled || button.dataset.action || ['sound-toggle', 'tutorial-sound', 'tutorial-next'].includes(button.id)) return;
 sound.play('click');
}, {capture: true});

function cardElement(card, hidden = false) {
 hidden ||= Boolean(card?.hidden);
 const el = document.createElement('div');
 el.className = `card ${hidden ? 'back' : card ? (card.red ? 'red' : 'black') : 'empty'}`;
 if (hidden) { el.textContent = '♠'; el.setAttribute('aria-label', '裏向きのカード'); }
 else if (card) {
  const corner = document.createElement('span'); corner.className = 'corner';
  const rank = document.createElement('b'); rank.textContent = card.rank;
  const suit = document.createElement('span'); suit.textContent = card.suit;
  corner.append(rank, suit);
  const center = document.createElement('span'); center.className = 'center-suit'; center.textContent = card.suit;
  el.append(corner, center); el.setAttribute('aria-label', `${card.rank} ${card.suit}`);
 } else { el.textContent = '＋'; el.setAttribute('aria-label', '未公開'); }
 if(card?.up&&!hidden){el.classList.add('up-card');el.title='表向き：相手にも見えています'}
 return el;
}

function render(s) {
 if(!state||state.hand!==s.hand||state.drawing!==s.drawing)drawSelection.clear();
 state = s;
 sessions.set(s.variant,s);stageNames=s.stages;
 document.body.dataset.variant=s.variant;
 document.body.classList.toggle('private-table',!s.boardGame);
 document.body.classList.toggle('many-cards',s.player.length>2);
 $('hand-number').textContent = `HAND ${String(s.hand).padStart(2, '0')}`;
 $('player-stack').textContent = s.stacks[0].toLocaleString();
 $('cpu-stack').textContent = s.stacks[1].toLocaleString();
 $('pot').textContent = s.pot.toLocaleString();
 $('player-dealer').hidden = s.dealer !== 0; $('cpu-dealer').hidden = s.dealer !== 1;
 $('player-cards').replaceChildren(...s.player.map((c,i) => {const el=cardElement(c);if(s.drawing){el.classList.add('selectable');el.tabIndex=0;el.setAttribute('role','button');el.setAttribute('aria-pressed','false');const toggle=()=>{if(working)return;if(drawSelection.has(i))drawSelection.delete(i);else drawSelection.add(i);el.classList.toggle('selected',drawSelection.has(i));el.setAttribute('aria-pressed',String(drawSelection.has(i)));sound.play('click');$('draw-submit').textContent=drawSelection.size?`${drawSelection.size}枚を交換する`:'交換せず進む'};el.onclick=toggle;el.onkeydown=e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();toggle()}}}return el}));
 $('cpu-cards').replaceChildren(...s.cpuHand.map(c=>cardElement(c)));
 if(s.boardGame){$('board').replaceChildren(...Array.from({length:5},(_,i)=>cardElement(s.board[i])))}else{const text=document.createElement('div');text.className='private-note';text.textContent=s.drawing?'交換したい手札を選んでください':selected.stud?'緑の枠は相手にも見える公開カード':'共有カードはありません';$('board').replaceChildren(text)}
 document.querySelector('.board-label').textContent=s.boardGame?'共有カード':s.stage;
 document.querySelector('.assumption').textContent=s.equityNote;
 $('hand-name').textContent = s.handName; $('cpu-name').textContent = s.cpuName;
 $('player-bet').textContent = s.bets[0] ? `◉ ${s.bets[0]}` : '';
 $('cpu-bet').textContent = s.bets[1] ? `◉ ${s.bets[1]}` : '';
 document.querySelectorAll('.stage-track > span').forEach((el, i) => {
  const current = s.stageIndex;
  el.classList.toggle('active', i === current); el.classList.toggle('passed', i < current);
 });
 const e = s.equity;
 $('compact-equity').textContent = `エクイティ ${percent(e.share)}`;
 $('share').textContent = e.share.toFixed(1);
 $('ring').style.setProperty('--equity', `${e.share}%`);
 for (const key of ['win', 'tie', 'loss']) {
  $(key).textContent = percent(e[key]); $(`${key}-bar`).style.width = `${e[key]}%`;
 }
 $('estimate-note').textContent = `${e.samples.toLocaleString()} 回試行 · 95%推定誤差 最大 ±${e.margin.toFixed(1)}ポイント`;
 $('history').replaceChildren(...stageNames.map((stage, i) => {
  const point = s.history[i]; const item = document.createElement('div'); item.className = `history-item ${point ? '' : 'pending'}`;
  const value = document.createElement('b'); value.textContent = point ? `${point.value.toFixed(1)}%` : '—';
  const track = document.createElement('div'); track.className = 'history-track';
  const bar = document.createElement('i'); bar.style.height = `${point ? Math.max(point.value, 2) : 0}%`; track.append(bar);
  const label = document.createElement('small'); label.textContent = stage; item.append(value, track, label);return item;
 }));
 $('log').replaceChildren(...s.log.map(line => {const li = document.createElement('li'); li.textContent = line; return li;}));
 $('log').scrollTop = $('log').scrollHeight;
 $('turn-label').textContent = s.done ? 'RESULT' : s.drawing?'カード交換':'あなたの番です';
 document.querySelector('.action-panel').dataset.outcome = s.done ? s.outcome : '';
 if (s.done) $('turn-label').textContent = ({win: 'YOU WIN', loss: 'MISAKI WIN', tie: 'DRAW'})[s.outcome] || 'RESULT';
 $('action-hint').textContent = s.done ? s.result : s.drawing?'手札をクリックして0〜5枚を選択':s.call ? `続けるには ${s.call} チップをコール` : 'チェックで進むかベットで仕掛けよう';
 refreshOpponentUI();
 if(s.drawing){const submit=document.createElement('button');submit.id='draw-submit';submit.className='primary';submit.textContent='交換せず進む';submit.onclick=()=>perform('Draw',[...drawSelection].sort());$('actions').replaceChildren(submit);return}
 const buttons = s.done ? [{key: s.canNext ? 'next' : 'reset', label: s.canNext ? '次のハンドへ →' : '新しいセッション →', primary: true}] : s.actions.map(key => ({key, label: ({fold:'フォールド',check:'チェック',call:`コール ${s.call}`,bet:`ベット ${s.betTo}`,raise:`レイズ 合計${s.betTo}`})[key], primary:key==='check'||key==='call'}));
 $('actions').replaceChildren(...buttons.map(({key,label,primary}) => {
  const b = document.createElement('button');b.textContent=label;b.dataset.action=key;b.className=primary?'primary':key==='fold'?'fold':'secondary';b.disabled=working;
  b.addEventListener('click',()=>perform(key==='next'?'NextHand':key==='reset'?'Reset':'Act',...(key==='next'||key==='reset'?[]:[key])));return b;
 }));
}

async function perform(method, ...args) {
 if (working) return;
 const previous = state;
 clearTimeout(rewardTimer);
 if (method === 'Act') sound.play(({call:'chips',bet:'chips',raise:'chips',check:'check',fold:'fold'})[args[0]] || 'click');
 else if(method==='Draw')sound.play('deal');
 else if (method !== 'Snapshot'&&method!=='OpenGame') sound.play('click');
 working = true; $('error').hidden = true; $('busy').hidden = false;
 document.querySelectorAll('#actions button, #reset, #home-button, .character-switch').forEach(b => b.disabled = true);
 try {
  const api = window.go?.main?.PokerApp;
  if (!api) throw new Error('ゲームに接続できませんでした。ゲーム一覧へ戻って開き直してください。');
  const next = await api[method](...args);
  working = false; render(next);
  if(method==='Reset')dealer.reset(next);else if(method!=='OpenGame'&&method!=='Snapshot')dealer.update(previous,next,{action:method==='Act'?args[0]:''});
  if (method !== 'Snapshot'&&method!=='OpenGame') {
   const from = method === 'Reset' ? {...previous, hand: -1, done: false} : previous;
   for (const effect of transitionEffects(from, next)) sound.play(effect.name, effect.delay);
  }
  if(method!=='Snapshot'&&method!=='OpenGame'&&isMatchVictory(previous,next)){
   const person=getOpponent();rewardTimer=setTimeout(()=>showReward(next,person),900);
  }
  return true;
 } catch (error) { $('error').textContent = String(error); $('error').hidden = false; sound.play('error'); return false; }
 finally { working = false; $('busy').hidden = true; $('reset').disabled = false; document.querySelectorAll('#actions button, #home-button, .character-switch').forEach(b=>b.disabled=false); }
}
$('reward-repeat').onclick=()=>$('dealer-repeat').click();
$('reward-settings').onclick=()=>$('dealer-settings').showModal();
$('reward-table').onclick=returnToTable;
$('reward-home').onclick=showHome;
$('reward-next').onclick=async()=>{
 if(working)return;const button=$('reward-next');button.disabled=true;dealer.stop();
 try{if(await perform(state.canNext?'NextHand':'Reset')){returnToTable();dealer.open(state)}}finally{button.disabled=false}
};
$('help').onclick=()=>$('help-dialog').showModal();
$('tutorial-launch').onclick=()=>{dealer.stop();tutorial.open({variant:selected})};
$('home-tutorial').onclick=()=>tutorial.open({from:'home',variant:selected});
$('home-start').onclick=enterGame;
$('home-button').onclick=showHome;
$('choose-variant').onclick=showVariants;
$('sound-toggle').onclick=$('tutorial-sound').onclick=toggleSound;
$('sound-volume').oninput=event=>{sound.setVolume(event.target.value);soundControls();};
$('sound-volume').onchange=()=>sound.play('check');
$('close-help').onclick=$('start-playing').onclick=()=>$('help-dialog').close();
$('reset').onclick=()=>{if(window.confirm('進行中のハンドを終了してチップを1,000に戻しますか？'))perform('Reset');};
soundControls();
updateHome();
createVariantSelector({onSelect:chooseVariant});
setupOpponentSwitch(person=>{dealer.setCharacter(person.id);if(state)render(state);updateHome();refreshOpponentUI();const greeting=dealer.greet();for(const p of document.querySelectorAll('.lobby-greeting'))p.textContent=greeting});
startTitleScreen();


window.loungeCanLeave=()=>!working;
window.loungeDispose=()=>{clearTimeout(rewardTimer);dealer.stop();window.dispatchEvent(new Event('pagehide'));};
