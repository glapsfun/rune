package lsp

// dispatch routes one incoming message to its handler. It returns true when the
// server should stop (after `exit`).
func (s *Server) dispatch(msg *Message) (stop bool) {
	switch msg.Method {
	case "initialize":
		s.initialize(msg.ID, msg.Params)
	case "initialized":
		// no-op notification
	case "shutdown":
		s.shutdown(msg.ID)
	case "exit":
		return true
	case "textDocument/didOpen":
		s.didOpen(msg.Params)
	case "textDocument/didChange":
		s.didChange(msg.Params)
	case "textDocument/didSave":
		s.didSave(msg.Params)
	case "textDocument/didClose":
		s.didClose(msg.Params)
	case "textDocument/completion":
		s.completion(msg.ID, msg.Params)
	case "textDocument/definition":
		s.definition(msg.ID, msg.Params)
	case "textDocument/hover":
		s.hover(msg.ID, msg.Params)
	case "textDocument/formatting":
		s.formatting(msg.ID, msg.Params)
	case "textDocument/documentSymbol":
		s.documentSymbol(msg.ID, msg.Params)
	case "$/cancelRequest":
		// best-effort: handled implicitly by version guarding
	default:
		if msg.IsRequest() {
			// Unknown request: respond with MethodNotFound so the client isn't
			// left waiting (notifications are simply ignored).
			s.conn.Write(NewErrorResponse(msg.ID, MethodNotFound, "unsupported method: "+msg.Method))
		}
	}
	return false
}
