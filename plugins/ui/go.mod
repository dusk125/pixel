module github.com/gopxl/pixel/v2/plugins/ui

go 1.21.2

replace (
	github.com/gopxl/pixel/v2 => ../../
	github.com/gopxl/pixel/v2/plugins/spritetools => ../spritetools
)

require (
	github.com/gopxl/mainthread/v2 v2.0.0
	github.com/gopxl/pixel/v2 v2.0.0-20231010025657-c49396e3e7db
	github.com/gopxl/pixel/v2/plugins/spritetools v0.0.0-00010101000000-000000000000
	github.com/inkyblackness/imgui-go v1.12.0
)

require (
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20221017161538-93cebf72946b // indirect
	github.com/go-gl/mathgl v1.1.0 // indirect
	github.com/gopxl/glhf/v2 v2.0.0-20231010004459-fc58b4a8f7d0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/image v0.13.0 // indirect
)
