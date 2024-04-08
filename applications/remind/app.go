package remind

import "github.com/aussiebroadwan/tony/framework"

func RegisterRemindApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "remind",
		// remind
		&RemindCommand{}, // [NOP]

		// remind <subcommand>
		framework.NewRoute(bot, "add", &RemindAddSubCommand{}),
		framework.NewRoute(bot, "del", &RemindDeleteSubCommand{}),
		framework.NewRoute(bot, "list", &RemindListSubCommand{}),
		framework.NewRoute(bot, "status", &RemindStatusSubCommand{}),
	)
}
