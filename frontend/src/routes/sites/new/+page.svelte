<script>
	import { PUBLIC_ROOT_DOMAIN } from '$app/env/public';
	const ROOT_DOMAIN = PUBLIC_ROOT_DOMAIN || 'mistatic.local';
	import { pb } from '#lib/pb.js';
	import { goto } from '$app/navigation';

	let name = $state('');
	let subdomain = $derived(
		name
			.toLowerCase()
			.replace(/[^a-z0-9-]/g, '')
			.replace(/-+/g, '-')
			.replace(/^-|-$/g, '')
	);
	let index_document = $state('index.html');
	let spa_fallback = $state(false);
	let is_public = $state(true);
	let loading = $state(false);
	let error = $state('');

	async function createSite() {
		loading = true;
		error = '';
		try {
			const record = await pb.collection('sites').create({
				name,
				subdomain,
				index_document: index_document.trim() || 'index.html',
				spa_fallback,
				is_public,
				user: pb.authStore.model?.id
			});
			goto(`/sites/${record.id}`);
		} catch (err) {
			console.error(err);
			const e = /** @type {Error} */ (err);
			error = e.message || 'Failed to create site. Subdomain might be taken.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="mx-auto mt-10 max-w-md">
	<div class="mb-4">
		<a href="/" class="text-sm text-blue-600 hover:underline dark:text-blue-400">← Back to sites</a>
	</div>

	<h1 class="mb-6 text-2xl font-semibold">Create a new site</h1>

	{#if error}
		<div
			class="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400"
		>
			{error}
		</div>
	{/if}

	<div class="card p-6">
		<form
			onsubmit={(e) => {
				e.preventDefault();
				createSite();
			}}
			class="space-y-4"
		>
			<div>
				<label for="name" class="mb-1 block text-sm font-medium">Site Name</label>
				<input
					id="name"
					type="text"
					bind:value={name}
					required
					class="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-[#30363d] dark:bg-[#0d1117] dark:focus:border-blue-500"
					placeholder="My Awesome Project"
				/>
			</div>

			<div>
				<span class="mb-1 block text-sm font-medium">URL Preview</span>
				<div
					class="rounded border border-gray-200 bg-gray-50 px-3 py-2 font-mono text-sm text-gray-500 dark:border-[#30363d] dark:bg-gray-800/50 dark:text-gray-400"
				>
					{subdomain ? subdomain : 'project'}.{ROOT_DOMAIN}
				</div>
			</div>

			<div>
				<label for="index_document" class="mb-1 block text-sm font-medium">Index Document</label>
				<input
					id="index_document"
					type="text"
					bind:value={index_document}
					class="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-[#30363d] dark:bg-[#0d1117] dark:focus:border-blue-500"
					placeholder="index.html"
				/>
				<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
					Default file served for root and directory requests (e.g. index.html, home.html)
				</p>
			</div>

			<div class="mt-2 flex items-center">
				<input id="spa_fallback" type="checkbox" bind:checked={spa_fallback} class="mr-2" />
				<label for="spa_fallback" class="text-sm"
					>Enable SPA Fallback (route 404s to {index_document.trim() || 'index.html'})</label
				>
			</div>

			<div class="mt-2 flex items-center">
				<input id="is_public" type="checkbox" bind:checked={is_public} class="mr-2" />
				<label for="is_public" class="text-sm"
					>Public Site (Uncheck to require authentication)</label
				>
			</div>

			<div class="border-t border-gray-100 pt-4 dark:border-[#30363d]">
				<button type="submit" disabled={loading} class="btn-primary w-full py-2">
					{loading ? 'Creating...' : 'Create site'}
				</button>
			</div>
		</form>
	</div>
</div>
