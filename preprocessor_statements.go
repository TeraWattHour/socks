package socks

import (
	"github.com/terawatthour/socks/internal/helpers"
)

// ---------------------- Extend Statement ----------------------

type ExtendStatement struct {
	Template string
	location helpers.Location
}

func (es *ExtendStatement) IsClosable() bool {
	return false
}

func (es *ExtendStatement) Location() helpers.Location {
	return es.location
}

func (es *ExtendStatement) Kind() string {
	return "extend"
}

// ---------------------- Template Statement ----------------------

type TemplateStatement struct {
	Template     string
	location     helpers.Location
	EndStatement *EndStatement
}

func (es *TemplateStatement) IsClosable() bool {
	return true
}

func (es *TemplateStatement) Location() helpers.Location {
	return es.location
}

func (es *TemplateStatement) Kind() string {
	return "template"
}

// ---------------------- Slot Statement ----------------------

type SlotStatement struct {
	Name         string
	location     helpers.Location
	Parent       Statement
	EndStatement *EndStatement
}

func (ss *SlotStatement) IsClosable() bool {
	return true
}

func (ss *SlotStatement) Location() helpers.Location {
	return ss.location
}

func (ss *SlotStatement) Kind() string {
	return "slot"
}

// ---------------------- Define Statement ----------------------

type DefineStatement struct {
	Name         string
	location     helpers.Location
	Parent       Statement
	EndStatement *EndStatement
}

func (es *DefineStatement) IsClosable() bool {
	return true
}

func (es *DefineStatement) Location() helpers.Location {
	return es.location
}

func (es *DefineStatement) Kind() string {
	return "define"
}
