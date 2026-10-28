const fs=require('fs'),path=require('path'),assert=require('assert/strict');
(async()=>{
 const source=fs.readFileSync(path.join(__dirname,'../frontend/rewards.js'),'utf8');const {isMatchVictory}=await import('data:text/javascript;base64,'+Buffer.from(source).toString('base64'));
 const previous={done:false,variant:'texas',hand:1,stacks:[1000,1000]},win={done:true,outcome:'win',variant:'texas',hand:1,stacks:[2000,0]};
 const cases=[['match win',previous,win,true],['ordinary hand win',previous,{...win,stacks:[1030,970]},false],['opponent has one chip',previous,{...win,stacks:[1999,1]},false],['unsettled all-in',previous,{...win,done:false},false],['player bust',previous,{...win,outcome:'loss',stacks:[0,2000]},false],['tie',previous,{...win,outcome:'tie'},false],['both empty',previous,{...win,stacks:[0,0]},false],['already shown result',win,win,false],['next match',win,{...win,hand:2},true],['missing stacks',previous,{...win,stacks:undefined},false],['other variant',win,{...win,variant:'omaha'},true],['already all-in before resolution',{...previous,stacks:[1990,0]},win,true]];
 for(const [name,before,after,expected] of cases)assert.equal(isMatchVictory(before,after),expected,name);
 console.log('Verified '+cases.length+' match victory boundary cases');
})().catch(e=>{console.error(e);process.exitCode=1});
