// Package inst implements basic types and functionality to describe and handle
// instruments (payload) of the CARIBIC container.
package inst

import (
	"car3-master/Go/state"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Instrument - a struct to characterize an instrument in the Conatiner-Lab.
type Instrument struct {
	ID        int         `json:"ID" yaml:"ID"`
	Name      string      `json:"Name" yaml:"Name"`
	Address   string      `json:"Address" yaml:"Address"`
	WUallowed bool        `json:"WU_allowed" yaml:"WU_allowed"`
	State     state.State //string
}

// NewInstr - instantiate a new instrument
func NewInstr() Instrument {
	i := Instrument{}
	i.ID = -1
	i.Name = "unknown"
	i.Address = "unknown"
	i.WUallowed = false
	i.State = state.Undefined // 0 // "Undefined"
	return i
}

// Payload - a mapping of instrument-ID (int) -> instrument type (struct).
type Payload map[int]Instrument

type instruments struct {
	Instruments []Instrument `json:"Payload" yaml:"Payload"`
}

// PayloadFromJSON - fill the payload map with instruments from a config file in json format.
func PayloadFromJSON(src string) (Payload, error) {
	var p = Payload{}
	var inst instruments

	jsonFile, err := os.Open(src)
	if err != nil {
		return p, err
	}
	defer jsonFile.Close()
	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return p, err
	}

	json.Unmarshal(jsonData, &inst)

	return unmarshalledToPayload(&inst, p) // p, nil
}

// PayloadFromYAML - fill the payload map with instruments from a config file in yaml format.
func PayloadFromYAML(src string) (Payload, error) {
	var p = Payload{}
	var inst instruments

	yamlFile, err := os.Open(src)
	if err != nil {
		return p, err
	}
	defer yamlFile.Close()
	yamlData, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return p, err
	}
	yaml.Unmarshal(yamlData, &inst)

	return unmarshalledToPayload(&inst, p) // p, nil
}

// unmarshalledToPayload a helper function to transfer instrument type to payload map
func unmarshalledToPayload(inst *instruments, p Payload) (Payload, error) {
	for i := 0; i < len(inst.Instruments); i++ {
		// check if key was already set
		if _, seen := p[inst.Instruments[i].ID]; seen {
			err := fmt.Sprintf("duplicate ID: %v\n", seen)
			return p, errors.New(err) // IDs must be unique!
		}
		// special treatment of Master and cRIO
		switch inst.Instruments[i].ID {
		case 0: // not allowed / wasn't parsed correctly
			continue // skip to next loop iteration
		case 1: // ID 1 == Master
			inst.Instruments[i].State = state.Measure
			inst.Instruments[i].WUallowed = true
		case 2: // ID 2 == cRIO
			inst.Instruments[i].WUallowed = true
		}

		p[inst.Instruments[i].ID] = inst.Instruments[i]
	}
	return p, nil
}

// AddressBytes - method to get 6-byte array that represents IP address (4 bytes)
// and UPD port (2 bytes). Address must be of type string with format "x.x.x.x:port".
func (i Instrument) AddressBytes() ([6]byte, bool) {
	parts := strings.Split(i.Address, ":")
	result := [6]byte{}
	if len(parts) != 2 {
		return result, false
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return result, false
	}

	copy(result[0:], net.IP.To4(net.ParseIP(parts[0])))
	binary.BigEndian.PutUint16(result[4:], uint16(port))

	return result, true
}

// ResolveUDPAddr method to get a *net.UDPAddr from the instrument's address string.
func (i Instrument) ResolveUDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", i.Address)
}

// String method to print the config of an instrument.
func (i Instrument) String() string {
	repr := fmt.Sprintf("ID:\t\t%v\n", i.ID)
	repr += fmt.Sprintf("Name:\t\t%s\n", i.Name)
	repr += fmt.Sprintf("Address:\t%s\n", i.Address)
	repr += fmt.Sprintf("WarmUp allowed:\t%v\n", i.WUallowed)
	repr += fmt.Sprintf("State:\t\t%s\n", i.State)
	return repr
}
