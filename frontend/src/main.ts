import "./style.css";
import pwpdfLogo from "./assets/pwpdflogo.png";
import soundtrackMp3Asset from "./assets/unrealsuperhero.mp3";
import soundtrackOggAsset from "./assets/unrealsuperhero.ogg";

import {
  EncryptPDF,
  GetApplicationState,
  SelectPDF,
  SetInputFile,
} from "../wailsjs/go/gui/App";
import {
  EventsOn,
  OnFileDrop,
  WindowSetMinSize,
  WindowSetSize,
} from "../wailsjs/runtime/runtime";

type PreparedDocument = {
  inputPath: string;
  fileName: string;
  defaultOutputPath: string;
  defaultOutputFilename: string;
  directory: string;
};

type ApplicationState = {
  version: string;
  pendingSelection?: PreparedDocument;
  pendingError?: string;
};

type OperationResult = {
  cancelled: boolean;
  outputPath?: string;
  outputFilename?: string;
  message?: string;
};

type AppState = {
  version: string;
  selectedDocument: PreparedDocument | null;
  password: string;
  passwordConfirmation: string;
  busy: boolean;
  error: string;
  success: OperationResult | null;
};

type SoundtrackSource = {
  src: string;
  type: string;
  label: string;
};

type LayoutMode = "default" | "compact" | "tight";

type Dimensions = {
  width: number;
  height: number;
};

const appState: AppState = {
  version: "",
  selectedDocument: null,
  password: "",
  passwordConfirmation: "",
  busy: false,
  error: "",
  success: null,
};

const soundtrackSources: SoundtrackSource[] = [
  { src: soundtrackMp3Asset, type: "audio/mpeg", label: "mp3" },
  { src: soundtrackOggAsset, type: 'audio/ogg; codecs="vorbis"', label: "ogg" },
];
const layoutModes: LayoutMode[] = ["default", "compact", "tight"];
const soundtrackProbe = document.createElement("audio");
const soundtrackAudio = new Audio();
soundtrackAudio.loop = true;
soundtrackAudio.preload = "none";

let soundtrackState: "off" | "on" | "error" = "off";
let soundtrackSourceIndex = resolvePreferredSoundtrackSourceIndex();
let soundtrackRecoveryInFlight = false;
let pendingLayoutSyncFrame: number | null = null;
let pendingWindowResizeFrame: number | null = null;
let lastWindowResizeSignature = "";
let currentUIScale = 1;

const appElement = document.querySelector<HTMLDivElement>("#app");
if (!appElement) {
  throw new Error("The app root could not be found.");
}
const appRoot = appElement;

appRoot.innerHTML = `
  <main class="shell">
    <section class="shell-chrome" aria-hidden="true">
      <div class="shell-chrome__lights">
        <span></span>
        <span></span>
        <span></span>
      </div>
      <img class="hero__logo" width="400px" src="${pwpdfLogo}" alt="pwpdf logo" />
      <div class="shell-chrome__label"></div>
    </section>

    <section class="hero">
      <div class="hero__topline">
        <div class="version-badge" id="version-badge">v0.1</div>
      </div>
      <h1><span class="hero__accent">Lock</span> the PDF. Keep the pixels.</h1>
      <p>Select or drop a file, enter a password twice, and save the encrypted copy without changing the document layout.</p>
      <div class="hero__chips">
        <div class="hero__chip-list" aria-hidden="true">
          <span class="hero-chip">AES-256</span>
          <span class="hero-chip">CLI + GUI</span>
          <span class="hero-chip">NO RASTERIZE</span>
        </div>
        <button
          class="hero__sound-toggle"
          id="sound-toggle"
          type="button"
          aria-pressed="false"
          data-state="off"
        >
          Sound Off
        </button>
      </div>
    </section>

    <section class="panel drop-panel" id="drop-panel">
      <div class="drop-panel__label">Drop a PDF here</div>
      <button class="primary-button" id="select-button" type="button">Select PDF</button>
    </section>

    <section class="panel details-panel" id="details-panel" hidden>
      <div class="panel-kicker">Input File</div>
      <div class="details-head">
        <div>
          <div class="details-label">Selected File</div>
          <div class="details-file" id="details-file"></div>
        </div>
        <button class="ghost-button" id="reset-button" type="button">Change PDF</button>
      </div>
      <div class="details-path" id="details-path"></div>
      <div class="details-output">
        <span>Default output:</span>
        <strong id="details-output"></strong>
      </div>
    </section>

    <section class="panel form-panel" id="form-panel" hidden>
      <div class="panel-kicker">Encryption Key</div>
      <label class="field">
        <span>Password</span>
        <input id="password-input" type="password" autocomplete="new-password" />
      </label>
      <label class="field">
        <span>Confirm Password</span>
        <input id="confirmation-input" type="password" autocomplete="new-password" />
      </label>
      <button class="primary-button wide-button" id="encrypt-button" type="button">Encrypt PDF</button>
    </section>

    <section class="status-panel" id="status-panel" hidden>
      <div class="status-panel__title" id="status-title"></div>
      <div class="status-panel__body" id="status-body"></div>
    </section>
  </main>
`;

