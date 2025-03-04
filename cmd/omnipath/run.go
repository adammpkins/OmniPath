package omnipath

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

					// Enhanced environment variables for better color support
					env := append(os.Environ(),
						"FORCE_COLOR=1",
						"TERM=xterm-256color",
						"COLORTERM=truecolor",
						"COMPOSE_FORCE_COLOR=1",
						"DOCKER_COLOR=1")

					// For Laravel Sail specifically, add more Docker-related vars
					if strings.Contains(strings.ToLower(s.Name), "sail") {
						env = append(env,
							"DOCKER_BUILDKIT=1",
							"LS_COLORS=rs=0:di=01;34:ln=01;36:mh=00:pi=40;33:so=01;35:do=01;35:bd=40;33;01:cd=40;33;01:or=40;31;01:mi=00:su=37;41:sg=30;43:ca=00:tw=30;42:ow=34;42:st=37;44:ex=01;32:*.tar=01;31:*.tgz=01;31:*.arc=01;31:*.arj=01;31:*.taz=01;31:*.lha=01;31:*.lz4=01;31:*.lzh=01;31:*.lzma=01;31:*.tlz=01;31:*.txz=01;31:*.tzo=01;31:*.t7z=01;31:*.zip=01;31:*.z=01;31:*.dz=01;31:*.gz=01;31:*.lrz=01;31:*.lz=01;31:*.lzo=01;31:*.xz=01;31:*.zst=01;31:*.tzst=01;31:*.bz2=01;31:*.bz=01;31:*.tbz=01;31:*.tbz2=01;31:*.tz=01;31:*.deb=01;31:*.rpm=01;31:*.jar=01;31:*.war=01;31:*.ear=01;31:*.sar=01;31:*.rar=01;31:*.alz=01;31:*.ace=01;31:*.zoo=01;31:*.cpio=01;31:*.7z=01;31:*.rz=01;31:*.cab=01;31:*.wim=01;31:*.swm=01;31:*.dwm=01;31:*.esd=01;31:*.avif=01;35:*.jpg=01;35:*.jpeg=01;35:*.mjpg=01;35:*.mjpeg=01;35:*.gif=01;35:*.bmp=01;35:*.pbm=01;35:*.pgm=01;35:*.ppm=01;35:*.tga=01;35:*.xbm=01;35:*.xpm=01;35:*.tif=01;35:*.tiff=01;35:*.png=01;35:*.svg=01;35:*.svgz=01;35:*.mng=01;35:*.pcx=01;35:*.mov=01;35:*.mpg=01;35:*.mpeg=01;35:*.m2v=01;35:*.mkv=01;35:*.webm=01;35:*.webp=01;35:*.ogm=01;35:*.mp4=01;35:*.m4v=01;35:*.mp4v=01;35:*.vob=01;35:*.qt=01;35:*.nuv=01;35:*.wmv=01;35:*.asf=01;35:*.rm=01;35:*.rmvb=01;35:*.flc=01;35:*.avi=01;35:*.fli=01;35:*.flv=01;35:*.gl=01;35:*.dl=01;35:*.xcf=01;35:*.xwd=01;35:*.yuv=01;35:*.cgm=01;35:*.emf=01;35:*.ogv=01;35:*.ogx=01;35:*.aac=00;36:*.au=00;36:*.flac=00;36:*.m4a=00;36:*.mid=00;36:*.midi=00;36:*.mka=00;36:*.mp3=00;36:*.mpc=00;36:*.ogg=00;36:*.ra=00;36:*.wav=00;36:*.oga=00;36:*.opus=00;36:*.spx=00;36:*.xspf=00;36:",
							"CLICOLOR=1",
							"CLICOLOR_FORCE=1")
					}
					c.Env = env

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
						buffer := make([]byte, 1024)
						for {
							n, err := reader.Read(buffer)
							if err != nil {
								if err != io.EOF {
									log.Printf("Error reading stdout for %s: %v", s.Name, err)
								}
								break
							}
							if n > 0 {
								mu.Lock()
								session.Output += string(buffer[:n])
								mu.Unlock()
							}
						}
					}()

					// Read stderr concurrently.
					go func() {
						reader := bufio.NewReader(stderrPipe)
						buffer := make([]byte, 1024)
						for {
							n, err := reader.Read(buffer)
							if err != nil {
								if err != io.EOF {
									log.Printf("Error reading stderr for %s: %v", s.Name, err)
								}
								break
							}
							if n > 0 {
								mu.Lock()
								session.Output += string(buffer[:n])
								mu.Unlock()
							}
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
