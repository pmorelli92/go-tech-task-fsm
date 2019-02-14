package domain

import (
	"errors"
	"fmt"
	"github.com/looplab/fsm"
	"time"
)

const (
	readyState = "ready"
	ridingState = "riding"
	batteryLowState = "batteryLow"
	bountyState = "bounty"
	collectedState = "collected"
	droppedState = "dropped"
	unknownState = "unknown"
)

const (
	startRideEvent = "startRideEvent"
	finishRideEvent = "finishRideEvent"
	batteryLowEvent = "batteryLowEvent"
	bountyEvent = "bountyEvent"
	collectedEvent = "collectedEvent"
	droppedEvent = "droppedEvent"
	readyEvent = "readyEvent"
	unknownEvent = "unknownEvent"
)

type UserRole int
const (
	EndUser	UserRole = iota + 1
	Hunter
	Admin
)

type Vehicle struct {
	battery           int
	fsm               *fsm.FSM
	LastChangeOfState time.Time
}

func (v *Vehicle) Battery() int {
	return v.battery
}

func NewVehicle() *Vehicle {

	vehicle := &Vehicle{battery:100, LastChangeOfState: time.Now() }
	vehicle.fsm = fsm.NewFSM(
		readyState,
		fsm.Events{

			{Name: startRideEvent, Src: []string{readyState}, Dst: ridingState},
			{Name: finishRideEvent, Src: []string{ridingState}, Dst: readyState},
			{Name: batteryLowEvent, Src: []string{ridingState}, Dst: batteryLowState},
			{Name: bountyEvent, Src: []string{readyState, batteryLowState}, Dst: bountyState},
			{Name: collectedEvent, Src: []string{bountyState}, Dst: collectedState},
			{Name: droppedEvent, Src: []string{collectedState}, Dst: droppedState},
			{Name: readyEvent, Src: []string{droppedState}, Dst: readyState},
			{Name: unknownEvent, Src: []string{readyState}, Dst: unknownState},
		},
		// Callbacks only gets executed when an event is triggered and not by forcing a change of state
		// This means we will have repeated logic on admin actions
		fsm.Callbacks{
			"enter_" + batteryLowState: func(e *fsm.Event) {
				e.FSM.SetState(bountyState)
			},
			"enter_state": func(e *fsm.Event) {
				vehicle.LastChangeOfState = time.Now()
			},
		},
	)

	return vehicle
}

func (v *Vehicle) GetCurrentState() string {
	return v.fsm.Current()
}

func (v * Vehicle) StartRide(role UserRole) error {
	var err error = nil
	switch role {
	case Hunter:
		fallthrough
	case EndUser:
		if v.fsm.Can(startRideEvent) {
			err = v.fsm.Event(startRideEvent)
		} else {
			errMsg := fmt.Sprintf("you cannot finish a ride right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	case Admin:
		v.fsm.SetState(ridingState)
		v.LastChangeOfState = time.Now()
	}

	return err
}

func (v * Vehicle) FinishRide(batteryLeft int, role UserRole) error {
	var err error = nil
	switch role {
	case Hunter:
		fallthrough
	case EndUser:
		if v.fsm.Can(finishRideEvent) {
			v.battery = batteryLeft
			if v.battery < 20 {
				err = v.fsm.Event(batteryLowEvent)
			} else {
				err = v.fsm.Event(finishRideEvent)
			}
		} else {
			errMsg := fmt.Sprintf("you cannot finish a ride right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	case Admin:
		v.battery = batteryLeft
		v.fsm.SetState(readyState)
		v.LastChangeOfState = time.Now()
	}

	return err
}

func (v * Vehicle) Collect(role UserRole) error {
	var err error = nil
	switch role {
	case EndUser:
		err = errors.New("you cannot collect the vehicle being end user")
	case Hunter:
		if v.fsm.Can(collectedEvent) {
			err = v.fsm.Event(collectedEvent)
		} else {
			errMsg := fmt.Sprintf("you cannot collect the vehicle right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	case Admin:
		v.fsm.SetState(collectedState)
		v.LastChangeOfState = time.Now()
	}
	return err
}

func (v * Vehicle) Drop(role UserRole) error {
	var err error = nil
	switch role {
	case EndUser:
		err = errors.New("you cannot drop the vehicle being end user")
	case Hunter:
		if v.fsm.Can(droppedEvent) {
			err = v.fsm.Event(droppedEvent)
		} else {
			errMsg := fmt.Sprintf("you cannot drop the vehicle right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	case Admin:
		v.fsm.SetState(droppedState)
		v.LastChangeOfState = time.Now()

	}
	return err
}

func (v * Vehicle) Ready(role UserRole) error {
	var err error = nil
	switch role {
	case EndUser:
		err = errors.New("you cannot set the vehicle as ready being end user")
	case Hunter:
		if v.fsm.Can(readyEvent) {
			v.battery = 100
			err = v.fsm.Event(readyEvent)
		} else {
			errMsg := fmt.Sprintf("you cannot set the vehicle ready right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	case Admin:
		v.battery = 100
		v.fsm.SetState(readyState)
		v.LastChangeOfState = time.Now()

	}
	return err
}


func (v * Vehicle) SetBatteryLow(role UserRole) error {
	var err error = nil
	if role == Admin {
		v.fsm.SetState(batteryLowState)
		v.LastChangeOfState = time.Now()
	} else {
		err = errors.New("only admin can set the vehicle on low battery")
	}
	return err
}


func (v * Vehicle) SetBounty(role UserRole) error {
	var err error = nil
	if role == Admin {
		v.fsm.SetState(bountyState)
		v.LastChangeOfState = time.Now()
	} else {
		err = errors.New("only admin can set the vehicle on bounty")
	}
	return err
}

func (v * Vehicle) Unknown(role UserRole) error {
	var err error = nil
	if role == Admin {
		if v.fsm.Can(unknownEvent) {
			err = v.fsm.Event(unknownEvent)
		} else {
			errMsg := fmt.Sprintf("you cannot set the vehicle to unknown right now the current state is %s", v.GetCurrentState())
			err = errors.New(errMsg)
		}
	} else {
		err = errors.New("only admin can set the vehicle to unknown")
	}
	return err
}

func SetVehiclesFromReadyToBounty(vehicles []*Vehicle) []error {
	var errs []error
	for _, v := range vehicles {
		if v.fsm.Can(bountyEvent) {
			if err := v.fsm.Event(bountyEvent); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

func SetVehiclesFromReadyToUnknown(vehicles []*Vehicle) []error {
	var errs []error
	for _, v := range vehicles {
		duration := time.Since(v.LastChangeOfState)
		if duration.Hours() >= 48 {
			if v.fsm.Can(unknownEvent) {
				if err := v.fsm.Event(unknownEvent); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	return errs
}