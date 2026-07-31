/**
 * @fileoverview Benchmark for measuring renderHistory performance.
 */

class MockElement {
  /**
   * @param {string} tagName
   */
  constructor(tagName) {
    this.tagName = tagName;
    this.children = [];
    this.className = '';
    this.innerHTML = '';
    this.appendChildCalls = 0;
  }

  /**
   * @param {MockElement|MockDocumentFragment} child
   */
  appendChild(child) {
    if (child instanceof MockDocumentFragment) {
      this.children.push(...child.children);
    } else {
      this.children.push(child);
    }
    this.appendChildCalls++;
  }
}

class MockDocumentFragment {
  constructor() {
    this.children = [];
  }

  /**
   * @param {MockElement} child
   */
  appendChild(child) {
    this.children.push(child);
  }
}

const mockDocument = {
  /**
   * @param {string} tagName
   * @return {MockElement}
   */
  createElement(tagName) {
    return new MockElement(tagName);
  },

  /**
   * @return {MockDocumentFragment}
   */
  createDocumentFragment() {
    return new MockDocumentFragment();
  },
};

/**
 * Original renderHistory implementation.
 * @param {MockElement} logsHistory
 * @param {Array<Object>} logs
 */
function renderHistoryOriginal(logsHistory, logs) {
  logsHistory.innerHTML = '';
  logs.forEach(item => {
    const div = mockDocument.createElement('div');
    div.className = 'history-item';
    div.innerHTML = `
      <div>
        <strong>${item.caregiver || 'Caregiver'}</strong>
        <p style="font-size: 12px; color: var(--text-muted);">${new Date(
          item.processed_at || Date.now()
        ).toLocaleTimeString()} - Shift Log</p>
      </div>
      <span class="badge">Verified Shift Log</span>
    `;
    logsHistory.appendChild(div);
  });
}

/**
 * Optimized renderHistory implementation using DocumentFragment.
 * @param {MockElement} logsHistory
 * @param {Array<Object>} logs
 */
function renderHistoryOptimized(logsHistory, logs) {
  logsHistory.innerHTML = '';
  const fragment = mockDocument.createDocumentFragment();
  logs.forEach(item => {
    const div = mockDocument.createElement('div');
    div.className = 'history-item';
    div.innerHTML = `
      <div>
        <strong>${item.caregiver || 'Caregiver'}</strong>
        <p style="font-size: 12px; color: var(--text-muted);">${new Date(
          item.processed_at || Date.now()
        ).toLocaleTimeString()} - Shift Log</p>
      </div>
      <span class="badge">Verified Shift Log</span>
    `;
    fragment.appendChild(div);
  });
  logsHistory.appendChild(fragment);
}

// Generate large test data
const logs = [];
for (let i = 0; i < 10000; i++) {
  logs.push({
    caregiver: `Caregiver ${i}`,
    processed_at: new Date().toISOString(),
  });
}

console.log('--- Establishing Baseline ---');
const logsHistoryOrig = new MockElement('div');
const startOrig = performance.now();
renderHistoryOriginal(logsHistoryOrig, logs);
const endOrig = performance.now();
const timeOrig = endOrig - startOrig;
console.log(`Original Appends on DOM: ${logsHistoryOrig.appendChildCalls}`);
console.log(`Original Render Time: ${timeOrig.toFixed(2)} ms`);

console.log('--- Measuring Optimized ---');
const logsHistoryOpt = new MockElement('div');
const startOpt = performance.now();
renderHistoryOptimized(logsHistoryOpt, logs);
const endOpt = performance.now();
const timeOpt = endOpt - startOpt;
console.log(`Optimized Appends on DOM: ${logsHistoryOpt.appendChildCalls}`);
console.log(`Optimized Render Time: ${timeOpt.toFixed(2)} ms`);

const pct = ((timeOrig - timeOpt) / timeOrig * 100).toFixed(1);
console.log(`Performance Improvement: ${pct}%`);
