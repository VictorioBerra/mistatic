package main

import (
	"fmt"
	"os"

	"mistatic/api"
	"mistatic/helpers"
	"mistatic/ui"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/osutils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("[warn] No .env file found (expected in prod):", err)
	}

	isGoRun := osutils.IsProbablyGoRun()

	app := pocketbase.New()

	// Initialize log workers
	helpers.InitLogger(5, 10000)

	// hooks.RegisterX(app)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		api.RegisterDeployRoute(se)
		api.RegisterFileRoutes(se)
		helpers.ServeStaticSite(se, ui.DistDirFS)
		return se.Next()
	})

	jsvm.MustRegister(app, jsvm.Config{
		HooksWatch: true,
		MigrationsDir: "pb_migrations",
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  isGoRun,
	})

	api.RegisterHooks(app)

	if err := app.Start(); err != nil {
		app.Logger().Error("app failed to start", "err", err)
		os.Exit(1)
	}
}
