package commandhandler

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	bt "github.com/dmsRosa6/bittorrent-client/internal/bittorrent"
)

type Command int
const (
    Unknown Command = iota
    Announce
	Info
    Trackers
	Help
    Exit
)

var commandArgs = map[Command]int{
	Unknown : 0,
	Announce : 1,
	Info : 0,
	Trackers : 0,
	Help : 0,
	Exit : 0,
}

var commandLookup = map[string]Command{
    "info":     Info,
    "trackers": Trackers,
    "exit":     Exit,
    "announce": Announce,
    "help":     Help,
}

var bencoder = bt.BEncoding{}

func (c Command) String() string {
    switch c {
    case Info:
        return "info"
    case Trackers:
        return "trackers"
    case Exit:
        return "exit"
	case Announce:
        return "announce"
    case Help:
        return "help"
    default:
        return "unknown"
    }
}


type Handler struct {
	CurrentTorrent *bt.Torrent
} 

func (r *Handler) ParseCommand(s string) Command{
	if cmd, ok := commandLookup[strings.ToLower(s)]; ok {
        return cmd
    }

    return Unknown
}

func (r *Handler) ExecuteCommand(command Command, args []string){
	var err error

	switch command {
    case Info:
	
		break
	case Trackers:

		break
	case Exit:
    
		break
	case Announce:
        err = r.announce(args)
		break
	case Help:
        r.help()
		break
    default:
        fmt.Println("Unkown command. type \"help\"")
    }

	if err != nil {
		handleError(err)
	}
}

func (r *Handler) help() {
	fmt.Println("Available commands:")
	fmt.Println("  info       - Display information about the currently loaded torrent")
	fmt.Println("  trackers   - List all trackers of the currently loaded torrent")
	fmt.Println("  announce   - (Optional) Send announce to trackers") 
	fmt.Println("  help       - Show this help message")
	fmt.Println("  exit       - Exit the Handler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Type the command name and press Enter, e.g.:")
	fmt.Println("    info")
	fmt.Println("    trackers")
}

// for now this expects a .torrent file, in the future enforce this better or make it so u can create torrents
func (r *Handler) announce(args []string) error{
	
	if len(args) != 1 {
		return fmt.Errorf("wrong number of arguments: got %d, expected %d", len(args), commandArgs[Announce])
	}

	path := args[0]

	buf, err := filePathToBytes(path)

	if err != nil{
		return err
	}

	torrent, err := bencoder.DecodeTorrent(buf)

	if err != nil{
		return err
	}

	return nil
}


// private 

func handleError(error error) {

}

// this expects a absolute path
// there is a single os call for the second part
func filePathToBytes(path string) ([]byte, error){

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
