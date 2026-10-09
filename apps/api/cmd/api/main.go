package main

import "github.com/paroki/domus/api/internal/config"

func main() {
	fiber := config.GetFiber()

	fiber.Listen(":8002")
}
