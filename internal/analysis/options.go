package analysis

// AnalyzeRequest asks the service to analyze one entry document.
//
// When Content is non-empty it is used as the entry document's source directly
// (the unsaved editor buffer); otherwise the entry is read through the service's
// SourceStore. Imported/mod files are always resolved through the store, so open
// overlays apply transitively (spec FR-003).
type AnalyzeRequest struct {
	URI     DocumentURI
	Content string
	Version int
}
