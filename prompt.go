package gobump

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-tty"
)

type promptResult int

const (
	promptResultNone promptResult = iota
	promptResultPatch
	promptResultMinor
	promptResultMajor
)

func promptTarget(currentVersion, target string) (promptResult, error) {
	tty, err := tty.Open()
	if err != nil {
		return promptResultNone, err
	}
	defer tty.Close()

	candidates := []struct {
		name   string
		config Config
		result promptResult
	}{
		{"patch", Config{PatchDelta: 1}, promptResultPatch},
		{"minor", Config{MinorDelta: 1}, promptResultMinor},
		{"major", Config{MajorDelta: 1}, promptResultMajor},
	}

	items := make([]string, len(candidates))
	promptResults := make(map[int]promptResult)

	for i, c := range candidates {
		newVersion, err := c.config.bumpedVersion(currentVersion)
		if err != nil {
			return promptResultNone, err
		}
		items[i] = fmt.Sprintf("%s (%s -> %s)", c.name, currentVersion, newVersion)
		promptResults[i] = c.result
	}

	prompt := promptui.Select{
		Label:    "Bump up " + target,
		HideHelp: true,
		Items:    items,
		Stdin:    tty.Input(),
		Stdout:   tty.Output(),
	}

	index, _, err := prompt.Run()

	if err != nil {
		return promptResultNone, err
	}

	return promptResults[index], nil
}
