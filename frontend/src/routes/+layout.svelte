<script>
    import { env } from '$env/dynamic/public';
    const APP_DOMAIN = env.PUBLIC_APP_DOMAIN || 'app.mistatic.local';
    import '../app.css';
    import { onMount } from 'svelte';
    import { pb } from '$lib/pb';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';

    let { children } = $props();
    let isDark = $state(false);
	let role = $state(pb.authStore.model?.role || '');
	let authReady = $state(false);
	let isAuthFlow = $derived($page.url.pathname === '/login' || $page.url.pathname === '/sso');
	let canViewManagement = $derived(role === 'Admin');

    onMount(() => {
        if (localStorage.theme === 'dark' || (!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
            isDark = true;
            document.documentElement.classList.add('dark');
        } else {
            document.documentElement.classList.remove('dark');
        }

        const unsubscribe = pb.authStore.onChange((_token, record) => {
            role = record?.role || '';
        });

        void (async () => {
            if (pb.authStore.isValid) {
                try {
                    const auth = await pb.collection('users').authRefresh();
                    role = auth.record.role || '';
                } catch {
                    pb.authStore.clear();
                }
            }

            authReady = true;
            if (!isAuthFlow && (!pb.authStore.isValid || role !== 'Admin')) {
                goto('/login');
            }
        })();

        return unsubscribe;
    });

    function logout() {
        pb.authStore.clear();
        goto('/login');
    }

    function toggleTheme() {
        isDark = !isDark;
        if (isDark) {
            document.documentElement.classList.add('dark');
            localStorage.theme = 'dark';
        } else {
            document.documentElement.classList.remove('dark');
            localStorage.theme = 'light';
        }
    }
</script>

<div class="min-h-screen">
	{#if canViewManagement}
    <header class="bg-[#24292f] dark:bg-[#161b22] dark:border-b dark:border-[#30363d] text-white p-4 flex items-center justify-between">
        <div class="flex items-center space-x-4">
            <a href="/" class="font-bold text-lg hover:text-gray-300">MiStatic</a>
        </div>
        <div class="flex items-center space-x-4">
            <a href={`http://${APP_DOMAIN}:8090/_/`} target="_blank" class="text-sm text-gray-300 hover:text-white">Admin UI</a>
            {#if pb.authStore.isValid}
                <button onclick={logout} class="text-sm text-gray-300 hover:text-white">Logout</button>
            {/if}
            <button onclick={toggleTheme} class="text-gray-300 hover:text-white" title="Toggle dark mode">
                {#if isDark}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
                {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" /></svg>
                {/if}
            </button>
        </div>
    </header>
    {/if}

    <main class="max-w-5xl mx-auto p-6">
        {#if isAuthFlow || (authReady && canViewManagement)}
            {@render children?.()}
        {/if}
    </main>
</div>
