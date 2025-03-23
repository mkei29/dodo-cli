package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/caarlos0/log"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
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
}

func CreateSearchCmd() *cobra.Command {
	opts := SearchArgs{}
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search for a string in the project files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.query = args
			return executeSearch(cmd, opts)
		},
	}
	searchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	searchCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/search/v1", "Server endpoint to search")
	return searchCmd
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
	args       *SearchArgs
	envArgs    *EnvArgs
	listStyles list.Styles
}

func initialModel(args SearchArgs, env EnvArgs) model {
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
	l := list.New(items, newSearchItemDelegate(), w-2, h-4) // Width: 30, Height: 10
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
		envArgs:         &env,
		args:            &args,
		listStyles:      listStyles,
	}
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

		switch msg.String() {
		case "/":
			// If the / key is pressed, just focus the text input and don't propagate the key event to the list and text input
			m.textInputActive = true
			m.textInput.Focus()
			return m, m.textInput.Cursor.BlinkCmd()
		case "up":
			m.textInputActive = false
			m.textInput.Blur()
		case "down":
			m.textInputActive = false
			m.textInput.Blur()
		}
	}

	// Update both textInput and list
	var cmd tea.Cmd
	var listCmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m model) updateEnter(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint: ireturn
	if !m.textInputActive {
		selectedItem, ok := m.list.SelectedItem().(searchItem)
		if !ok {
			m.errorMessage = "No item selected"
			return m, nil
		}
		err := openBrowser(selectedItem.url)
		if err != nil {
			m.errorMessage = fmt.Sprintf("failed to open browser: %s", err)
		}
		return m, tea.Quit
	}

	m.textInput.Blur()
	query := m.textInput.Value()

	records, err := sendSearchRequest(m.envArgs, m.args.endpoint, query)
	if err != nil {
		m.errorMessage = fmt.Sprintf("failed to execute search: %s", err)
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

// Implementations for the list item UI
type searchItemDelegate struct {
	styles searchItemDelegateStyles
}

func newSearchItemDelegate() searchItemDelegate {
	return searchItemDelegate{
		styles: newSearchItemStyles(),
	}
}

func (d searchItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// This function assume that the item has searchItem type.
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
		lines[1] = fmt.Sprintf("%s...", lastLine[:len(lastLine)-5])
	}
	adjustedDescription := strings.Join(lines, "\n")

	isSelected := index == m.Index()
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
	// Return the height of each list item including spacing
	return 4
}

func (d searchItemDelegate) Spacing() int {
	// Return the spacing between list items
	return 1
}

func (d searchItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
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

func executeSearch(_ *cobra.Command, args SearchArgs) error {
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	env := NewEnvArgs()
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	p := tea.NewProgram(initialModel(args, env))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run the program: %w", err)
	}
	return nil
}

func sendSearchRequest(env *EnvArgs, uri, query string) ([]SearchRecord, error) {
	body := SearchPostRequest{
		Query: query,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create a new upload request from body: %w", err)
	}
	bearer := "Bearer " + env.APIKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a request to the server: %w", err)
	}
	defer resp.Body.Close()

	data := SearchPostResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}
	return data.Records, nil
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
