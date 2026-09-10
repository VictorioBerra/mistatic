/// <reference path="../pb_data/types.d.ts" />
//
// App settings (SMTP, S3, backups, rate limits) and initial superuser.
//
migrate(
  (app) => {
    // ─────────────────────────────────────────────────────────────────────────
    // APP SETTINGS
    // ─────────────────────────────────────────────────────────────────────────

    let settings = app.settings();

    settings.meta.appName = $os.getenv("POCKETBASE_META_APP_NAME");
    settings.meta.appUrl = $os.getenv("POCKETBASE_META_APP_URL");
    settings.meta.senderName = $os.getenv("POCKETBASE_META_SENDER_NAME");
    settings.meta.senderAddress = $os.getenv("POCKETBASE_META_SENDER_EMAIL");

    settings.trustedProxy.headers = ["CF-Connecting-IP"];

    settings.rateLimits.enabled =
      $os.getenv("POCKETBASE_RATE_LIMITS_ENABLED") !== "false";

    try {
      app.save(settings);
    } catch (err) {
      console.error("Failed to save settings:", err);
    }

    // ─────────────────────────────────────────────────────────────────────────
    // SUPERUSER
    // ─────────────────────────────────────────────────────────────────────────

    const email = $os.getenv("POCKETBASE_SUPERUSER_EMAIL");
    const pass = $os.getenv("POCKETBASE_SUPERUSER_PASSWORD");

    if (email && pass) {
      try {
        const superusers = app.findCollectionByNameOrId("_superusers");
        const superuser = new Record(superusers);
        superuser.set("email", email);
        superuser.setPassword(pass);
        app.save(superuser);
        console.log("Successfully created superuser from environment variables.");
      } catch (err) {
        console.error("Failed to auto-create superuser:", err);
      }
    }
  },
  (app) => {
    // no down migration
  },
);
