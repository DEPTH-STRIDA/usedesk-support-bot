package tg

func (b *Bot) SetMsgRoutes() error {

	b.HandlMsgRoute("/start", HandleStartMessage())
	b.HandlMsgRoute("/help", HandleStartMessage())

	return nil
}

func (b *Bot) SetCallBackRoutes() error {

	// b.HandlMsgRoute("/start", HandleStartMessage())

	return nil
}
