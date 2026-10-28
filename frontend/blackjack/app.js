import {getOpponent,refreshOpponentUI,setupOpponentSwitch} from './opponents.js';
import {createDealer} from './dealer.js';
import {createMusic} from './music.js';
import {createSound} from './audio.js';
import {startTitleScreen} from './intro.js';
const $=id=>document.getElementById(id),sound=createSound(),music=createMusic();
const dealer=createDealer({character:getOpponent().id,onSpeaking:value=>music.setDucked(value)});
let state,working=false,bet=50,rewardTimer;
const pct=n=>n.toFixed(1)+'%';
function cardElement(card){const el=document.createElement('div');el.className='card '+(!card?'empty':card.hidden?'back':card.red?'red':'black');if(!card){el.textContent='＋';el.setAttribute('aria-label','まだカードはありません')}else if(card.hidden){el.textContent='♠';el.setAttribute('aria-label','非公開の裏カード')}else{const corner=document.createElement('span');corner.className='corner';const rank=document.createElement('b');rank.textContent=card.rank;const suit=document.createElement('span');suit.textContent=card.suit;corner.append(rank,suit);const center=document.createElement('span');center.className='center-suit';center.textContent=card.suit;el.append(corner,center);el.setAttribute('aria-label',card.rank+' '+card.suit)}return el}
function cards(id,hand){$(id).replaceChildren(...(hand.length?hand:[null,null]).map(cardElement))}
function render(s){
 state=s;document.body.dataset.phase=s.phase;cards('player-cards',s.player);cards('dealer-cards',s.dealer);$('dealer-name').textContent=getOpponent().name;
 $('player-total').textContent=s.player.length?(s.total>21?'バースト '+s.total:(s.soft?'ソフト ':'合計 ')+s.total):'ベットを選択';
 $('dealer-total').textContent=!s.dealer.length?'カード待ち':s.done?(s.dealerTotal>21?'バースト '+s.dealerTotal:(s.dealerSoft?'ソフト ':'合計 ')+s.dealerTotal):'裏カードは非公開';
 $('round-number').textContent='HAND '+String(s.hand).padStart(2,'0');$('shoe-note').textContent=`残り ${s.remaining}枚 / 6デッキ`;
 $('table-bet').textContent=s.bet?'BET '+s.bet:'BLACKJACK LOUNGE';$('chips').textContent=s.chips.toLocaleString();$('goal-progress').value=Math.min(s.chips,s.goal);$('goal-percent').textContent=Math.min(100,Math.floor(s.chips/s.goal*100))+'%';
 for(const key of ['wins','losses','ties'])$(key).textContent=s[key];
 $('turn-label').textContent=s.phase==='betting'?'ベットを選ぼう':s.done?s.cleared?'SESSION CLEAR':({win:'YOU WIN',loss:'DEALER WIN',tie:'PUSH'})[s.outcome]:'あなたの番';
 $('result-text').textContent=s.phase==='betting'?'カードを配って勝負を始めます':s.done?s.result:'ヒットで引くか スタンドで止める';
 document.querySelector('.bj-action-panel').dataset.outcome=s.done?s.outcome:'';
 $('profit-note').textContent=s.done?`このハンドの収支 ${s.profit>=0?'+':''}${s.profit}チップ`:'最初の2枚ならダブルダウンできます';
 const predicting=s.phase==='playing';$('bust').textContent=predicting?pct(s.odds.bust):'—';$('stand-win').textContent=predicting?pct(s.odds.win):'—';$('sample-note').textContent=predicting?`${s.odds.samples.toLocaleString()}回試行 / 95%推定誤差 最大±${s.odds.margin.toFixed(1)}ポイント`:'配札後のあなたの番に表示';
 for(const key of ['win','tie','loss']){$(key).textContent=predicting?pct(s.odds[key]):'—';$(key+'-bar').style.width=predicting?s.odds[key]+'%':'0%'}
 $('log').replaceChildren(...s.log.map(text=>{const li=document.createElement('li');li.textContent=text;return li}));
 $('bet-controls').hidden=s.phase!=='betting';$('bet-options').replaceChildren(...[10,20,50,100,200].map(value=>{const b=document.createElement('button');b.textContent=value;b.setAttribute('aria-pressed',String(bet===value));b.disabled=value>s.chips||working;b.onclick=()=>{bet=value;render(state)};return b}));
 if(bet>s.chips){bet=[200,100,50,20,10].find(n=>n<=s.chips)||10;for(const b of $('bet-options').children)b.setAttribute('aria-pressed',String(Number(b.textContent)===bet))}
 let controls=s.phase==='betting'?[{id:'deal',text:bet+'チップで配る',primary:true}]:s.done?[{id:s.canNext?'next':'reset',text:s.canNext?'次のハンドへ →':'新しいセッション →',primary:true}]:s.actions.map(id=>({id,text:{hit:'ヒット +1枚',stand:'スタンド',double:'ダブル ×2'}[id],primary:id==='stand'}));
 $('actions').replaceChildren(...controls.map(({id,text,primary})=>{const b=document.createElement('button');b.dataset.action=id;b.textContent=text;b.className=primary?'primary':'secondary';b.disabled=working||(id==='deal'&&s.chips<bet);b.onclick=()=>perform(id==='deal'?'Deal':id==='next'?'NextHand':id==='reset'?'Reset':'Act',...(id==='deal'?[bet]:['next','reset'].includes(id)?[]:[id]));return b}));
 $('home-play-title').textContent=s.hand?'ゲームに戻る':'ゲーム開始';refreshOpponentUI();
}
async function perform(method,...args){if(working)return false;working=true;clearTimeout(rewardTimer);$('error').hidden=true;$('busy').hidden=false;document.querySelectorAll('#actions button,#bet-options button,#reset,#home-button,.character-switch').forEach(b=>b.disabled=true);const before=state;
 try{const api=window.go?.main?.BlackjackApp;if(!api)throw Error('デスクトップアプリから起動してください');const next=await api[method](...args);working=false;render(next);
  if(method==='Deal'){sound.play('deal');dealer.update(before,next)}else if(method==='Act'){sound.play(args[0]==='stand'?'check':'deal');dealer.update(before,next)}else if(method==='Reset')dealer.reset(next);
  if(next.done&&(!before?.done||before.hand!==next.hand))sound.play(next.outcome,.18);
  if(next.cleared&&!before?.cleared){const person=getOpponent();rewardTimer=setTimeout(()=>showReward(next,person),900)}return true;
 }catch(e){$('error').textContent=String(e);$('error').hidden=false;sound.play('error');return false}finally{working=false;$('busy').hidden=true;if(state)render(state);$('reset').disabled=false;$('home-button').disabled=false;document.querySelectorAll('.character-switch').forEach(b=>b.disabled=false)}}
