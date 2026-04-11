package adapter

import (
	"hop.top/kit/toolspec"
)

// enrichAfterMaps derives After maps from Workflow.Steps for any
// workflow that has Steps but no After map. This bridges the gap
// between LLM/tldr sources (which produce Steps) and the generic
// adapter (which reads After maps).
func enrichAfterMaps(spec *toolspec.ToolSpec) {
	if spec == nil {
		return
	}
	for i := range spec.Workflows {
		wf := &spec.Workflows[i]
		if len(wf.After) > 0 || len(wf.Steps) < 2 {
			continue
		}
		wf.After = make(map[string][]string)
		for j := 0; j < len(wf.Steps)-1; j++ {
			cur := wf.Steps[j]
			next := wf.Steps[j+1]
			wf.After[cur] = append(wf.After[cur], next)
		}
		// Also link last step back to first (cycle hint).
		last := wf.Steps[len(wf.Steps)-1]
		first := wf.Steps[0]
		if _, ok := wf.After[last]; !ok {
			wf.After[last] = []string{first}
		}
	}
}

// enrichingSource wraps a toolspec.Source and enriches resolved
// specs by deriving After maps from Steps.
type enrichingSource struct {
	inner toolspec.Source
}

func (s *enrichingSource) Resolve(tool string) (*toolspec.ToolSpec, error) {
	spec, err := s.inner.Resolve(tool)
	if err != nil || spec == nil {
		return spec, err
	}
	enrichAfterMaps(spec)
	return spec, nil
}
