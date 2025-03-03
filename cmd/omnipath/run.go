package omnipath

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	detect "github.com/adammpkins/OmniPath/internal/detect"
	"github.com/adammpkins/OmniPath/internal/tui"
	"github.com/adammpkins/OmniPath/internal/tui/multiplexer"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run selected service(s) interactively (if interactive) or in foreground (if non-interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		// Get services from detect.
		detectServices := detect.GetServices()
		if len(detectServices) == 0 {
			log.Println("No run commands detected. Please try running the project manually.")
			return
		}

		// Convert detect.Service to tui.Service.
		var allServices []tui.Service
		for _, ds := range detectServices {
			allServices = append(allServices, tui.Service{
				Name:        ds.Name,
				Command:     ds.Command,
				Interactive: ds.Interactive,
			})
		}

		var selectedServices []tui.Service
		// If more than one service is available, prompt for selection.
		if len(allServices) > 1 {
			selected, err := tui.RunMultiSelect(allServices)
			if err != nil {
				log.Fatalf("Error selecting service: %v", err)
			}
			if len(selected) == 0 {
				log.Println("No service selected.")
				return
			}
			selectedServices = selected
		} else {
			selectedServices = []tui.Service{allServices[0]}
		}

		// Split selected services into interactive and non-interactive.
		var interactiveServices []tui.Service
		var nonInteractiveServices []tui.Service
		for _, s := range selectedServices {
			if s.Interactive {
				interactiveServices = append(interactiveServices, s)
			} else {
				nonInteractiveServices = append(nonInteractiveServices, s)
			}
		}

		// Run non-interactive services in the foreground.
		for _, s := range nonInteractiveServices {
			log.Printf("Launching non-interactive service %s: %s\n", s.Name, s.Command)
			c := exec.Command("sh", "-c", s.Command)
			// Attach standard input/output so the command's output is visible.
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Stdin = os.Stdin
			c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := c.Run(); err != nil {
				log.Printf("Error running %s: %v", s.Name, err)
			}
		}

		// Launch interactive services using the multiplexer.
		if len(interactiveServices) > 0 {
			var sessions []*tui.Session // Use pointers for live updates.
			var wg sync.WaitGroup
			var mu sync.Mutex

			for _, s := range interactiveServices {
				wg.Add(1)
				go func(s tui.Service) {
					defer wg.Done()
					log.Printf("Launching interactive service %s: %s\n", s.Name, s.Command)
					c := exec.Command("sh", "-c", s.Command)
					c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

					stdoutPipe, err := c.StdoutPipe()
					if err != nil {
						log.Printf("Error obtaining stdout for %s: %v", s.Name, err)
						return
					}
					stderrPipe, err := c.StderrPipe()
					if err != nil {
						log.Printf("Error obtaining stderr for %s: %v", s.Name, err)
						return
					}
					stdinPipe, err := c.StdinPipe()
					if err != nil {
						log.Printf("Error obtaining stdin for %s: %v", s.Name, err)
						return
					}

					if err := c.Start(); err != nil {
						log.Printf("Error starting %s: %v", s.Name, err)
						return
					}

					session := &tui.Session{
						Name:   s.Name,
						Stdin:  stdinPipe,
						Output: "",
						Cmd:    c,
					}

					// Read stdout concurrently.
					go func() {
						reader := bufio.NewReader(stdoutPipe)
						for {
							line, err := reader.ReadString('\n')
							if err != nil {
								if err != io.EOF {
									log.Printf("Error reading stdout for %s: %v", s.Name, err)
								}
								break
							}
							mu.Lock()
							session.Output += line
							mu.Unlock()
						}
					}()
					// Read stderr concurrently.
					go func() {
						reader := bufio.NewReader(stderrPipe)
						for {
							line, err := reader.ReadString('\n')
							if err != nil {
								if err != io.EOF {
									log.Printf("Error reading stderr for %s: %v", s.Name, err)
								}
								break
							}
							mu.Lock()
							session.Output += line
							mu.Unlock()
						}
					}()

					mu.Lock()
					sessions = append(sessions, session)
					mu.Unlock()
				}(s)
			}
			wg.Wait()

			if len(sessions) == 0 {
				log.Fatalf("No interactive sessions available due to errors starting processes.")
			}

			if err := multiplexer.RunMultiplexer(sessions); err != nil {
				log.Fatalf("Error running multiplexer: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
