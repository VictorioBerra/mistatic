<script>
	import { pb } from '#lib/pb.js';
	import { page } from '$app/state';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function login() {
		loading = true;
		error = '';
		try {
			const auth = await pb.collection('users').authWithPassword(email, password);

			// Return to redirect param if exists, otherwise /
			const requestedRedirect = page.url.searchParams.get('redirect');
			if (auth.record.role === 'Reader' && !requestedRedirect) {
				pb.authStore.clear();
				error = 'Reader accounts can only sign in from a private site.';
				return;
			}
			const redirect = requestedRedirect || '/';
			window.location.href = redirect;
		} catch {
			error = 'Invalid email or password';
		} finally {
			loading = false;
		}
	}
</script>

<div class="mx-auto mt-20 max-w-md">
	<div class="card p-8">
		<div class="mb-6 text-center">
			<h1 class="text-2xl font-bold">Log in to MiStatic</h1>
			<p class="mt-1 text-sm text-gray-500">Manage your static sites securely</p>
		</div>

		{#if error}
			<div
				class="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600 dark:border-red-800 dark:bg-red-900/20 dark:text-red-400"
			>
				{error}
			</div>
		{/if}

		<form
			onsubmit={(e) => {
				e.preventDefault();
				login();
			}}
			class="space-y-4"
		>
			<div>
				<label for="email" class="mb-1 block text-sm font-medium">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="w-full rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none dark:border-[#30363d] dark:bg-[#0d1117]"
				/>
			</div>

			<div>
				<label for="password" class="mb-1 block text-sm font-medium">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="w-full rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none dark:border-[#30363d] dark:bg-[#0d1117]"
				/>
			</div>

			<div class="pt-2">
				<button type="submit" disabled={loading} class="btn-primary w-full py-2">
					{loading ? 'Logging in...' : 'Sign In'}
				</button>
			</div>
		</form>
	</div>
</div>
