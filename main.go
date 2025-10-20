package main

import (
	"context"
	"os"
	"path"

	"github.com/sebastianrakel/bogo/types"
	
	"github.com/charmbracelet/huh"

	"go.yaml.in/yaml/v4"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/skratchdot/open-golang/open"
)

const VERSION = "0.1.0"

var config *types.Config

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}

	cmd := &cobra.Command{
		Use:   "bogo",
		Short: "cli bookmark manager",
		Version: VERSION,
	}

	cmd.AddCommand(cmdEntries())

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

func cmdEntriesOpen() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open",
		Short:   "open entry",
		Aliases: []string{"o"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, store := range config.Stores {
				var options []huh.Option[*types.Entry]
				entries, err := store.ListEntry()
				if err != nil {
					return err
				}

				for _, entry := range entries  {
					options = append(options, huh.NewOption[*types.Entry](entry.Title, &entry))
				}

				var selected *types.Entry

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[*types.Entry]().
							Title("Choose bookmark").
							Options(
								options...
							).
							Value(&selected)),
				)

				err = form.Run()
				if err != nil {
					return err
				}

				if (selected != nil) {
					err = open.Run(selected.Url)
					if err != nil {
						return err
					}					
				}
			}
			return nil
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
