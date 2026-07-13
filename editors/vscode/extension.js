// Rune VS Code extension: a thin client that launches `rune lsp` over stdio and
// wires it to Runefile documents. All language intelligence lives in the Rune
// binary (one parser, one analyzer, one formatter) — this file only starts the
// server and forwards requests.
const { workspace } = require("vscode");
const { LanguageClient, TransportKind } = require("vscode-languageclient/node");

/** @type {import('vscode-languageclient/node').LanguageClient | undefined} */
let client;

function activate(_context) {
  const config = workspace.getConfiguration("rune");
  const command = config.get("path") || "rune";

  const serverOptions = {
    run: { command, args: ["lsp"], transport: TransportKind.stdio },
    debug: { command, args: ["lsp", "--log-level", "debug"], transport: TransportKind.stdio },
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
  return client.start();
}

function deactivate() {
  return client ? client.stop() : undefined;
}

module.exports = { activate, deactivate };
