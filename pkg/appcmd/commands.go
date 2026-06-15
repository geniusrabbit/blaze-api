package appcmd

// ICommands is a list of commands
type ICommands []ICommand

// Get returns a command by name
func (c ICommands) Get(name string) ICommand {
	for _, cmd := range c {
		if cmd.Cmd() == name {
			return cmd
		}
	}
	return nil
}
