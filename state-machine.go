package state_machine

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type IStateMachine interface {
	GetName() string
	Load(filePath string) error
	ProcessTransition(currentState, nextState string, obj any) error
	runCheckFunction(handlers []string, obj any) error
	runOnErrorFunction(handlers []string, obj any) error
	runOnSuccessFunction(handlers []string, obj any) error
	AddCheckFunction(name string, handler HandlerFunc)
	AddOnErrorFunction(name string, handler HandlerFunc)
	AddOnSuccessFunction(name string, handler HandlerFunc)
	AddExecuteFunction(handler HandlerExecFunction)
	getCheckFunction(name string) HandlerFunc
	getOnErrorFunction(name string) HandlerFunc
	getOnSuccessFunction(name string) HandlerFunc
	AddStateMachineToTrigger(stateMachine IStateMachine) IStateMachine
}

func NewStateMachine() IStateMachine {
	return &StateMachine{
		MapStates:              make(map[string]map[string]Handlers),
		StateMachinesToTrigger: make(map[string]IStateMachine),
		CheckHandlers:          make(map[string]HandlerFunc),
		OnSuccessHandlers:      make(map[string]HandlerFunc),
		OnErrorHandlers:        make(map[string]HandlerFunc),
	}
}

// StateMachine ...
type StateMachine struct {
	filePath               string
	Name                   string                         `json:"name"`
	States                 []State                        `json:"states"`
	MapStates              map[string]map[string]Handlers `json:"map_states"`
	Execute                HandlerExecFunction            `json:"execute"`
	CheckHandlers          map[string]HandlerFunc         `json:"check_handlers"`
	OnSuccessHandlers      map[string]HandlerFunc         `json:"on_success_handlers"`
	OnErrorHandlers        map[string]HandlerFunc         `json:"on_error_handlers"`
	StateMachinesToTrigger map[string]IStateMachine       `json:"state_machines_to_trigger"`
}

type State struct {
	Name        string       `json:"name"`
	Transitions []Transition `json:"transitions"`
}

type Transition struct {
	// Name
	Name string `json:"name"`
	// Check
	Check []string `json:"check"`
	// On Success
	OnSuccess []string `json:"on_success"  mapstructure:"on_success"`
	// On Error
	OnError []string `json:"on_error" mapstructure:"on_error"`
}

type Handlers struct {
	//Update Status
	updateStatus string
	// Check
	Check []string `json:"check"`
	// On Success
	OnSuccess []string `json:"on_success"`
	// On Error
	OnError []string `json:"on_error"`
}

type HandlerExecFunction func(currentState, nextState string, arg any) (success bool, err error)
type HandlerFunc func(arg any, optArg ...string) (success bool, err error)

func (sm *StateMachine) Load(filePath string) error {
	// Read the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	//var test any
	viper.SetConfigFile(filePath)

	if err = viper.ReadInConfig(); err != nil {
		return err
	}

	if err = viper.Unmarshal(&sm); err != nil {
		return err
	}

	var handlers Handlers
	// initialize the state machine
	for _, state := range sm.States {

		if sm.MapStates[state.Name] == nil {
			sm.MapStates[state.Name] = make(map[string]Handlers)
		}

		for _, transition := range state.Transitions {
			// add check handlers
			handlers.Check = append(handlers.Check, transition.Check...)
			// add on_success handlers
			handlers.OnSuccess = append(handlers.OnSuccess, transition.OnSuccess...)

			// add on_error handlers
			handlers.OnError = append(handlers.OnError, transition.OnError...)

			sm.MapStates[state.Name][transition.Name] = handlers
		}
	}

	return nil
}

func checkIfFunctionWithArguments(input string) (success bool, function string, arguments []string) {
	parts := strings.Split(input, "(")
	if len(parts) != 2 {
		return false, "", nil
	}

	functionName := parts[0]
	argumentsString := parts[1]
	argumentsString = strings.TrimSuffix(argumentsString, ")")
	arguments = strings.Split(argumentsString, ",")
	for i, arg := range arguments {
		arguments[i] = strings.TrimSpace(arg)
	}
	return true, functionName, arguments
}

func (sm *StateMachine) GetName() string {
	return sm.Name
}

func (sm *StateMachine) getCheckFunction(name string) HandlerFunc {
	// it's an internal function
	return sm.CheckHandlers[name]
}
func (sm *StateMachine) getOnErrorFunction(name string) HandlerFunc {
	return sm.OnErrorHandlers[name]
}

func (sm *StateMachine) getOnSuccessFunction(name string) HandlerFunc {
	// it's an internal function

	return sm.OnSuccessHandlers[name]
}

func (sm *StateMachine) AddCheckFunction(name string, handler HandlerFunc) {
	sm.CheckHandlers[name] = handler
}

func (sm *StateMachine) AddOnErrorFunction(name string, handler HandlerFunc) {
	sm.OnErrorHandlers[name] = handler
}

func (sm *StateMachine) AddOnSuccessFunction(name string, handler HandlerFunc) {
	sm.OnSuccessHandlers[name] = handler
}

func (sm *StateMachine) AddExecuteFunction(handler HandlerExecFunction) {
	sm.Execute = handler
}

func (sm *StateMachine) runCheckFunction(handlers []string, obj any) error {
	for _, handler := range handlers {
		isIntFunc, funIntFunc, argIntFunc := checkIfFunctionWithArguments(handler)

		if isIntFunc {
			handler = funIntFunc
		}

		handlerFunc := sm.getCheckFunction(handler)

		_, err := handlerFunc(obj, argIntFunc...)

		if err != nil {
			return err
		}

	}
	return nil
}

func (sm *StateMachine) runOnErrorFunction(handlers []string, obj any) error {
	for _, handler := range handlers {
		handlerFunc := sm.getOnErrorFunction(handler)
		_, err := handlerFunc(obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sm *StateMachine) runOnSuccessFunction(handlers []string, obj any) error {
	for _, handler := range handlers {
		isFuncArg, nameFunc, argFunc := checkIfFunctionWithArguments(handler)

		if smTrigger, isSMTrigger := sm.StateMachinesToTrigger[nameFunc]; isSMTrigger {
			err := smTrigger.ProcessTransition(argFunc[0], argFunc[1], obj)
			if err != nil {
				return err
			}

		} else {
			if isFuncArg {
				handler = nameFunc
			}
			handlerFunc := sm.getCheckFunction(handler)

			_, err := handlerFunc(obj, argFunc...)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sm *StateMachine) AddStateMachineToTrigger(stateMachine IStateMachine) IStateMachine {
	sm.StateMachinesToTrigger[stateMachine.GetName()] = stateMachine

	return sm
}

func (sm *StateMachine) ProcessTransition(currentState, nextState string, obj any) error {

	// Get handlers
	handlers, exitTransition := sm.MapStates[currentState][nextState]
	if !exitTransition {
		return fmt.Errorf("don't exit transition")
	}

	err := sm.runCheckFunction(handlers.Check, obj)
	if err != nil {
		return err
	}

	success, err := sm.Execute(currentState, nextState, obj)
	if err != nil {
		return err
	}

	if success {
		err := sm.runOnSuccessFunction(handlers.OnSuccess, obj)
		if err != nil {
			return err
		}
	}

	defer func() {
		if err != nil {
			err := sm.runOnErrorFunction(handlers.OnError, obj)
			if err != nil {
				return
			}
		}
	}()

	return nil
}
