"""Synthesize card dialogue with the reference edition's approved local profiles.

Run using the reference .voice-lab Python (librosa/numpy/soundfile installed).
Cache keys include spoken text, all profile settings, mode and pitch processing.
"""
from pathlib import Path
import json, hashlib, urllib.request, urllib.parse, io, wave, math, time
import numpy as np
import librosa
ROOT=Path(__file__).resolve().parents[1]
FRONT=ROOT/'frontend'
BASE='http://127.0.0.1:10113'
profiles=json.loads((ROOT/'docs/five-cast-voice-profiles.json').read_text('utf-8'))
plans=json.loads((ROOT/'scripts/five-cast-voice-plan.json').read_text('utf-8'))
aliases=json.loads((FRONT/'asset-map.json').read_text('utf-8'))
out=FRONT/'cast-voice';out.mkdir(exist_ok=True)
readings={'MISAKI':'みさき','AYAKA':'あやか','RINA':'りな','SHIZUE':'しずえ','KATHY':'キャシー','美咲':'みさき','静江':'しずえ'}
def post(path,value):
    request=urllib.request.Request(BASE+path,data=json.dumps(value,ensure_ascii=False).encode(),headers={'Content-Type':'application/json'})
    for attempt in range(3):
        try:return urllib.request.urlopen(request,timeout=240).read()
        except Exception:
            if attempt==2:raise
            time.sleep(2)
def save(path,obj):path.parent.mkdir(parents=True,exist_ok=True);path.write_text(json.dumps(obj,ensure_ascii=False,indent=2),'utf-8')
records=[];cache={};done=0
for plan in plans:
    for line in plan['lines']:
        cid=plan.get('character',line.get('character'));profile=profiles[cid]
        spoken=line['text']
        for a,b in readings.items():spoken=spoken.replace(a,b)
        files={}
        for mode in ['expressive','calm']:
            params=dict(profile['parameters']);params['volumeScale']=1.0
            if mode=='calm':params['intonationScale']*=.85
            signature=hashlib.sha256(json.dumps([cid,spoken,profile['style_id'],params,profile['postPitchScale'],'v1'],ensure_ascii=False,sort_keys=True).encode()).hexdigest()
            path=out/(signature+'.wav')
            if not path.exists():
                query=json.loads(post('/audio_query?'+urllib.parse.urlencode({'text':spoken,'speaker':profile['style_id']}),{}))
                query.update(params)
                raw=post('/synthesis?speaker='+str(profile['style_id']),query)
                if profile['postPitchScale']!=1.0:
                    with wave.open(io.BytesIO(raw),'rb') as w:
                        rate=w.getframerate();channels=w.getnchannels();assert channels==1 and w.getsampwidth()==2
                        samples=np.frombuffer(w.readframes(w.getnframes()),dtype='<i2').astype(np.float32)/32768
                    shifted=librosa.effects.pitch_shift(samples,sr=rate,n_steps=12*math.log2(profile['postPitchScale']))
                    # Keep exact duration and reserve headroom without changing normal loudness.
                    shifted=librosa.util.fix_length(shifted,size=len(samples));peak=float(np.max(np.abs(shifted)))
                    if peak>.98:shifted*=.98/peak
                    buffer=io.BytesIO()
                    with wave.open(buffer,'wb') as w:
                        w.setnchannels(1);w.setsampwidth(2);w.setframerate(rate);w.writeframes(np.rint(shifted*32767).astype('<i2').tobytes())
                    raw=buffer.getvalue()
                temp=path.with_suffix('.tmp');temp.write_bytes(raw);temp.replace(path)
                done+=1
                if done%10==0:print(f'Generated {done} new clips: {cid} {line["id"]} {mode}',flush=True)
            raw=path.read_bytes()
            with wave.open(io.BytesIO(raw),'rb') as w:duration=w.getnframes()/w.getframerate();assert duration>.15
            sha=hashlib.sha256(raw).hexdigest()
            if plan['game']:
                folder='voice' if cid=='misaki' else 'voice-'+cid
                virtual=f'assets/{folder}/{line["id"]}-{mode}.wav'
                aliases[plan['game']+'/'+virtual]=path.relative_to(FRONT).as_posix()
            else:virtual=path.relative_to(FRONT).as_posix()
            files[mode]={'path':virtual,'sha256':sha,'duration':duration,'style_id':profile['style_id'],'intonationScale':params['intonationScale'],'postPitchScale':profile['postPitchScale']}
            records.append({'character':cid,'text':line['text'],'mode':mode,'path':path.relative_to(FRONT).as_posix(),'sha256':sha,'signature':signature,'duration':duration})
        line['files']=files
    name='voiceLines' if plan['game'] else 'lobbyLines'
    (ROOT/plan['module']).write_text('export const '+name+' = '+json.dumps(plan['lines'],ensure_ascii=False,indent=2)+';\n','utf-8')
    manifest={'provider':'AivisSpeech','license':'ACML 1.0','profiles':profiles,'lines':plan['lines']}
    # A direct file supersedes the old shared manifest alias.
    aliases.pop(plan['manifest'],None)
    save(FRONT/plan['manifest'],manifest)
    save(FRONT/'asset-map.json',aliases)
    print(f'Completed {plan["module"]}',flush=True)
save(ROOT/'docs/five-cast-voice-build.json',{'reference':str(Path('C:/work/game-test/reversi-queens')),'profiles':profiles,'clips':records,'uniqueFiles':len({r['path'] for r in records})})
print(f'Complete: {len(records)} mapped clips / {len({r["path"] for r in records})} distinct recordings',flush=True)
