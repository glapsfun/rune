package mcpserver

// authorized reports whether a task may be CALLED in this session (Q3/FR-028):
//   - private tasks are never registered as tools (filtered upstream);
//   - non-destructive tasks are callable;
//   - destructive ([confirm]) tasks require explicit approval (AllowDestructive);
//   - an operator AllowList, when set, further narrows callable tasks.
func (s *Server) authorized(taskName string) bool {
	if len(s.opts.AllowList) > 0 && !inList(s.opts.AllowList, taskName) {
		return false
	}
	if s.isDestructive(taskName) && !s.opts.AllowDestructive {
		return false
	}
	return true
}

// isDestructive looks up whether the named task is marked destructive.
func (s *Server) isDestructive(taskName string) bool {
	for _, t := range s.engine.Tasks() {
		if t.Name == taskName {
			return t.Destructive
		}
	}
	return false
}

func inList(list []string, name string) bool {
	for _, n := range list {
		if n == name {
			return true
		}
	}
	return false
}
