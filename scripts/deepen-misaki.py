"""Regenerate MISAKI only; preserve timing, other characters, and original asset URLs."""
import argparse,array,hashlib,io,json,wave
from pathlib import Path
from voicevox_core.blocking import Onnxruntime,OpenJtalk,Synthesizer,VoiceModelFile
p=argparse.ArgumentParser();p.add_argument('--runtime',type=Path,required=True);args=p.parse_args()
root=Path(__file__).resolve().parents[1];front=root/'frontend';runtime=args.runtime.resolve()
aliases=json.loads((front/'asset-map.json').read_text(encoding='utf8'));cache={}
onnx=Onnxruntime.load_once(filename=str(runtime/'onnxruntime/lib/voicevox_onnxruntime.dll'))
synth=Synthesizer(onnx,OpenJtalk(str(runtime/'dict/open_jtalk_dic_utf_8-1.11')),acceleration_mode='CPU',cpu_num_threads=4)
with VoiceModelFile.open(str(runtime/'models/vvms/2.vvm')) as model:synth.load_voice_model(model)
profiles={'reward':(1.02,0,1.08),'welcome':(1.02,0,1.05),'deal':(1.06,0,1.03),'win':(1.08,.015,1.16),'loss':(.99,-.005,.96),'tie':(1.04,.005,1.07),'draw':(1.01,0,1.02),'reveal':(1.02,.005,1.10),'raise':(1.06,.01,1.14),'streakWin':(1.07,.012,1.16),'streakLoss':(.98,-.005,.96),'fold':(1.0,0,1.0)}
def render(line,mode):
 kind=line.get('kind','welcome');key=(line['text'],kind,mode)
 if key in cache:return cache[key]
 q=synth.create_audio_query(line['text'],16);q.speed_scale,q.pitch_scale,q.intonation_scale=profiles[kind] if mode=='expressive' else (1.03,0,.85)
 q.pitch_scale-=.055;q.intonation_scale*=.93
 for phrase in q.accent_phrases:
  if phrase.pause_mora is not None:phrase.pause_mora['vowel_length']=min(phrase.pause_mora['vowel_length'],max(.10,min(.40,phrase.pause_mora['vowel_length']*.45)))
 q.pre_phoneme_length=.02;q.post_phoneme_length=.06;q.output_sampling_rate=24000;q.output_stereo=False
 raw=synth.synthesis(q,16)
 with wave.open(io.BytesIO(raw),'rb') as f:params=f.getparams();samples=array.array('h',f.readframes(f.getnframes()))
 gain=min(.85*32767/max(abs(x) for x in samples),1.8);samples=array.array('h',(max(-32767,min(32767,round(x*gain))) for x in samples))
 buf=io.BytesIO()
 with wave.open(buf,'wb') as f:f.setparams(params);f.writeframes(samples.tobytes())
 data=buf.getvalue();meta={'sha256':hashlib.sha256(data).hexdigest(),'duration':len(samples)/24000,'pitchScale':q.pitch_scale,'intonationScale':q.intonation_scale}
 cache[key]=(data,meta);print(len(cache),line['id'],mode,flush=True);return data,meta
for game in ['poker','blackjack','baccarat']:
 for manifest,module in [('manifest.json','voice-lines.js'),('reward-manifest.json','misaki-reward-lines.js')]:
  virtual=game+'/assets/voice/'+manifest;obj=json.loads((front/aliases[virtual]).read_text(encoding='utf8'))
  for line in obj['lines']:
   for mode in ['expressive','calm']:
    data,meta=render(line,mode);dest='shared/'+meta['sha256']+'.wav';(front/dest).write_bytes(data);aliases[game+'/'+line['files'][mode]['path']]=dest;line['files'][mode].update(meta)
  raw=json.dumps(obj,ensure_ascii=False,indent=2).encode();dest='shared/'+hashlib.sha256(raw).hexdigest()+'.json';(front/dest).write_bytes(raw);aliases[virtual]=dest
  (front/game/module).write_text('export const voiceLines = '+json.dumps(obj['lines'],ensure_ascii=False,indent=2)+';\n',encoding='utf8')
obj=json.loads((front/'lobby-voice/manifest.json').read_text(encoding='utf8'))
for line in obj['lines']:
 if line['character']!='misaki':continue
 for mode in ['expressive','calm']:
  data,meta=render(line,mode);(front/line['files'][mode]['path']).write_bytes(data);line['files'][mode].update(meta)
(front/'lobby-voice/manifest.json').write_text(json.dumps(obj,ensure_ascii=False,indent=2),encoding='utf8')
(front/'lobby-lines.js').write_text('export const lobbyLines = '+json.dumps(obj['lines'],ensure_ascii=False,indent=2)+';\n',encoding='utf8')
(front/'asset-map.json').write_text(json.dumps(aliases,ensure_ascii=False,indent=2),encoding='utf8')
# Only shared files no longer reachable through any game's map are removed.
used=set(aliases.values())
for file in (front/'shared').iterdir():
 if file.is_file() and file.relative_to(front).as_posix() not in used:file.unlink()
print('Complete:',len(cache),'MISAKI recordings',flush=True)
