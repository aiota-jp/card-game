// Only a settled match victory unlocks the summer celebration.
export function isMatchVictory(previous,next){
 return next?.done===true && next.outcome==='win' && next.stacks?.[1]===0 && next.stacks?.[0]>0 && (!previous?.done || previous.hand!==next.hand || previous.variant!==next.variant);
}
