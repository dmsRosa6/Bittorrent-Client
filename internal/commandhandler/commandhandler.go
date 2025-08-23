package commandhandler

import (
	"fmt"
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
	comm := strings.ToLower(s)

	switch comm {
    case Info.String():
        return Info
    case Trackers.String():
        return Trackers
    case Exit.String():
        return Exit
	case Announce.String():
        return Announce
	case Help.String():
        return Help
    default:
        return Unknown
    }
}

func (r *Handler) ExecuteCommand(command Command){

	switch command {
    case Info:
	
		break
	case Trackers:

		break
	case Exit:
    
		break
	case Announce:
        
		break
	case Help:
        r.help()
		break
    default:
        fmt.Println("Unkown command. type \"help\"")
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