const versionBadge = queryElement<HTMLDivElement>("#version-badge");
const shellElement = queryElement<HTMLElement>(".shell");
const dropPanel = queryElement<HTMLDivElement>("#drop-panel");
const selectButton = queryElement<HTMLButtonElement>("#select-button");
const soundToggleButton = queryElement<HTMLButtonElement>("#sound-toggle");
const detailsPanel = queryElement<HTMLDivElement>("#details-panel");
const detailsFile = queryElement<HTMLDivElement>("#details-file");
const detailsPath = queryElement<HTMLDivElement>("#details-path");
const detailsOutput = queryElement<HTMLDivElement>("#details-output");
const resetButton = queryElement<HTMLButtonElement>("#reset-button");
const formPanel = queryElement<HTMLDivElement>("#form-panel");
const passwordInput = queryElement<HTMLInputElement>("#password-input");
const confirmationInput = queryElement<HTMLInputElement>("#confirmation-input");
const encryptButton = queryElement<HTMLButtonElement>("#encrypt-button");
const statusPanel = queryElement<HTMLDivElement>("#status-panel");
const statusTitle = queryElement<HTMLDivElement>("#status-title");
const statusBody = queryElement<HTMLDivElement>("#status-body");

selectButton.addEventListener("click", (event) => {
  event.stopPropagation();
  void handleSelectPDF();
});

soundToggleButton.addEventListener("click", () => {
  void toggleSoundtrack();
});

dropPanel.addEventListener("click", () => {
  void handleSelectPDF();
});

resetButton.addEventListener("click", () => {
  void handleSelectPDF();
});

passwordInput.addEventListener("input", (event) => {
  appState.password = (event.target as HTMLInputElement).value;
  render();
});

confirmationInput.addEventListener("input", (event) => {
  appState.passwordConfirmation = (event.target as HTMLInputElement).value;
  render();
});

encryptButton.addEventListener("click", () => {
  void handleEncryptPDF();
});

EventsOn("app:prepared-document", (documentSelection: PreparedDocument) => {
  applySelection(documentSelection);
});

EventsOn("app:error-message", (message: string) => {
  appState.error = message;
  appState.success = null;
  render();
});

OnFileDrop((_x: number, _y: number, paths: string[]) => {
  const firstPath = paths?.[0];
  if (firstPath) {
    void loadDocument(firstPath);
  }
}, true);

window.addEventListener("resize", () => {
  scheduleLayoutSync();
});

soundtrackAudio.addEventListener("error", () => {
  if (soundtrackState !== "on" || soundtrackRecoveryInFlight) {
    soundtrackState = "error";
    updateSoundToggleButton();
    return;
  }

  soundtrackRecoveryInFlight = true;
  void recoverSoundtrackAfterError();
});

updateSoundToggleButton();

void initialise();

async function initialise(): Promise<void> {
  try {
    const initialState = (await GetApplicationState()) as ApplicationState;
    appState.version = initialState.version ?? "dev";

    if (initialState.pendingSelection) {
      applySelection(initialState.pendingSelection);
    }
    if (initialState.pendingError) {
      appState.error = initialState.pendingError;
      appState.success = null;
    }
  } catch (error) {
    appState.error = normaliseError(error);
  }

  render();
}

async function handleSelectPDF(): Promise<void> {
  if (appState.busy) {
    return;
  }

  appState.busy = true;
  appState.error = "";
  appState.success = null;
  render();

  try {
    const selectedDocument = (await SelectPDF()) as PreparedDocument | null;
    if (selectedDocument) {
      applySelection(selectedDocument);
    }
  } catch (error) {
    appState.error = normaliseError(error);
  } finally {
    appState.busy = false;
    render();
  }
}

async function loadDocument(path: string): Promise<void> {
  appState.busy = true;
  appState.error = "";
  appState.success = null;
  render();

  try {
    const selectedDocument = (await SetInputFile(path)) as PreparedDocument;
    applySelection(selectedDocument);
  } catch (error) {
    appState.error = normaliseError(error);
  } finally {
    appState.busy = false;
    render();
  }
}

