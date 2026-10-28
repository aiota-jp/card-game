export function startTitleScreen({duration=0}={}){
 const screen=document.getElementById('title-screen');if(!screen)return;
 const siblings=[...document.body.children].filter(el=>el!==screen&&el.tagName!=='SCRIPT').map(el=>[el,el.inert]);
 for(const [el] of siblings)el.inert=true;
 let timer,finished=false;
 function finish(){if(finished)return;finished=true;clearTimeout(timer);screen.hidden=true;document.body.classList.remove('is-intro');for(const [el,inert] of siblings)el.inert=inert;document.removeEventListener('keydown',onKey);document.getElementById('variant-title').focus({preventScroll:true})}
 function onKey(event){if(event.key==='Escape'){event.preventDefault();finish()}}
 document.getElementById('title-continue').onclick=finish;document.addEventListener('keydown',onKey);
 document.getElementById('title-continue').focus({preventScroll:true});timer=setTimeout(finish,duration);
 return finish;
}
