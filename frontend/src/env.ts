import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	PUBLIC_APP_DOMAIN: { public: true, schema: (input) => input ?? '' },
	PUBLIC_ROOT_DOMAIN: { public: true, schema: (input) => input ?? '' }
});
