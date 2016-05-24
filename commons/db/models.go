package db

// App is a application using xim.
type App struct {
	Name string
	Desc string
}

// Channel is a app's messaging channel.
type Channel struct {
	App     string
	Channel string
	Owner   string
}
