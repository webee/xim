package main

import "xim/apps/xchat/logic"

func startLogic() {
	config := logic.NewConfig(
		&logic.Config{
			Debug:     args.debug,
			BrokerURL: args.brokerURL,
		})
	logic.Start(config)
}
