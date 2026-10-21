/// <reference path="../pb_data/types.d.ts" />

migrate((app) => {
  const sites = app.findCollectionByNameOrId("sites");

  const indexDocField = new TextField({
    name: "index_document",
    required: false,
  });
  sites.fields.add(indexDocField);

  app.save(sites);
}, (app) => {
  const sites = app.findCollectionByNameOrId("sites");

  sites.fields.removeByName("index_document");

  app.save(sites);
});

