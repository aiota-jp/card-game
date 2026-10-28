"""Original 16-bar lounge instrumental, 96 BPM, F major; no sampled recordings."""
from pathlib import Path
import numpy as np, wave, hashlib, json
ROOT=Path(__file__).resolve().parents[1]
SR=32000;BEAT=60/96;DURATION=16*4*BEAT
mix=np.zeros((int(SR*DURATION),2),dtype=np.float64)
rng=np.random.default_rng(20261028)
def add(y,start,pan=0):
 n=int(start*SR);length=min(len(y),len(mix)-n)
 if length<=0:return
 mix[n:n+length,0]+=y[:length]*np.sqrt((1-pan)/2)
 mix[n:n+length,1]+=y[:length]*np.sqrt((1+pan)/2)
def piano(midi,start,length,gain=.08,pan=-.18):
 t=np.arange(int(length*SR))/SR;f=440*2**((midi-69)/12)
 tone=sum(a*np.sin(2*np.pi*f*h*t)*np.exp(-t*(1.9+h*.7)) for h,a in [(1,1),(2,.3),(3,.13),(4,.05)])
 attack=np.minimum(1,t/.006);release=np.minimum(1,(length-t)/.035)
 add(tone*attack*release*gain,start,pan)
def bass(midi,start):
 t=np.arange(int(BEAT*.85*SR))/SR;f=440*2**((midi-69)/12)
 y=(np.sin(2*np.pi*f*t)+.18*np.sin(4*np.pi*f*t))*np.minimum(1,t/.008)*np.exp(-t*7)*.22
 add(y,start,.08)
chords=[[53,57,60,64],[50,53,57,60],[43,50,53,57],[48,52,55,58]]
melody=[None,72,76,74,72,None,69,67,None,69,72,76,74,72,None,69]
for bar in range(16):
 base=bar*4*BEAT;chord=chords[bar%4]
 for step in range(4):
  bass(chord[0]-12+(7 if step==3 else 0),base+step*BEAT)
  if step in [0,2]:
   for j,n in enumerate(chord):piano(n+12,base+(step+.12)*BEAT+j*.012,BEAT*1.25,.035)
  # Brushed percussion, quiet and fully synthesized.
  t=np.arange(int(.07*SR))/SR
  y=rng.standard_normal(len(t));y=np.concatenate(([0],np.diff(y)))*np.exp(-t*55)*.012
  add(y,base+(step+.66)*BEAT,.3)
  if step in [1,3]:add(y*.8,base+step*BEAT,-.3)
 if bar>=4:
  for step in range(4):
   note=melody[((bar-4)*4+step)%len(melody)]
   if note:piano(note,base+(step+.02)*BEAT,BEAT*.85,.07,.22)
# Short boundary fades prevent a pop on looping.
fade=int(.025*SR);mix[:fade]*=np.linspace(0,1,fade)[:,None];mix[-fade:]*=np.linspace(1,0,fade)[:,None]
peak=float(np.max(np.abs(mix)));mix*=.75/peak
path=ROOT/'frontend/lounge-bgm.wav'
with wave.open(str(path),'wb') as w:w.setnchannels(2);w.setsampwidth(2);w.setframerate(SR);w.writeframes(np.rint(mix*32767).astype('<i2').tobytes())
(ROOT/'docs/lobby-bgm.json').write_text(json.dumps({'title':'Five at the Lounge','original':True,'bpm':96,'key':'F major','bars':16,'duration':DURATION,'sampleRate':SR,'channels':2,'sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'source':'scripts/build-lobby-bgm.py','samples':'synthesized oscillators and seeded noise only'},indent=2),'utf-8')
print(f'Original lounge music: {DURATION}s, peak .75, {path.stat().st_size} bytes')
