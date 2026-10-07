// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
"use strict";

// Execute the shipped client unchanged. Only DOM, time and HTTP are fixtures;
// these tests make no claim about a browser engine or learned-model quality.
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

const webDirectory = path.join(__dirname, "../web");
const appSource = fs.readFileSync(path.join(webDirectory, "app.js"), "utf8");
const html = fs.readFileSync(path.join(webDirectory, "index.html"), "utf8");
const schema = "riidolaya-live-claims-v1";
const headNames = ["response_requested", "current_activity_claimed", "completion_claimed"];
const classNames = ["true", "false", "unknown"];

function element(children = {}, collections = {}) {
  const attributes = new Map();
  const listeners = new Map();
  return {
    dataset: {}, textContent: "", value: "", hidden: false, disabled: false,
    querySelector(selector) {
      assert.ok(Object.hasOwn(children, selector), `unsupported fixture selector: ${selector}`);
      return children[selector];
    },
    querySelectorAll(selector) {
      assert.ok(Object.hasOwn(collections, selector), `unsupported fixture collection: ${selector}`);
      return collections[selector];
    },
    setAttribute(name, value) { attributes.set(name, String(value)); },
    getAttribute(name) { return attributes.get(name) ?? null; },
    removeAttribute(name) {
      attributes.delete(name);
      if (name.startsWith("data-")) delete this.dataset[name.slice(5)];
    },
    addEventListener(type, listener) {
      if (!listeners.has(type)) listeners.set(type, []);
      listeners.get(type).push(listener);
    },
    dispatch(type) {
      assert.ok(listeners.has(type), `missing client event listener: ${type}`);
      for (const listener of listeners.get(type)) listener({ type, target: this });
    },
    focus() { this.focused = true; },
    setSelectionRange(start, end) { this.selection = [start, end]; },
  };
}

function documentFixture() {
  const elements = {};
  for (const id of ["claim-text", "byte-counter", "input-error", "clear-text", "run-status",
    "run-status-text", "result-announcement", "inference-details", "model-details"]) {
    assert.ok(html.includes(`id="${id}"`), `client element missing from shipped HTML: ${id}`);
    elements[`#${id}`] = element();
  }
  assert.match(html, /class="results"/);
  elements[".results"] = element();
  const articles = [...html.matchAll(/<article\b[^>]*data-head="([^"]+)"[^>]*>([\s\S]*?)<\/article>/g)];
  assert.deepEqual(articles.map(match => match[1]), headNames, "shipped cards must match response heads");
  const cards = articles.map(([, head, markup]) => {
    const title = markup.match(/<h3\b[^>]*>([^<]+)<\/h3>/);
    assert.ok(title, `card title missing: ${head}`);
    const children = {};
    for (const selector of [".decision-value", ".raw-candidate strong", ".candidate-confidence", ".decision-note"]) {
      const className = selector.split(" ")[0].slice(1);
      assert.ok(markup.includes(`class="${className}"`), `card element missing: ${selector}`);
      children[selector] = element();
    }
    children.h3 = element();
    children.h3.textContent = title[1];
    const rowClasses = [...markup.matchAll(/class="probability-row" data-class="([^"]+)"/g)].map(match => match[1]);
    assert.deepEqual(rowClasses, classNames, `shipped probability rows: ${head}`);
    const rows = rowClasses.map(className => {
      const row = element({ meter: element(), ".probability-value": element() });
      row.dataset.class = className;
      children[`[data-class="${className}"]`] = row;
      return row;
    });
    const card = element(children, { ".probability-row": rows });
    card.dataset.head = head;
    return card;
  });
  const examples = [...html.matchAll(/data-example="([^"]+)"/g)].map(([, name]) => {
    const button = element();
    button.dataset.example = name;
    return button;
  });
  assert.deepEqual(examples.map(button => button.dataset.example), ["request", "activity", "completion"]);
  const document = element(elements, { "[data-head]": cards, "[data-example]": examples });
  return { document, elements, cards, examples };
}

function clockFixture() {
  let now = 0;
  let nextID = 0;
  const timers = new Map();
  return {
    setTimeout(callback, milliseconds) {
      const id = ++nextID;
      timers.set(id, { callback, time: now + milliseconds });
      return id;
    },
    clearTimeout(id) { timers.delete(id); },
    advance(milliseconds) {
      const end = now + milliseconds;
      for (;;) {
        const next = [...timers.entries()].filter(([, timer]) => timer.time <= end)
          .sort((a, b) => a[1].time - b[1].time || a[0] - b[0])[0];
        if (!next) break;
        const [id, timer] = next;
        timers.delete(id);
        now = timer.time;
        timer.callback();
      }
      now = end;
    },
  };
}

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((fulfill, fail) => { resolve = fulfill; reject = fail; });
  return { promise, resolve, reject };
}

