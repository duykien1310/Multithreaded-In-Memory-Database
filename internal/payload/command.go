package payload

type Command struct {
	Cmd  string
	Args []string
}

type Task struct {
	Command *Command
	ReplyCh chan []byte
}
