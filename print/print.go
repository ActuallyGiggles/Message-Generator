package print

import (
	"Message-Generator/stats"
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

var started string
var errorChannel chan error

func Page(title string) {
	print("\033[H\033[2J")
	if title == "Exited" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightRed)).WithFullWidth().Println(title)
	} else if title == "Started" {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgGreen)).WithFullWidth().Println(title)
	} else {
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithFullWidth().Println(title)
	}
	pterm.Println()
}

func Success(message string) {
	t := time.Now()
	pterm.Success.Println(message + "\n" + t.Format(time.RFC822))
	stats.Log(t, message)
}

func Error(message string) {
	t := time.Now()
	pterm.Error.Println(message + "\n" + t.Format(time.RFC822))
	stats.Log(t, message)

	errorChannel <- errors.New(message)
}

func Info(message string) {
	t := time.Now()
	pterm.Info.Println(message + "\n" + t.Format(time.RFC822))
}

func InfoNoTime(message string) {
	pterm.Info.Println(message)
}

func Warning(message string) {
	t := time.Now()
	pterm.Warning.Println(message + "\n" + t.Format(time.RFC822))
	stats.Log(t, message)

	errorChannel <- errors.New(message)
}

func ProgressBar(title string, total int) (pb *pterm.ProgressbarPrinter) {
	pb, _ = pterm.DefaultProgressbar.WithTotal(total).WithTitle(title).WithRemoveWhenDone(true).Start()
	return
}

func Spinner(title string) (sp *pterm.SpinnerPrinter) {
	sp, _ = pterm.DefaultSpinner.Start(title)
	return
}

func Started(text string, errorChan chan error) {
	t := time.Now()
	stats.Log(t, text)
	pterm.Info.Println(text)
	started = text
	errorChannel = errorChan
}

func TerminalInput(cancel context.CancelFunc) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := strings.ToLower(scanner.Text())

		fmt.Println()

		switch input {
		default:
			Info("Not a command")
		case "exit":
			cancel()
		case "force exit":
			os.Exit(0)
		case "clear":
			Page("Twitch Message Generator")
		case "help":
			t := fmt.Sprintln("[started] for when the program started")
			t += fmt.Sprintln("[help] for list of commands")
			t += fmt.Sprintln("[clear] to clear the screen")
			t += fmt.Sprintln("[exit] to exit the program")
			Info(t)
		case "started":
			Info(started)
		}
	}
}

func ClearScreen() {
	print("\033[H\033[2J")
}

func Table(data [][]string) {
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
}
