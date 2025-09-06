package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
	"github.com/toritoritori29/dodo-cli/src/openapi"
)

const (
	PrimaryColorLight = "#F3A200"
	PrimaryColorDark  = "#FFCC66"
	TextColorLight0   = "#000000"
	TextColorLight1   = "#333333"
	TextColorDark0    = "#FFFFFF"
	TextColorDark1    = "#DADADA"
	TextColorLightDim = "#666666"
	TextColorDarkDim  = "#888888"
	MaxWidth          = 100
)

type SearchArgs struct {
	query    []string // search query
	endpoint string   // server endpoint to search
	debug    bool     // enable debug mode
	format   string   // output format (tui, json)
}

// Implement LoggingConfig and PrinterConfig interface for UploadArgs.
func (opts *SearchArgs) DisableLogging() bool {
	return opts.format == FormatJSON
}

func (opts *SearchArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *SearchArgs) EnableColor() bool {
	return true
}

func (opts *SearchArgs) EnablePrinter() bool {
	return true
}

func CreateSearchCmd() *cobra.Command {
	opts := SearchArgs{}
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search for a string in the project files.",
		RunE: func(_ *cobra.Command, _ []string) error {
			printer := NewPrinter(ErrorLevel)
			env := NewEnvArgs()
			err := CheckArgsAndEnvForSearch(opts, env)
			if err != nil {
				return printer.HandleError(err)
			}
			printer = NewPrinterFromArgs(&opts)
			if err := searchCmdEntrypoint(opts, env); err != nil {
				return printer.HandleError(err)
			}
			return nil
		},
	}
	searchCmd.Flags().StringArrayVarP(&opts.query, "query", "q", nil, "Search query (for json output only)")
	searchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	searchCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/search/v1", "Server endpoint for search")
	searchCmd.Flags().StringVar(&opts.format, "format", FormatTUI, "Output format (tui, json)")
	return searchCmd
}

func CheckArgsAndEnvForSearch(args SearchArgs, env EnvArgs) error {
	if env.APIKey == "" {
		return errors.New("DODO_API_KEY is not set")
	}
	if args.endpoint == "" {
		return errors.New("no endpoint provided")
	}
	if args.format != FormatTUI && args.format != FormatJSON {
		return fmt.Errorf("unknown format: %s", args.format)
	}

	// TUI
	if args.format == FormatTUI && args.query != nil {
		return errors.New("you cannot provide query arguments in TUI mode")
	}

	// JSON
	if args.format == FormatJSON && args.debug {
		return errors.New("debug mode is not supported in json format")
	}
	return nil
}

func searchCmdEntrypoint(args SearchArgs, env EnvArgs) error {
	switch args.format {
	case FormatTUI:
		return executeSearchTUI(args, env)
	case FormatJSON:
		return executeSearchJSON(args, env)
	default:
		return fmt.Errorf("unknown format: %s", args.format)
	}
}

// JSON implementation.
func executeSearchJSON(args SearchArgs, env EnvArgs) error {
	endpoint, err := NewEndpoint(args.endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}
	query := strings.Join(args.query, " ")
	records, err := sendSearchRequest(&env, endpoint, query)
	if err != nil {
		return fmt.Errorf("failed to execute the search: %w", err)
	}

	output := openapi.SearchPostResponse{
		Records: records,
	}
	outputBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the search response: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(outputBytes))
	return nil
}

// TUI implementation.
func executeSearchTUI(args SearchArgs, env EnvArgs) error {
	model, err := initialModel(args, env)
	if err != nil {
		return fmt.Errorf("failed to initialize the model: %w", err)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run the program: %w", err)
	}
	return nil
}

type searchItem struct {
	title       string
	description string
	url         string
	domain      string
}

func (i searchItem) FilterValue() string { return i.title }

func (i searchItem) Title() string { return i.title }

func (i searchItem) Description() string { return i.description }

type model struct {
	textInput       textinput.Model
	textInputActive bool
	list            list.Model
	choices         []list.Item
	selected        map[int]struct{}
	errorMessage    string

	// configurations
	endpoint   Endpoint
	args       *SearchArgs
	envArgs    *EnvArgs
	listStyles list.Styles
}

