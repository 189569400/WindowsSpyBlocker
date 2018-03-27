package firewall

import (
	"github.com/crazy-max/WindowsSpyBlocker/app/menu"
	"github.com/crazy-max/WindowsSpyBlocker/app/utils/data"
	"github.com/fatih/color"
)

func menuWin7(args ...string) (err error) {
	menuCommands := []menu.CommandOption{
		{
			Description: "Add extra rules",
			Color:       color.FgHiYellow,
			Function:    addWin7Extra,
		},
		{
			Description: "Add spy rules",
			Color:       color.FgHiYellow,
			Function:    addWin7Spy,
		},
		{
			Description: "Add update rules",
			Color:       color.FgHiYellow,
			Function:    addWin7Update,
		},
	}

	menuOptions := menu.NewOptions("Firewall rules for Windows 7", "'menu' for help [telemetry-firewall-win7]> ", 0, "")

	menuN := menu.NewMenu(menuCommands, menuOptions)
	menuN.Start()
	return
}

func addWin7Extra(args ...string) error {
	addRules(data.OS_WIN7, data.RULES_EXTRA)
	return nil
}

func addWin7Spy(args ...string) error {
	addRules(data.OS_WIN7, data.RULES_SPY)
	return nil
}

func addWin7Update(args ...string) error {
	addRules(data.OS_WIN7, data.RULES_UPDATE)
	return nil
}
