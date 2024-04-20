package snailrace

// AchievementCallback defines a callback function type for when an achievement
// is unlocked. Important as we dont want this package to depend on any other
// package.
type AchievementCallback func(userId string, achievementName string) bool

func HostRace(stateCb StateChangeCallback, achievementCb AchievementCallback, messageId, channelId string) {

}
