package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/log"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

type DocArgs struct {
	debug    bool
	noColor  bool
	endpoint string
	format   string
}

// Implement LoggingConfig and PrinterConfig interface for DocsArgs.
func (opts *DocArgs) DisableLogging() bool {
	return opts.format == FormatJSON
}

func (opts *DocArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *DocArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *DocArgs) EnablePrinter() bool {
	return true
}

func CreateDocCmd() *cobra.Command {
	opts := DocArgs{}
	docsCmd := &cobra.Command{
		Use:          "doc",
		Short:        "List the documentation in your organization",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			printer := NewPrinter(ErrorLevel)
			env := NewEnvArgs()
			err := CheckArgsAndEnvForDocs(&opts, &env)
			if err != nil {
				printer.PrintError(err)
				return err
			}

			if err := docCmdEntrypoint(&opts, &env); err != nil {
				printer.PrintError(err)
				return err
			}
			return nil
		},
	}
	docsCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	docsCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	docsCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/projects/v1", "The endpoint of the dodo API server")
	docsCmd.Flags().StringVar(&opts.format, "format", FormatTUI, "Output format for the command. Supported formats: {tui, json}")

	return docsCmd
}

func docCmdEntrypoint(args *DocArgs, env *EnvArgs) error {
	if err := InitLogger(args); err != nil {
		return err
	}

	log.Debugf("Sending request to the %s", args.endpoint)
	orgs, err := NewProjectFromAPI(env, args.endpoint)
	if err != nil {
		return err
	}
	log.Debugf("Received %d projects from the server", len(orgs))

	if len(orgs) == 0 {
		return errors.New("no projects found in your organization. Please create a project first: https://www.dodo-doc.com/")
	}

	switch args.format {
	case FormatTUI:
		return renderProjectsWithTUI(orgs)
	case FormatJSON:
		return renderProjectsWithJSON(orgs)
	default:
		return fmt.Errorf("unknown format: %s", args.format)
	}
}

func CheckArgsAndEnvForDocs(args *DocArgs, env *EnvArgs) error {
	if env.APIKey == "" {
		return errors.New("DODO_API_KEY environment variable is not set")
	}
	if args.endpoint == "" {
		return errors.New("endpoint is not set")
	}
	if args.format != FormatTUI && args.format != FormatJSON {
		return fmt.Errorf("unknown format: %s", args.format)
	}

	if args.format == "json" && args.debug {
		return errors.New("debug mode is not supported in json format")
	}

	if args.format == "tui" && args.noColor {
		return errors.New("no-color options is not supported in tui format")
	}
	return nil
}

// TUI implementation.
type DocTUIModel struct {
	list list.Model
}

func (m DocTUIModel) Init() tea.Cmd {
	return nil
}

func (m DocTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEnter {
			return m.updateEnter()
		}
	}

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	return m, listCmd
}

func (m DocTUIModel) updateEnter() (tea.Model, tea.Cmd) { //nolint:ireturn
	selectedItem, ok := m.list.SelectedItem().(DocOutput)
	if !ok {
		return m, tea.Quit
	}
	if err := openBrowser(selectedItem.URL); err != nil {
		return m, tea.Quit
	}
	return m, tea.Quit
}

func (m DocTUIModel) View() string {
	return m.list.View()
}

func NewDocTUIModel(orgs []Project) DocTUIModel { //nolint:funlen
	items := make([]list.Item, 0, len(orgs))

	for _, org := range orgs {
		items = append(items, DocOutput{
			Slug:        org.Slug,
			ProjectName: org.ProjectName,
			IsPublic:    org.IsPublic,
			ProjectID:   org.ProjectID,
			URL:         org.BaseURL,
		})
	}

	w, h, _ := term.GetSize(os.Stdout.Fd())
	delegate := list.NewDefaultDelegate()

	// Unselected Styles
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: TextColorLight0, Dark: TextColorDark0}).
		Padding(0, 0, 0, 2)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: TextColorLightDim, Dark: TextColorDarkDim}).Padding(0, 0, 0, 2)

	// Selected Styles
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: PrimaryColorLight, Dark: PrimaryColorDark,
		}).
		Foreground(lipgloss.AdaptiveColor{
			Light: PrimaryColorLight,
			Dark:  PrimaryColorDark,
		}).
		Bold(true).Padding(0, 0, 0, 1)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: PrimaryColorLight, Dark: PrimaryColorDark,
		}).
		Foreground(lipgloss.AdaptiveColor{
			Light: TextColorLight0,
			Dark:  TextColorDark0,
		}).Padding(0, 0, 0, 1)

	// Other Styles
	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: TextColorLight1, Dark: TextColorDark1}).
		Padding(0, 0, 0, 2)

	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: TextColorLightDim, Dark: TextColorDarkDim}).Padding(0, 0, 0, 2)

	l := list.New(items, delegate, w-1, h-1)
	l.Title = ""

	// Overwrite the filter input styles.
	filterInput := textinput.New()
	filterInput.Prompt = "Filter: "
	filterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: PrimaryColorLight, Dark: PrimaryColorDark}).Bold(true)
	filterInput.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: TextColorLight0, Dark: TextColorDark0})
	filterInput.CharLimit = 64
	l.FilterInput = filterInput

	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.DisableQuitKeybindings()

	return DocTUIModel{
		list: l,
	}
}

func renderProjectsWithTUI(orgs []Project) error {
	p := tea.NewProgram(NewDocTUIModel(orgs))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run the program: %w", err)
	}
	return nil
}

// JSON implementation.
func renderProjectsWithJSON(orgs []Project) error {
	outputs := make([]DocOutput, 0, len(orgs))
	for _, org := range orgs {
		outputs = append(outputs, DocOutput{
			Slug:        org.Slug,
			ProjectName: org.ProjectName,
			IsPublic:    org.IsPublic,
			ProjectID:   org.ProjectID,
			URL:         org.BaseURL,
		})
	}
	b, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the output: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
	return nil
}

// Utility that describes the list item for TUI and JSON output.
type DocOutput struct {
	Slug        string `json:"slug,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectID   string `json:"project_id"`
	IsPublic    bool   `json:"is_public,omitempty"`
	URL         string `json:"url,omitempty"`
}

func (d DocOutput) Title() string {
	return d.ProjectName
}

func (d DocOutput) Description() string {
	return d.URL
}

func (d DocOutput) FilterValue() string {
	return d.ProjectName
}
