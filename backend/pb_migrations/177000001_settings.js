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

    settings.smtp.enabled = true;
    settings.smtp.host = $os.getenv("POCKETBASE_SMTP_HOST");
    settings.smtp.port = $os.getenv("POCKETBASE_SMTP_PORT");
    settings.smtp.username = $os.getenv("POCKETBASE_SMTP_USERNAME");
    settings.smtp.password = $os.getenv("POCKETBASE_SMTP_PASSWORD");
    // I found that passing "false" explicitly did not work, so we use null instead
    // Despite the docs https://pocketbase.io/jsvm/interfaces/core.SMTPConfig.html#tls
    settings.smtp.tls =
      $os.getenv("POCKETBASE_SMTP_TLS") == ""
        ? null
        : $os.getenv("POCKETBASE_SMTP_TLS") == "true";

    settings.trustedProxy.headers = ["CF-Connecting-IP"];

    settings.rateLimits.enabled =
      $os.getenv("POCKETBASE_RATE_LIMITS_ENABLED") !== "false";

    // Only configure S3 if ALBUMS_S3_ENDPOINT is set (production mode)
    if ($os.getenv("ALBUMS_S3_ENDPOINT")) {
      // File storeage S3
      settings.s3.enabled = true;
      settings.s3.endpoint = $os.getenv("ALBUMS_S3_ENDPOINT");
      settings.s3.region = $os.getenv("ALBUMS_S3_REGION");
      settings.s3.accessKey = $os.getenv("ALBUMS_S3_ACCESS_KEY_ID");
      settings.s3.secret = $os.getenv("ALBUMS_S3_SECRET_ACCESS_KEY");
      settings.s3.bucket = $os.getenv("ALBUMS_S3_BUCKET_NAME");

      // Backups
      settings.backups.s3.enabled = true;
      settings.backups.s3.endpoint = $os.getenv("BACKUPS_S3_ENDPOINT");
      settings.backups.s3.region = $os.getenv("BACKUPS_S3_REGION");
      settings.backups.s3.accessKey = $os.getenv("BACKUPS_S3_ACCESS_KEY_ID");
      settings.backups.s3.secret = $os.getenv("BACKUPS_S3_SECRET_ACCESS_KEY");
      settings.backups.s3.bucket = $os.getenv("BACKUPS_S3_BUCKET_NAME");
      settings.backups.cron = $os.getenv("BACKUPS_CRON") || "0 0 * * *"; // default to daily at midnight
      settings.backups.cronMaxKeep = parseInt(
        $os.getenv("BACKUPS_CRON_MAX_KEEP") || "7",
      ); // default to keep 7 backups
    }

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
        console.log(
          "Successfully created superuser from environment variables.",
        );
      } catch (err) {
        console.error("Failed to auto-create superuser:", err);
      }
    }
  };,
  (app) => {
    // no down migration
  },
);
