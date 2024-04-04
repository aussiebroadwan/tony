package framework

type ApplicationRuleType int

const (
	ApplicationRuleTypeModeration ApplicationRuleType = iota
	ApplicationRuleTypeReactions
)

type ApplicationRule interface {
	// Name of the rule
	Name() string
	GetType() ApplicationRuleType

	// Test the rule against content
	Test(content string) error

	// What action should be taken if the rule is violated
	Action(ctx *Context, violation error)
}

type ActionableRule struct {
	Channel string
	Rule    ApplicationRule
}

func Rule(channel string, rule ApplicationRule) ActionableRule {
	return ActionableRule{
		Channel: channel,
		Rule:    rule,
	}
}