function clearReward(){clearTimeout(rewardTimer);$('reward-screen').hidden=true;document.body.classList.remove('is-reward')}
function home(){if(working)return;clearReward();dealer.stop();$('game-screen').hidden=true;$('variant-screen').hidden=false;document.body.classList.add('is-home');refreshOpponentUI();window.scrollTo(0,0);$('variant-title').focus()}
function table(){clearReward();$('variant-screen').hidden=true;$('game-screen').hidden=false;document.body.classList.remove('is-home');window.scrollTo(0,0);$('game-screen').focus()}
async function enter(){if(working)return;$('home-start').disabled=true;try{if(await perform('Snapshot')){table();dealer.open(state)}}finally{$('home-start').disabled=false}}
function showReward(result,person){if(state!==result||$('game-screen').hidden)return;dealer.stop();$('game-screen').hidden=true;$('variant-screen').hidden=true;$('reward-screen').hidden=false;document.body.classList.add('is-home','is-reward');$('reward-image').src=person.reward;$('reward-image').alt=person.name+'の夏服姿';$('reward-opponent').textContent=person.name+'からのお祝い';$('reward-result').textContent=result.chips.toLocaleString()+'チップでクリア';dealer.reward();$('reward-line').textContent=$('dealer-line').textContent;window.scrollTo(0,0);$('reward-title').focus()}
function musicControls(){$('music-toggle').textContent=music.enabled?'BGM ON':'BGM OFF';$('music-toggle').setAttribute('aria-pressed',String(music.enabled));$('music-volume').value=music.volume;$('music-value').textContent=music.volume+'%'}
function soundControls(){$('sound-toggle').textContent=sound.enabled?'♫ 音 ON':'♪ 音 OFF';$('sound-toggle').setAttribute('aria-pressed',String(sound.enabled));$('sound-volume').value=sound.volume;$('volume-value').textContent=sound.volume+'%'}
$('music-toggle').onclick=()=>{music.setEnabled(!music.enabled);musicControls()};$('music-volume').oninput=e=>{music.setVolume(e.target.value);musicControls()};$('sound-toggle').onclick=()=>{sound.setEnabled(!sound.enabled);soundControls()};$('sound-volume').oninput=e=>{sound.setVolume(e.target.value);soundControls()};
document.addEventListener('pointerdown',()=>{sound.unlock();music.unlock()},{capture:true});document.addEventListener('keydown',e=>{if(e.key==='Enter'||e.key===' '){sound.unlock();music.unlock()}},{capture:true});document.addEventListener('click',e=>{const b=e.target.closest('button');if(b&&!b.disabled)sound.play('click')},{capture:true});
$('home-start').onclick=enter;$('home-button').onclick=home;$('help').onclick=()=>$('help-dialog').showModal();$('close-help').onclick=()=>$('help-dialog').close();$('reset').onclick=()=>{if(!working&&window.confirm('持ち点を1,000に戻して新しいセッションを始めますか？'))perform('Reset')};
$('reward-repeat').onclick=()=>$('dealer-repeat').click();$('reward-settings').onclick=()=>$('dealer-settings').showModal();$('reward-home').onclick=home;$('reward-table').onclick=()=>{dealer.stop();table()};$('reward-new').onclick=async()=>{if(await perform('Reset')){dealer.stop();table();dealer.open(state)}};
setupOpponentSwitch(person=>{dealer.setCharacter(person.id);if(state)render(state);refreshOpponentUI();const text=dealer.greet();for(const p of document.querySelectorAll('.lobby-greeting'))p.textContent=text});

