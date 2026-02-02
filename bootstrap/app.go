package bootstrap

type Application struct {
	Env *Env
}

type ApplicationInterface interface {
	App()
}

func (app *Application) App() {
	app.Env = NewEnv()
}

func App() Application {
	app := &Application{}
	app.Env = NewEnv()
	return *app
}
