package main

import "online-subs/internal/initializers"

// @title Subscriptions Service API
// @version 1.0
// @description API for managing user subscriptions.
// @schemes http
// @BasePath /
// @tag.name subscriptions
// @produce json
// @consume json
func main() {
	initializers.RunSubsService()
}
