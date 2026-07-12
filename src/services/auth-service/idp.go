package main

import (
	"log"
	"sync"
	"time"
)

type IdpStateMachine struct {
	mu           sync.RWMutex
	currentState IdpState
	lastFailure  time.Time
	failureCount int
	localDBTTL   time.Duration
}

var (
	globalIdpState = &IdpStateMachine{
		currentState: IdpPrimary,
		localDBTTL:   3600 * time.Second,
	}
)

const (
	idpDegradeThreshold  = 3
	idpRecoveryThreshold = 2
)

func (m *IdpStateMachine) CurrentState() IdpState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentState
}

func (m *IdpStateMachine) RecordFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastFailure = time.Now()
	m.failureCount++

	if m.failureCount >= idpDegradeThreshold {
		m.failureCount = 0
		switch m.currentState {
		case IdpPrimary:
			m.currentState = IdpFallbackLocalDB
			log.Printf("IdP degraded to %s", m.currentState)
		case IdpFallbackLocalDB:
			m.currentState = IdpFallbackReadonly
			log.Printf("IdP degraded to %s", m.currentState)
		case IdpFallbackReadonly:
			m.currentState = IdpEmergencyBypass
			log.Printf("IdP degraded to %s", m.currentState)
		case IdpEmergencyBypass:
		}
	}
}

func (m *IdpStateMachine) RecordSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.failureCount = 0

	if m.currentState != IdpPrimary {
		m.currentState = IdpPrimary
		log.Printf("IdP recovered to %s", m.currentState)
	}
}

func (m *IdpStateMachine) IsDegraded() bool {
	return m.CurrentState() != IdpPrimary
}

func (m *IdpStateMachine) RequiresMFA() bool {
	return m.CurrentState() == IdpEmergencyBypass
}
