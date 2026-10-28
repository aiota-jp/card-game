import {localizeOpponent} from './opponents.js';
const texasSteps = [
 {title: 'まずは1ハンド練習しよう。', subtitle: '約2分 / チップは減りません', text: '自分の手札2枚と中央に出る共有カード5枚。その7枚から最も強い5枚を作り相手と比べます。この練習では同じカードで流れを体験できます。', focus: 'あなたの手札', board: 0, action: '練習をはじめる →', note: '練習用の配札です。実際の対戦のカードやチップには影響しません。'},
 {title: '01 手札を見て参加する。', subtitle: 'プリフロップ / 手札2枚', text: '今回の手札はA♠とK♠。MISAKIは10あなたは5チップを出しています。コールは「相手と同じ額まで追加する」操作です。あと5チップを出して参加してみましょう。', focus: 'コールを体験', board: 0, pot: 15, action: 'コール 5', sound: 'chips', note: 'フォールドなら降りるレイズなら賭け額を増やします。'},
 {title: '02 共有カードが3枚公開。', subtitle: 'フロップ / 共有カード3枚', text: 'Q♠・J♠・2♥が出ました。手札と合わせてスペードが4枚。あと1枚スペードが出ればフラッシュが完成します。相手が追加で賭けていないのでチェックで進めます。', focus: 'フラッシュまであと1枚', board: 3, pot: 20, action: 'チェックして次のカードへ', sound: 'check', note: 'チェックは追加チップを出さずに順番を渡す操作です。'},
 {title: '03 次の1枚で役が完成。', subtitle: 'ターン / 共有カード4枚', text: '10♠が公開されました！ A♠・K♠・Q♠・J♠・10♠でロイヤルフラッシュ。最も強い役ができました。強い役ができたときも公開情報を見直して判断しましょう。', focus: 'ロイヤルフラッシュ！', board: 4, pot: 20, action: 'チェックしてリバーへ', sound: 'check', note: 'この練習の配札では相手の手札にかかわらず勝ちが確定しています。'},
 {title: '04 最後のカードと最後の選択。', subtitle: 'リバー / 共有カード5枚', text: '3♦が公開されました。これ以上カードは増えません。ベットは「このラウンドで最初にチップを賭ける」操作です。今回は20チップをベットして相手の反応を見ます。', focus: 'ベットを体験', board: 5, pot: 20, action: 'ベット 20 → MISAKIがコール', sound: 'chips', note: '実際の対戦ではMISAKIがフォールドすることもあります。'},
 {title: '05 手札を開いて勝負！', subtitle: 'ショーダウン / 勝敗と精算', text: 'MISAKIは9♥・9♦のワンペア。あなたのロイヤルフラッシュが勝ち60チップのポットを獲得します。途中でどちらかがフォールドした場合は残った側がポットを獲得します。', focus: 'YOU WIN · ポット60獲得', board: 5, pot: 60, reveal: true, action: '勝率の読み方も確認 →', sound: 'click', note: 'ポットには自分が賭けたチップも含まれます。利益が60という意味ではありません。'},
 {title: '06 右側の数字は判断のヒント。', subtitle: '勝率とエクイティ', text: '勝ち・引き分け・負けの割合は見えているカードから計算します。大きい数字のエクイティは引き分けで受け取る取り分も含めた値です。', focus: '例：勝ち50% ＋ 引き分け10% ÷ 2 ＝ 55%', board: 5, reveal: true, action: '準備OK！ 対戦に戻る →', sound: 'click', note: '相手はランダムな手札を想定。高い勝率でも勝利を保証するものではありません。'},
];

