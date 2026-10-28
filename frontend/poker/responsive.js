// Collapse secondary information only when crossing a layout breakpoint.
// The user can still open it without changing the table or starting a hand.
const compact = window.matchMedia('(max-height: 900px), (max-width: 879px)');
const details = document.getElementById('secondary-panels');
function adapt() {
 document.body.classList.toggle('compact-layout', compact.matches);
 details.open = !compact.matches;
}
details.addEventListener('toggle', () => {if(details.open){const log=document.getElementById('log');log.scrollTop=log.scrollHeight;}});
compact.addEventListener('change', adapt);
adapt();

