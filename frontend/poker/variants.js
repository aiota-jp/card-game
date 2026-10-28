export const variants = [
 {id:'texas', name:'テキサスホールデム', english:'TEXAS HOLD’EM', icon:'♠', count:'2 + 5', kind:'共有カード', description:'手札2枚と共有カード5枚から最も強い5枚を選ぶ。基本から学べる定番のルール。', facts:['手札2枚＋共有カード5枚','7枚から最も強い5枚を選ぶ','チュートリアル・MISAKIとの対戦・勝率表示'], available:true},
 {id:'omaha', name:'オマハ', english:'OMAHA', icon:'♦', count:'4 + 5', kind:'共有カード', description:'代表的な形式では手札4枚。手札から2枚共有カードから3枚を必ず使って役を作る。', facts:['代表的な形式：手札4枚＋共有カード5枚','手札2枚と共有カード3枚を必ず使う','手札の使い方がホールデムと異なる']},
 {id:'draw', name:'ファイブカードドロー', english:'FIVE-CARD DRAW', icon:'♣', count:'5 ↺', kind:'カード交換', description:'最初に5枚を配り不要なカードを交換。手札の中だけで役を作るおなじみのポーカー。', facts:['最初の手札は5枚','不要なカードを交換して役を目指す','相手の手札は基本的に見えない']},
 {id:'stud', name:'セブンカードスタッド', english:'SEVEN-CARD STUD', icon:'♥', count:'7', kind:'公開＋非公開', description:'表向きと裏向きのカードを各自に配る。相手の公開カードを見ながら自分の7枚で勝負する。', facts:['各自に計7枚のカード','表向き・裏向きのカードを配る','自分の7枚から最も強い5枚を選ぶ']},
 {id:'razz', name:'ラズ', english:'RAZZ', icon:'♧', count:'7 ↓', kind:'低い手を競う', description:'スタッド系のポーカー。通常の強い役を目指す代わりに決められたルールで低い手を競う。', facts:['スタッド系のゲーム','低い手を作ることを目指す','通常のハイハンドとは勝敗の基準が異なる']},
 {id:'short', name:'ショートデック', english:'SHORT DECK', icon:'♤', count:'36', kind:'36枚デッキ', description:'2〜5を除いた36枚で遊ぶポーカー。カードの構成が変わり役の順位が異なるルールもある。', facts:['2〜5を除く36枚を使用','通常の52枚デッキとは確率が異なる','役の順位などは採用ルールによって異なる']},
];

export function createVariantSelector({onSelect}) {
 const $ = id => document.getElementById(id);
 $('variant-grid').replaceChildren(...variants.map(variant=>{
  variant.available=true;
  const button=document.createElement('button');button.className=`variant-card${variant.available?' available':''}`;
  button.id=`variant-${variant.id}`;button.dataset.variant=variant.id;
  const header=document.createElement('span');header.className='variant-card-header';
  const english=document.createElement('span');english.textContent=variant.english;
  const status=document.createElement('span');status.className='variant-status';status.textContent='プレイ可能';header.append(english,status);
  const art=document.createElement('span');art.className='variant-card-art';
  const icon=document.createElement('span');icon.textContent=variant.icon;
  const count=document.createElement('b');count.textContent=variant.count;
  const kind=document.createElement('small');kind.textContent=variant.kind;art.append(icon,count,kind);
  const name=document.createElement('strong');name.className='variant-name';name.textContent=variant.name;
  const description=document.createElement('span');description.className='variant-description';description.textContent=variant.description;
  const action=document.createElement('span');action.className='variant-action';action.textContent=variant.available?'選択して進む →':'特徴を見る ↗';
  button.append(header,art,name,description,action);
  button.onclick=()=>onSelect(variant.id);
  return button;
 }));
}

export function variantMeta(id){
 const base=variants.find(v=>v.id===id)||variants[0];
 const extra={
  texas:{board:true,stages:['プリフロップ','フロップ','ターン','リバー'],summary:'手札2枚＋共有カード5枚から最強の5枚。ブラインド5/10。',guides:['手札2枚を確認','共有カードが3→1→1枚','7枚から最強の5枚で勝負']},
  omaha:{board:true,stages:['プリフロップ','フロップ','ターン','リバー'],summary:'オマハ・ハイ。手札4枚から必ず2枚共有カードから必ず3枚を使います。ブラインド5/10。',guides:['手札4枚を確認','共有カードが3→1→1枚','手札2枚＋共有3枚を必ず使う']},
  draw:{board:false,stages:['交換前','カード交換','交換後'],summary:'手札5枚。交換前のベット後0〜5枚を1回だけ交換。最後のベット後に5枚で勝負します。ブラインド5/10。',guides:['手札5枚でベット','不要なカードを0〜5枚選んで交換','最後のベットと手札の公開']},
  stud:{board:false,stud:true,stages:['3枚目','4枚目','5枚目','6枚目','7枚目'],summary:'セブンカードスタッド。最初に裏2枚＋表1枚4〜6枚目は表7枚目は裏。アンティ5最初は低い公開カードがブリングイン5。その後は公開手の強い側から行動。',guides:['裏2枚＋表1枚で開始','表3枚→最後に裏1枚を追加','7枚から最強の5枚を選ぶ']},
  razz:{board:false,stud:true,stages:['3枚目','4枚目','5枚目','6枚目','7枚目'],summary:'A-to-5のラズ。Aは1ストレートとフラッシュは無視。重複のない低い5枚を目指します。アンティ5最初は高い公開カードがブリングイン5。その後は公開手の低い側から行動。',guides:['裏2枚＋表1枚で開始','ペアを避けて低いカードを集める','最良は5-4-3-2-Aのロー']},
  short:{board:true,stages:['プリフロップ','フロップ','ターン','リバー'],summary:'6〜Aだけの36枚。フラッシュはフルハウスより強いルールです。A-6-7-8-9もストレート。アンティ5＋ボタンの追加ブラインド5。',guides:['6〜Aの36枚デッキ','手札2枚＋共有カード5枚','フラッシュがフルハウスより強い']},
 }[base.id];
 return {...base,...extra};
}

