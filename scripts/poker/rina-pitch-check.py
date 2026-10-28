import argparse,json,math
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--before',type=Path,required=True);args=p.parse_args();count=0
for name in ['manifest.json','reward-manifest.json']:
 before=json.loads((args.before/name).read_text(encoding='utf8'));after=json.loads((Path('frontend/assets/voice-rina')/name).read_text(encoding='utf8'))
 assert len(before['lines'])==len(after['lines'])
 for old,new in zip(before['lines'],after['lines']):
  assert old['id']==new['id'] and old['text']==new['text'] and old['kind']==new['kind']
  for mode in ['expressive','calm']:
   a,b=old['files'][mode],new['files'][mode];x,y=a['voice'],b['voice']
   assert x['styleId']==y['styleId'] and x['speedScale']==y['speedScale'] and x['intonationScale']==y['intonationScale']
   assert math.isclose(y['pitchScale']-x['pitchScale'],.05,abs_tol=1e-8)
   assert a['pauses']==b['pauses'] and a['duration']==b['duration'] and a['sha256']!=b['sha256']
   count+=1
assert count==82
report={'files':count,'pitchScaleIncrease':.05,'textPreserved':True,'voiceStylePreserved':True,'speedPreserved':True,'intonationPreserved':True,'pauseAndDurationPreserved':True}
Path('docs/rina-pitch-check.json').write_text(json.dumps(report,indent=2),encoding='utf8');print(json.dumps(report))
