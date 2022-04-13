package httpoc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// errInvalidTraceParent indicates Parent parsing error.
var errInvalidTraceParent = errors.New("invalid_trace_parent")

const (
	// headerTraceParent defines https://www.w3.org/TR/trace-context/ spec's Traceparent header name.
	headerTraceParent = "Traceparent"
	// headerTraceResponse defines https://www.w3.org/TR/trace-context/ spec's Traceresponse header name.
	headerTraceResponse = "TraceResponse"

	// traceIDSize defines the size of TraceID in bytes.
	traceIDSize = 16
	// traceParentIDSize defines the size of TraceParentID in bytes.
	traceParentIDSize = 8

	traceParentFormat = "%02x-%032x-%016x-%02x"
	logTraceID        = "id"
	logTraceParentID  = "span"
	xidSize           = 12
)

// traceID represents trace id.
type traceID [traceIDSize]byte

// String returns ID's hex string representation.
func (i *traceID) String() string {
	return hex.EncodeToString(i[:])
}

// traceParentID represents parent id (span id).
type traceParentID [traceParentIDSize]byte

// String returns TraceParentID's hex string representation.
func (p *traceParentID) String() string {
	return hex.EncodeToString(p[:])
}

// traceParent represents https://www.w3.org/TR/baggage/ spec's Traceparent.
type traceParent struct {
	// Version is the version number, 0 for current spec.
	Version byte
	// Flags is a bit field for trace flags, only "sampled" flag defined in current spec.
	Flags byte
	// TraceID is fixed 16 byte trace id.
	TraceID *traceID
	// ParentID is a fixed 8 byte trace parent id (span id).
	ParentID *traceParentID
}

// newTraceParent return a new Parent instance with random generated TraceID and ParentID.
func newTraceParent() *traceParent {
	p := &traceParent{}

	p.ParentID = &traceParentID{}
	if _, err := rand.Read(p.ParentID[:]); err != nil {
		panic(err)
	}

	p.TraceID = &traceID{}
	copy(p.TraceID[:], xid.New().Bytes())

	copy(p.TraceID[xidSize:], p.ParentID[traceParentIDSize-(traceIDSize-xidSize):])

	return p
}

// parseTraceParent builds Parent from string representation defined in https://www.w3.org/TR/trace-context/ spec.
func parseTraceParent(str string) (p *traceParent, err error) {
	p = &traceParent{}

	var tid, pid []byte

	n, err := fmt.Sscanf(strings.ToLower(str), traceParentFormat, &p.Version, &tid, &pid, &p.Flags)
	if n != 4 || err != nil {
		return nil, errInvalidTraceParent
	}

	p.TraceID = &traceID{}
	copy(p.TraceID[:], tid)

	p.ParentID = &traceParentID{}
	copy(p.ParentID[:], pid)

	return p, nil
}

// String returns Parent string representation, as defined in https://www.w3.org/TR/trace-context/ spec.
func (p *traceParent) String() string {
	return fmt.Sprintf(traceParentFormat, p.Version, p.TraceID[:], p.ParentID[:], p.Flags)
}

// MarshalZerologObject is a zerolog object marshaller for Parent.
func (p *traceParent) MarshalZerologObject(e *zerolog.Event) {
	e.Hex(logTraceID, p.TraceID[:])
	e.Hex(logTraceParentID, p.ParentID[:])
}
