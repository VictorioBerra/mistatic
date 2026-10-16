/// <reference path="../pb_data/types.d.ts" />

migrate((app) => {
  const sites = app.findCollectionByNameOrId("sites");

  // 1. Add enable_request_logging bool field (defaults to false / opt-in)
  const loggingField = new BoolField({
    name: "enable_request_logging",
  });
  sites.fields.add(loggingField);

  // 2. Remove old disable_request_logging field
  sites.fields.removeByName("disable_request_logging");

  app.save(sites);
}, (app) => {
  const sites = app.findCollectionByNameOrId("sites");

  const loggingField = new BoolField({
    name: "disable_request_logging",
  });
  sites.fields.add(loggingField);

  sites.fields.removeByName("enable_request_logging");

  app.save(sites);
});
