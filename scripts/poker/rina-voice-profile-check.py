import argparse,json,math
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--before',type=Path,required=True);args=p.parse_args();count=0;max_duration_difference=0
for name in ['manifest.json','reward-manifest.json']:
 before=json.loads((args.before/name).read_text(encoding='utf8'));after=json.loads((Path('frontend/assets/voice-rina')/name).read_text(encoding='utf8'))
 assert after['styleId']==18 and len(before['lines'])==len(after['lines'])
 for old,new in zip(before['lines'],after['lines']):
  assert old['id']==new['id'] and old['text']==new['text'] and old['kind']==new['kind']
  for mode in ['expressive','calm']:
   a,b=old['files'][mode],new['files'][mode];x,y=a['voice'],b['voice']
   assert y['queryStyleId']==16 and y['styleId']==18 and x['speedScale']==y['speedScale'] and x['intonationScale']==y['intonationScale']
   assert math.isclose(y['pitchScale']-x['pitchScale'],.06,abs_tol=1e-8)
   assert a['sha256']!=b['sha256'] and a['duration']==b['duration'] and a['pauses']==b['pauses'];max_duration_difference=max(max_duration_difference,abs(a['duration']-b['duration']));count+=1
assert count==82
report={'files':count,'styleId':18,'pitchScaleIncrease':.06,'textPreserved':True,'speedSettingPreserved':True,'intonationSettingPreserved':True,'timingPreserved':True,'maxDurationDifferenceSeconds':round(max_duration_difference,3)}
Path('docs/rina-voice-profile-check.json').write_text(json.dumps(report,indent=2),encoding='utf8');print(json.dumps(report))