const lesson=[
 {title:'01 カードの合計を知ろう',text:'21以下でディーラーより高い合計を目指します。Aは1または11、絵札はすべて10です。',own:['A','6'],dealer:['9','?'],action:'ヒットを練習 →',note:'A＋6はソフト17。Aを11として数えています。'},
 {title:'02 ヒットで1枚引く',text:'今回は10＋6の16。ヒットを押すとカードを1枚追加します。この練習では3が出ます。',own:['10','6'],dealer:['9','?'],action:'ヒット +1枚',note:'21を超えるとその時点で負けです。'},
 {title:'03 スタンドで止める',text:'合計19になりました。スタンドで止めるとディーラーが裏カードを開きます。この例では9＋7＋2で18となりあなたの勝ちです。',own:['10','6','3'],dealer:['9','7','2'],action:'ダブルダウンを知ろう →',note:'ディーラーは16以下で引き17以上なら止まります。'},
 {title:'04 ダブルダウン',text:'最初の2枚に限りベットを2倍にして1枚だけ引けます。5＋6の11から10を引くと21。追加した後は自動的に勝負へ進みます。',own:['5','6','10'],dealer:['9','7','2'],action:'ブラックジャックとは？ →',note:'追加ベット分の持ち点が必要です。'},
 {title:'05 最初の2枚で21',text:'Aと10点札の2枚がブラックジャックです。通常の勝利より多い3:2配当。50チップを賭けて勝つと利益は75チップです。',own:['A','K'],dealer:['9','8'],action:'確率表示を見よう →',note:'双方ブラックジャックなら引き分けです。'},
 {title:'06 確率を判断のヒントに',text:'バースト確率は次の1枚で21を超える割合。スタンド勝率は今止めた場合の推定です。実際の裏カードは使いません。',own:['10','9'],dealer:['9','?'],action:'練習完了 →',note:'実際の対戦で1,000チップからスタート。2,000チップで夏服のお祝いです。'}
];let step=0;
function lessonCard(rank){if(rank==='?')return {hidden:true};return {rank,suit:'♠',red:false}}
function drawLesson(){const s=lesson[step];$('tutorial-title').textContent=s.title;$('tutorial-text').textContent=s.text;$('tutorial-note').textContent=s.note;cards('tutorial-player',s.own.map(lessonCard));cards('tutorial-dealer',s.dealer.map(lessonCard));$('tutorial-counter').textContent=`${step+1} / ${lesson.length}`;$('tutorial-progress').value=step+1;$('tutorial-back').disabled=step===0;$('tutorial-next').textContent=s.action}
function openLesson(){dealer.stop();step=0;drawLesson();$('tutorial-dialog').showModal()}
$('home-tutorial').onclick=$('tutorial-launch').onclick=openLesson;$('close-tutorial').onclick=()=>$('tutorial-dialog').close();$('tutorial-back').onclick=()=>{if(step>0){step--;drawLesson()}};$('tutorial-next').onclick=()=>{sound.play(step===1?'deal':'click');if(step<lesson.length-1){step++;drawLesson()}else $('tutorial-dialog').close()};
musicControls();soundControls();startTitleScreen();

window.loungeCanLeave=()=>!working;
window.loungeDispose=()=>{clearTimeout(rewardTimer);dealer.stop();window.dispatchEvent(new Event('pagehide'));};
