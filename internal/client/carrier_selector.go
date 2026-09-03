// ==============================================================================
// XDNS
// Author: tajirax
// Github: https://github.com/WhiteDNS/XDNS
// Year: 2026
// ==============================================================================
// carrier_selector.go — adaptive DNS-record-type ("carrier") selection with
// automatic fallback. Rotates over the configured query types but steers away
// from carriers whose responses have stopped coming back (blocked or poisoned on
// the current network), and periodically re-explores them so a recovered carrier
// is picked up again. On a clean network where every carrier works it behaves
// exactly like the original round-robin; with a single configured type it is a
// no-op passthrough.
// ==============================================================================
package client

import (
	"sync"
	"sync/atomic"
	"time"

	Enums "xdns-go/internal/enums"
)

const (
	// How often the active-carrier set is recomputed (cheap TryLock on the hot path).
	carrierEvalInterval = 3 * time.Second
	// Minimum sends on a carrier before its success rate is trusted; below this it
	// is always kept in rotation so new/recovered carriers get explored.
	carrierMinSamples = 20
	// A sampled carrier is dropped from rotation when its success rate falls below
	// this (e.g. blocked or fully poisoned → responses stop coming back).
	carrierMinSuccessRate = 0.15
	// carrierMaxWeight bounds preference for a mature fast carrier. The lower
	// weight floor keeps every usable carrier in periodic rotation, which is
	// important when a censor changes policy after the initial scan.
	carrierMaxWeight = 4
	carrierRTTFloor  = 20 * time.Millisecond
)

type carrierSelector struct {
	types   []uint16
	typeIdx map[uint16]int
	cursor  atomic.Uint32

	// Per-type send/success counters, halved on each eval so stale data fades.
	sent    []atomic.Uint64
	success []atomic.Uint64
	// rttEWMA stores an exponentially weighted response time in nanoseconds.
	// It is deliberately per carrier and per resolver path (via forPath), so a
	// slow HTTPS answer through one resolver cannot demote that record type on a
	// different resolver where it is fast.
	rttEWMA []atomic.Int64

	active       atomic.Pointer[[]int] // type indices currently in rotation
	lastEvalNano atomic.Int64
	evalMu       sync.Mutex

	nowFn  func() time.Time
	pathMu sync.Mutex
	paths  map[string]*carrierSelector
}

func newCarrierSelector(types []uint16, nowFn func() time.Time) *carrierSelector {
	norm := normalizeRuntimeQueryTypes(types)
	s := &carrierSelector{
		types:   norm,
		typeIdx: make(map[uint16]int, len(norm)),
		sent:    make([]atomic.Uint64, len(norm)),
		success: make([]atomic.Uint64, len(norm)),
		rttEWMA: make([]atomic.Int64, len(norm)),
		nowFn:   nowFn,
		paths:   make(map[string]*carrierSelector),
	}
	for i, t := range norm {
		if _, dup := s.typeIdx[t]; !dup {
			s.typeIdx[t] = i
		}
	}
	all := make([]int, len(norm))
	for i := range norm {
		all[i] = i
	}
	s.active.Store(&all)
	return s
}

func (s *carrierSelector) forPath(path string) *carrierSelector {
	if s == nil || path == "" {
		return s
	}
	s.pathMu.Lock()
	defer s.pathMu.Unlock()
	if child := s.paths[path]; child != nil {
		return child
	}
	child := newCarrierSelector(s.types, s.nowFn)
	// Child selectors never need grandchildren.
	child.paths = nil
	s.paths[path] = child
	return child
}

func (s *carrierSelector) nextForPath(path string) uint16 {
	if path == "" {
		return s.next()
	}
	// Keep aggregate telemetry warm while the returned decision comes from the
	// path-specific selector.
	_ = s.next()
	return s.forPath(path).next()
}

func (s *carrierSelector) recordSuccessForPath(path string, qType uint16, rtt time.Duration) {
	s.recordSuccess(qType, rtt)
	if path != "" {
		s.forPath(path).recordSuccess(qType, rtt)
	}
}

func (s *carrierSelector) now() time.Time {
	if s != nil && s.nowFn != nil {
		return s.nowFn()
	}
	return time.Now()
}

