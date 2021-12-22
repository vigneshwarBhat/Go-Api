package main

import (
	"Product-Api/app"
)

func main() {
	a := app.App{}
	a.Initialize()
	a.Run("8080")
}
