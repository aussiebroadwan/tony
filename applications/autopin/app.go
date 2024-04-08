package autopin

import "github.com/aussiebroadwan/tony/framework"

const autopinThreshold = 5

func RegisterAutopinApp(bot *framework.Bot) framework.Route {
	return framework.NewRoute(bot, "autopin", &AutopinApp{})
}
