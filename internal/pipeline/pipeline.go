// Package pipeline chains multiple io.WriteCloser stages together so that
// output from a cron job flows through each stage in sequence.
package pipeline

import "io"

// Stage is any io.WriteCloser that can participate in a pipeline.
type Stage = io.WriteCloser

// Pipeline writes to the first stage; each stage is responsible for forwarding
// data to the next via its own internal writer. Close flushes and closes all
// stages in reverse order so that buffered data propagates downstream before
// the downstream stage is torn down.
type Pipeline struct {
	stages []Stage
}

// New creates a Pipeline from the provided stages. Stages are closed in
// reverse order when Close is called. At least one stage must be provided.
func New(stages ...Stage) (*Pipeline, error) {
	if len(stages) == 0 {
		return nil, errNoStages
	}
	return &Pipeline{stages: stages}, nil
}

// Write sends p to the first stage in the pipeline.
func (p *Pipeline) Write(data []byte) (int, error) {
	return p.stages[0].Write(data)
}

// Close closes all stages in reverse order, collecting the first error
// encountered. Subsequent stages are still closed even if an earlier one
// returns an error.
func (p *Pipeline) Close() error {
	var first error
	for i := len(p.stages) - 1; i >= 0; i-- {
		if err := p.stages[i].Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// Len returns the number of stages in the pipeline.
func (p *Pipeline) Len() int { return len(p.stages) }

// errors
type pipelineError string

func (e pipelineError) Error() string { return string(e) }

const errNoStages pipelineError = "pipeline: at least one stage is required"
