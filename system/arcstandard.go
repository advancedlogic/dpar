// Copyright 2015 The dpar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package system

import (
	"errors"
	"fmt"
	"strings"
)

// Archetype transitions + interface validation.
var archetypeASShift Transition = asShift{}
var archetypeASLeftArc Transition = asLeftArc{"<archetype>"}
var archetypeASRightArc Transition = asRightArc{"<archetype>"}

// Assert TransitionSystem/TransitionSerializer conformance.
var _ TransitionSystem = NewArcStandard()
var _ TransitionSerializer = NewArcStandard()

type ArcStandard struct {
	archetypeTransitions TransitionSet
}

func NewArcStandard() *ArcStandard {
	trans := map[Transition]interface{}{
		archetypeASShift:    nil,
		archetypeASLeftArc:  nil,
		archetypeASRightArc: nil,
	}

	return &ArcStandard{trans}
}

func (ts *ArcStandard) IsTerminal(c Configuration) bool {
	return len(c.Buffer) == 0
}

func (ts *ArcStandard) PossibleTransitions(configuration Configuration) TransitionSet {
	possible := make(TransitionSet)

	for trans := range ts.archetypeTransitions {
		if trans.IsPossible(configuration) {
			possible[trans] = nil
		}
	}

	return possible
}

func (ts *ArcStandard) SerializeTransition(t Transition) (string, error) {
	switch t := t.(type) {
	case asLeftArc:
		return fmt.Sprintf("LEFT_ARC %s", t.relation), nil
	case asRightArc:
		return fmt.Sprintf("RIGHT_ARC %s", t.relation), nil
	case asShift:
		return "SHIFT", nil
	default:
		return "", fmt.Errorf("Not a transition of the arc standard system: %T", t)
	}
}

func (ts *ArcStandard) DeserializeTransition(transString string) (Transition, error) {
	parts := strings.SplitN(transString, " ", 2)

	if len(parts) == 0 {
		return nil, errors.New("Empty transition")
	}

	switch parts[0] {
	default:
		return nil, fmt.Errorf("Unknown transition: %s", parts[0])
	case "LEFT_ARC":
		if len(parts) == 1 {
			return nil, errors.New("Left-Arc transition requires label argument")
		}
		return asLeftArc{parts[1]}, nil
	case "RIGHT_ARC":
		if len(parts) == 1 {
			return nil, errors.New("Right-Arc transition requires label argument")
		}
		return asRightArc{parts[1]}, nil
	case "SHIFT":
		return asShift{}, nil
	}
}

type asLeftArc struct {
	relation string
}

func (l asLeftArc) IsPossible(c Configuration) bool {
	stackSize := len(c.Stack)
	return len(c.Stack) != 0 && len(c.Buffer) != 0 &&
		c.Stack[stackSize-1] != 0
}

func (l asLeftArc) Apply(c *Configuration) {
	stackSize := len(c.Stack)
	head := c.Buffer[0]
	dependant := c.Stack[stackSize-1]
	dependency := Dependency{head, l.relation, dependant}

	c.AddDependency(&dependency)
	c.Stack = c.Stack[:stackSize-1]
}

type asRightArc struct {
	relation string
}

func (r asRightArc) IsPossible(c Configuration) bool {
	return len(c.Stack) != 0 && len(c.Buffer) != 0
}

func (r asRightArc) Apply(c *Configuration) {
	stackSize := len(c.Stack)
	dependant := c.Buffer[0]
	head := c.Stack[stackSize-1]
	dependency := Dependency{head, r.relation, dependant}

	c.AddDependency(&dependency)
	c.Stack = c.Stack[:stackSize-1]
	c.Buffer[0] = head
}

type asShift struct{}

func (s asShift) IsPossible(c Configuration) bool {
	return len(c.Buffer) != 0
}

func (s asShift) Apply(c *Configuration) {
	token := c.Buffer[0]
	c.Buffer = c.Buffer[1:]
	c.Stack = append(c.Stack, token)
}

type ArcStandardOracle struct {
	dependentHeadMapping map[uint]Dependency
}

func NewArcStandardOracle(goldDependencies DependencySet) Guide {
	oracle := ArcStandardOracle{goldDependencies.CreateDependentHeadMapping()}
	return &oracle
}

func (o *ArcStandardOracle) BestTransition(c *Configuration) Transition {
	stackSize := len(c.Stack)
	if stackSize != 0 {
		stackTip := c.Stack[stackSize-1]
		bufferHead := c.Buffer[0]

		la := asLeftArc{o.dependentHeadMapping[stackTip].Relation}
		if la.IsPossible(*c) && o.dependentHeadMapping[stackTip].Head == bufferHead {
			return la
		}

		ra := asRightArc{o.dependentHeadMapping[bufferHead].Relation}
		if ra.IsPossible(*c) && o.dependentHeadMapping[bufferHead].Head == stackTip &&
			!o.neededForAttachment(c, bufferHead) {
			return ra
		}
	}

	return asShift{}
}

func (o *ArcStandardOracle) neededForAttachment(c *Configuration, token uint) bool {
	for _, bufToken := range c.Buffer {
		if o.dependentHeadMapping[bufToken].Head == token {
			return true
		}
	}

	return false
}
