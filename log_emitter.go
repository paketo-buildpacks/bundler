package bundler

import (
	"io"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type LogEmitter struct {
	// Emitter is embedded and therefore delegates all of its functions to the
	// LogEmitter.
	scribe.Emitter
}

func NewLogEmitter(output io.Writer) LogEmitter {
	return LogEmitter{
		Emitter: scribe.NewEmitter(output),
	}
}

func (l LogEmitter) Environment(env packit.Environment) {
	l.Process("Configuring environment")
	l.Subprocess("%s", scribe.NewFormattedMapFromEnvironment(env))
	l.Break()
}