// next returns the record type for the next tunnel query and counts it as sent.
func (s *carrierSelector) next() uint16 {
	if s == nil || len(s.types) == 0 {
		return Enums.DNS_RECORD_TYPE_TXT
	}
	if len(s.types) == 1 {
		s.sent[0].Add(1)
		return s.types[0]
	}
	s.maybeEval()
	active := s.active.Load()
	if active == nil || len(*active) == 0 {
		idx := int(s.cursor.Add(1)-1) % len(s.types)
		s.sent[idx].Add(1)
		return s.types[idx]
	}
	ti := (*active)[int(s.cursor.Add(1)-1)%len(*active)]
	s.sent[ti].Add(1)
	return s.types[ti]
}

// recordSuccess credits a carrier when one of its responses decodes successfully.
func (s *carrierSelector) recordSuccess(qType uint16, rtt time.Duration) {
	if s == nil {
		return
	}
	if i, ok := s.typeIdx[qType]; ok {
		s.success[i].Add(1)
		if rtt <= 0 {
			return
		}
		// CAS makes the hot response path lock-free. A 7:1 EWMA smooths normal
		// DNS jitter without hiding a sustained transport regression.
		next := rtt.Nanoseconds()
		for {
			old := s.rttEWMA[i].Load()
			if old > 0 {
				next = (old*7 + next) / 8
			}
			if s.rttEWMA[i].CompareAndSwap(old, next) {
				return
			}
		}
	}
}

func (s *carrierSelector) maybeEval() {
	now := s.now()
	if last := s.lastEvalNano.Load(); last != 0 && now.Sub(time.Unix(0, last)) < carrierEvalInterval {
		return
	}
	if !s.evalMu.TryLock() {
		return
	}
	defer s.evalMu.Unlock()
	if last := s.lastEvalNano.Load(); last != 0 && now.Sub(time.Unix(0, last)) < carrierEvalInterval {
		return
	}
	s.evalLocked()
	s.lastEvalNano.Store(now.UnixNano())
}

func (s *carrierSelector) evalLocked() {
	next := make([]int, 0, len(s.types)*carrierMaxWeight)
	bestIdx, bestRate := 0, -1.0
	bestScore := 0.0
	type carrierScore struct {
		index int
		score float64
		keep  bool
	}
	scores := make([]carrierScore, 0, len(s.types))
	for i := range s.types {
		sent := s.sent[i].Load()
		succ := s.success[i].Load()
		// Decay so stale results fade and a dropped carrier's send count shrinks
		// back under carrierMinSamples, forcing periodic re-exploration.
		s.sent[i].Store(sent / 2)
		s.success[i].Store(succ / 2)

		keep := sent < carrierMinSamples || float64(succ) >= carrierMinSuccessRate*float64(sent)
		if keep {
			// Bayesian smoothing avoids promoting a carrier that has one lucky
			// reply. Until enough samples arrive, all carriers retain equal
			// exploration weight.
			score := 1.0
			if sent >= carrierMinSamples {
				delivery := float64(succ+3) / float64(sent+6)
				rtt := time.Duration(s.rttEWMA[i].Load())
				if rtt < carrierRTTFloor {
					rtt = carrierRTTFloor
				}
				score = delivery / float64(rtt)
			}
			scores = append(scores, carrierScore{index: i, score: score, keep: true})
			if score > bestScore {
				bestScore = score
			}
		}
		if sent > 0 {
			if rate := float64(succ) / float64(sent); rate > bestRate {
				bestRate, bestIdx = rate, i
			}
		}
	}
	if len(scores) == 0 {
		// Never leave the rotation empty: keep the best-performing carrier so the
		// tunnel always has a carrier to send on.
		next = []int{bestIdx}
	} else {
		for _, candidate := range scores {
			weight := 1
			if bestScore > 0 && candidate.score < 1 {
				// A mature score is in duration^-1 units; only use preference after
				// the warm-up value has been replaced by a measured one.
				weight = max(1, min(carrierMaxWeight, int(candidate.score/bestScore*float64(carrierMaxWeight)+0.5)))
			}
			for copies := 0; copies < weight; copies++ {
				next = append(next, candidate.index)
			}
		}
	}
	s.active.Store(&next)
}
