import argparse,array,hashlib,json,math,wave
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('--character',choices=['misaki','ayaka','rina'],default='misaki');args=parser.parse_args();folder='voice' if args.character=='misaki' else 'voice-'+args.character;root=Path('.');manifest=json.loads((root/'frontend/assets'/folder/'manifest.json').read_text(encoding='utf8'));report=[];hashes=set()
for line in manifest['lines']:
 for mode,info in line['files'].items():
  assert all(p['after']<=.4 and p['after']<=p['before'] for p in info.get('pauses',[]))
  path=root/'frontend'/info['path'];raw=path.read_bytes();digest=hashlib.sha256(raw).hexdigest()
  assert digest==info['sha256'] and digest not in hashes;hashes.add(digest)
  with wave.open(str(path),'rb') as f:
   assert f.getnchannels()==1 and f.getsampwidth()==2 and f.getframerate()==24000
   samples=array.array('h',f.readframes(f.getnframes()));duration=len(samples)/24000
  peak=max(abs(x) for x in samples)/32768;rms=math.sqrt(sum((x/32768)**2 for x in samples)/len(samples))
  assert .01<peak<.99 and rms>.005 and abs(duration-info['duration'])<.002
  report.append({'id':line['id'],'mode':mode,'duration':round(duration,3),'peak':round(peak,4),'rms':round(rms,4)})
assert len(report)==80
(root/('docs/voice-audio-check.json' if args.character=='misaki' else 'docs/voice-audio-check-'+args.character+'.json')).write_text(json.dumps({'lines':40,'files':80,'credit':manifest['credit'],'checks':report},ensure_ascii=False,indent=2),encoding='utf8')
print(f"Verified {len(report)} WAV files; {sum(r['duration'] for r in report):.1f} seconds total")
