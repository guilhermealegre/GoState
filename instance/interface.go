package state_machine

type IStateMachine interface {
	GetName() string
	Load(filePath string) error
	ProcessTransition(nextState string, obj any) error
	AddCheckFunction(name string, handler HandlerFunc)
	AddOnErrorFunction(name string, handler HandlerFunc)
	AddOnSuccessFunction(name string, handler HandlerFunc)
	AddExecuteFunction(handler HandlerExecFunction)
	AddCurrentStateFunction(handler CurrenctStateFunc)
	AddStateMachineToTrigger(name string, stateMachine IStateMachine) IStateMachine
	AddAdapterFunction(name string, handler HandlerAdaptersFunction)
	AddFilterFunction(name string, handler HandlerFilterFunction)
}
