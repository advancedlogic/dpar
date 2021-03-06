// Copyright 2015 The dpar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package system

import (
	"github.com/danieldk/conllx"
)

type Guide interface {
	BestTransition(configuration *Configuration) Transition
}

type Parser interface {
	Parse(tokens []conllx.Token) (DependencySet, error)
}

var _ Parser = &GreedyParser{}

type GreedyParser struct {
	transitionSystem TransitionSystem
	guide            Guide
}

func NewGreedyParser(ts TransitionSystem, guide Guide) *GreedyParser {
	return &GreedyParser{ts, guide}
}

func (p *GreedyParser) Parse(tokens []conllx.Token) (DependencySet, error) {
	c, err := NewConfiguration(tokens)
	if err != nil {
		return nil, err
	}

	return p.parseConfiguration(&c), nil
}

func (p *GreedyParser) parseConfiguration(c *Configuration) DependencySet {
	for !p.transitionSystem.IsTerminal(*c) {
		p.guide.BestTransition(c).Apply(c)
	}

	return c.Dependencies()
}
