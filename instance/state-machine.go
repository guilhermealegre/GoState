package state_machine

import (
	"bitbucket.org/asadventure/be-infrastructure-lib/errors"
	"github.com/spf13/viper"
	"os"
	"strings"
	"regexp"
)


func NewStateMachine() IStateMachine {
	return &StateMachine{
		MapStates:              	make(map[string]map[string]Handlers),
		stateMachinesToTriggerMap: 	make(map[string]IStateMachine),
		CheckHandlers:          	make(map[string]HandlerFunc),
		OnSuccessHandlers:      	make(map[string]HandlerFunc),
		OnErrorHandlers:      		make(map[string]HandlerFunc),
		AdapterHandlers: 		    make(map[string]HandlerAdaptersFunction),
		FilterHandlers:         	make(map[string]HandlerFilterFunction),
	}
}

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

	// initialize the state machine
	for _, state := range sm.States {

		if sm.MapStates[state.Name] == nil {
			sm.MapStates[state.Name] = make(map[string]Handlers)
		}

		for _, transition := range state.Transitions {
			var handlers Handlers
			// add check handlers
			for _, check := range transition.Check{
				funcName, args := splitFunctionAndArguments(check.Func)
				handlers.Check = append(handlers.Check, CheckStruct{
					Func: funcName,
					FuncArg: args,
				})
			}
			// add on_success handlers
			for _, onSuccess := range transition.OnSuccess{
				funcName, args := splitFunctionAndArguments(onSuccess.Func)
				handlers.OnSuccess = append(handlers.OnSuccess, OnSuccessStruct{
					Func: funcName,
					FuncArg: args,
					Adapter: onSuccess.Adapter,
					Filter: onSuccess.Filter,
					IsStateMachine: onSuccess.IsStateMachine,
					IgnoreError: onSuccess.IgnoreError,
				})
			}
			// add on_error handlers
			for _, onError := range transition.OnError{
				funcName, args := splitFunctionAndArguments(onError.Func)
				handlers.OnError = append(handlers.OnError, OnErrorStruct{
					Func: funcName,
					FuncArg: args,
				})
			}

			sm.MapStates[state.Name][transition.Name] = handlers
		}
	}

	return nil
}


func (sm *StateMachine) GetName() string {
	return sm.Name
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
	sm.execute = handler
}

func (sm *StateMachine) AddStateMachineToTrigger(name string, stateMachine IStateMachine) IStateMachine {
	sm.stateMachinesToTriggerMap[name] = stateMachine
	return sm
}

func (sm *StateMachine) AddAdapterFunction(name string, handler HandlerAdaptersFunction) {
	sm.AdapterHandlers[name] = handler
}


func (sm *StateMachine) AddFilterFunction(name string, handler HandlerFilterFunction) {
	sm.FilterHandlers[name] = handler
}

func (sm *StateMachine) AddCurrentStateFunction(handler CurrenctStateFunc) {
	sm.currectState = handler
}


func (sm *StateMachine) ProcessTransition(nextState string, obj any) (err error) {

	// Get handlers
	currentState, err := sm.currectState(obj)
	if err != nil {
		return err
	}

	handlers, exitTransition := sm.MapStates[currentState][nextState]
	if !exitTransition {
		return errors.ErrorInStateMachineTransition().Formats(currentState, nextState, sm.Name)
	}

	err = sm.runCheckFunction(handlers.Check, obj)
	if err != nil {
		return err
	}

	err = sm.execute(nextState, obj)
	if err != nil {
		return err
	}

	err = sm.runOnSuccessFunction(handlers.OnSuccess, obj)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = sm.runOnErrorFunction(handlers.OnError, obj)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (sm *StateMachine) getCheckFunction(name string) HandlerFunc {
	// it's an internal function
	return sm.CheckHandlers[name]
}
func (sm *StateMachine) getOnErrorFunction(name string) HandlerFunc {
	return sm.OnErrorHandlers[name]
}

func (sm *StateMachine) getOnSuccessFunction(name string) HandlerFunc {
	return sm.OnSuccessHandlers[name]
}

func (sm *StateMachine) getAdapterFunction(name string) HandlerAdaptersFunction {
	return sm.AdapterHandlers[name]
}

func (sm *StateMachine) getFilterFunction(name string) HandlerFilterFunction {
	return sm.FilterHandlers[name]
}

func (sm *StateMachine) getStateMachineToTrigger(name string) IStateMachine {
	return sm.stateMachinesToTriggerMap[name]
}

func (sm *StateMachine) runCheckFunction(handlers []CheckStruct, obj any) error {
	for _, handler := range handlers {
		handlerFunc := sm.getCheckFunction(handler.Func)

		success, err := handlerFunc(obj, handler.FuncArg...)

		if !success {
			return errors.ErrorInStateMachineTransition()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (sm *StateMachine) runOnErrorFunction(handlers []OnErrorStruct, obj any) error {
	for _, handler := range handlers {
		handlerFunc := sm.getOnErrorFunction(handler.Func)
		_, err := handlerFunc(obj)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sm *StateMachine) runOnSuccessFunction(handlers []OnSuccessStruct, obj any) error {
	for _, handler := range handlers {
		objs := []any{obj}

		adapter := sm.getAdapterFunction(handler.Adapter)
		if adapter != nil {
			newObjs, err := adapter(obj)
			if err != nil {
				return err
			}
			objs = newObjs
		}

		filter := sm.getFilterFunction(handler.Filter)
		if filter != nil {
			newObjs, err := filter(objs)
			if err != nil {
				return err
			}
			objs = newObjs
		}

		for _, obj := range objs {
			if handler.IsStateMachine {
				smTrigger := sm.getStateMachineToTrigger(handler.Func)
				if smTrigger != nil {
					err := smTrigger.ProcessTransition(handler.FuncArg[1], obj)
					if err != nil && !handler.IgnoreError {
						return err
					}
				}
			} else {

				handlerFunc := sm.getOnSuccessFunction(handler.Func)
				_, err := handlerFunc(obj, handler.FuncArg...)
				if err != nil && !handler.IgnoreError {
					return err
				}
			}

		}
	}

	return nil
}

func splitFunctionAndArguments(input string) (function string, arguments []string) {
	pattern := regexp.MustCompile(`^\w+\((?:\w+(?:,\s*\w+)*)?\)$`)
	if !pattern.MatchString(input) {
		return input, nil
	}

	parts := strings.Split(input, "(")
	functionName := parts[0]
	argumentsString := parts[1]
	argumentsString = strings.TrimSuffix(argumentsString, ")")
	arguments = strings.Split(argumentsString, ",")
	for i, arg := range arguments {
		arguments[i] = strings.TrimSpace(arg)
	}

	return functionName, arguments
}