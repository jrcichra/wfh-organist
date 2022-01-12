package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var prom_last_ms = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "wfh_organist_last_ms",
		Help: "Last ms reported",
	},
)

// Not implemented yet

var prom_notes_on = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_notes_on",
		Help: "# of Notes on",
	},
	[]string{"key", "channel", "velocity"},
)

var prom_notes_off = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_notes_off",
		Help: "# of Notes off",
	},
	[]string{"key", "channel"},
)

var prom_program_change = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_program_change",
		Help: "# of Program change",
	},
	[]string{"program", "channel"},
)

var prom_aftertouch = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_aftertouch",
		Help: "# of Aftertouch",
	},
	[]string{"pressure", "channel"},
)

var prom_control_change = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_control_change",
		Help: "# of Control change",
	},
	[]string{"controller", "value", "channel"},
)

var prom_note_off_velocity = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_note_off_velocity",
		Help: "# of Note off velocity",
	},
	[]string{"key", "velocity", "channel"},
)

var prom_pitchbend = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_pitchbend",
		Help: "# of Pitchbend",
	},
	[]string{"value", "absvalue", "channel"},
)

var prom_poly_aftertouch = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "wfh_organist_poly_aftertouch",
		Help: "# of Poly aftertouch",
	},
	[]string{"key", "pressure", "channel"},
)