async function handleEncryptPDF(): Promise<void> {
  if (!appState.selectedDocument || appState.busy) {
    return;
  }

  appState.busy = true;
  appState.error = "";
  appState.success = null;
  render();

  try {
    const result = (await EncryptPDF({
      inputPath: appState.selectedDocument.inputPath,
      password: appState.password,
      passwordConfirmation: appState.passwordConfirmation,
    })) as OperationResult;

    if (!result.cancelled) {
      appState.success = result;
      appState.password = "";
      appState.passwordConfirmation = "";
    }
  } catch (error) {
    appState.error = normaliseError(error);
  } finally {
    appState.busy = false;
    render();
  }
}

function applySelection(selection: PreparedDocument): void {
  appState.selectedDocument = selection;
  appState.password = "";
  appState.passwordConfirmation = "";
  appState.error = "";
  appState.success = null;
  render();
}

function render(): void {
  const versionLabel = `v${appState.version}`;
  if (versionBadge.textContent !== versionLabel) {
    versionBadge.textContent = versionLabel;
  }

  shellElement.dataset.stage = appState.selectedDocument ? "selected" : "idle";
  dropPanel.hidden = !!appState.selectedDocument;
  detailsPanel.hidden = !appState.selectedDocument;
  formPanel.hidden = !appState.selectedDocument;

  if (passwordInput.value !== appState.password) {
    passwordInput.value = appState.password;
  }
  if (confirmationInput.value !== appState.passwordConfirmation) {
    confirmationInput.value = appState.passwordConfirmation;
  }

  if (appState.selectedDocument) {
    detailsFile.textContent = appState.selectedDocument.fileName;
    detailsPath.textContent = appState.selectedDocument.inputPath;
    detailsPath.title = appState.selectedDocument.inputPath;
    detailsOutput.textContent = appState.selectedDocument.defaultOutputFilename;
    detailsOutput.title = appState.selectedDocument.defaultOutputFilename;
  } else {
    detailsFile.textContent = "";
    detailsPath.textContent = "";
    detailsPath.title = "";
    detailsOutput.textContent = "";
    detailsOutput.title = "";
  }

  const canEncrypt =
    !!appState.selectedDocument &&
    !!appState.password &&
    !!appState.passwordConfirmation &&
    !appState.busy;

  selectButton.disabled = appState.busy;
  resetButton.disabled = appState.busy;
  passwordInput.disabled = appState.busy || !appState.selectedDocument;
  confirmationInput.disabled = appState.busy || !appState.selectedDocument;
  encryptButton.disabled = !canEncrypt;
  encryptButton.textContent = appState.busy ? "Encrypting..." : "Encrypt PDF";

  if (appState.error) {
    statusPanel.hidden = false;
    statusPanel.dataset.variant = "error";
    statusTitle.textContent = "Something needs attention";
    statusBody.textContent = appState.error;
    statusBody.title = appState.error;
    scheduleWindowResizeIfNeeded();
    return;
  }

  if (appState.success) {
    statusPanel.hidden = false;
    statusPanel.dataset.variant = "success";
    statusTitle.textContent = appState.success.message ?? "Encrypted PDF saved successfully.";
    statusBody.textContent = appState.success.outputPath ?? "";
    statusBody.title = appState.success.outputPath ?? "";
    scheduleWindowResizeIfNeeded();
    return;
  }

  statusPanel.hidden = true;
  statusPanel.dataset.variant = "";
  statusTitle.textContent = "";
  statusBody.textContent = "";
  statusBody.title = "";
  scheduleWindowResizeIfNeeded();
}

async function toggleSoundtrack(): Promise<void> {
  if (soundtrackState === "on") {
    soundtrackAudio.pause();
    soundtrackAudio.currentTime = 0;
    soundtrackState = "off";
    updateSoundToggleButton();
    return;
  }

  soundtrackAudio.volume = 0.55;

  const started = await tryStartSoundtrack(soundtrackSourceIndex);
  if (started) {
    soundtrackState = "on";
  } else {
    soundtrackState = "error";
  }

  updateSoundToggleButton();
}

function updateSoundToggleButton(): void {
  soundToggleButton.dataset.state = soundtrackState;
  soundToggleButton.setAttribute("aria-pressed", soundtrackState === "on" ? "true" : "false");

  if (soundtrackState === "on") {
    soundToggleButton.textContent = "Sound On";
    soundToggleButton.title = "Stop soundtrack";
    return;
  }

  if (soundtrackState === "error") {
    soundToggleButton.textContent = "Sound Error";
    soundToggleButton.title = "The runtime could not play the bundled soundtrack";
    return;
  }

  soundToggleButton.textContent = "Sound Off";
  soundToggleButton.title = "Start soundtrack";
}

