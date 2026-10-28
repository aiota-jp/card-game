export const opponents={
 misaki:{reward:'assets/misaki-summer.png',id:'misaki',name:'MISAKI',japanese:'美咲',full:'assets/misaki-full.png',portrait:'assets/dealer.png'},
 ayaka:{reward:'assets/ayaka-summer.png',id:'ayaka',name:'AYAKA',japanese:'あやか',full:'assets/ayaka-full.png',portrait:'assets/ayaka-portrait.png'},
 rina:{reward:'assets/rina-summer.png',id:'rina',name:'RINA',japanese:'リナ',full:'assets/rina-full.png',portrait:'assets/rina-portrait.png'},
 shizue:{reward:'assets/shizue-summer.png',id:'shizue',name:'SHIZUE',japanese:'静江',full:'assets/shizue-full.png',portrait:'assets/shizue-portrait.png'},
 kathy:{reward:'assets/kathy-summer.png',id:'kathy',name:'KATHY',japanese:'キャシー',full:'assets/kathy-full.png',portrait:'assets/kathy-portrait.png'}
};
const order=Object.keys(opponents);const next=()=>order[(order.indexOf(selected)+1)%order.length];
let selected='misaki';try{const saved=localStorage.getItem('cardlounge.lounge.character');if(Object.hasOwn(opponents,saved))selected=saved}catch{}
export const getOpponent=()=>opponents[selected];
export function opponentText(text){return String(text).replace(/MISAKI|AYAKA|RINA|SHIZUE|KATHY/g,getOpponent().name).replace(/美咲|あやか|リナ|静江|キャシー/g,getOpponent().japanese)}
export function localizeOpponent(root=document){
 const walker=document.createTreeWalker(root,NodeFilter.SHOW_TEXT);let node;
 while((node=walker.nextNode())){if(node.parentElement?.closest('script,style'))continue;const value=opponentText(node.nodeValue);if(value!==node.nodeValue)node.nodeValue=value}
 for(const el of root.querySelectorAll('[aria-label],[alt],[title]'))for(const attribute of ['aria-label','alt','title']){const value=el.getAttribute(attribute);if(value)el.setAttribute(attribute,opponentText(value))}
}
export function refreshOpponentUI(){
 localizeOpponent();const person=getOpponent(),other=opponents[next()];document.body.dataset.character=selected;
 for(const img of document.querySelectorAll('.lobby-figure img')){if(img.getAttribute('src')!==person.full)img.src=person.full;img.alt=`対戦相手 ${person.name}の全身画像`}
 for(const img of document.querySelectorAll('.dealer-portrait')){if(img.getAttribute('src')!==person.portrait)img.src=person.portrait;img.alt=`対戦相手 ${person.name}の上半身`}
 for(const heading of document.querySelectorAll('.lobby-host-name h2')){heading.replaceChildren(document.createTextNode(person.name+' '));const small=document.createElement('small');small.textContent=person.japanese;heading.append(small)}
 for(const button of document.querySelectorAll('.character-switch')){button.setAttribute('aria-label',`${person.name}の画像をクリックして${other.name}に切り替える`);button.title=`クリックで${other.name}に切り替え`}
 for(const hint of document.querySelectorAll('.character-switch-hint'))hint.textContent=`画像をクリックで ${other.name}へ ↔`;
}
export function setupOpponentSwitch(onChange){
 for(const button of document.querySelectorAll('.character-switch'))button.onclick=()=>{selected=next();try{localStorage.setItem('cardlounge.lounge.character',selected)}catch{}refreshOpponentUI();onChange(getOpponent())};
 refreshOpponentUI();
}