function createDemo(source = appSource) {
  const dom = documentFixture();
  const clock = clockFixture();
  const requests = [];
  const fetch = (url, options) => {
    const headers = deferred();
    const body = deferred();
    const request = {
      url, options, jsonReads: 0,
      resolveHeaders(status = 200) {
        headers.resolve({ ok: status >= 200 && status < 300, status,
          json: () => { request.jsonReads += 1; return body.promise; } });
      },
      resolveBody: body.resolve,
      respond(data, status = 200) { this.resolveHeaders(status); body.resolve(data); },
      reject: headers.reject,
    };
    requests.push(request);
    // Deliberately allow an aborted request to resolve: already-arrived headers
    // or a delayed JSON decode must still be guarded by the client's sequence.
    return headers.promise;
  };
  vm.runInNewContext(source, { document: dom.document, window: clock, fetch, TextEncoder,
    AbortController, Intl }, { filename: "app.js", timeout: 1000 });
  assert.equal(requests.length, 1);
  assert.equal(requests[0].url, "/api/status");
  return {
    ...dom, clock, requests, statusRequest: requests[0],
    get hints() { return requests.filter(request => request.url === "/api/hints"); },
    get phase() { return dom.elements["#run-status"].dataset.phase; },
    input(text) {
      dom.elements["#claim-text"].value = text;
      dom.elements["#claim-text"].dispatch("input");
    },
  };
}

function statusFixture(maxBytes = 4096) {
  return { schema, model: { name: "Client status fixture; not a research model", training_steps: 0 },
    limits: { max_text_bytes: maxBytes } };
}

function predictionFixture(states = ["unknown", "false", "unknown"]) {
  return { schema, state: "ready", input_bytes: 7, inference_us: 1200,
    model: { name: "Client response fixture; not a research model", training_steps: 0 },
    prediction: { heads: headNames.map((head, index) => ({ head, state: states[index],
      winner: classNames[index], probabilities: [[0.6, 0.1, 0.3], [0.1, 0.8, 0.1], [0.1, 0.2, 0.7]][index],
      confidence: [0.6, 0.8, 0.7][index], margin: [0.3, 0.7, 0.5][index],
      unknown_reason: index === 0 ? "low_confidence" : "" })) } };
}

const settle = () => new Promise(resolve => setImmediate(resolve));

async function startDemo(maxBytes = 4096) {
  const demo = createDemo();
  demo.statusRequest.respond(statusFixture(maxBytes));
  await settle();
  assert.equal(demo.phase, "idle");
  return demo;
}

function displaySnapshot(demo) {
  return { phase: demo.phase, busy: demo.elements[".results"].getAttribute("aria-busy"),
    model: demo.elements["#model-details"].textContent,
    announcement: demo.elements["#result-announcement"].textContent,
    inference: demo.elements["#inference-details"].textContent,
    cards: demo.cards.map(card => ({ state: card.dataset.state,
      decision: card.querySelector(".decision-value").textContent,
      candidate: card.querySelector(".raw-candidate strong").textContent,
      rows: card.querySelectorAll(".probability-row").map(row => ({ value: row.querySelector("meter").value,
        text: row.querySelector(".probability-value").textContent, winner: row.dataset.winner })) })) };
}

function assertReset(demo) {
  assert.equal(demo.phase, "idle");
  assert.equal(demo.elements[".results"].getAttribute("aria-busy"), "false");
  assert.equal(demo.elements["#result-announcement"].textContent, "");
  assert.equal(demo.elements["#inference-details"].textContent, "실제 모델 · 입력 대기");
  for (const card of demo.cards) {
    assert.equal(card.dataset.state, "empty");
    assert.equal(card.querySelector(".decision-value").textContent, "입력 대기");
    assert.equal(card.querySelector(".raw-candidate strong").textContent, "—");
    assert.equal(card.querySelector(".candidate-confidence").textContent, "");
    for (const row of card.querySelectorAll(".probability-row")) {
      assert.equal(row.querySelector("meter").value, 0);
      assert.equal(row.querySelector(".probability-value").textContent, "—");
      assert.equal(row.dataset.winner, undefined);
    }
  }
}

