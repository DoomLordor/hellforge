package defaults

import (
	"context"
	"fmt"
	"os"

	apiserver "github.com/DoomLordor/hellforge/api-server"
)

func Main(configurator apiserver.Configurator) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := apiserver.NewAPIServer(ctx, configurator)
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}

	err = server.Run(ctx)
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(-1)
	}
	os.Exit(0)
}
