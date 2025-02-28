# OmniPath

OmniPath is a smart directory-based automation tool for developers. It streamlines common tasks such as navigating to your Git repository, serving project documentation, opening dependency docs, and even auto-detecting & running your project. With a clean, dark-themed interface for your README, OmniPath brings all your project-related documentation and operations into one unified CLI.

## Features

- **repo**: Opens your project's Git remote (e.g., GitHub repository) in your default browser.
- **readme**: Serves your project's `README.md` as an HTML page with a dark theme and smooth fonts.
- **docs**: Detects dependencies (for example, via a `composer.json` in Laravel apps) and opens the corresponding dependency documentation in your browser. If multiple dependencies are found, you'll be prompted to select one.
- **run**: Auto-detects the project type (Go, JavaScript, Python, etc.) and runs the appropriate command for the project.


## File Structure

    OmniPath/
    ├── go.mod
    ├── cmd/
    │   └── omnipath/
    │       ├── main.go         // Main entry point (ideally moved to a separate cmd/main for production)
    │       ├── root.go         // Sets up the Cobra CLI
    │       ├── repo.go         // Command to open the Git repository
    │       ├── docs.go         // Command to open dependency documentation
    │       ├── readme.go       // Command to serve README.md as HTML
    │       └── run.go          // Command to auto-detect and run the project
    └── internal/
        ├── browser/
        │   └── browser.go      // Functions to open URLs in the default browser (cross-platform)
        ├── git/
        │   └── remote.go       // Functions to retrieve and parse Git remote URLs
        ├── docs/
        │   └── local.go        // Serves README.md as HTML with dark styling
        ├── detect/
        │   └── project.go      // Detects the project type and executes the run command
        └── dependency/
            └── dependency.go   // Detects project dependencies and returns documentation URLs
## Installation

1. **Clone the repository:**

   ~~~bash
   git clone https://github.com/adammpkins/OmniPath.git
   cd OmniPath
   ~~~

2. **Install dependencies:**

   Ensure you have [Go](https://golang.org/dl/) (version 1.20 or later) installed. Then, run:

   ~~~bash
   go mod tidy
   ~~~

3. **Build the project:**

   ~~~bash
   go build -o omnipath ./cmd/omnipath
   ~~~

   Alternatively, you can run it directly:

   ~~~bash
   go run ./cmd/main/main.go [command]
   ~~~

   > **Note:** For production, consider moving your `main.go` to a separate directory (e.g., `cmd/main/`) so that all CLI library code remains in the `cmd/omnipath` package.

## Usage

OmniPath provides several subcommands:

- **Open Repository:**

  ~~~bash
  omnipath repo
  ~~~

  Opens your project's Git remote URL in your default web browser.

- **Serve README:**

  ~~~bash
  omnipath readme
  ~~~

  Serves the `README.md` file from the project root as an HTML page with dark styling. It automatically opens your default browser to display the content.

- **Open Dependency Documentation:**

  ~~~bash
  omnipath docs
  ~~~

  Detects dependencies (for example, via a `composer.json` for Laravel apps) and opens the corresponding dependency documentation in your browser. If multiple dependencies are detected, you'll be prompted to select one.

- **Run Project:**

  ~~~bash
  omnipath run
  ~~~

  Auto-detects your project's type and executes the appropriate run command (e.g., `go run .`, `npm start`, or `python main.py`).

## Customization

- **Styling the README:**  
  The dark theme and smooth font styling for your README are defined in the HTML template within `internal/docs/local.go`. Feel free to tweak the CSS to match your preferences.

- **Extending Dependency Detection:**  
  Modify `internal/dependency/dependency.go` to add support for more dependencies and their corresponding documentation URLs.

## Contributing

Contributions are welcome! Please fork the repository, make your changes, and submit a pull request. Open an issue if you have any questions or suggestions.

## License

This project is licensed under the [MIT License](LICENSE).

---

Happy coding with OmniPath!