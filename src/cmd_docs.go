package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"
)

type DocsArgs struct {
	debug    bool
	noColor  bool
	endpoint string
	format   string
}

// Implement LoggingConfig and PrinterConfig interface for TouchArgs.
func (opts *DocsArgs) DisableLogging() bool {
	return false
}

func (opts *DocsArgs) EnableDebugMode() bool {
	return opts.debug
}

func (opts *DocsArgs) EnableColor() bool {
	return !opts.noColor
}

func (opts *DocsArgs) EnablePrinter() bool {
	return true
}

func CreateDocsCmd() *cobra.Command {
	opts := DocsArgs{}
	docsCmd := &cobra.Command{
		Use:           "docs",
		Short:         "List the documentation in your organization",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return docsCmdEntrypoint(&opts)
		},
	}
	docsCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode if set this flag")
	docsCmd.Flags().BoolVar(&opts.noColor, "no-color", false, "Disable color output")
	docsCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/projects/v1", "The endpoint of the dodo API server")
	docsCmd.Flags().StringVar(&opts.format, "format", "tui", "Output format for the command. Supported formats: {tui, json}")
	return docsCmd
}

func docsCmdEntrypoint(opts *DocsArgs) error {
	env := NewEnvArgs()
	orgs, err := NewProjectFromAPI(env, opts.endpoint)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		return errors.New("no projects found in your organization")
	}

	switch opts.format {
	case "tui":
		return renderProjectsWithTUI(orgs)
	case "json":
		return renderProjectsWithJSON(orgs)
	default:
		return fmt.Errorf("unknown format: %s", opts.format)
	}
}

// TUI implementation.
type DocsTUIModel struct {
	list list.Model
}

func (m DocsTUIModel) Init() tea.Cmd {
	return nil
}

func (m DocsTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn
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

func (m DocsTUIModel) updateEnter() (tea.Model, tea.Cmd) { //nolint:ireturn
	selectedItem, ok := m.list.SelectedItem().(DocsOutput)
	if !ok {
		return m, tea.Quit
	}
	if err := openBrowser(selectedItem.URL); err != nil {
		return m, tea.Quit
	}
	return m, tea.Quit
}

func (m DocsTUIModel) View() string {
	return m.list.View()
}

func NewDocsTUIModel(orgs []Project) DocsTUIModel {
	items := make([]list.Item, 0, len(orgs))

	for _, org := range orgs {
		items = append(items, DocsOutput{
			Slug:        org.Slug,
			ProjectName: org.ProjectName,
			IsPublic:    org.IsPublic,
			ProjectID:   org.ProjectID,
			URL:         org.BaseURL,
		})
	}

	w, h, _ := term.GetSize(os.Stdout.Fd())
	l := list.New(items, list.NewDefaultDelegate(), w-1, h-1)
	l.Title = ""

	// l.Styles = listStyles
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.DisableQuitKeybindings()

	return DocsTUIModel{
		list: l,
	}
}

func renderProjectsWithTUI(orgs []Project) error {
	p := tea.NewProgram(NewDocsTUIModel(orgs))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run the program: %w", err)
	}
	return nil
}

// JSON implementation.
func renderProjectsWithJSON(orgs []Project) error {
	outputs := make([]DocsOutput, 0, len(orgs))
	for _, org := range orgs {
		outputs = append(outputs, DocsOutput{
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
type DocsOutput struct {
	Slug        string `json:"slug,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectID   string `json:"project_id"`
	IsPublic    bool   `json:"is_public,omitempty"`
	URL         string `json:"url,omitempty"`
}

func (d DocsOutput) Title() string {
	return d.ProjectName
}

func (d DocsOutput) Description() string {
	return d.URL
}

func (d DocsOutput) FilterValue() string {
	return d.ProjectName
}
