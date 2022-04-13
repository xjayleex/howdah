package test

import "howdah/pb"

type AgentCommand pb.AgentCommand

func HandleCommand(cmd *AgentCommand) {
	switch c := cmd.Command.(type) {
	case *pb.AgentCommand_ExecutionCommand:
	case *pb.AgentCommand_PackageCommand:
	case *pb.AgentCommand_RegistrationCommand:
	}
}

func handlePackageCommand (command *pb.AgentCommand_PackageCommand) {
	command.PackageCommand.
}