func initialModel(args SearchArgs, env EnvArgs) (model, error) {
	endpoint, err := NewEndpoint(args.endpoint)
	if err != nil {
		return model{}, fmt.Errorf("invalid endpoint: %w", err)
	}
	listStyles := initListStyles()

	w, h, _ := term.GetSize(os.Stdout.Fd())
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = w

	// Set appropriate dimensions for the list
	items := []list.Item{}
	listDelegate := newSearchItemDelegate(false)
	l := list.New(items, listDelegate, w-2, h-4) // Width: 30, Height: 10
	l.Title = ""
	l.Styles = listStyles
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return model{
		textInput:       ti,
		textInputActive: true,
		list:            l,
		choices:         items,
		selected:        make(map[int]struct{}),
		errorMessage:    "",
		endpoint:        endpoint,
		envArgs:         &env,
		args:            &args,
		listStyles:      listStyles,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn
	if msg, ok := msg.(tea.KeyMsg); ok {
		m.errorMessage = ""
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEnter {
			return m.updateEnter(msg)
		}
		if msg.Type == tea.KeyEscape {
			return m.updateEscape()
		}

		switch msg.String() {
		case "/":
			return m.updateEscape()
		case "up":
			return m.updateKeyUpDown(msg)
		case "down":
			return m.updateKeyUpDown(msg)
		}
	}

	// Update both textInput and list
	var cmd tea.Cmd
	var listCmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m model) updateKeyUpDown(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn
	if m.textInputActive {
		m.textInputActive = false
		m.list.SetDelegate(newSearchItemDelegate(true))
		m.list.ResetSelected()
		m.textInput.Blur()
		return m, nil
	}
	// Update both textInput and list
	var cmd tea.Cmd
	var listCmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m model) updateEscape() (tea.Model, tea.Cmd) { //nolint: ireturn
	m.textInputActive = true
	m.list.SetDelegate(newSearchItemDelegate(false))
	m.textInput.Focus()
	return m, m.textInput.Cursor.BlinkCmd()
}

func (m model) updateEnter(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn
	if !m.textInputActive {
		// In case the text input is not active, it means the list is focused.
		selectedItem, ok := m.list.SelectedItem().(searchItem)
		if !ok {
			m.errorMessage = "No item is selected"
			return m, nil
		}
		err := openBrowser(selectedItem.url)
		if err != nil {
			m.errorMessage = fmt.Sprintf("Failed to open the browser: %s", err)
		}
		return m, nil
	}

	// In case the text input is active.
	m.list.SetDelegate(newSearchItemDelegate(true))
	m.textInput.Blur()
	query := m.textInput.Value()

	records, err := sendSearchRequest(m.envArgs, m.endpoint, query)
	if err != nil {
		m.errorMessage = fmt.Sprintf("Failed to execute the search: %s", err)
	}

	// Update the list
	items := make([]list.Item, len(records))
	for i, record := range records {
		var domain string
		parsedURL, err := url.Parse(record.Url)

		if err == nil {
			domain = parsedURL.Hostname()
		} else {
			domain = "unknown"
		}

		items[i] = searchItem{
			title:       record.Title,
			description: record.Contents,
			url:         record.Url,
			domain:      domain,
		}
	}
	m.list.SetItems(items)
	m.list.ResetSelected()
	m.textInputActive = false

	var cmd tea.Cmd
	var listCmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m model) View() string {
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: PrimaryColorLight, Dark: PrimaryColorDark}).
		Bold(true)
	text := fmt.Sprintf("%s %s\n\n", searchStyle.Render("Search >"), m.textInput.View())
	text += m.list.View() + "\n"

	if m.errorMessage != "" {
		text += m.listStyles.StatusBar.Render(fmt.Sprintf("Error: %s\n", m.errorMessage))
	} else {
		text += m.listStyles.StatusBar.Render(fmt.Sprintln("Press Enter to search, ↑/↓ to navigate, / to focus search field, and Ctrl-C exit."))
	}
	return text
}

// Implementations for the list item UI.
type searchItemDelegate struct {
	focused bool
	styles  searchItemDelegateStyles
}

func newSearchItemDelegate(focused bool) searchItemDelegate {
	return searchItemDelegate{
		focused: focused,
		styles:  newSearchItemStyles(),
	}
}

func (d searchItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// This function assumes that the item is of type searchItem.
	sitem, ok := item.(searchItem)
	if !ok {
		return
	}

	// params
	var (
		title       string
		description string
		hostname    string
	)

	width := min(m.Width(), MaxWidth) - 1
	lines := strings.Split(ansi.Hardwrap(sitem.description, width, false), "\n")
	if len(lines) > 2 {
		lines = lines[:2]
		lastLine := lines[1]
		lines[1] = lastLine[:len(lastLine)-5] + "..."
	}
	adjustedDescription := strings.Join(lines, "\n")

	isSelected := index == m.Index() && d.focused
	if isSelected {
		title = d.styles.SelectedTitle.Render(sitem.title)
		description = d.styles.SelectedDescription.Render(adjustedDescription)
		hostname = d.styles.SelectedHostname.Render(sitem.domain)
	} else {
		title = d.styles.Title.Render(sitem.title)
		description = d.styles.Description.Render(adjustedDescription)
		hostname = d.styles.Hostname.Render(sitem.domain)
	}
	fmt.Fprintf(w, "%s %s\n", title, hostname)
	fmt.Fprintf(w, "%s\n", description)
}

func (d searchItemDelegate) Height() int {
	// Return the height of each list item, including spacing
	return 4
}

func (d searchItemDelegate) Spacing() int {
	// Return the spacing between list items
	return 1
}

func (d searchItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	// Handle item-level updates if necessary
	return nil
}

func initListStyles() list.Styles {
	styles := list.DefaultStyles()
	styles.StatusBar = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	return styles
}

type searchItemDelegateStyles struct {
	Title               lipgloss.Style
	SelectedTitle       lipgloss.Style
	Hostname            lipgloss.Style
	SelectedHostname    lipgloss.Style
	Description         lipgloss.Style
	SelectedDescription lipgloss.Style
}

func newSearchItemStyles() searchItemDelegateStyles {
	return searchItemDelegateStyles{
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: TextColorLight0,
				Dark:  TextColorDark0,
			}).
			Bold(true).
			Padding(0, 0, 0, 2),
		SelectedTitle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: PrimaryColorLight,
				Dark:  PrimaryColorDark,
			}).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: PrimaryColorLight,
				Dark:  PrimaryColorDark,
			}).
			Padding(0, 0, 0, 1),
		Hostname: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: TextColorLightDim,
				Dark:  TextColorDarkDim,
			}).
			Padding(0, 0, 0, 1),
		SelectedHostname: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: TextColorLightDim,
				Dark:  TextColorDarkDim,
			}).
			Padding(0, 0, 0, 1),
		Description: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: TextColorLight1,
				Dark:  TextColorDark1,
			}).
			Padding(0, 0, 0, 2),
		SelectedDescription: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
				Light: TextColorLight0,
				Dark:  TextColorDark0,
			}).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: PrimaryColorLight,
				Dark:  PrimaryColorDark,
			}).
			Padding(0, 0, 0, 1),
	}
}

func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = errors.New("unsupported platform")
	}
	if err != nil {
		return errors.New("failed to open the document in the browser")
	}
	return nil
}
