package main

import (
	"context"
	"os"
	"path"

	"github.com/sebastianrakel/bogo/types"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"go.yaml.in/yaml/v4"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/fang"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

const VERSION = "0.1.0"

var config *types.Config

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:     "bogo",
		Short:   "cli bookmark manager",
		Version: VERSION,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdEntriesOpen().Execute()
		},
	}

	cmd.AddCommand(cmdEntries())
	cmd.AddCommand(cmdStores())

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}

func loadConfig() error {
	configFilePath := path.Join(xdg.ConfigHome, "bogo", "config.yaml")

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	return nil
}

func cmdStores() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stores",
		Short:   "manage stores",
		Aliases: []string{"s"},
	}

	cmd.AddCommand(cmdStoresList())

	return cmd
}

func cmdEntries() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entry",
		Short:   "manage entries",
		Aliases: []string{"e"},
	}

	cmd.Flags().String("store", "", "which store should be used")

	cmd.AddCommand(cmdEntriesAdd())
	cmd.AddCommand(cmdEntriesOpen())

	return cmd
}

func listEntries() error {
	var options []huh.Option[*types.Entry]
	for _, store := range config.Stores {
		entries, err := store.ListEntry()
		if err != nil {
			return err
		}

		for _, entry := range entries {
			options = append(options, huh.NewOption[*types.Entry](entry.GetTitle(), &entry))
		}

	}

	var selected *types.Entry
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*types.Entry]().
				Title("Choose bookmark").
				Options(
					options...,
				).
				Value(&selected)),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	if selected != nil {
		err = open.Run(selected.Url)
		if err != nil {
			return err
		}
	}
	return nil
}

func cmdEntriesOpen() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open",
		Short:   "open entry",
		Aliases: []string{"o"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listEntries()
		},
	}

	return cmd
}

func cmdEntriesAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "create entry",
		Aliases: []string{"c"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var selectedStore *types.LocalStore
			if len(config.Stores) == 1 {
				for _, store := range config.Stores {
					selectedStore = &store
				}
			}
			title, err := cmd.Flags().GetString("title")
			if err != nil {
				return err
			}

			tags, err := cmd.Flags().GetStringArray("tags")
			if err != nil {
				return err
			}

			entry := types.Entry{
				Title: title,
				Url:   args[0],
				Tags:  tags,
			}

			err = selectedStore.EntryAdd(&entry)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("title", "", "title for entry")
	cmd.Flags().StringArray("tags", []string{}, "tags for the entry")

	return cmd
}

type item struct {
	title string
	path  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.path }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func cmdStoresList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list stores",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []list.Item{}

			for storename, store := range config.Stores {
				items = append(items, item{title: storename, path: store.Path})
			}

			m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
			m.list.Title = "Stores"
			p := tea.NewProgram(m, tea.WithAltScreen())

			_, err := p.Run()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
