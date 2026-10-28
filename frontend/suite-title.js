export function startSuiteTitle(onEnter,{duration=5000}={}){
 const screen=document.getElementById('suite-title'),button=document.getElementById('suite-title-enter'),image=document.getElementById('suite-title-image');
 let done=false,timer,started=false,ready=false;
 function finish(){if(done)return;done=true;clearTimeout(timer);document.removeEventListener('visibilitychange',visible);document.removeEventListener('keydown',key);screen.hidden=true;document.body.classList.remove('is-suite-intro');document.querySelector('.suite-bar').hidden=false;document.getElementById('hub').hidden=false;document.getElementById('hub-title').focus({preventScroll:true});onEnter();}
 function visible(){if(done||started||!ready||document.hidden)return;started=true;screen.classList.add('counting');timer=setTimeout(finish,duration);button.focus({preventScroll:true});}
 function key(e){if(e.key==='Escape'){e.preventDefault();finish();}}
 button.onclick=finish;document.addEventListener('keydown',key);document.addEventListener('visibilitychange',visible);
 // Start the five seconds only after the image and native window are visible.
 const loaded=()=>{ready=true;visible();};
 if(image.complete)loaded();else{image.addEventListener('load',loaded,{once:true});image.addEventListener('error',loaded,{once:true});}
 return finish;
}
