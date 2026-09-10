<script>
    import { page } from '$app/stores';
    import { onMount } from 'svelte';
    import { pb } from '$lib/pb';

    let siteId = /** @type {string} */ ($page.params.site_id);
    let deploymentId = /** @type {string} */ ($page.params.deployment_id);

    /** @type {any[]} */
    let files = $state([]);
    let loading = $state(true);
    let error = $state("");

    let currentPath = $state(""); // "" means root

    let visibleItems = $derived.by(() => {
        return files.filter(f => {
            const parts = f.path.split('/');
            parts.pop();
            const dir = parts.join('/');
            return dir === currentPath;
        }).sort((a, b) => {
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
            const res = await fetch(`/api/mistatic/deploy/${siteId}/${deploymentId}/files`, {
                headers: { 'Authorization': pb.authStore.token }
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || 'Failed to load files');
            }
            files = await res.json();
        } catch (err) {
            const e = /** @type {Error} */ (err);
            error = e.message;
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadFiles();
    });

    /**
     * @param {string} path
     */
    async function deleteFile(path) {
        if (!confirm(`Delete ${path}?`)) return;
        try {
            const res = await fetch(`/api/mistatic/deploy/${siteId}/${deploymentId}/files/delete?path=${encodeURIComponent(path)}`, {
                method: 'DELETE',
                headers: { 'Authorization': pb.authStore.token }
            });
            if (!res.ok) throw new Error('Delete failed');
            await loadFiles();
        } catch (err) {
            alert('Failed to delete file');
        }
    }

    /** @type {HTMLInputElement | null} */
    let fileInput = $state(null);
    let isUploading = $state(false);
    let isDragging = $state(false);

    /**
     * @param {File} file
     */
    async function uploadFile(file) {
        isUploading = true;
        const targetPath = currentPath ? `${currentPath}/${file.name}` : file.name;
        const formData = new FormData();
        formData.append('file', file);
        formData.append('path', targetPath);

        try {
            const res = await fetch(`/api/mistatic/deploy/${siteId}/${deploymentId}/files/upload`, {
                method: 'POST',
                headers: { 'Authorization': pb.authStore.token },
                body: formData
            });
            
            if (!res.ok) throw new Error('Upload failed');
            await loadFiles();
        } catch (err) {
            alert('Failed to upload file');
        } finally {
            isUploading = false;
            if (fileInput) fileInput.value = "";
        }
    }

    /** @param {DragEvent} e */
    function handleDrop(e) {
        e.preventDefault();
        isDragging = false;
        if (!e.dataTransfer || !e.dataTransfer.files.length) return;
        uploadFile(e.dataTransfer.files[0]);
    }

    /** @param {Event} e */
    function handleFileSelect(e) {
        const target = /** @type {HTMLInputElement} */ (e.target);
        if (!target.files || !target.files.length) return;
        uploadFile(target.files[0]);
    }
</script>

<div class="mb-4">
    <a href={`/sites/${siteId}`} class="text-blue-600 dark:text-blue-400 hover:underline text-sm">&larr; Back to site deployments</a>
</div>

<h1 class="text-2xl font-bold mb-6 flex items-center">
    File Browser: <span class="font-mono text-lg font-normal ml-2 text-gray-500">{deploymentId}</span>
</h1>

{#if error}
    <p class="text-red-500">{error}</p>
{:else if loading}
    <p>Loading files...</p>
{:else}
    <div class="space-y-6 max-w-5xl">
        <!-- Breadcrumbs -->
        <div class="flex items-center space-x-2 text-sm font-medium text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800/50 p-3 rounded-lg border border-gray-200 dark:border-[#30363d]">
            <button class="hover:text-blue-600 dark:hover:text-blue-400 flex items-center" onclick={() => currentPath = ""}>
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path></svg>
                Root
            </button>
            {#each breadcrumbs as crumb}
                <span class="text-gray-400">/</span>
                <button class="hover:text-blue-600 dark:hover:text-blue-400" onclick={() => currentPath = crumb.path}>
                    {crumb.name}
                </button>
            {/each}
        </div>

        <!-- Dropzone -->
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="card p-6 border-dashed border-2 text-center transition-colors cursor-pointer {isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 dark:border-[#30363d] hover:bg-gray-50 dark:hover:bg-gray-800/50 dark:bg-[#0d1117]'}"
            ondragover={(e) => { e.preventDefault(); isDragging = true; }}
            ondragleave={(e) => { e.preventDefault(); isDragging = false; }}
            ondrop={handleDrop}
            onclick={() => fileInput?.click()}>
            
            <input type="file" class="hidden" bind:this={fileInput} onchange={handleFileSelect} />
            
            {#if isUploading}
                <div class="py-4">
                    <p class="font-medium text-blue-600 dark:text-blue-400 mb-1">Uploading file...</p>
                </div>
            {:else}
                <div class="py-4 flex flex-col items-center">
                    <svg class="h-8 w-8 text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                    </svg>
                    <p class="font-medium mb-1 text-sm">Drop a file here to upload to <span class="font-bold">{currentPath ? currentPath : 'Root'}</span> folder</p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">Or click to browse</p>
                </div>
            {/if}
        </div>

        <!-- File List -->
        <div class="card overflow-hidden">
            <table class="w-full text-left border-collapse">
                <thead>
                    <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-[#d0d7de] dark:border-[#30363d]">
                        <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400">Name</th>
                        <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400 text-right">Size</th>
                        <th class="py-2 px-4 font-medium text-sm text-gray-600 dark:text-gray-400 text-right">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {#if currentPath !== ""}
                        <tr class="border-b border-[#d0d7de] dark:border-[#30363d] hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer" onclick={() => {
                            const parts = currentPath.split('/');
                            parts.pop();
                            currentPath = parts.join('/');
                        }}>
                            <td class="py-2 px-4 text-sm font-medium flex items-center">
                                <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"></path></svg>
                                ..
                            </td>
                            <td class="py-2 px-4 text-sm text-right text-gray-500"></td>
                            <td class="py-2 px-4 text-right"></td>
                        </tr>
                    {/if}

                    {#if visibleItems.length === 0}
                        <tr>
                            <td colspan="3" class="py-6 px-4 text-center text-sm text-gray-500">This folder is empty.</td>
                        </tr>
                    {/if}
                    {#each visibleItems as file}
                        <tr class="border-b border-[#d0d7de] dark:border-[#30363d] last:border-0 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                            <td class="py-2 px-4 text-sm font-medium">
                                {#if file.is_dir}
                                    <button class="flex items-center text-blue-600 dark:text-blue-400 hover:underline" onclick={() => currentPath = file.path}>
                                        <svg class="w-4 h-4 mr-2 text-blue-500" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>
                                        {file.name}
                                    </button>
                                {:else}
                                    <div class="flex items-center text-gray-800 dark:text-gray-200">
                                        <svg class="w-4 h-4 mr-2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path></svg>
                                        {file.name}
                                    </div>
                                {/if}
                            </td>
                            <td class="py-2 px-4 text-sm text-right text-gray-500">
                                {file.is_dir ? '--' : (file.size / 1024).toFixed(1) + ' KB'}
                            </td>
                            <td class="py-2 px-4 text-right">
                                <button class="text-red-500 hover:text-red-700 text-xs font-medium" onclick={() => deleteFile(file.path)}>
                                    Delete
                                </button>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    </div>
{/if}
