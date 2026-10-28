"""Compare rebuilt voice timing with a build-time backup; no engine required."""
import argparse,json,wave
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--before',type=Path,required=True);args=p.parse_args();root=Path(__file__).resolve().parents[1];characters=[]
for folder in ['voice','voice-ayaka']:
 manifest=json.loads((root/'frontend/assets'/folder/'manifest.json').read_text(encoding='utf8'));old_seconds=new_seconds=0.;changed=pauses=0;max_old=max_new=0.
 for line in manifest['lines']:
  for mode,info in line['files'].items():
   def duration(path):
    with wave.open(str(path),'rb') as f:return f.getnframes()/f.getframerate()
   old=duration(args.before/folder/Path(info['path']).name);new=duration(root/'frontend'/info['path']);old_seconds+=old;new_seconds+=new
   assert new<old;changed+=1
   for pause in info['pauses']:
    assert 0<pause['after']<=.4 and pause['after']<=pause['before'];pauses+=1;max_old=max(max_old,pause['before']);max_new=max(max_new,pause['after'])
 assert changed==80 and pauses>0
 characters.append({'folder':folder,'changedFiles':changed,'originalSeconds':round(old_seconds,3),'newSeconds':round(new_seconds,3),'removedPauseSeconds':round(old_seconds-new_seconds,3),'pauseCount':pauses,'maxOriginalPauseSeconds':max_old,'maxNewPauseSeconds':max_new})
report={'characters':characters,'unchangedPlaybackRate':1.794,'openingDelayMs':80,'resultDelayMs':180}
(root/'docs/voice-pause-check.json').write_text(json.dumps(report,indent=2),encoding='utf8');print(json.dumps(report))
