/// <reference path="../pb_data/types.d.ts" />

const adminOwnsSite = "@request.auth.role = 'Admin' && @request.auth.id = user.id";
const adminOwnsRelatedSite = "@request.auth.role = 'Admin' && @request.auth.id = site.user.id";

migrate((app) => {
  const users = app.findCollectionByNameOrId("users");
  const roleField = new SelectField({
    name: "role",
    maxSelect: 1,
    required: false,
    values: ["Admin", "Reader"],
  });
  users.fields.add(roleField);
  app.save(users);

  app.db()
    .newQuery("UPDATE users SET role = 'Admin' WHERE role = '' OR role IS NULL")
    .execute();

  roleField.required = true;
  users.createRule = "@request.auth.role = 'Admin'";
  users.listRule = "@request.auth.role = 'Admin'";
  users.viewRule = "@request.auth.role = 'Admin'";
  users.updateRule = "@request.auth.role = 'Admin' && @request.body.role:isset = false";
  users.deleteRule = "@request.auth.role = 'Admin'";
  app.save(users);

  const sites = app.findCollectionByNameOrId("sites");
  sites.listRule = adminOwnsSite;
  sites.viewRule = adminOwnsSite;
  sites.createRule = adminOwnsSite;
  sites.updateRule = adminOwnsSite;
  sites.deleteRule = adminOwnsSite;
  app.save(sites);

  for (const name of ["deployments", "custom_domains"]) {
    const collection = app.findCollectionByNameOrId(name);
    collection.listRule = adminOwnsRelatedSite;
    collection.viewRule = adminOwnsRelatedSite;
    collection.createRule = adminOwnsRelatedSite;
    collection.updateRule = adminOwnsRelatedSite;
    collection.deleteRule = adminOwnsRelatedSite;
    app.save(collection);
  }

  const requestLogs = app.findCollectionByNameOrId("request_logs");
  requestLogs.listRule = adminOwnsRelatedSite;
  requestLogs.viewRule = adminOwnsRelatedSite;
  requestLogs.deleteRule = adminOwnsRelatedSite;
  app.save(requestLogs);
}, (app) => {
  const users = app.findCollectionByNameOrId("users");
  users.createRule = null;
  users.listRule = "id = @request.auth.id";
  users.viewRule = "id = @request.auth.id";
  users.updateRule = "id = @request.auth.id";
  users.deleteRule = "id = @request.auth.id";
  users.fields.removeByName("role");
  app.save(users);

  const sites = app.findCollectionByNameOrId("sites");
  sites.listRule = "@request.auth.id = user.id";
  sites.viewRule = "@request.auth.id = user.id";
  sites.createRule = "@request.auth.id = user.id";
  sites.updateRule = "@request.auth.id = user.id";
  sites.deleteRule = "@request.auth.id = user.id";
  app.save(sites);

  for (const name of ["deployments", "custom_domains"]) {
    const collection = app.findCollectionByNameOrId(name);
    collection.listRule = "@request.auth.id = site.user.id";
    collection.viewRule = "@request.auth.id = site.user.id";
    collection.createRule = "@request.auth.id = site.user.id";
    collection.updateRule = "@request.auth.id = site.user.id";
    collection.deleteRule = "@request.auth.id = site.user.id";
    app.save(collection);
  }

  const requestLogs = app.findCollectionByNameOrId("request_logs");
  requestLogs.listRule = "@request.auth.id = site.user.id";
  requestLogs.viewRule = "@request.auth.id = site.user.id";
  requestLogs.deleteRule = "@request.auth.id = site.user.id";
  app.save(requestLogs);
});