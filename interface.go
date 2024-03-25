package state_machine

type IStateMachine interface {
	GetName() string
	Load(filePath string) error
	ProcessTransition(nextState string, obj any) (success bool, err error)
	AddCheckFunction(name string, handler HandlerFunc)
	AddOnErrorFunction(name string, handler HandlerFunc)
	AddOnSuccessFunction(name string, handler HandlerFunc)
	AddExecuteFunction(handler HandlerExecFunction)
	AddCurrentStateFunction(handler CurrentStateFunc)
	AddStateMachineToTrigger(name string, stateMachine IStateMachine) IStateMachine
	AddAdapterFunction(name string, handler HandlerAdapterFunction)
	AddFilterFunction(name string, handler HandlerFilterFunction)
}
