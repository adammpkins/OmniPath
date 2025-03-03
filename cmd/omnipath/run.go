package omnipath

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"sync"
	"syscall"

	"github.com/adammpkins/OmniPath/internal/detect"
	"github.com/adammpkins/OmniPath/internal/tui"
	"github.com/adammpkins/OmniPath/internal/tui/multiplexer"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run selected service(s) interactively in fallback non-PTY mode",
	Run: func(cmd *cobra.Command, args []string) {
		// Detect available services.
		services := detect.GetServices()
		if len(services) == 0 {
			log.Println("No run commands detected. Please try running the project manually.")
			return
		}

		var selectedServices []tui.Service
		// If there are multiple services, let the user select.
		if len(services) > 1 {
			var serviceList []tui.Service
			for _, svc := range services {
				serviceList = append(serviceList, tui.Service{
					Name:    svc.Name,
					Command: svc.Command,
				})
			}
			selected, err := tui.RunMultiSelect(serviceList)
			if err != nil {
				log.Fatalf("Error selecting service: %v", err)
			}
			if len(selected) == 0 {
				log.Println("No service selected.")
				return
			}
			selectedServices = selected
		} else {
			// For a single service, always force interactive mode.
			svc := services[0]
			selectedServices = []tui.Service{{
				Name:    svc.Name,
				Command: svc.Command,
			}}
		}

		// Launch all selected services interactively in fallback mode.
		var sessions []*tui.Session // Use pointers so updates propagate.
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, s := range selectedServices {
			wg.Add(1)
			go func(s tui.Service) {
				defer wg.Done()
				log.Printf("Launching %s interactively: %s\n", s.Name, s.Command)
				c := exec.Command("sh", "-c", s.Command)
				// Set process group so all children can be terminated.
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

				// Create a new session (as a pointer).
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
			log.Fatalf("No sessions available to run multiplexer due to errors starting processes.")
		}

		// Run the multiplexer UI with the sessions.
		if err := multiplexer.RunMultiplexer(sessions); err != nil {
			log.Fatalf("Error running multiplexer: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
