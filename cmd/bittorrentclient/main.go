package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	handler "github.com/dmsRosa6/bittorrent-client/internal/commandhandler"
)

func main() {
	r := &handler.Handler{}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("BitTorrent Client. Type 'help' for commands, 'exit' to quit.")

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		cmd := r.ParseCommand(input)

		if cmd == handler.Exit {
			fmt.Println("Exiting. Bye!")
			break
		}

		r.ExecuteCommand(cmd)
	}
}