function resolvePreferredSoundtrackSourceIndex(): number {
  const preferredIndex = soundtrackSources.findIndex((source) => soundtrackProbe.canPlayType(source.type) !== "");
  return preferredIndex === -1 ? 0 : preferredIndex;
}

function setSoundtrackSource(index: number): void {
  soundtrackSourceIndex = index;
  soundtrackAudio.src = soundtrackSources[index].src;
  soundtrackAudio.load();
}

async function tryStartSoundtrack(startIndex: number): Promise<boolean> {
  const attemptedIndexes = new Set<number>();
  let lastError: unknown = null;

  for (let offset = 0; offset < soundtrackSources.length; offset += 1) {
    const index = (startIndex + offset) % soundtrackSources.length;
    if (attemptedIndexes.has(index)) {
      continue;
    }

    attemptedIndexes.add(index);
    setSoundtrackSource(index);

    try {
      await soundtrackAudio.play();
      return true;
    } catch (error) {
      lastError = error;
      console.error(`Unable to start soundtrack using ${soundtrackSources[index].label}.`, error);
    }
  }

  console.error("Unable to start the soundtrack in the current runtime.", lastError);
  return false;
}

async function recoverSoundtrackAfterError(): Promise<void> {
  const started = await tryStartSoundtrack(soundtrackSourceIndex + 1);
  soundtrackRecoveryInFlight = false;

  soundtrackState = started ? "on" : "error";
  updateSoundToggleButton();
}

function scheduleLayoutSync(): void {
  if (pendingLayoutSyncFrame !== null) {
    return;
  }

  pendingLayoutSyncFrame = window.requestAnimationFrame(() => {
    pendingLayoutSyncFrame = null;
    syncLayoutMode();
  });
}

function scheduleWindowResizeIfNeeded(): void {
  const nextSignature = buildWindowResizeSignature();
  if (nextSignature === lastWindowResizeSignature) {
    return;
  }

  lastWindowResizeSignature = nextSignature;
  scheduleWindowResize();
}

function buildWindowResizeSignature(): string {
  const selectionKey = appState.selectedDocument?.inputPath ?? "";

  if (appState.error) {
    return `${selectionKey}|error|${appState.error}`;
  }

  if (appState.success) {
    return `${selectionKey}|success|${appState.success.outputPath ?? appState.success.message ?? ""}`;
  }

  return `${selectionKey}|idle`;
}

function scheduleWindowResize(): void {
  if (pendingWindowResizeFrame !== null) {
    return;
  }

  pendingWindowResizeFrame = window.requestAnimationFrame(() => {
    pendingWindowResizeFrame = window.requestAnimationFrame(() => {
      pendingWindowResizeFrame = null;
      void syncWindowToContent();
    });
  });
}

async function syncWindowToContent(): Promise<void> {
  const screenBudget = readAvailableScreenSize();
  applyUIScale(resolveUIScale(screenBudget));
  const minimumWindowWidth = Math.ceil(480 * currentUIScale);
  const minimumWindowHeight = Math.ceil(500 * currentUIScale);
  const maxWindowWidth = Math.max(480, screenBudget.width - 48);
  const maxWindowHeight = Math.max(500, screenBudget.height - 72);

  try {
    WindowSetMinSize(minimumWindowWidth, minimumWindowHeight);
  } catch {
    // Ignore sizing calls when the runtime is not available outside Wails.
  }

  applyWindowFitMode(true);
  applyBestLayoutMode({
    width: maxWindowWidth,
    height: maxWindowHeight,
  });

  const shellSize = readElementSize(shellElement);
  const appPadding = readAppPadding();
  const desiredWidth = clamp(shellSize.width + appPadding.width + 8, minimumWindowWidth, maxWindowWidth);
  const desiredHeight = clamp(shellSize.height + appPadding.height + 18, minimumWindowHeight, maxWindowHeight);
  applyWindowFitMode(false);

  try {
    WindowSetSize(desiredWidth, desiredHeight);
  } catch {
    // Ignore sizing calls when the runtime is not available outside Wails.
  }
}

function syncLayoutMode(): void {
  const previousScale = currentUIScale;
  const previousLayoutMode = currentLayoutMode();
  applyUIScale(resolveUIScale(readAvailableScreenSize()));
  const viewport = readViewportSize();
  const hasSelection = appState.selectedDocument !== null;
  applyLayoutMode(pickLayoutMode(viewport, hasSelection));

  if (currentUIScale !== previousScale || currentLayoutMode() !== previousLayoutMode) {
    scheduleWindowResize();
  }
}

