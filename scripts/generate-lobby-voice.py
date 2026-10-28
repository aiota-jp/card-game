"""Build-time VOICEVOX synthesis. Keeps each greeting continuous, including its invitation."""
import argparse,array,hashlib,io,json,wave
from pathlib import Path
from voicevox_core.blocking import Onnxruntime,OpenJtalk,Synthesizer,VoiceModelFile
p=argparse.ArgumentParser();p.add_argument('--runtime',type=Path,required=True);args=p.parse_args()
root=Path(__file__).resolve().parents[1];runtime=args.runtime.resolve();out=root/'frontend/lobby-voice';out.mkdir(exist_ok=True)
onnx=Onnxruntime.load_once(filename=str(runtime/'onnxruntime/lib/voicevox_onnxruntime.dll'))
synth=Synthesizer(onnx,OpenJtalk(str(runtime/'dict/open_jtalk_dic_utf_8-1.11')),acceleration_mode='CPU',cpu_num_threads=4)
with VoiceModelFile.open(str(runtime/'models/vvms/2.vvm')) as model:synth.load_voice_model(model)
lines=json.loads((root/'scripts/lobby-lines.json').read_text(encoding='utf8'))
for line in lines:
 line['files']={};character=line['character'];style=15 if character=='ayaka' else 18 if character=='rina' else 16
 for mode in ['expressive','calm']:
  q=synth.create_audio_query(line['text'],16 if character=='rina' else style)
  q.speed_scale,q.pitch_scale,q.intonation_scale=(1.02,0,1.05) if mode=='expressive' else (1.03,0,.85)
  if character=='misaki':q.pitch_scale-=.055;q.intonation_scale*=.93
  elif character=='ayaka':q.pitch_scale+=.04;q.intonation_scale*=1.04
  elif character=='rina':q.pitch_scale+=.035;q.intonation_scale*=.98;q.speed_scale*=1.06
  for phrase in q.accent_phrases:
   if phrase.pause_mora is not None:phrase.pause_mora['vowel_length']=min(phrase.pause_mora['vowel_length'],max(.10,min(.40,phrase.pause_mora['vowel_length']*.45)))
  q.pre_phoneme_length=.02;q.post_phoneme_length=.06;q.output_sampling_rate=24000;q.output_stereo=False
  raw=synth.synthesis(q,style)
  with wave.open(io.BytesIO(raw),'rb') as f:params=f.getparams();samples=array.array('h',f.readframes(f.getnframes()))
  peak=max(abs(x) for x in samples);gain=min(.85*32767/peak,1.8);samples=array.array('h',(max(-32767,min(32767,round(x*gain))) for x in samples))
  dest=out/(line['id']+'-'+mode+'.wav')
  with wave.open(str(dest),'wb') as f:f.setparams(params);f.writeframes(samples.tobytes())
  data=dest.read_bytes();line['files'][mode]={'path':'lobby-voice/'+dest.name,'sha256':hashlib.sha256(data).hexdigest(),'duration':len(samples)/24000}
  print(dest.name,round(len(samples)/24000,2),flush=True)
(out/'manifest.json').write_text(json.dumps({'credit':'VOICEVOX:九州そら','lines':lines},ensure_ascii=False,indent=2),encoding='utf8')
(root/'frontend/lobby-lines.js').write_text('export const lobbyLines = '+json.dumps(lines,ensure_ascii=False,indent=2)+';\n',encoding='utf8')
