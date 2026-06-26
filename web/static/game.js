// Front-end for War. It owns NO game logic — every decision is made by the Go
// server. This file only: (1) asks the server for the next step, (2) draws the
// result on the felt. The Go engine remains the single source of truth.

const SUIT_SYMBOL = { Spades: "♠", Clubs: "♣", Hearts: "♥", Diamonds: "♦" };
const RED_SUITS = new Set(["Hearts", "Diamonds"]);

const els = {
  countYou: document.getElementById("count-you"),
  countCpu: document.getElementById("count-cpu"),
  cardYou: document.getElementById("card-you"),
  cardCpu: document.getElementById("card-cpu"),
  flipBtn: document.getElementById("flip-btn"),
  message: document.getElementById("message"),
  newGame: document.getElementById("new-game"),
  felt: document.querySelector(".felt"),
};

let gameOver = false;

// ── API calls ─────────────────────────────────────────
// fetch() returns a Promise; await unwraps it. These two tiny functions are
// the entire contract with the Go server.
async function apiNew() {
  const res = await fetch("/api/new", { method: "POST" });
  return res.json();
}
async function apiStep() {
  const res = await fetch("/api/step", { method: "POST" });
  return res.json();
}

// ── Rendering ─────────────────────────────────────────
function cardHTML(card) {
  if (!card) return '<div class="card back"></div>';
  const sym = SUIT_SYMBOL[card.suit] ?? "?";
  const red = RED_SUITS.has(card.suit) ? "red" : "";
  const rank = faceName(card.rank);
  return `<div class="card revealed ${red}">
            <span class="rank">${rank}</span><span class="suit">${sym}</span>
          </div>`;
}

function faceName(rank) {
  return { 11: "J", 12: "Q", 13: "K", 14: "A" }[rank] ?? String(rank);
}

function render(result, { youCard = true } = {}) {
  els.countYou.textContent = result.count1;
  els.countCpu.textContent = result.count2;
  els.cardCpu.innerHTML = cardHTML(result.card2);

  // For "your" slot we re-create the flip button after a resolved round so it
  // stays clickable; on reveal we show the face.
  els.cardYou.innerHTML = cardHTML(result.card1);

  els.message.textContent = result.message;
  els.message.classList.toggle("win", result.outcome === "GameOver");

  if (result.outcome === "WarStart") {
    els.felt.classList.remove("war");
    void els.felt.offsetWidth; // reflow so the animation can re-trigger
    els.felt.classList.add("war");
  }

  if (result.outcome === "GameOver") {
    gameOver = true;
    showFlipButton(false);
  }
}

// Swap the "your" slot back to a clickable face-down card for the next round.
function armNextRound() {
  if (gameOver) return;
  els.cardYou.innerHTML =
    '<button class="card back clickable" id="flip-btn" aria-label="Flip your card"></button>';
  document.getElementById("flip-btn").addEventListener("click", onFlip);
}

function showFlipButton(show) {
  const btn = document.getElementById("flip-btn");
  if (btn) btn.disabled = !show;
}

// ── Event handlers ────────────────────────────────────
async function onFlip() {
  if (gameOver) return;
  const result = await apiStep();
  render(result);
  // After a decided round (or a war step), re-arm the clickable card so the
  // player can play the next one. War steps also re-arm (you click through).
  if (result.outcome !== "GameOver") {
    setTimeout(armNextRound, 250);
  }
}

async function onNewGame() {
  const result = await apiNew();
  gameOver = false;
  els.felt.classList.remove("war");
  els.cardCpu.innerHTML = '<div class="card back"></div>';
  armNextRound();
  els.countYou.textContent = result.count1;
  els.countCpu.textContent = result.count2;
  els.message.textContent = "Click your card to play the first round.";
  els.message.classList.remove("win");
}

// ── Wire up ───────────────────────────────────────────
els.flipBtn.addEventListener("click", onFlip);
els.newGame.addEventListener("click", onNewGame);

// Start a fresh game on load.
onNewGame();
