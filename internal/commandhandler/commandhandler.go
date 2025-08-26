package commandhandler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	bt "github.com/dmsosa6/bittorrent-client/internal/bittorrent"

	session "github.com/dmsosa6/bittorrent-client/internal/session"
)

type Command int

type CommandHelp struct {
	Description string
	Usage       string
}

var commandHelp = map[Command]CommandHelp{
	Info: {
		"Display information about a torrent",
		"info <infohash|name>",
	},
	Announce: {
		"Add a torrent from a .torrent file",
		"announce /path/to/file.torrent",
	},
	List: {
		"List all torrents",
		"list",
	},
	Exit: {
		"Exit the client",
		"exit",
	},
	Help: {
		"Show this help message",
		"help",
	},
}

const (
	Unknown Command = iota
	Announce
	Info
	List
	Load
	Help
	Exit
)

var commandArgs = map[Command][]int{
	Unknown:  {0},
	Announce: {1},
	Info:     {1},
	List:     {0},
	Help:     {0, 1},
	Exit:     {0},
	Load:     {1},
}

var commandLookup = map[string]Command{
	"info":     Info,
	"exit":     Exit,
	"announce": Announce,
	"help":     Help,
	"list":     List,
	"load":     Load,
}

type CommandHelp struct {
	Description string
	Usage       string
}

var commandHelp = map[Command]CommandHelp{
	Info: {
		"Display information about a torrent",
		"info <infohash|name>",
	},
	Announce: {
		"Add a torrent from a .torrent file",
		"announce /path/to/file.torrent",
	},
	List: {
		"List all torrents",
		"list",
	},
	Exit: {
		"Exit the client",
		"exit",
	},
	Help: {
		"Show this help message",
		"help",
	},
}

var bencoder = bt.BEncoding{}

func (c Command) String() string {
	switch c {
	case Info:
		return "info"
	case List:
		return "list"
	case Exit:
		return "exit"
	case Announce:
		return "announce"
	case Load:
		return "load"
	case Help:
		return "help"
	default:
		return "unknown"
	}
}

type Handler struct {
	CurrentTorrent *bt.Torrent
}

func (r *Handler) ParseCommand(s string) Command {

	if cmd, ok := commandLookup[strings.ToLower(s)]; ok {
		return cmd
	}

	return Unknown
}

func (r *Handler) ExecuteCommand(command Command, args []string, s session.Session) {
	var err error

	switch command {
	case Info:
		err = r.info(args, s)
		break
	case List:
		err = r.list(s)
		break
	case Announce:
		err = r.announce(args, s)
		break
	case Help:
		err = r.help(args)
		break
	case Load:
		err = r.load(args, s)
		break
	default:
		fmt.Println("Unkown command. type \"help\"")
	}

	if err != nil {
		handleError(err)
	}
}

func (r *Handler) help(args []string) error {
	if len(args) == 1 {
		commandString := args[0]
		command := r.ParseCommand(commandString)

		if command == Unknown {
			return fmt.Errorf("%s is a unknown command, type 'help' for a list of commands", commandString)
		}

		fmt.Printf("%s usage: %s", command, commandHelp[command])
	}

	fmt.Println("Available commands:")
	for cmd, h := range commandHelp {
		fmt.Printf("  %-10s - %s\n", cmd.String(), h.Description)
	}

	return nil
}

func (r *Handler) list(s session.Session) error {
	if len(s.Torrents) == 0 {
		return errors.New("no torrents to show")
	}

	fmt.Println("List of torrents:")

	for _, v := range s.Torrents {
		fmt.Printf(v.HexStringInfohash())
	}

	return nil
}

func (r *Handler) info(args []string, s session.Session) error {
	key := args[0]

	torrent, ok := s.Torrents[key]

	if !ok {
		return fmt.Errorf("no torrents with infohash: %s\n", key)
	}

	fmt.Println(torrent.Details())

	return nil
}

// for now this expects a .torrent file, in the future enforce this better or make it so u can create torrents
func (r *Handler) load(args []string, s session.Session) error {

	path := args[0]

	buf, err := filePathToBytes(path)

	if err != nil {
		return err
	}

	torrent, err := bencoder.DecodeTorrent(buf)

	if err != nil {
		return err
	}

	s.SetCurrTorrent(torrent)

	return nil
}

func (r *Handler) announce(args []string, s session.Session) error {

	return nil
}

// private
// If needed use a logger, probabily not
func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
	}
}

// this expects a absolute path
// there is a single os call for the second part
func filePathToBytes(path string) ([]byte, error) {

	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	bs := make([]byte, stat.Size())

	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return bs, nil
}

func validateArgs(command Command, args []string) error {
	expected := commandArgs[command]
	if !slices.Contains(expected, len(args)) {
		return fmt.Errorf(
			"wrong number of arguments for %s: got %d, expected one of %v",
			command, len(args), expected,
		)
	}
	return nil
}
