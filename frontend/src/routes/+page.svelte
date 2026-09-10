<script>
    import { env } from '$env/dynamic/public';
    import { pb } from '$lib/pb';
    import { getSiteUrl } from '$lib/utils';
    import { onMount } from 'svelte';
    
    const ROOT_DOMAIN = env.PUBLIC_ROOT_DOMAIN || 'mistatic.local';
    
    /** @type {any[]} */
    let sites = $state([]);
    let loading = $state(true);

    onMount(async () => {
        if (!pb.authStore.isValid) return;
        try {
            sites = await pb.collection('sites').getFullList({ sort: '-created' });
        } catch (err) {
            console.error(err);
            // silently fail and let layout guard redirect
        } finally {
            loading = false;
        }
    });
</script>

<div class="flex justify-between items-center mb-6">
    <h1 class="text-2xl font-semibold">Sites</h1>
    <a href="/sites/new" class="btn-primary">New Site</a>
</div>

{#if loading}
    <p>Loading...</p>
{:else if sites.length === 0}
    <div class="card p-8 text-center text-gray-500 dark:text-gray-400">
        <p class="mb-4">You don't have any sites yet.</p>
        <a href="/sites/new" class="btn-primary">Create your first site</a>
    </div>
{:else}
    <div class="card overflow-hidden">
        <table class="w-full text-left border-collapse">
            <thead>
                <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-[#d0d7de] dark:border-[#30363d]">
                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Name</th>
                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">URL</th>
                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400 text-right">Actions</th>
                </tr>
            </thead>
            <tbody>
                {#each sites as site (site.id)}
                    <tr class="border-b border-[#d0d7de] dark:border-[#30363d] last:border-0 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                        <td class="py-3 px-4 font-medium">
                            <a href={`/sites/${site.id}`} class="text-blue-600 dark:text-blue-400 hover:underline">{site.name}</a>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm">
                            <div class="flex flex-col space-y-1">
                                <a href={getSiteUrl(site.subdomain, ROOT_DOMAIN)} target="_blank" class="hover:underline">{site.subdomain}.{ROOT_DOMAIN}</a>
                                <a href="/site/{site.subdomain}/" target="_blank" class="text-xs text-gray-500 hover:underline">/site/{site.subdomain}/</a>
                            </div>
                        </td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-right font-medium space-x-3">
                            <a href={getSiteUrl(site.subdomain, ROOT_DOMAIN)} target="_blank" class="btn-default text-xs text-gray-500 hover:text-gray-700 dark:text-gray-300 dark:hover:text-white">Visit</a>
                            <a href={`/sites/${site.id}`} class="btn-default text-xs">Manage</a>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
{/if}
