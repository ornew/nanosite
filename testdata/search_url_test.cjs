const fs = require('node:fs');
const vm = require('node:vm');
const assert = require('node:assert/strict');
let inputListener;
const node = () => ({ children: [], append(...children) { this.children.push(...children); }, addEventListener() {}, focus() {} });
const theme = node(), navigation = node(), input = node(), results = node();
input.addEventListener = (event, cb) => { if (event === 'input') inputListener = cb; };
results.replaceChildren = () => { results.children = []; };
const context = {
 document: {
  documentElement: { dataset: {} },
  querySelector: selector => ({'#theme': theme, '#navigation': navigation, '#search': input, '#search-results': results}[selector]),
  scripts: [{ src: 'https://example.com/project/assets/search-index.js' }],
  createElement: node, addEventListener() {}
 },
 matchMedia: () => ({ matches: false, addEventListener() {} }),
 localStorage: { setItem() {} }, URL,
 NANOSITE_SEARCH: [{title: 'Unicode', text: 'Unicode', url: 'docs/name%20%23%25/index.html#%E6%97%A5%E6%9C%AC%E8%AA%9E'}]
};
vm.runInNewContext(fs.readFileSync(process.argv[2], 'utf8'), context);
input.value = 'Unicode'; inputListener();
const link = results.children[1].children[0].children[0];
assert.equal(link.href, 'https://example.com/project/docs/name%20%23%25/index.html#%E6%97%A5%E6%9C%AC%E8%AA%9E');
assert.equal(decodeURIComponent(new URL(link.href).hash.slice(1)), '日本語');
console.log('Search preserves encoded paths, deployment prefixes and heading fragments');

// A custom template can retain search while omitting default theme/navigation controls.
context.document.querySelector = selector => ({'#search': input, '#search-results': results}[selector]);
vm.runInNewContext(fs.readFileSync(process.argv[2], 'utf8'), context);
inputListener();
assert.equal(results.children[1].children[0].children[0].href, link.href);

context.NANOSITE_SEARCH[0].url = 'javascript:example/index.html#heading';
inputListener();
assert.equal(results.children[1].children[0].children[0].href, 'https://example.com/project/javascript:example/index.html#heading');
