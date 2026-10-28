// Original instrumental: 118 BPM, D minor. All audio is generated locally.
export const musicDuration = 16 * 4 * 60 / 118;
export async function renderMusic(Audio = window.OfflineAudioContext || window.webkitOfflineAudioContext) {
 const rate=32000, ctx=new Audio(2,Math.ceil(musicDuration*rate),rate), beat=60/118;
 const mix=ctx.createGain(), compressor=ctx.createDynamicsCompressor();
 mix.gain.value=.65;compressor.threshold.value=-14;compressor.ratio.value=4;mix.connect(compressor);compressor.connect(ctx.destination);
 const hz=n=>440*2**((n-69)/12);
 function note(midi,t,d,level,type='triangle',cutoff=1800){
  const o=ctx.createOscillator(),f=ctx.createBiquadFilter(),g=ctx.createGain();o.type=type;o.frequency.value=hz(midi);f.type='lowpass';f.frequency.value=cutoff;
  g.gain.setValueAtTime(0,t);g.gain.linearRampToValueAtTime(level,t+.008);g.gain.exponentialRampToValueAtTime(.00001,t+d);
  o.connect(f);f.connect(g);g.connect(mix);o.start(t);o.stop(Math.min(musicDuration,t+d+.01));
 }
 const noise=ctx.createBuffer(1,rate*.2,rate),data=noise.getChannelData(0);let seed=982451653;
 for(let i=0;i<data.length;i++){seed=(Math.imul(seed,1664525)+1013904223)|0;data[i]=seed/2147483648;}
 function drum(t,open=false,snare=false){
  const s=ctx.createBufferSource(),f=ctx.createBiquadFilter(),g=ctx.createGain();s.buffer=noise;f.type='highpass';f.frequency.value=snare?1400:6500;
  const d=snare?.12:open?.13:.035;g.gain.setValueAtTime(snare?.16:open?.07:.045,t);g.gain.exponentialRampToValueAtTime(.00001,t+d);
  s.connect(f);f.connect(g);g.connect(mix);s.start(t);s.stop(Math.min(musicDuration,t+d));
 }
 const chords=[[50,57,60,65],[46,53,58,62],[43,50,58,62],[45,52,57,61]];
 // An original, spacious motif; no quoted melody or sampled recording.
 const motif=[74,null,69,72,null,77,76,null,69,65,null,67,72,null,64,69];
 for(let bar=0;bar<16;bar++){
  const chord=chords[bar%4],base=bar*4*beat;
  for(let b=0;b<4;b++){
   const t=base+b*beat,o=ctx.createOscillator(),g=ctx.createGain();o.frequency.setValueAtTime(145,t);o.frequency.exponentialRampToValueAtTime(43,t+.12);g.gain.setValueAtTime(.48,t);g.gain.exponentialRampToValueAtTime(.00001,t+.23);o.connect(g);g.connect(mix);o.start(t);o.stop(t+.24);
   if(b%2===1)drum(t,false,true);
   drum(t+beat/2,true);
  }
  for(let step=0;step<8;step++){
   const t=base+step*beat/2;drum(t);note(chord[0]-12+(step===6?12:0),t+beat*.1,beat*.32,.20,'sawtooth',450);
   if(step%2===1)for(const n of chord)note(n+12,t,beat*.28,.025,'sawtooth',1200);
   if(bar>=4){const m=motif[(bar%2)*8+step];if(m!==null)note(m+(bar>=12?-12:0),t,beat*.38,.055,'triangle',2200);}
  }
 }
 return ctx.startRendering();
}
const read=(key,fallback)=>{try{return localStorage.getItem(key)??fallback}catch{return fallback}};
export function createMusic(){
 let enabled=read('cardlounge.music.enabled','true')==='true',volume=Math.max(0,Math.min(100,Number(read('cardlounge.music.volume','25'))||0));
 let ctx,gain,source,buffer,rendering,unlocked=false,failed=false,ducked=false;
 function save(){try{localStorage.setItem('cardlounge.music.enabled',String(enabled));localStorage.setItem('cardlounge.music.volume',String(volume))}catch{}}
 function level(){if(gain)gain.gain.setTargetAtTime(enabled?volume/100*(ducked?.25:1):0,ctx.currentTime,.08)}
 async function unlock(){
  unlocked=true;if(!enabled||volume===0||document.hidden||failed)return;
  try{
   if(!ctx){const Audio=window.AudioContext||window.webkitAudioContext;if(!Audio)return;ctx=new Audio();gain=ctx.createGain();gain.gain.value=0;gain.connect(ctx.destination)}
   if(ctx.state==='suspended')await ctx.resume();
   if(!buffer){rendering??=renderMusic();buffer=await rendering;}
   if(!source){source=ctx.createBufferSource();source.buffer=buffer;source.loop=true;source.loopEnd=musicDuration;source.connect(gain);source.start()}
   level();
  }catch{failed=true;/* A missing audio device never blocks play. */}
 }
 document.addEventListener('visibilitychange',()=>{if(document.hidden){if(ctx?.state==='running')ctx.suspend().catch(()=>{})}else if(unlocked)unlock()});
 window.addEventListener('pagehide',()=>{if(ctx)ctx.close().catch(()=>{})});
 return {get enabled(){return enabled},get volume(){return volume},unlock,
  setDucked(value){ducked=Boolean(value);level()},
  setEnabled(value){enabled=Boolean(value);save();level();if(enabled)unlock()},
  setVolume(value){volume=Math.max(0,Math.min(100,Number(value)||0));save();level();if(enabled&&volume>0)unlock()}};
}
