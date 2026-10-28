from pathlib import Path
import zipfile
ROOT=Path(__file__).resolve().parents[1]
dest=ROOT/'card-game-lounge.zip'
temp=ROOT/'card-game-lounge.zip.tmp'
excluded={'.tools','build','node_modules','__pycache__','.git','.history'}
files=[p for p in ROOT.rglob('*') if p.is_file() and not any(x in excluded for x in p.relative_to(ROOT).parts) and p not in [dest,temp] and p.suffix not in ['.log','.tmp']]
with zipfile.ZipFile(temp,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=6) as z:
 for p in files:z.write(p,Path('CardGame')/p.relative_to(ROOT))
with zipfile.ZipFile(temp) as z:assert z.testzip() is None
temp.replace(dest)
print(f'Packaged {len(files)} files: {dest.stat().st_size:,} bytes')
