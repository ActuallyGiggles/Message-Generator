package markov

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// StartInstructions details intructions to start markov.
//
//	WriteInterval: How often to trigger a write cycle. If left blank, will be 10 minutes.
//	IntervalUnit: What unit to use for the WriteInterval. If left blank, will be 10 minutes.
//	SeparationKey: What string should act as a separator. (E.g. a " ")
//	StartKey: What string can be used to mark the beginning of a message. (E.g. "!-")
//	EndKey: What string can be used to mark the end of a message. (E.g. "-!")
//	ShouldZip: Whether or not to zip the markov-chains folder every six hours.
//	DefluffTriggerValue: What value amount is too little to keep and therefore should be defluffed.
//	ErrorTracker: If you want to recieve errors from write operations, provide a channel.
//	Debug: Print logs of stuffs.
type StartInstructions struct {
	WriteInterval int
	IntervalUnit  string

	SeparationKey string
	StartKey      string
	EndKey        string

	ShouldZip           bool
	DefluffTriggerValue int

	ErrorChannel chan error
	Debug        bool
}

// OutputInstructions details instructions on how to make an output.
//
//	Chain: What chain to use.
//	Method: What method to use.
//		"LikelyBeginning": Start with a likely beginning word.
//		"TargetedBeginning": Start with a specific beginning word.
//		"TargetedMiddle": Generate a message with a specific middle word. (yet to implement)
//		"TargetedEnding": End with a specific ending word.
//		"LikelyEnding": End with a likely ending word.
type OutputInstructions struct {
	Chain  string
	Method string
	Target string
}

type worker struct {
	Name    string
	Chain   chain
	ChainMx sync.Mutex
	Intake  int
}

type chain struct {
	Parents []parent
}

type parent struct {
	Word         string
	Grandparents []grandparent
	Children     []child
}

type child struct {
	Word  string
	Value int
}

type grandparent struct {
	Word  string
	Value int
}

// WorkerStats contains the name of the chain the worker is responsible for and the intake amount in that worker.
type WorkerStats struct {
	ChainResponsibleFor string
	Intake              int
}

type PeakIntakeStruct struct {
	Chain  string    `json:"chain"`
	Amount int       `json:"amount"`
	Time   time.Time `json:"time"`
}

// A Choice contains a generic item and a weight controlling the frequency with
// which it will be selected.
type Choice struct {
	Weight int
	Word   string
}

type encode struct {
	Encoder        *json.Encoder
	File           *os.File
	ContinuedEntry bool
}

type Progress struct {
	IsDone         bool
	CurrentProcess string
	Progress       int
	Total          int
}

type Statistics struct {
	// Total times
	TotalStartTime time.Time
	TotalUptime    time.Duration

	// Session times
	SessionStartTime time.Time
	SessionUptime    time.Duration

	// Inputs
	TotalInputs   int
	SessionInputs int

	// Outputs
	TotalOutputs   int
	SessionOutputs int

	// Write
	NextWriteTime  time.Time
	TimeUntilWrite time.Duration

	// Zip
	NextZipTime  time.Time
	TimeUntilZip time.Duration

	// Defluff
	NextDefluffTime  time.Time
	TimeUntilDefluff time.Duration

	Workers         int
	PeakChainIntake PeakIntakeStruct
	Durations       []report
}

type report struct {
	ProcessName string
	Duration    string
}