function lesson(variant){
 const steps=texasSteps.map(step=>({...step}));
 let own=['As','Ks'],board=['Qs','Js','10s','2h','3d'],cpu=['9h','9d'],after=null,ownCounts=null,cpuName='MISAKI · ワンペア';
 // The Texas animation reveals Q,J,2 before the turn 10.
 board=['Qs','Js','2h','10s','3d'];
 if(variant.id==='omaha'){
  own=['As','Ks','2h','3d'];board=['Qs','Js','9c','10s','9d'];cpu=['9h','9s','6c','7c'];cpuName='MISAKI · フォーカード';
  steps[0].text='オマハは手札4枚。役を作るときは手札から必ず2枚共有カードから必ず3枚を使います。この練習はオマハ・ハイです。';
  steps[1].text='手札はA♠・K♠・2♥・3♦。4枚すべてを役に使えるわけではありません。まずは5チップをコールして参加します。';
  steps[2].text='Q♠・J♠・9♣が公開。A♠とK♠を選び共有カード3枚と組み合わせます。手札1枚だけや共有カード5枚だけで役は作れません。';
  steps[3].text='10♠が出ました。手札A♠・K♠と共有Q♠・J♠・10♠の「2枚＋3枚」でロイヤルフラッシュです。';
  steps[4].text='最後に9♦。手札と共有カードの使い方は最後まで同じです。20をベットしMISAKIがコールした場合を見てみましょう。';
  steps[5].text='MISAKIの9のフォーカードよりあなたのロイヤルフラッシュが強いため勝ちです。オマハでも必ず手札2枚と共有3枚を使って比較します。';
 }
 if(variant.id==='short'){
  board=['Qs','Js','6h','10s','7d'];
  steps[0].text='ショートデックは2〜5を除いた36枚で遊びます。このアプリは「フラッシュがフルハウスより強い」ルール。A-6-7-8-9もストレートとして扱います。';
  steps[1].text='全員のアンティとボタンの追加ブラインドでポット15。今回はボタン側のMISAKIに5をコールする練習です。手札はA♠・K♠です。';
  steps[2].text='Q♠・J♠・6♥が公開。36枚のデッキなので通常のホールデムとは残りカードの構成と勝率が変わります。';
  steps[4].text='最後は7♦。この配札ではロイヤルフラッシュが完成しています。ベット20にMISAKIがコールする流れを見てみましょう。';
 }
 if(variant.stud){
  own=variant.id==='razz'?['As','2d','3c','4h','5s','Kh','Qd']:['As','Ad','Ac','Kc','Kd','2c','3d'];
  cpu=variant.id==='razz'?['2h','3d','4c','5h','6s','Jd','Qc']:['Qs','Qd','9h','9c','2d','3h','4h'];board=[];ownCounts=[3,3,4,5,6,7,7];
  cpuName=variant.id==='razz'?'MISAKI · 6-5-4-3-2 ロー':'MISAKI · ツーペア';
  steps[0].text=variant.summary;steps[0].focus='表向きのカードは相手にも見えます';
  steps[1].text='最初は裏向き2枚と表向き1枚。MISAKIがブリングイン5を出した練習例です。アンティはすでにポットに入りコール5で続けられます。';
  steps[1].subtitle='3枚目 / アンティとブリングイン';
  steps[2].title='02 4枚目は表向きで追加。';steps[2].subtitle='4枚目 / 公開カードを確認';
  steps[2].text=variant.id==='razz'?'4♥が追加され低いカードがそろってきました。Aは1として扱いペアになると不利です。相手の公開カードも判断と勝率計算に反映されます。':'K♣が表向きで追加。相手もこのカードを見ることができます。公開カードだけで相手の役を決めつけず裏向きのカードもあることを意識しましょう。';
  steps[2].focus='緑の枠は公開カード';
  steps[3].title='03 5枚目で役の候補ができる。';steps[3].subtitle='5枚目 / 役を比較';
  steps[3].text=variant.id==='razz'?'5♠が出て5-4-3-2-Aの最良のローが完成。ラズではストレートやフラッシュは不利になりません。':'K♦が加わりAのスリーカード＋Kのペアでフルハウスが完成しました。';
  steps[3].focus=variant.id==='razz'?'最良のロー：5-4-3-2-A':'フルハウスが完成';steps[3].note='後からカードが増えても自分の7枚のうち最良の5枚を選べます。';
  steps[4].title='04 6枚目は表7枚目は裏。';steps[4].subtitle='6〜7枚目 / 最後のベット';
  steps[4].text='6枚目までが表向き最後の7枚目は自分だけに見えるカードです。各段階のベットを経て最後に20をベットしMISAKIがコールした例を確認します。';
  steps[4].focus='最後の1枚は非公開';steps[4].action='最後まで進める →';
  steps[5].text=variant.id==='razz'?'あなたの5-4-3-2-AローはMISAKIの6ローより低く勝ちます。最も大きいカードが同じなら、次に大きいカードから順に比べます。':'あなたのフルハウスはMISAKIのツーペアより強く勝ちです。共有カードはなくそれぞれ自分の7枚から5枚を選んで比べます。';
  steps.forEach(s=>s.board=0);
 }
 if(variant.id==='draw'){
  own=['As','Ad','4c','7h','9d'];after=['As','Ad','Ac','Kd','Kh'];board=[];cpu=['Qs','Qd','8c','8h','2d'];cpuName='MISAKI · ツーペア';
  steps[0].text='5枚を配ってベットし不要なカードを0〜5枚1回だけ交換します。その後もう一度ベットして勝負。交換しない「スタンドパット」も選べます。';
  steps[1].text='Aが2枚のワンペアです。まずはコール5で参加します。交換するカードはこのあとの交換画面で選びます。';
  steps[2].title='02 不要なカードを選んで交換。';steps[2].subtitle='カード交換 / クリックして選択';steps[2].text='Aのペアを残して4♣・7♥・9♦を選んでください。枠が光ったカードが交換対象です。3枚を選ぶと次へ進めます。';steps[2].focus='4♣・7♥・9♦の3枚を選択';steps[2].action='選んだ3枚を交換';steps[2].sound='deal';steps[2].note='実際の対戦では0枚（交換なし）から5枚まで自由に選べます。';
  steps[3].title='03 新しいカードを確認。';steps[3].subtitle='交換後 / 最後の5枚';steps[3].text='この練習ではA♣・K♦・K♥に入れ替わりAのスリーカードとKのペアでフルハウスになりました。実際の交換カードはランダムです。';steps[3].focus='フルハウスが完成';steps[3].note='交換は1回だけです。ここからは5枚を変えずにベットします。';
  steps[4].text='交換後に最後のベットを行います。20をベットして相手がコールした例を確認しましょう。';steps[4].subtitle='交換後 / 最後のベット';
  steps[5].text='あなたのフルハウスはMISAKIのツーペアより強くポットを獲得します。全員の交換が終わり最後のベットがそろってから手札を公開します。';
  steps[6].note='ドローの表示値は「現在の5枚を維持した場合の勝率」です。交換前の数値は交換後の予測ではありません。';
  steps.forEach(s=>s.board=0);
 }
 steps[0].subtitle=variant.name+' / 約2分';
 if(!variant.stud&&variant.id!=='texas')steps[5].focus='YOU WIN · ポット60獲得';
 return {steps,own,board,cpu,after,ownCounts,cpuName,variant};
}

