from pathlib import Path
import json,hashlib,io,wave
import numpy as np
ROOT=Path(__file__).resolve().parents[1]
F=ROOT/'frontend'
aliases=json.loads((F/'asset-map.json').read_text('utf-8'))
build=json.loads((ROOT/'docs/five-cast-voice-build.json').read_text('utf-8'))
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
checks=0
for game in ['poker','blackjack','baccarat']:
 for cid in ['misaki','ayaka','rina','shizue','kathy']:
  folder='voice' if cid=='misaki' else 'voice-'+cid
  for name in ['manifest.json','reward-manifest.json']:
   manifest=json.loads((F/f'{game}/assets/{folder}/{name}').read_text('utf-8'))
   expected=40 if name=='manifest.json' else 1
   assert len(manifest['lines'])==expected,(game,cid,name)
   for line in manifest['lines']:
    for mode,file in line['files'].items():
     p=F/aliases[game+'/'+file['path']]
     assert sha(p)==file['sha256']
     assert file['style_id']==build['profiles'][cid]['style_id']
     assert file['postPitchScale']==(1.1 if cid=='misaki' else 1.0)
     checks+=1
audio=[]
for rel in sorted({x['path'] for x in build['clips']}):
 p=F/rel
 with wave.open(str(p),'rb') as w:
  assert w.getsampwidth()==2 and w.getnchannels()==1
  duration=w.getnframes()/w.getframerate()
  samples=np.frombuffer(w.readframes(w.getnframes()),dtype='<i2').astype(float)/32768
 peak=float(np.max(np.abs(samples)));rms=float(np.sqrt(np.mean(samples*samples)))
 assert duration>.15 and peak>.01 and rms>.0001,(rel,duration,peak,rms)
 clipped=int(np.count_nonzero(np.abs(samples)>=.999))
 assert clipped/max(1,len(samples))<.001,(rel,'clipping')
 audio.append({'path':rel,'duration':duration,'peak':peak,'rms':rms,'clippedSamples':clipped})
backup=Path('C:/work/game-org/CardGame-backup-20261028-104744')
preserved=[]
oldaliases=json.loads((backup/'frontend/asset-map.json').read_text('utf-8'))
for key,target in oldaliases.items():
 if any(key.endswith('/assets/'+suffix+'.png') for cid in ['misaki','ayaka','rina'] for suffix in [cid+'-full',cid+'-summer',cid+'-portrait','dealer']):
  assert sha(F/aliases[key])==sha(backup/'frontend'/target),key
  preserved.append(key)
for cid in ['shizue']:
 for suffix,source in [('full',f'fullbody/{cid}_normal.png'),('summer',f'fullbody/{cid}_summer.png')]:
  assert sha(F/f'cast-images/{cid}-{suffix}.png')==sha(Path('C:/work/game-org/reversi-queens/assets/characters')/source)
# Kathy artwork was explicitly updated on 2026-10-28. Validate adopted images.
for rel,digest in json.loads((ROOT/'docs/kathy-younger-images.json').read_text('utf-8'))['images'].items():
 assert sha(ROOT/rel)==digest,rel
assert len(build['clips'])==1260
report={'mappedClips':len(build['clips']),'gameFileChecks':checks,'uniqueRecordings':len(audio),'preservedImageAliases':preserved,'unchangedReferenceImagesMatch':True,'updatedKathyImagesMatch':True,'audio':audio}
(ROOT/'docs/five-cast-asset-validation.json').write_text(json.dumps(report,ensure_ascii=False,indent=2),'utf-8')
print(json.dumps({k:v for k,v in report.items() if k!='audio'},ensure_ascii=True))