function applyBestLayoutMode(screenBudget: Dimensions): void {
  let fallbackLayoutMode: LayoutMode = "tight";

  for (const mode of layoutModes) {
    applyLayoutMode(mode);

    const shellSize = readElementSize(shellElement);
    const appPadding = readAppPadding();
    const requiredWidth = shellSize.width + appPadding.width + 8;
    const requiredHeight = shellSize.height + appPadding.height + 18;

    if (
      requiredWidth <= screenBudget.width &&
      requiredHeight <= screenBudget.height
    ) {
      return;
    }

    fallbackLayoutMode = mode;
  }

  applyLayoutMode(fallbackLayoutMode);
}

function pickLayoutMode(viewport: Dimensions, hasSelection: boolean): LayoutMode {
  if (hasSelection) {
    if (viewport.width <= 720 || viewport.height <= 560) {
      return "tight";
    }
    if (viewport.width <= 920 || viewport.height <= 700) {
      return "compact";
    }
    return "default";
  }

  if (viewport.width <= 560 || viewport.height <= 480) {
    return "tight";
  }
  if (viewport.width <= 680 || viewport.height <= 620) {
    return "compact";
  }
  return "default";
}

function applyLayoutMode(mode: LayoutMode): void {
  if (currentLayoutMode() === mode) {
    return;
  }

  if (mode === "default") {
    delete appRoot.dataset.layout;
    delete shellElement.dataset.layout;
    return;
  }

  appRoot.dataset.layout = mode;
  shellElement.dataset.layout = mode;
}

function currentLayoutMode(): LayoutMode {
  return (appRoot.dataset.layout as LayoutMode | undefined) ?? "default";
}

function applyWindowFitMode(enabled: boolean): void {
  if (enabled) {
    appRoot.dataset.windowFit = "true";
    return;
  }

  delete appRoot.dataset.windowFit;
}

function readViewportSize(): Dimensions {
  return {
    width: Math.max(window.innerWidth || 0, document.documentElement.clientWidth || 0),
    height: Math.max(window.innerHeight || 0, document.documentElement.clientHeight || 0),
  };
}

function readAvailableScreenSize(): Dimensions {
  return {
    width: window.screen.availWidth || window.screen.width || window.innerWidth || 1280,
    height: window.screen.availHeight || window.screen.height || window.innerHeight || 720,
  };
}

function readElementSize(element: HTMLElement): Dimensions {
  const bounds = element.getBoundingClientRect();
  const logicalWidth = Math.max(element.scrollWidth, element.clientWidth, element.offsetWidth);
  const logicalHeight = Math.max(element.scrollHeight, element.clientHeight, element.offsetHeight);

  if (currentUIScale !== 1) {
    return {
      width: Math.ceil(logicalWidth * currentUIScale),
      height: Math.ceil(logicalHeight * currentUIScale),
    };
  }

  return {
    width: Math.ceil(Math.max(logicalWidth, bounds.width)),
    height: Math.ceil(Math.max(logicalHeight, bounds.height)),
  };
}

function readAppPadding(): Dimensions {
  const appStyles = window.getComputedStyle(appRoot);

  return {
    width: Math.ceil(parseFloat(appStyles.paddingLeft) + parseFloat(appStyles.paddingRight)),
    height: Math.ceil(parseFloat(appStyles.paddingTop) + parseFloat(appStyles.paddingBottom)),
  };
}

function clamp(value: number, minimum: number, maximum: number): number {
  return Math.min(Math.max(value, minimum), maximum);
}

function resolveUIScale(screenSize: Dimensions): number {
  if (screenSize.width < 1920 || screenSize.height < 1080) {
    return 0.7;
  }

  return 1;
}

function applyUIScale(scale: number): void {
  if (currentUIScale === scale) {
    return;
  }

  currentUIScale = scale;
  appRoot.style.setProperty("--ui-scale", String(scale));
}

function queryElement<T extends Element>(selector: string): T {
  const element = document.querySelector<T>(selector);
  if (!element) {
    throw new Error(`Missing required element: ${selector}`);
  }
  return element;
}

function normaliseError(error: unknown): string {
  if (typeof error === "string") {
    return error;
  }

  if (error && typeof error === "object" && "message" in error) {
    return String((error as { message?: unknown }).message ?? "Unexpected error");
  }

  return "Unexpected error";
}