export function createTutorial({cardElement,sound,onClose}){
 const $=id=>document.getElementById(id),dialog=$('tutorial-dialog');
 let index=0,origin='game',finished=false,active;const selected=new Set();
 const parse=code=>{let rank=code.slice(0,-1);if(rank==='T')rank='10';const suit={s:'♠',h:'♥',d:'♦',c:'♣'}[code.slice(-1)];return {rank,suit,red:suit==='♥'||suit==='♦'}};
 const validSelection=()=>selected.size===3&&[2,3,4].every(i=>selected.has(i));
 function draw(){
  const {steps,variant}=active,step=steps[index],exchange=variant.id==='draw'&&index===2;
  $('tutorial-counter').textContent=`${index+1} / 7`;$('tutorial-progress').value=index+1;
  for(const [id,text] of [['tutorial-title',step.title],['tutorial-subtitle',step.subtitle],['tutorial-text',step.text],['tutorial-focus',step.focus],['tutorial-note',step.note]])$(id).textContent=text;
  $('tutorial-pot').textContent=`練習用ポット ${step.pot??0}`;
  const own=index>=3&&active.after?active.after:active.own.slice(0,active.ownCounts?.[index]??active.own.length);
  $('tutorial-player').replaceChildren(...own.map((code,i)=>{
   const card=parse(code);card.up=variant.stud&&i>=2&&i<=5;const el=cardElement(card);
   if(exchange){el.classList.add('selectable');el.tabIndex=0;el.setAttribute('role','button');el.setAttribute('aria-pressed',String(selected.has(i)));el.classList.toggle('selected',selected.has(i));const toggle=()=>{selected.has(i)?selected.delete(i):selected.add(i);el.classList.toggle('selected',selected.has(i));el.setAttribute('aria-pressed',String(selected.has(i)));$('tutorial-next').disabled=!validSelection();sound.play('click')};el.onclick=toggle;el.onkeydown=e=>{if(e.key==='Enter'||e.key===' '){e.preventDefault();toggle()}}}return el;
  }));
  if(variant.board)$('tutorial-board').replaceChildren(...Array.from({length:5},(_,i)=>cardElement(i<step.board?parse(active.board[i]):null)));
  else{const text=document.createElement('small');text.textContent=variant.stud?'各自のカードで役を作ります':'共有カードはありません';$('tutorial-board').replaceChildren(text)}
  const cpuCount=active.ownCounts?.[index]??active.cpu.length;
  $('tutorial-cpu').replaceChildren(...active.cpu.slice(0,cpuCount).map((code,i)=>{const c=parse(code);c.up=variant.stud&&i>=2&&i<=5;return cardElement(c,!step.reveal&&!c.up)}));
  $('tutorial-cpu-label').textContent=step.reveal?active.cpuName:variant.stud?'MISAKI · 緑の枠は公開カード':'MISAKI · 手札は非公開';
  $('tutorial-player-label').textContent='あなた · '+(variant.stud?'緑の枠は公開カード':'自分だけの手札');
  $('tutorial-next').textContent=index===6&&origin==='home'?'準備OK！ 開始画面へ →':step.action;
  $('tutorial-next').disabled=exchange&&!validSelection();$('tutorial-back').disabled=index===0;
  dialog.dataset.step=String(index);dialog.dataset.variant=variant.id;localizeOpponent(dialog);
 }
 function close(done=false){finished=done;if(done){try{localStorage.setItem(`poker.tutorial.completed.v1.${active.variant.id}`,'1')}catch{}}dialog.close()}
 $('tutorial-next').onclick=()=>{if($('tutorial-next').disabled)return;sound.play(active.steps[index].sound||'click');if(index===6){close(true);return}index++;selected.clear();draw();if([2,3,4].includes(index))sound.play('deal',.12);if(index===5)sound.play('win',.18);$('tutorial-title').focus({preventScroll:true})};
 $('tutorial-back').onclick=()=>{if(index>0){index--;selected.clear();draw()}};
 $('close-tutorial').onclick=()=>close();dialog.addEventListener('close',()=>onClose?.({completed:finished,origin,variant:active.variant.id}));
 return {open({from='game',variant}={}){if(dialog.open)return;active=lesson(variant);origin=from;finished=false;index=0;selected.clear();const help=$('help-dialog');if(help.open)help.close();draw();dialog.showModal();$('tutorial-next').focus()}};
}

