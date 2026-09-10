<script>
    import { env } from '$env/dynamic/public';
    const ROOT_DOMAIN = env.PUBLIC_ROOT_DOMAIN || 'mistatic.local';
    import { pb } from '$lib/pb';
    import { goto } from '$app/navigation';
    
    let name = $state('');
    let subdomain = $derived(name.toLowerCase().replace(/[^a-z0-9-]/g, '').replace(/-+/g, '-').replace(/^-|-$/g, ''));
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
                spa_fallback,
                is_public,
                user: pb.authStore.model.id
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

<div class="max-w-md mx-auto mt-10">
    <div class="mb-4">
        <a href="/" class="text-blue-600 dark:text-blue-400 hover:underline text-sm">&larr; Back to sites</a>
    </div>

    <h1 class="text-2xl font-semibold mb-6">Create a new site</h1>

    {#if error}
        <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 px-4 py-3 rounded mb-4 text-sm">
            {error}
        </div>
    {/if}

    <div class="card p-6">
        <form onsubmit={(e) => { e.preventDefault(); createSite(); }} class="space-y-4">
            <div>
                <label for="name" class="block text-sm font-medium mb-1">Site Name</label>
                <input id="name" type="text" bind:value={name} required class="w-full border border-gray-300 dark:border-[#30363d] dark:bg-[#0d1117] rounded px-3 py-2 text-sm focus:outline-none focus:border-blue-500 dark:focus:border-blue-500" placeholder="My Awesome Project" />
            </div>

            <div>
                <label class="block text-sm font-medium mb-1">URL Preview</label>
                <div class="bg-gray-50 dark:bg-gray-800/50 text-gray-500 dark:text-gray-400 px-3 py-2 rounded border border-gray-200 dark:border-[#30363d] text-sm font-mono">
                    {subdomain ? subdomain : 'project'}.{ROOT_DOMAIN}
                </div>
            </div>

            <div class="flex items-center mt-2">
                <input id="spa_fallback" type="checkbox" bind:checked={spa_fallback} class="mr-2" />
                <label for="spa_fallback" class="text-sm">Enable SPA Fallback (route 404s to index.html)</label>
            </div>

            <div class="flex items-center mt-2">
                <input id="is_public" type="checkbox" bind:checked={is_public} class="mr-2" />
                <label for="is_public" class="text-sm">Public Site (Uncheck to require authentication)</label>
            </div>

            <div class="pt-4 border-t border-gray-100 dark:border-[#30363d]">
                <button type="submit" disabled={loading} class="btn-primary w-full py-2">
                    {loading ? 'Creating...' : 'Create site'}
                </button>
            </div>
        </form>
    </div>
</div>
