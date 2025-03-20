package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/caarlos0/log"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

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

type item struct {
	title       string
	description string
	url         string
}

func (i item) FilterValue() string { return i.title }

func (i item) Title() string { return i.title }

func (i item) Description() string { return i.description }

type model struct {
	textInput textinput.Model
	list      list.Model
	choices   []list.Item
	selected  map[int]struct{}
	args      *SearchArgs
	envArgs   *EnvArgs
	// configurations
}

func initialModel(args SearchArgs, env EnvArgs) model {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	// Set appropriate dimensions for the list
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 50, 40) // Width: 30, Height: 10
	l.Title = ""
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return model{
		textInput: ti,
		list:      l,
		choices:   items,
		selected:  make(map[int]struct{}),
		envArgs:   &env,
		args:      &args,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEnter {
			return m.updateEnter(msg)
		}
		switch msg.String() {
		case "/":
			m.textInput.Focus()
			query := m.textInput.Value()
			if query != "" {
				m.list.SetItems(m.choices)
			}
			m.list.CursorUp()
		case "down":
			m.list.CursorDown()
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

func (m model) updateEnter(msg tea.Msg) (tea.Model, tea.Cmd) {
	query := m.textInput.Value()
	records, err := sendSearchRequest(m.envArgs, m.args.endpoint, query)
	if err != nil {
		log.Errorf("failed to execute search: %s", err)
	}

	// Update the list
	items := make([]list.Item, len(records))
	for i, record := range records {
		items[i] = item{
			title:       record.Title,
			description: record.Contents,
			url:         record.Url,
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
	return fmt.Sprintf(
		"Search: %s\n%s\n",
		m.textInput.View(),
		docStyle.Render(m.list.View()),
	)
}

func executeSearch(cmd *cobra.Command, args SearchArgs) error {
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
		fmt.Printf("Error starting program: %s\n", err)
		return err
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
