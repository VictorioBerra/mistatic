<script>
    import { env } from '$env/dynamic/public';
    const ROOT_DOMAIN = env.PUBLIC_ROOT_DOMAIN || 'mistatic.local';
    import { page } from '$app/stores';
    import { onMount } from 'svelte';
    import { pb } from '$lib/pb';
    import { getSiteUrl } from '$lib/utils';
    
    import JSZip from 'jszip';
    
    let siteId = /** @type {string} */ ($page.params.id);
    
    /** @type {any} */
    let site = $state(null);
    /** @type {any[]} */
    let deployments = $state([]);
    /** @type {any[]} */
    let domains = $state([]);
    /** @type {any[]} */
    let requestLogs = $state([]);
    let logsPage = $state(1);
    let logsTotalPages = $state(1);
    const logsPerPage = 20;
    
    let isDragging = $state(false);
    let isUploading = $state(false);
    
    /** @type {File | null} */
    let pendingFile = $state(null);
    /** @type {string[]} */
    let zipContents = $state([]);
    
    /** @type {HTMLInputElement | null} */
    let fileInput = $state(null);
    
    async function loadData() {
        site = await pb.collection('sites').getOne(siteId);
        deployments = await pb.collection('deployments').getFullList({ filter: `site="${siteId}"`, sort: '-created' });
        domains = await pb.collection('custom_domains').getFullList({ filter: `site="${siteId}"` });
        
        await loadLogs();
    }

    async function loadLogs() {
        if (!site) return;
        try {
            const logsResult = await pb.collection('request_logs').getList(logsPage, logsPerPage, { filter: `site="${site.id}"`, sort: '-created' });
            requestLogs = logsResult.items;
            logsTotalPages = logsResult.totalPages;
        } catch (err) {
            console.error('Failed to load request logs', err);
        }
    }

    onMount(() => {
        loadData().catch(console.error);
    });

    /** @param {File} file */
    async function stageFile(file) {
        pendingFile = file;
        zipContents = [];
        
        if (file.name.endsWith('.zip')) {
            zipContents = ["Loading contents..."];
            try {
                const zip = new JSZip();
                const contents = await zip.loadAsync(file);
                const rootItems = new Set();
                Object.keys(contents.files).forEach(path => {
                    const parts = path.split('/').filter(p => p);
                    if (parts.length > 0) {
                        rootItems.add(parts[0]);
                    }
                });
                const itemsArr = Array.from(rootItems);
                zipContents = itemsArr.slice(0, 10);
                if (itemsArr.length > 10) {
                    zipContents.push(`... and ${itemsArr.length - 10} more`);
                }
            } catch (e) {
                console.error(e);
                zipContents = ["(Could not read zip contents)"];
            }
        } else {
            zipContents = [file.name];
        }
    }

    /** @param {File} file */
    async function uploadFile(file) {
        isUploading = true;
        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await fetch(`/api/mistatic/deploy/${siteId}`, {
                method: 'POST',
                headers: {
                    'Authorization': pb.authStore.token
                },
                body: formData
            });
            
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Upload failed');
            }
            
            await loadData();
        } catch (err) {
            console.error(err);
            const e = /** @type {Error} */ (err);
            alert(e.message);
        } finally {
            isUploading = false;
        }
    }

    /** @param {DragEvent} e */
    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        
        if (!e.dataTransfer || !e.dataTransfer.files.length) return;
        stageFile(e.dataTransfer.files[0]);
    }

    /** @param {Event} e */
    function handleFileSelect(e) {
        const target = /** @type {HTMLInputElement} */ (e.target);
        if (!target.files || !target.files.length) return;
        stageFile(target.files[0]);
        target.value = ''; // Reset input
    }
</script>

<div class="mb-4">
    <a href="/" class="text-blue-600 dark:text-blue-400 hover:underline text-sm">&larr; Back to sites</a>
</div>

