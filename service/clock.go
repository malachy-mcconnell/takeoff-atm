package service

import (
	"github.com/malachy-mcconnell/takeoff-atm/domain"
	"time"
)

type Clock struct {
	startTime      time.Time
	running        bool
	duration       time.Duration
	channel        <-chan *domain.Account
	sessionAccount *domain.Account
}

func NewSessionClock(duration time.Duration, c <-chan *domain.Account, sessionAccount *domain.Account) *Clock {
	clock := Clock{}
	clock.startTime = time.Now()
	clock.running = true
	clock.duration = duration
	clock.channel = c
	clock.sessionAccount = sessionAccount
	return &clock
}
