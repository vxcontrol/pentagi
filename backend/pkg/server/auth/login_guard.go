package auth

import (
	"sync"
	"time"
)

const maxGuardEntries = 4096

type guardEntry struct {
	attempts    int
	windowEnds  time.Time
	lockedUntil time.Time
}

func (e *guardEntry) freeAt() time.Time {
	if e.lockedUntil.After(e.windowEnds) {
		return e.lockedUntil
	}
	return e.windowEnds
}

// One instance holds one kind of key: mixing account keys and address keys into
// a single guard gives them one shared limit.
type LoginGuard struct {
	mu      sync.Mutex
	entries map[string]*guardEntry
	limit   int
	window  time.Duration
	lockout time.Duration
	now     func() time.Time
}

func NewLoginGuard(limit int, window, lockout time.Duration) *LoginGuard {
	return &LoginGuard{
		entries: make(map[string]*guardEntry),
		limit:   limit,
		window:  window,
		lockout: lockout,
		now:     time.Now,
	}
}

func (g *LoginGuard) Take(key string) (time.Duration, bool) {
	if g == nil || key == "" {
		return 0, true
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	now := g.now()
	entry := g.entryLocked(key, now)
	if entry == nil {
		return 0, true
	}

	if now.Before(entry.lockedUntil) {
		return entry.lockedUntil.Sub(now), false
	}

	if now.After(entry.windowEnds) {
		entry.attempts = 0
		entry.windowEnds = now.Add(g.window)
	}

	entry.attempts++
	if entry.attempts > g.limit {
		entry.attempts = 0
		entry.lockedUntil = now.Add(g.lockout)
		return g.lockout, false
	}

	return 0, true
}

func (g *LoginGuard) Reset(key string) {
	if g == nil || key == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.entries, key)
}

func (g *LoginGuard) Relax(key string) {
	if g == nil || key == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	entry, ok := g.entries[key]
	if !ok {
		return
	}

	entry.attempts /= 2
	if entry.attempts == 0 && !g.now().Before(entry.lockedUntil) {
		delete(g.entries, key)
	}
}

func (g *LoginGuard) Refund(key string) {
	if g == nil || key == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	entry, ok := g.entries[key]
	if !ok || entry.attempts == 0 {
		return
	}

	entry.attempts--
	if entry.attempts == 0 && !g.now().Before(entry.lockedUntil) {
		delete(g.entries, key)
	}
}

func (g *LoginGuard) entryLocked(key string, now time.Time) *guardEntry {
	if entry, ok := g.entries[key]; ok {
		return entry
	}

	if len(g.entries) >= maxGuardEntries {
		g.sweepLocked(now)
	}
	if len(g.entries) >= maxGuardEntries {
		g.evictCheapestLocked()
	}
	if len(g.entries) >= maxGuardEntries {
		return nil
	}

	entry := &guardEntry{}
	g.entries[key] = entry

	return entry
}

func (g *LoginGuard) sweepLocked(now time.Time) {
	for key, entry := range g.entries {
		if now.After(entry.freeAt()) {
			delete(g.entries, key)
		}
	}
}

func (g *LoginGuard) evictCheapestLocked() {
	var (
		victim         string
		victimFree     time.Time
		victimAttempts int
		locked         bool
	)

	for key, entry := range g.entries {
		entryLocked := !entry.lockedUntil.IsZero()
		free := entry.freeAt()
		sameClass := locked == entryLocked
		switch {
		case victim == "":
		case locked && !entryLocked:
		case sameClass && entry.attempts < victimAttempts:
		case sameClass && entry.attempts == victimAttempts && free.Before(victimFree):
		default:
			continue
		}
		victim, victimFree, victimAttempts, locked = key, free, entry.attempts, entryLocked
	}

	if victim != "" {
		delete(g.entries, victim)
	}
}
