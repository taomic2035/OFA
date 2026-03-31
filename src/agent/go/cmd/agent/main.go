package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ofa/agent/internal/client"
	"github.com/ofa/agent/internal/executor"
	pb "github.com/ofa/agent/proto"
)

func main() {
	// Load configuration
	centerAddr := os.Getenv("CENTER_ADDR")
	if centerAddr == "" {
		centerAddr = "localhost:9090"
	}

	agentName := os.Getenv("AGENT_NAME")
	if agentName == "" {
		agentName = "test-agent"
	}

	// Create executor with built-in skills
	exec := executor.NewExecutor()
	exec.RegisterSkill(&executor.TextProcessSkill{})
	exec.RegisterSkill(&executor.JSONProcessSkill{})
	exec.RegisterSkill(&executor.CalculatorSkill{})
	exec.RegisterSkill(&executor.EchoSkill{})

	// Create agent client
	cfg := &client.Config{
		CenterAddr:   centerAddr,
		Name:         agentName,
		Type:         pb.AgentType_AGENT_TYPE_FULL,
		DeviceInfo:   getDeviceInfo(),
		Capabilities: exec.GetCapabilities(),
		TaskHandler: func(ctx context.Context, task *pb.TaskAssignment) (*pb.TaskResult, error) {
			log.Printf("Executing task: %s (skill: %s)", task.TaskId, task.SkillId)
			return exec.Execute(ctx, task)
		},
		MessageHandler: func(ctx context.Context, msg *pb.Message) error {
			log.Printf("Received message: %s from %s (action: %s)", msg.MsgId, msg.FromAgent, msg.Action)
			return nil
		},
	}

	agent, err := client.NewAgentClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Connect to Center
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := agent.Connect(ctx, cfg); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	log.Printf("Agent started: %s", agent.GetAgentID())

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down agent...")
	agent.Disconnect()
	log.Println("Agent stopped")
}

func getDeviceInfo() *pb.DeviceInfo {
	return &pb.DeviceInfo{
		Os:           "Linux",
		OsVersion:    "5.15.0",
		Model:        "Desktop",
		Manufacturer: "Generic",
		TotalMemory:  16 * 1024 * 1024 * 1024,
		CpuCores:     8,
		Arch:         "x86_64",
	}
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}