test("rapid input sends only the final whole value after the quiet period", async () => {
  const demo = await startDemo();
  demo.input("Original first sample.");
  demo.clock.advance(179);
  assert.equal(demo.hints.length, 0);
  demo.input("Original final sample.");
  demo.clock.advance(179);
  assert.equal(demo.hints.length, 0);
  assert.equal(demo.phase, "waiting");
  demo.clock.advance(1);
  assert.equal(demo.hints.length, 1);
  const { options } = demo.hints[0];
  assert.deepEqual(JSON.parse(options.body), { text: "Original final sample." });
  assert.equal(options.method, "POST");
  assert.equal(options.headers["Content-Type"], "application/json");
  assert.equal(options.cache, "no-store");
  assert.equal(demo.phase, "loading");
  assert.equal(demo.elements[".results"].getAttribute("aria-busy"), "true");
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 1);
});

test("editing aborts in-flight fetch and a late response cannot overwrite the newer result", async () => {
  const demo = await startDemo();
  demo.input("Original older sample.");
  demo.clock.advance(180);
  const older = demo.hints[0];
  demo.input("Original newer sample.");
  assert.equal(older.options.signal.aborted, true);
  demo.clock.advance(180);
  demo.hints[1].respond(predictionFixture(["true", "false", "unknown"]));
  await settle();
  assert.equal(demo.phase, "ready");
  const current = displaySnapshot(demo);
  const stale = predictionFixture(["false", "true", "true"]);
  stale.model.name = "Older response fixture";
  older.respond(stale);
  await settle();
  assert.equal(older.jsonReads, 0, "stale fetch must be discarded before decoding its body");
  assert.deepEqual(displaySnapshot(demo), current);
});

test("a response already decoding JSON cannot overwrite a later input", async () => {
  const demo = await startDemo();
  demo.input("Original slow-body sample.");
  demo.clock.advance(180);
  const older = demo.hints[0];
  older.resolveHeaders();
  await settle();
  assert.equal(older.jsonReads, 1);
  demo.input("Original quick-body sample.");
  assert.equal(older.options.signal.aborted, true);
  demo.clock.advance(180);
  demo.hints[1].respond(predictionFixture(["true", "false", "unknown"]));
  await settle();
  const current = displaySnapshot(demo);
  older.resolveBody(predictionFixture(["false", "true", "true"]));
  await settle();
  assert.deepEqual(displaySnapshot(demo), current);
});

test("IME composition suppresses intermediate requests and analyzes the committed text once", async () => {
  const demo = await startDemo();
  const input = demo.elements["#claim-text"];
  input.dispatch("compositionstart");
  demo.input("ㅎ");
  demo.clock.advance(1000);
  demo.input("한");
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 0);
  assert.equal(demo.elements["#run-status-text"].textContent, "한글 입력 중");
  input.dispatch("compositionend");
  // Browsers may emit a final input event after compositionend.
  demo.input("한");
  demo.clock.advance(179);
  assert.equal(demo.hints.length, 0);
  demo.clock.advance(1);
  assert.equal(demo.hints.length, 1);
  assert.deepEqual(JSON.parse(demo.hints[0].options.body), { text: "한" });
});

test("UTF-8 byte limits accept the exact boundary and reject whole oversized text", async () => {
  const demo = await startDemo(7);
  demo.input("한😀"); // 3 + 4 UTF-8 bytes; three UTF-16 code units.
  assert.equal(demo.elements["#byte-counter"].textContent, "7 / 7 bytes");
  demo.clock.advance(180);
  assert.deepEqual(JSON.parse(demo.hints[0].options.body), { text: "한😀" });
  demo.input("한😀x");
  assert.equal(demo.hints[0].options.signal.aborted, true);
  assert.equal(demo.elements["#byte-counter"].textContent, "8 / 7 bytes");
  assert.equal(demo.elements["#byte-counter"].dataset.overLimit, "true");
  assert.equal(demo.elements["#claim-text"].getAttribute("aria-invalid"), "true");
  assert.equal(demo.elements["#input-error"].hidden, false);
  assert.equal(demo.phase, "error");
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 1, "oversized text must not be truncated and sent");
  demo.input("한");
  assert.equal(demo.elements["#claim-text"].getAttribute("aria-invalid"), null);
  assert.equal(demo.elements["#input-error"].hidden, true);
  demo.clock.advance(180);
  assert.deepEqual(JSON.parse(demo.hints[1].options.body), { text: "한" });
});

