package types

import (
	"errors"
	"os"

	"github.com/sebastianrakel/bogo/helper"
	"go.yaml.in/yaml/v4"
)

type LocalStore struct {
	entries []Entry `yaml:"-"`
	Path string `yaml:"path"`
}

func (l LocalStore) GetPath() string {
	return helper.ReplaceTilde(l.Path)
}

func (l *LocalStore) load() error {
	_, err := os.Stat(l.GetPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	
	data, err := os.ReadFile(l.GetPath())
	if err != nil {
		return err
	}

	var entries []Entry
	err = yaml.Unmarshal(data, &entries)
	if err != nil {
		return err
	}
	
	l.entries = entries

	return nil
}

func (l *LocalStore) save() error {
	data, err := yaml.Marshal(l.entries)
	if err != nil {
		return err

	}
	
	err = os.WriteFile(l.GetPath(), data, 0600)
	return err
}

func (l LocalStore) ListEntry() ([]Entry, error) {
	err := l.load()
	if err != nil {
		return nil, err
	}

	return l.entries, nil
}

func (l LocalStore) EntryAdd(entry *Entry) error {
	err := l.load()
	if err != nil {
		return err
	}
	
	l.entries = append(l.entries, *entry)

	return l.save()
}
