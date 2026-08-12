// Rune VS Code extension: a thin client that launches `rune lsp` over stdio and
// wires it to Runefile documents. All language intelligence lives in the Rune
// binary (one parser, one analyzer, one formatter) — this file only starts the
// server and forwards requests.
const { workspace, window, commands, env, Uri } = require("vscode");
const { execFile } = require("node:child_process");
const { LanguageClient } = require("vscode-languageclient/node");

const INSTALL_DOCS_URL = "https://github.com/rune-task-runner/rune/blob/main/docs/installation.md";
const RESTART_DEBOUNCE_MS = 400;

/** @type {import('vscode-languageclient/node').LanguageClient | undefined} */
let client;

// All start/stop transitions run through this chain, one at a time — two rapid
// rune.path changes must not race two probes into two live clients.
let transitions = Promise.resolve();

function enqueueTransition(task) {
  transitions = transitions.then(task, task);
  // Observe failures so a rejected transition is not an unhandled rejection;
  // the chain itself keeps running (task is used as both handlers above).
  transitions.catch(() => {});
  return transitions;
}

// Probe the configured executable before handing it to the language client:
// vscode-languageclient's own failure mode for a missing command is a cryptic
// "server crashed 5 times" toast. Resolves to {kind: "ok"}, {kind: "missing"}
// (no such binary), {kind: "outdated"} (ran and exited non-zero — a Rune
// release without `lsp`), {kind: "unresponsive"} (hit the 5s probe timeout),
// or {kind: "unlaunchable", detail} (spawn failed for another reason, e.g.
// EACCES on a non-executable file or ENOTDIR on a bad path).
function probeServer(command) {
  return new Promise((resolve) => {
    execFile(command, ["lsp", "--help"], { timeout: 5000 }, (error) => {
      if (!error) return resolve({ kind: "ok" });
      if (error.code === "ENOENT") return resolve({ kind: "missing" });
      if (error.killed) return resolve({ kind: "unresponsive" });
      if (typeof error.code === "number") return resolve({ kind: "outdated" });
      return resolve({ kind: "unlaunchable", detail: String(error.code || error.message) });
    });
  });
}

// One toast at a time (contract L9): the settings editor commits intermediate
// values while the user types, and each failed re-probe would otherwise stack
// another notification.
let errorToastVisible = false;

async function explainUnusableBinary(command, probe) {
  if (errorToastVisible) return;
  errorToastVisible = true;
  try {
    const messages = {
      missing: `Rune binary not found ("${command}"). The Rune extension needs the rune executable on your PATH, or set "rune.path" to its location.`,
      outdated: `The Rune binary at "${command}" cannot run "rune lsp" — it is likely an older release. Please upgrade Rune, or point "rune.path" at a newer binary.`,
      unresponsive: `The Rune binary at "${command}" did not respond within 5 seconds. If the path is correct, retry once the machine settles (e.g. after a first-launch antivirus scan).`,
      unlaunchable: `The Rune binary at "${command}" could not be launched (${probe.detail}). Check that "rune.path" points at the rune executable and that it is executable.`,
    };
    const choice = await window.showErrorMessage(messages[probe.kind], "Retry", "Install instructions", "Open Settings");
    if (choice === "Retry") {
      void enqueueTransition(restartClient);
    } else if (choice === "Install instructions") {
      env.openExternal(Uri.parse(INSTALL_DOCS_URL));
    } else if (choice === "Open Settings") {
      commands.executeCommand("workbench.action.openSettings", "rune.path");
    }
  } finally {
    errorToastVisible = false;
  }
}

async function startClient() {
  const config = workspace.getConfiguration("rune");
  const command = config.get("path") || "rune";

  const probe = await probeServer(command);
  if (probe.kind !== "ok") {
    // Fire-and-forget: the notification's promise settles only when the user
    // interacts with the toast, and activation must not pend on that. The
    // client is never constructed, so there is no crash-restart loop.
    void explainUnusableBinary(command, probe);
    return;
  }

  // No `transport:` here — an Executable talks stdio by default, and setting
  // TransportKind.stdio makes vscode-languageclient append `--stdio` to the
  // args, which older rune releases reject as an unknown flag (exit 2,
  // crash-restart loop). rune ≥ 0.4.3 accepts --stdio, but not passing it
  // keeps the extension working with every LSP-capable rune.
  const serverOptions = {
    run: { command, args: ["lsp"] },
    debug: { command, args: ["lsp", "--log-level", "debug"] },
  };

  const clientOptions = {
    documentSelector: [{ scheme: "file", language: "runefile" }],
    synchronize: {
      // Notify the server when Runefiles change on disk (e.g. imported files
      // edited outside the editor).
      fileEvents: workspace.createFileSystemWatcher("**/{Runefile,.runefile,*.rune}"),
    },
  };

  client = new LanguageClient("rune", "Rune Language Server", serverOptions, clientOptions);
  try {
    await client.start();
  } catch (err) {
    // A dead client must not block later restarts (the next rune.path change
    // or Retry stops nothing and probes fresh).
    client = undefined;
    throw err;
  }
}

// Stop whatever is running (even a client whose server since crashed out) and
// start against the current configuration.
async function restartClient() {
  if (client) {
    const old = client;
    client = undefined;
    await old.stop().catch(() => {});
  }
  await startClient();
}

function activate(context) {
  // Restart when rune.path changes — switching binaries takes effect without
  // a window reload. Debounced: the settings editor writes intermediate
  // values while the user is still typing.
  let debounce;
  context.subscriptions.push(
    workspace.onDidChangeConfiguration((e) => {
      if (!e.affectsConfiguration("rune.path")) return;
      clearTimeout(debounce);
      debounce = setTimeout(() => {
        void enqueueTransition(restartClient);
      }, RESTART_DEBOUNCE_MS);
    }),
    { dispose: () => clearTimeout(debounce) },
  );
  return enqueueTransition(startClient);
}

function deactivate() {
  return client ? client.stop() : undefined;
}

module.exports = { activate, deactivate };