{#if !site}
    <p>Loading...</p>
{:else}
    <div class="flex justify-between items-end mb-6">
        <div>
            <h1 class="text-2xl font-bold">{site.name}</h1>
            <div class="flex flex-col space-y-1 mt-1">
                <a href={getSiteUrl(site.subdomain, ROOT_DOMAIN)} target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline text-sm">
                    {site.subdomain}.{ROOT_DOMAIN}
                </a>
                <a href="/site/{site.subdomain}/" target="_blank" class="text-gray-500 dark:text-gray-400 hover:underline text-xs">
                    Alternative: /site/{site.subdomain}/
                </a>
            </div>
        </div>
        <div class="flex items-center space-x-3">
            <a href={getSiteUrl(site.subdomain, ROOT_DOMAIN)} target="_blank" class="btn-default flex items-center space-x-2 py-1.5">
                <span>Visit Site</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
            </a>
        </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="md:col-span-2 space-y-6">
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div class="card p-8 border-dashed border-2 text-center transition-colors cursor-pointer {isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 dark:border-[#30363d] hover:bg-gray-50 dark:hover:bg-gray-800/50 dark:bg-[#0d1117]'}"
                ondragover={(e) => { e.preventDefault(); isDragging = true; }}
                ondragleave={(e) => { e.preventDefault(); isDragging = false; }}
                ondrop={handleDrop}
                onclick={() => fileInput?.click()}>
                
                <input type="file" class="hidden" bind:this={fileInput} onchange={handleFileSelect} />
                
                {#if isUploading}
                    <div class="py-8">
                        <p class="font-medium text-blue-600 dark:text-blue-400 mb-2">Deploying your site...</p>
                        <p class="text-sm text-gray-500 dark:text-gray-400">Processing files and updating pointers.</p>
                    </div>
                {:else if pendingFile}
                    <div class="py-6">
                        <svg class="mx-auto h-12 w-12 text-blue-500 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        <p class="font-medium text-gray-900 dark:text-gray-100 mb-2">Ready to deploy</p>
                        <div class="text-sm text-gray-500 dark:text-gray-400 font-mono mb-4">
                            {#each zipContents as item}
                                <div>{item}</div>
                            {/each}
                        </div>
                        <div class="space-x-3">
                            <button class="btn-default" onclick={(e) => { e.stopPropagation(); pendingFile = null; zipContents = []; fileInput.value = ''; }}>Cancel</button>
                            <button class="btn-primary" onclick={(e) => { e.stopPropagation(); uploadFile(pendingFile); }}>Deploy</button>
                        </div>
                    </div>
                {:else}
                    <div class="py-8">
                        <svg class="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                        </svg>
                        <p class="font-medium mb-1">Drag and drop your file or .zip here</p>
                        <p class="text-sm text-gray-500 dark:text-gray-400">Upload a single file or a zip of your static site</p>
                    </div>
                {/if}
            </div>

            <div>
                <h2 class="text-lg font-semibold mb-3">Deployments</h2>
                <div class="card overflow-hidden">
                    <table class="w-full text-left border-collapse">
                        <thead>
                            <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-[#d0d7de] dark:border-[#30363d]">
                                <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">ID</th>
                                <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Status</th>
                                <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Date</th>
                                <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400 text-right">State</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#if deployments.length === 0}
                                <tr>
                                    <td colspan="4" class="py-4 px-4 text-center text-sm text-gray-500 dark:text-gray-400">No deployments yet. Upload a file or zip to get started.</td>
                                </tr>
                            {/if}
                            {#each deployments as dep}
                                <tr class="border-b border-[#d0d7de] dark:border-[#30363d] last:border-0 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                                    <td class="py-3 px-4 font-mono text-xs">{dep.id}</td>
                                    <td class="py-3 px-4 text-sm">
                                        <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {dep.status === 'success' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}">
                                            {dep.status}
                                        </span>
                                    </td>
                                    <td class="py-3 px-4 text-sm text-gray-500">{new Date(dep.created).toLocaleString()}</td>
                                    <td class="py-3 px-4 text-right">
                                                                                {#if site.active_deployment === dep.id}
                                            <span class="text-xs font-medium text-green-600 flex items-center justify-end mb-2">
                                                <span class="w-2 h-2 rounded-full bg-green-500 mr-1.5"></span> Active
                                            </span>
                                        {/if}
                                        {#if dep.status === 'success'}
                                            <select class="btn-default text-xs py-1 px-2 pr-6 appearance-none bg-white dark:bg-[#0d1117]" onchange={async (e) => {
                                                const target = /** @type {HTMLSelectElement} */ (e.target);
                                                const action = target.value;
                                                target.value = ''; // reset
                                                if (!action) return;
                                                
                                                if (action === 'activate') {
                                                    await pb.collection('sites').update(site.id, { active_deployment: dep.id });
                                                    await loadData();
                                                } else if (action === 'delete') {
                                                    if (confirm('Delete this deployment?')) {
                                                        await pb.collection('deployments').delete(dep.id);
                                                        await loadData();
                                                    }
                                                } else if (action === 'download') {
                                                    const res = await fetch(`/api/mistatic/deploy/${site.id}/${dep.id}/download`, {
                                                        headers: { 'Authorization': pb.authStore.token }
                                                    });
                                                    if (res.ok) {
                                                        const blob = await res.blob();
                                                        const url = window.URL.createObjectURL(blob);
                                                        const a = document.createElement('a');
                                                        a.href = url;
                                                        a.download = `${dep.id}.zip`;
                                                        a.click();
                                                        window.URL.revokeObjectURL(url);
                                                    } else {
                                                        alert('Download failed');
                                                    }
                                                } else if (action === 'browse') {
                                                    window.location.href = `/sites/${site.id}/deployments/${dep.id}`;
                                                }
                                            }}>
                                                <option value="">Actions...</option>
                                                {#if site.active_deployment !== dep.id}
                                                    <option value="activate">Rollback / Make Active</option>
                                                {/if}
                                                <option value="browse">Browse Files</option>
                                                <option value="download">Download Zip</option>
                                                {#if site.active_deployment !== dep.id}
                                                    <option value="delete">Delete</option>
                                                {/if}
                                            </select>
                                        {/if}
                                    </td>
                                </tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
            </div>

            <div class="mt-8">
                <div class="flex justify-between items-center mb-3">
                    <h2 class="text-lg font-semibold">Request Logs</h2>
                    <button class="text-xs text-blue-600 hover:underline" onclick={() => { logsPage = 1; loadLogs(); }}>Refresh</button>
                </div>
                <div class="card overflow-hidden">
                    <div class="overflow-x-auto">
                        <table class="w-full text-left border-collapse">
                            <thead>
                                <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-[#d0d7de] dark:border-[#30363d]">
                                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Date</th>
                                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Method</th>
                                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Path</th>
                                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Duration</th>
                                    <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">IP</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#if requestLogs.length === 0}
                                    <tr>
                                        <td colspan="5" class="py-4 px-4 text-center text-sm text-gray-500 dark:text-gray-400">No requests recorded yet.</td>
                                    </tr>
                                {/if}
                                {#each requestLogs as log}
                                    <tr class="border-b border-[#d0d7de] dark:border-[#30363d] last:border-0 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                                        <td class="py-3 px-4 text-sm whitespace-nowrap">{new Date(log.created).toLocaleString()}</td>
                                        <td class="py-3 px-4 text-sm font-mono">{log.method}</td>
                                        <td class="py-3 px-4 text-sm truncate max-w-[200px]" title={log.path}>{log.path}</td>
                                        <td class="py-3 px-4 text-sm">{log.duration_ms}ms</td>
                                        <td class="py-3 px-4 text-sm text-gray-500">{log.ip}</td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                    
                    {#if logsTotalPages > 1}
                        <div class="flex items-center justify-between px-4 py-3 border-t border-[#d0d7de] dark:border-[#30363d] bg-gray-50 dark:bg-gray-800/50">
                            <button 
                                class="btn-default text-xs py-1 px-3 disabled:opacity-50 disabled:cursor-not-allowed"
                                disabled={logsPage === 1}
                                onclick={() => { logsPage--; loadLogs(); }}
                            >
                                Previous
                            </button>
                            <span class="text-xs text-gray-500">Page {logsPage} of {logsTotalPages}</span>
                            <button 
                                class="btn-default text-xs py-1 px-3 disabled:opacity-50 disabled:cursor-not-allowed"
                                disabled={logsPage >= logsTotalPages}
                                onclick={() => { logsPage++; loadLogs(); }}
                            >
                                Next
                            </button>
                        </div>
                    {/if}
                </div>
            </div>
        </div>

        <div class="space-y-6">
            <div class="card p-4">
                <h3 class="font-semibold mb-3">Settings</h3>
                <label class="flex items-center space-x-2 text-sm cursor-pointer">
                    <input type="checkbox" class="rounded border-gray-300 dark:border-[#30363d]" 
                        checked={site.disable_request_logging} 
                        onchange={async (e) => {
                            const target = /** @type {HTMLInputElement} */ (e.target);
                            try {
                                site = await pb.collection('sites').update(site.id, { disable_request_logging: target.checked });
                            } catch (err) {
                                alert('Failed to update setting');
                                target.checked = !target.checked;
                            }
                        }}
                    />
                    <span>Disable request logging</span>
                </label>
            </div>

            <div class="card p-4">
                <h3 class="font-semibold mb-3">Custom Domains</h3>
                {#if domains.length === 0}
                    <p class="text-sm text-gray-500 mb-3">No custom domains configured.</p>
                {:else}
                    <ul class="space-y-2 mb-3">
                        {#each domains as domain}
                            <li class="flex justify-between items-center text-sm border border-gray-200 dark:border-[#30363d] rounded px-2 py-1">
                                <span>{domain.domain}</span>
                                <button class="text-red-500 hover:text-red-700 text-xs" onclick={async () => {
                                    await pb.collection('custom_domains').delete(domain.id);
                                    await loadData();
                                }}>Remove</button>
                            </li>
                        {/each}
                    </ul>
                {/if}
                
                <form onsubmit={async (e) => {
                    e.preventDefault();
                    const target = /** @type {HTMLFormElement} */ (e.target);
                    const input = /** @type {HTMLInputElement} */ (target.elements.namedItem('domain'));
                    if(input && input.value) {
                        try {
                            await pb.collection('custom_domains').create({ site: site.id, domain: input.value });
                            input.value = '';
                            await loadData();
                        } catch (err) {
                            const errorObj = /** @type {Error} */ (err);
                            alert(errorObj.message);
                        }
                    }
                }} class="flex space-x-2">
                    <input name="domain" type="text" placeholder="example.com" class="w-full border border-gray-300 dark:border-[#30363d] dark:bg-[#0d1117] rounded px-2 py-1 text-sm focus:outline-none focus:border-blue-500" />
                    <button type="submit" class="btn-default text-sm">Add</button>
                </form>
            </div>
            
            <div class="card p-4 border-red-200 dark:border-red-900/50">
                <h3 class="font-semibold text-red-600 mb-3">Danger Zone</h3>
                <button class="w-full py-1.5 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-800/50 rounded-md text-sm font-medium hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors"
                    onclick={async () => {
                        if (confirm('Are you sure you want to delete this site?')) {
                            await pb.collection('sites').delete(site.id);
                            window.location.href = '/';
                        }
                    }}>Delete Site</button>
            </div>
        </div>
    </div>
{/if}
