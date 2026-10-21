<script>
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { pb } from '#lib/pb.js';

	let siteId = /** @type {string} */ (page.params.site_id);
	let deploymentId = /** @type {string} */ (page.params.deployment_id);

	/** @type {any[]} */
	let files = $state([]);
	let loading = $state(true);
	let error = $state('');

	let currentPath = $state(''); // "" means root

	let visibleItems = $derived.by(() => {
		return files
			.filter((f) => {
				const parts = f.path.split('/');
				parts.pop();
				const dir = parts.join('/');
				return dir === currentPath;
			})
			.sort((a, b) => {
				if (a.is_dir && !b.is_dir) return -1;
				if (!a.is_dir && b.is_dir) return 1;
				return a.name.localeCompare(b.name);
			});
	});

	let breadcrumbs = $derived.by(() => {
		if (!currentPath) return [];
		const parts = currentPath.split('/');
		return parts.map((p, i) => {
			return {
				name: p,
				path: parts.slice(0, i + 1).join('/')
			};
		});
	});

	async function loadFiles() {
		loading = true;
		try {
			files = await pb.send(`/api/mistatic/deploy/${siteId}/${deploymentId}/files`, {});
		} catch (err) {
			const e = /** @type {any} */ (err);
			error = e?.data?.error || e?.message || 'Failed to load files';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadFiles();
	});
</script>

<div class="mb-4">
	<a href={`/sites/${siteId}`} class="text-sm text-blue-600 hover:underline dark:text-blue-400"
		>← Back to site deployments</a
	>
</div>

<div class="mb-6 flex items-center justify-between">
	<h1 class="flex items-center text-2xl font-bold">
		Deployment Manifest: <span class="ml-2 font-mono text-lg font-normal text-gray-500"
			>{deploymentId}</span
		>
	</h1>
	<span
		class="inline-flex items-center rounded border border-gray-300 bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-300"
	>
		Immutable Snapshot
	</span>
</div>

{#if error}
	<p class="text-red-500">{error}</p>
{:else if loading}
	<p>Loading files...</p>
{:else}
	<div class="max-w-5xl space-y-6">
		<!-- Breadcrumbs -->
		<div
			class="flex items-center space-x-2 rounded-lg border border-gray-200 bg-gray-50 p-3 text-sm font-medium text-gray-600 dark:border-[#30363d] dark:bg-gray-800/50 dark:text-gray-300"
		>
			<button
				class="flex items-center hover:text-blue-600 dark:hover:text-blue-400"
				onclick={() => (currentPath = '')}
			>
				<svg class="mr-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
					></path></svg
				>
				Root
			</button>
			{#each breadcrumbs as crumb (crumb.path)}
				<span class="text-gray-400">/</span>
				<button
					class="hover:text-blue-600 dark:hover:text-blue-400"
					onclick={() => (currentPath = crumb.path)}
				>
					{crumb.name}
				</button>
			{/each}
		</div>

		<!-- File List (Read-Only) -->
		<div class="card overflow-hidden">
			<table class="w-full border-collapse text-left">
				<thead>
					<tr
						class="border-b border-[#d0d7de] bg-gray-50 dark:border-[#30363d] dark:bg-gray-800/50"
					>
						<th class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400">Name</th>
						<th class="px-4 py-2 text-right text-sm font-medium text-gray-600 dark:text-gray-400"
							>Size</th
						>
					</tr>
				</thead>
				<tbody>
					{#if currentPath !== ''}
						<tr
							class="cursor-pointer border-b border-[#d0d7de] hover:bg-gray-50 dark:border-[#30363d] dark:hover:bg-gray-800/50"
							onclick={() => {
								const parts = currentPath.split('/');
								parts.pop();
								currentPath = parts.join('/');
							}}
						>
							<td class="flex items-center px-4 py-2 text-sm font-medium" colspan="2">
								<svg
									class="mr-2 h-4 w-4 text-gray-400"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
									><path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"
									></path></svg
								>
								..
							</td>
						</tr>
					{/if}

					{#if visibleItems.length === 0}
						<tr>
							<td colspan="2" class="px-4 py-6 text-center text-sm text-gray-500"
								>This folder is empty.</td
							>
						</tr>
					{/if}
					{#each visibleItems as file (file.path)}
						<tr
							class="border-b border-[#d0d7de] last:border-0 hover:bg-gray-50 dark:border-[#30363d] dark:hover:bg-gray-800/50"
						>
							<td class="px-4 py-2 text-sm font-medium">
								{#if file.is_dir}
									<button
										class="flex items-center text-blue-600 hover:underline dark:text-blue-400"
										onclick={() => (currentPath = file.path)}
									>
										<svg class="mr-2 h-4 w-4 text-blue-500" fill="currentColor" viewBox="0 0 20 20"
											><path
												d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"
											></path></svg
										>
										{file.name}
									</button>
								{:else}
									<div class="flex items-center text-gray-800 dark:text-gray-200">
										<svg
											class="mr-2 h-4 w-4 text-gray-400"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
											><path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
											></path></svg
										>
										{file.name}
									</div>
								{/if}
							</td>
							<td class="px-4 py-2 text-right font-mono text-sm text-gray-500">
								{file.is_dir ? '--' : (file.size / 1024).toFixed(1) + ' KB'}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{/if}
