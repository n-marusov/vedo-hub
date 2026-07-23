// Temporary referential-integrity analysis for design .pen files
const fs = require('fs');

function walkRefs(node, acc) {
  if (!node || typeof node !== 'object') return;
  if (node.ref) {
    const descKeys = node.descendants ? Object.keys(node.descendants) : [];
    acc.push({
      ref: node.ref,
      id: node.id,
      name: node.name || '-',
      enabled: node.enabled === false ? false : true,
      hasDescendants: !!node.descendants,
      descendantKeys: descKeys,
      descendantSample: node.descendants
        ? descKeys.map((k) => {
            const v = node.descendants[k];
            const val = v.content !== undefined ? v.content : JSON.stringify(v);
            return k + '=' + String(val).slice(0, 40);
          })
        : []
    });
  }
  for (const k of Object.keys(node)) {
    const v = node[k];
    if (Array.isArray(v)) v.forEach((c) => walkRefs(c, acc));
    else if (v && typeof v === 'object') walkRefs(v, acc);
  }
}

const target = process.argv[2];
const mode = process.argv[3] || 'all'; // 'all' or 'broken'
function loadLibIds() {
  const txt = fs.readFileSync('design/ui-kit.lib.pen', 'utf-8');
  const ids = txt.match(/"id"\s*:\s*"[A-Za-z0-9_]+"/g) || [];
  return new Set(ids.map((s) => s.replace(/^.*"id"\s*:\s*"/, '').replace(/"$/, '')));
}
const libIds = loadLibIds();

if (mode === 'libids') {
  console.log([...libIds].sort().join('\n'));
  process.exit(0);
}
console.error('[audit] library component ids loaded:', libIds.size);

const d = JSON.parse(fs.readFileSync(target, 'utf-8'));
const refs = [];
walkRefs(d, refs);
for (const r of refs) {
  const broken = !r.ref.startsWith('B:')
    ? false
    : !libIds.has(r.ref.slice(2));
  if (mode === 'broken' && !broken) continue;
  const flag = broken ? ' *** BROKEN ***' : '';
  console.log(`${r.ref} @${r.id} name="${r.name}" enabled=${r.enabled}${flag}`);
  if (r.descendantKeys.length) {
    console.log('   descendants: ' + r.descendantSample.join(' | '));
  }
}
