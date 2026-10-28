// Original effects synthesized locally. No network, audio files or autoplay.
export const effectNames = ['click', 'check', 'chips', 'fold', 'deal', 'win', 'loss', 'tie', 'error'];

export function scheduleEffect(ctx, destination, name, start = ctx.currentTime, volume = 0.4) {
 const level = Math.max(0, Math.min(1, volume));
 function tone(freq, offset, duration, strength = 0.12, type = 'sine', end = freq) {
  const oscillator = ctx.createOscillator();
  const envelope = ctx.createGain();
  oscillator.type = type;
  oscillator.frequency.setValueAtTime(freq, start + offset);
  oscillator.frequency.exponentialRampToValueAtTime(Math.max(20, end), start + offset + duration);
  envelope.gain.setValueAtTime(0, start + offset);
  envelope.gain.linearRampToValueAtTime(strength * level, start + offset + 0.006);
  envelope.gain.exponentialRampToValueAtTime(0.00001, start + offset + duration);
  oscillator.connect(envelope); envelope.connect(destination);
  oscillator.start(start + offset); oscillator.stop(start + offset + duration + 0.01);
 }
 function shuffle(offset) {
  const length = Math.ceil(ctx.sampleRate * 0.065);
  const buffer = ctx.createBuffer(1, length, ctx.sampleRate);
  const samples = buffer.getChannelData(0);
  for (let i = 0; i < length; i++) samples[i] = (Math.random() * 2 - 1) * (1 - i / length);
  const source = ctx.createBufferSource(), filter = ctx.createBiquadFilter(), gain = ctx.createGain();
  source.buffer = buffer; filter.type = 'bandpass'; filter.frequency.value = 2300; filter.Q.value = 0.7;
  gain.gain.setValueAtTime(0.14 * level, start + offset);
  gain.gain.exponentialRampToValueAtTime(0.00001, start + offset + 0.065);
  source.connect(filter); filter.connect(gain); gain.connect(destination); source.start(start + offset);
 }
 switch (name) {
  case 'click': tone(620, 0, 0.045, 0.085, 'sine', 380); break;
  case 'check': tone(310, 0, 0.085, 0.16, 'sine', 145); break;
  case 'chips': tone(1350, 0, 0.09, 0.13); tone(1880, 0.045, 0.13, 0.09); tone(1130, 0.09, 0.1, 0.07); break;
  case 'fold': shuffle(0); tone(250, 0, 0.11, 0.08, 'triangle', 110); break;
  case 'deal': shuffle(0); shuffle(0.085); tone(430, 0.085, 0.045, 0.045); break;
  case 'win': [523.25, 659.25, 783.99, 1046.5].forEach((f, i) => tone(f, i * 0.105, 0.38, 0.13)); tone(523.25, 0.315, 0.5, 0.06); break;
  case 'loss': [392, 329.63, 261.63].forEach((f, i) => tone(f, i * 0.13, 0.3, 0.09, 'triangle')); break;
  case 'tie': tone(440, 0, 0.24, 0.1); tone(554.37, 0.14, 0.28, 0.1); break;
  case 'error': tone(220, 0, 0.1, 0.1, 'triangle'); tone(196, 0.13, 0.14, 0.1, 'triangle'); break;
 }
}

function stored(key, fallback) { try { return localStorage.getItem(key) ?? fallback; } catch { return fallback; } }
function persist(key, value) { try { localStorage.setItem(key, String(value)); } catch { /* Settings remain available for this session. */ } }

export function createSound() {
 let context;
 let master;
 let enabled = stored('cardlounge.sound.enabled', 'true') === 'true';
 const savedVolume = Number(stored('cardlounge.sound.volume', '40'));
 let volume = Number.isFinite(savedVolume) ? Math.max(0, Math.min(100, savedVolume)) : 40;
 let latest = 0;
 function unlock() {
  if (!enabled || volume === 0) return;
  try {
   const Audio = window.AudioContext || window.webkitAudioContext;
   if (!Audio) return;
   if (!context) {
    context = new Audio(); master = context.createGain();
    master.gain.value = volume / 100; master.connect(context.destination);
   }
   if (context.state === 'suspended') context.resume().catch(() => {});
  } catch { /* Audio failure never blocks the game. */ }
 }
 return {
  get enabled() { return enabled; }, get volume() { return volume; },
  unlock,
  setEnabled(value) { enabled = Boolean(value); persist('cardlounge.sound.enabled', enabled); if (master) master.gain.setTargetAtTime(enabled ? volume / 100 : 0, context.currentTime, 0.01); if (enabled) unlock(); },
  setVolume(value) { volume = Math.max(0, Math.min(100, Number(value) || 0)); persist('cardlounge.sound.volume', volume); if (master) master.gain.setTargetAtTime(enabled ? volume / 100 : 0, context.currentTime, 0.01); },
  play(name, delay = 0) {
   if (!enabled || volume === 0 || !effectNames.includes(name)) return;
   unlock(); if (!context) return;
   // Avoid a click storm while still allowing deal + outcome as a sequence.
   const now = performance.now(); if (name === 'click' && now - latest < 45) return;
   latest = now;
   try { scheduleEffect(context, master, name, context.currentTime + 0.012 + delay, 1); } catch { /* Silent fallback. */ }
  },
 };
}

// Initial snapshots and repeated renders are silent; only new state is sounded.
export function transitionEffects(previous, next) {
 if (!previous) return [];
 const effects = [];
 const newHand = next.hand !== previous.hand;
 if (newHand || next.board.length > previous.board.length || (next.player?.length || 0) > (previous.player?.length || 0)) effects.push({name: 'deal', delay: 0.05});
 if (next.done && (!previous.done || newHand)) effects.push({name: next.outcome || 'tie', delay: effects.length ? 0.26 : 0.12});
 return effects;
}
