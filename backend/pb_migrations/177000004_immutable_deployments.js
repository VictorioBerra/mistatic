/// <reference path="../pb_data/types.d.ts" />

migrate((app) => {
  const deployments = app.findCollectionByNameOrId("deployments");

  // 1. Add deployment_artifact file field
  const artifactField = new FileField({
    name: "deployment_artifact",
    maxSelect: 1,
    maxSize: 524288000, // 500 MB limit
    mimeTypes: [
      "application/zip",
      "application/x-zip-compressed",
      "application/octet-stream",
    ],
  });
  deployments.fields.add(artifactField);

  // 2. Set updateRule to null to enforce immutable deployments
  deployments.updateRule = null;

  app.save(deployments);
}, (app) => {
  const deployments = app.findCollectionByNameOrId("deployments");

  deployments.fields.removeByName("deployment_artifact");
  deployments.updateRule = "@request.auth.role = 'Admin' && @request.auth.id = site.user.id";

  app.save(deployments);
});
