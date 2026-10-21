<script>
	import { PUBLIC_APP_DOMAIN } from '$app/env/public';
	const APP_DOMAIN = PUBLIC_APP_DOMAIN || 'app.mistatic.local';
	import '../app.css';
	import { onMount } from 'svelte';
	import { pb } from '#lib/pb.js';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	let { children } = $props();
	let isDark = $state(false);
	let role = $state(pb.authStore.model?.role || '');
	let authReady = $state(false);
	let isAuthFlow = $derived(page.url.pathname === '/login' || page.url.pathname === '/sso');
	let canViewManagement = $derived(role === 'Admin');

	onMount(() => {
		if (
			localStorage.theme === 'dark' ||
			(!('theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)
		) {
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

<div class="flex min-h-screen flex-col">
	{#if canViewManagement}
		<header
			class="flex items-center justify-between bg-[#24292f] p-4 text-white dark:border-b dark:border-[#30363d] dark:bg-[#161b22]"
		>
			<div class="flex items-center space-x-4">
				<a href="/" class="text-lg font-bold hover:text-gray-300">MiStatic</a>
			</div>
			<div class="flex items-center space-x-4">
				<a
					href={`http://${APP_DOMAIN}:8090/_/`}
					target="_blank"
					class="text-sm text-gray-300 hover:text-white">Admin UI</a
				>
				{#if pb.authStore.isValid}
					<button onclick={logout} class="text-sm text-gray-300 hover:text-white">Logout</button>
				{/if}
				<button
					onclick={toggleTheme}
					class="text-gray-300 hover:text-white"
					title="Toggle dark mode"
				>
					{#if isDark}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
							></path>
						</svg>
					{:else}
						<svg
							xmlns="http://www.w3.org/2000/svg"
							class="h-5 w-5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
							></path>
						</svg>
					{/if}
				</button>
			</div>
		</header>
	{/if}

	<main class="mx-auto w-full max-w-5xl flex-1 p-6">
		{#if isAuthFlow || (authReady && canViewManagement)}
			{@render children?.()}
		{/if}
	</main>

	{#if canViewManagement || isAuthFlow}
		<footer
			class="mt-auto border-t border-[#d0d7de] py-4 text-xs text-gray-500 dark:border-[#30363d] dark:text-gray-400"
		>
			<div class="mx-auto flex max-w-5xl items-center justify-between px-6">
				<span class="font-medium">MiStatic</span>
				<a
					href="https://github.com/VictorioBerra/mistatic"
					target="_blank"
					rel="noopener noreferrer"
					class="inline-flex items-center space-x-1.5 text-gray-500 transition-colors hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200"
				>
					<svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
						<path
							fill-rule="evenodd"
							d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
							clip-rule="evenodd"
						/>
					</svg>
					<span>GitHub</span>
				</a>
			</div>
		</footer>
	{/if}
</div>
