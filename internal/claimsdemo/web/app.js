"use strict";

(() => {
  const SCHEMA = "riidolaya-live-claims-v1";
  const DEFAULT_MAX_BYTES = 4096;
  const DEBOUNCE_MS = 180;
  const classes = ["true", "false", "unknown"];
  const classLabels = { true: "있음", false: "없음", unknown: "불명" };
  const decisionLabels = { true: "있음", false: "없음", unknown: "판정 보류" };
  const examples = {
    request: "수정한 목차가 자연스러운지 한 번 봐줄 수 있어?",
    activity: "지금 두 가지 화면 구성을 나란히 놓고 간격을 비교하고 있어.",
    completion: "소개 페이지의 문구를 다듬고 최종 파일을 저장했어.",
  };
  const textInput = document.querySelector("#claim-text");
  const byteCounter = document.querySelector("#byte-counter");
  const inputError = document.querySelector("#input-error");
  const clearButton = document.querySelector("#clear-text");
  const results = document.querySelector(".results");
  const status = document.querySelector("#run-status");
  const statusText = document.querySelector("#run-status-text");
  const announcement = document.querySelector("#result-announcement");
  const inferenceDetails = document.querySelector("#inference-details");
  const modelDetails = document.querySelector("#model-details");
  const cards = [...document.querySelectorAll("[data-head]")];
  const encoder = new TextEncoder();
  const numberFormat = new Intl.NumberFormat("ko-KR");
  let maxBytes = DEFAULT_MAX_BYTES;
  let sequence = 0;
  let timer = null;
  let activeRequest = null;
  let composing = false;

  function setStatus(phase, message, busy = false) {
    status.dataset.phase = phase;
    statusText.textContent = message;
    results.setAttribute("aria-busy", String(busy));
  }

  function resetCards(label = "입력 대기", note = "문장을 입력하면 실제 모델 결과가 나타납니다.") {
    for (const card of cards) {
      card.dataset.state = "empty";
      card.querySelector(".decision-value").textContent = label;
      card.querySelector(".raw-candidate strong").textContent = "—";
      card.querySelector(".candidate-confidence").textContent = "";
      card.querySelector(".decision-note").textContent = note;
      for (const row of card.querySelectorAll(".probability-row")) {
        row.querySelector("meter").value = 0;
        row.querySelector(".probability-value").textContent = "—";
        row.removeAttribute("data-winner");
      }
    }
  }

  function invalidateRequest() {
    sequence += 1;
    window.clearTimeout(timer);
    timer = null;
    if (activeRequest) activeRequest.abort();
    activeRequest = null;
    return sequence;
  }

  function updateInputState() {
    const bytes = encoder.encode(textInput.value).length;
    byteCounter.textContent = `${numberFormat.format(bytes)} / ${numberFormat.format(maxBytes)} bytes`;
    byteCounter.dataset.overLimit = String(bytes > maxBytes);
    clearButton.disabled = textInput.value.length === 0;
    inputError.hidden = true;
    inputError.textContent = "";
    textInput.removeAttribute("aria-invalid");
    return bytes;
  }

  function showError(message, inputInvalid = false) {
    resetCards("분석 불가", "문제를 해결한 뒤 다시 입력해보세요.");
    setStatus("error", "분석할 수 없어요");
    inputError.textContent = message;
    inputError.hidden = false;
    if (inputInvalid) textInput.setAttribute("aria-invalid", "true");
    inferenceDetails.textContent = "추론 결과 없음";
    announcement.textContent = message;
  }

  function handleInput() {
    const currentSequence = invalidateRequest();
    const bytes = updateInputState();
    announcement.textContent = "";
    if (bytes > maxBytes) {
      showError(`입력이 ${numberFormat.format(maxBytes)} bytes를 넘었습니다. 문장을 줄여주세요. 입력을 잘라 보내지 않습니다.`, true);
      return;
    }
    if (!textInput.value.trim()) {
      resetCards();
      setStatus("idle", "문장을 기다리고 있어요");
      inferenceDetails.textContent = "실제 모델 · 입력 대기";
      return;
    }
    resetCards("분석 대기", composing ? "한글 입력을 마치면 분석합니다." : "새 문장을 실제 모델로 분석합니다.");
    inferenceDetails.textContent = "새 입력 · 분석 대기";
    setStatus("waiting", composing ? "한글 입력 중" : "입력 반영 중", true);
    if (!composing) {
      const text = textInput.value;
      timer = window.setTimeout(() => infer(text, currentSequence), DEBOUNCE_MS);
    }
  }

  function validHead(head) {
    return head && classes.includes(head.state) && classes.includes(head.winner)
      && Array.isArray(head.probabilities) && head.probabilities.length === 3
      && head.probabilities.every(value => Number.isFinite(value) && value >= 0 && value <= 1)
      && Number.isFinite(head.confidence) && head.confidence >= 0 && head.confidence <= 1
      && Number.isFinite(head.margin) && head.margin >= 0 && head.margin <= 1;
  }

  function unknownReason(head) {
    const reason = String(head.unknown_reason || "");
    if (reason.includes("confidence")) return "확신도가 기준에 못 미쳐 판정을 보류했습니다.";
    if (reason.includes("margin")) return "후보 점수 차이가 작아 판정을 보류했습니다.";
    if (head.winner === "unknown") return "모델의 최상위 후보가 불명이라 판정을 보류했습니다.";
    return "판정 기준을 충족하지 못해 판정을 보류했습니다.";
  }

  function renderModel(model) {
    if (!model || typeof model.name !== "string") return;
    const steps = Number.isSafeInteger(model.training_steps) && model.training_steps >= 0
      ? ` · ${numberFormat.format(model.training_steps)} 학습 스텝` : "";
    modelDetails.textContent = `${model.name} · 로컬 추론${steps}`;
  }

  function renderPrediction(data) {
    if (data.schema !== SCHEMA) throw new Error("invalid_response");
    renderModel(data.model);
    if (data.state === "empty" && data.prediction === null) {
      resetCards();
      setStatus("idle", "문장을 기다리고 있어요");
      inferenceDetails.textContent = "실제 모델 · 입력 대기";
      return;
    }
    if (data.state !== "ready" || !Array.isArray(data.prediction?.heads)) throw new Error("invalid_response");
    const heads = cards.map(card => data.prediction.heads.find(head => head.head === card.dataset.head));
    if (!heads.every(validHead)) throw new Error("invalid_response");
    const summary = [];
    cards.forEach((card, index) => {
      const head = heads[index];
      card.dataset.state = head.state;
      card.querySelector(".decision-value").textContent = decisionLabels[head.state];
      card.querySelector(".raw-candidate strong").textContent = classLabels[head.winner];
      card.querySelector(".candidate-confidence").textContent = ` · ${(head.confidence * 100).toFixed(1)}%`;
      card.querySelector(".decision-note").textContent = head.state === "unknown" ? unknownReason(head)
        : head.state === "true" ? "이 신호가 있다고 분류했습니다." : "이 신호가 없다고 분류했습니다.";
      for (const [classIndex, value] of head.probabilities.entries()) {
        const row = card.querySelector(`[data-class="${classes[classIndex]}"]`);
        row.querySelector("meter").value = value;
        row.querySelector(".probability-value").textContent = `${(value * 100).toFixed(1)}%`;
        row.dataset.winner = String(head.winner === classes[classIndex]);
      }
      summary.push(`${card.querySelector("h3").textContent}: ${decisionLabels[head.state]}`);
    });
    const duration = Number.isFinite(data.inference_us) && data.inference_us >= 0
      ? data.inference_us < 1000 ? `${numberFormat.format(Math.round(data.inference_us))} µs`
        : `${(data.inference_us / 1000).toFixed(1)} ms`
      : "시간 미제공";
    const bytes = Number.isSafeInteger(data.input_bytes) && data.input_bytes >= 0
      ? `${numberFormat.format(data.input_bytes)} bytes` : "";
    inferenceDetails.textContent = `실제 모델 · ${duration}${bytes ? ` · ${bytes}` : ""}`;
    setStatus("ready", "새 문장 분석 완료");
    announcement.textContent = `분석 완료. ${summary.join(". ")}.`;
  }

  async function infer(text, currentSequence) {
    if (currentSequence !== sequence || composing) return;
    const controller = new AbortController();
    activeRequest = controller;
    timer = null;
    setStatus("loading", "실제 모델 분석 중", true);
    try {
      const response = await fetch("/api/hints", {
        method: "POST",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify({ text }),
        signal: controller.signal,
        cache: "no-store",
      });
      if (currentSequence !== sequence) return;
      const data = await response.json();
      if (currentSequence !== sequence) return;
      if (!response.ok) {
        const code = data.error?.code;
        if (response.status === 413 || code === "text_too_large" || code === "input_too_large") {
          showError(`입력이 허용 크기를 넘었습니다. ${numberFormat.format(maxBytes)} bytes 이내로 줄여주세요.`, true);
        } else {
          showError("모델 분석에 실패했습니다. 잠시 후 문장을 다시 입력해주세요.");
        }
        return;
      }
      renderPrediction(data);
    } catch (error) {
      if (currentSequence !== sequence || error.name === "AbortError") return;
      showError(error.message === "invalid_response"
        ? "모델 결과의 형식을 확인할 수 없습니다. 서버 연결을 확인해주세요."
        : "로컬 서버에 연결할 수 없습니다. 데모 서버가 실행 중인지 확인해주세요.");
    } finally {
      if (currentSequence === sequence) activeRequest = null;
    }
  }

  textInput.addEventListener("input", handleInput);
  textInput.addEventListener("compositionstart", () => {
    composing = true;
    handleInput();
  });
  textInput.addEventListener("compositionend", () => {
    composing = false;
    handleInput();
  });
  clearButton.addEventListener("click", () => {
    composing = false;
    textInput.value = "";
    handleInput();
    textInput.focus();
  });
  document.querySelectorAll("[data-example]").forEach(button => {
    button.addEventListener("click", () => {
      composing = false;
      textInput.value = examples[button.dataset.example];
      handleInput();
      textInput.focus();
      textInput.setSelectionRange(textInput.value.length, textInput.value.length);
    });
  });

  async function loadStatus() {
    const initialSequence = sequence;
    try {
      const response = await fetch("/api/status", { headers: { "Accept": "application/json" }, cache: "no-store" });
      const data = await response.json();
      if (!response.ok || data.schema !== SCHEMA) throw new Error("status_unavailable");
      renderModel(data.model);
      if (Number.isSafeInteger(data.limits?.max_text_bytes) && data.limits.max_text_bytes > 0) {
        maxBytes = Math.min(DEFAULT_MAX_BYTES, data.limits.max_text_bytes);
        if (sequence === initialSequence) updateInputState();
        else if (encoder.encode(textInput.value).length > maxBytes) handleInput();
        else byteCounter.textContent = `${numberFormat.format(encoder.encode(textInput.value).length)} / ${numberFormat.format(maxBytes)} bytes`;
      }
      if (sequence === initialSequence) setStatus("idle", "문장을 기다리고 있어요");
    } catch {
      if (sequence === initialSequence) setStatus("error", "모델 연결을 확인해주세요");
    }
  }

  updateInputState();
  loadStatus();
})();