test("a late server limit invalidates an oversized request already in progress", async () => {
  const demo = createDemo();
  demo.input("한😀x");
  demo.clock.advance(180);
  const pending = demo.hints[0];
  demo.statusRequest.respond(statusFixture(7));
  await settle();
  assert.equal(pending.options.signal.aborted, true);
  assert.equal(demo.phase, "error");
  assert.equal(demo.elements["#byte-counter"].textContent, "8 / 7 bytes");
  pending.respond(predictionFixture());
  await settle();
  assert.equal(demo.phase, "error");
  assert.equal(pending.jsonReads, 0);
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 1);
});

test("clear removes the displayed result, cancels pending work and focuses the empty input", async () => {
  const demo = await startDemo();
  demo.input("Original displayed sample.");
  demo.clock.advance(180);
  demo.hints[0].respond(predictionFixture());
  await settle();
  assert.equal(demo.phase, "ready");
  demo.elements["#clear-text"].dispatch("click");
  assertReset(demo);
  demo.input("Original pending sample.");
  demo.clock.advance(180);
  const pending = demo.hints[1];
  demo.elements["#clear-text"].dispatch("click");
  assert.equal(pending.options.signal.aborted, true);
  assert.equal(demo.elements["#claim-text"].value, "");
  assert.equal(demo.elements["#claim-text"].focused, true);
  assert.equal(demo.elements["#clear-text"].disabled, true);
  assert.equal(demo.elements["#byte-counter"].textContent, "0 / 4,096 bytes");
  assertReset(demo);
  pending.respond(predictionFixture());
  await settle();
  assertReset(demo);
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 2);
});

test("whitespace clears previous probabilities without calling inference", async () => {
  const demo = await startDemo();
  demo.input("Original probability sample.");
  demo.clock.advance(180);
  demo.hints[0].respond(predictionFixture());
  await settle();
  demo.input("\u2003\u3000\t");
  assertReset(demo);
  demo.clock.advance(1000);
  assert.equal(demo.hints.length, 1);
});

test("fixture rendering keeps raw candidates separate from final abstention", async () => {
  const demo = await startDemo();
  demo.input("Original rendering sample.");
  demo.clock.advance(180);
  // Reverse response order to exercise head identifiers, not array positions.
  const response = predictionFixture();
  response.prediction.heads.reverse();
  demo.hints[0].respond(response);
  await settle();
  const [requested, activity, completion] = demo.cards;
  assert.equal(requested.dataset.state, "unknown");
  assert.equal(requested.querySelector(".decision-value").textContent, "판정 보류");
  assert.equal(requested.querySelector(".raw-candidate strong").textContent, "있음");
  assert.equal(requested.querySelector(".candidate-confidence").textContent, " · 60.0%");
  assert.match(requested.querySelector(".decision-note").textContent, /확신도/);
  const trueRow = requested.querySelector('[data-class="true"]');
  assert.equal(trueRow.querySelector("meter").value, 0.6);
  assert.equal(trueRow.querySelector(".probability-value").textContent, "60.0%");
  assert.equal(trueRow.dataset.winner, "true");
  assert.equal(requested.querySelector('[data-class="unknown"]').dataset.winner, "false");
  assert.equal(activity.querySelector(".decision-value").textContent, "없음");
  assert.equal(completion.querySelector(".raw-candidate strong").textContent, "불명");
  assert.equal(demo.elements[".results"].getAttribute("aria-busy"), "false");
  assert.match(demo.elements["#model-details"].textContent, /fixture; not a research model/);
  assert.match(demo.elements["#result-announcement"].textContent, /응답 요청: 판정 보류/);
  assert.equal(demo.elements["#inference-details"].textContent, "실제 모델 · 1.2 ms · 7 bytes");
});

test("an invalid response clears prior results instead of retaining a valid-looking decision", async () => {
  const demo = await startDemo();
  demo.input("Original valid sample.");
  demo.clock.advance(180);
  demo.hints[0].respond(predictionFixture());
  await settle();
  demo.input("Original invalid-response sample.");
  demo.clock.advance(180);
  const invalid = predictionFixture();
  invalid.prediction.heads[0].probabilities[0] = 1.1;
  demo.hints[1].respond(invalid);
  await settle();
  assert.equal(demo.phase, "error");
  assert.equal(demo.elements[".results"].getAttribute("aria-busy"), "false");
  assert.equal(demo.elements["#input-error"].hidden, false);
  assert.match(demo.elements["#input-error"].textContent, /형식/);
  for (const card of demo.cards) {
    assert.equal(card.dataset.state, "empty");
    assert.equal(card.querySelector(".decision-value").textContent, "분석 불가");
    assert.equal(card.querySelector(".raw-candidate strong").textContent, "—");
  }
});
