<script>
    import { pb } from '$lib/pb';
    import { goto } from '$app/navigation';
    import { page } from '$app/stores';

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
            const requestedRedirect = $page.url.searchParams.get('redirect');
            if (auth.record.role === 'Reader' && !requestedRedirect) {
                pb.authStore.clear();
                error = 'Reader accounts can only sign in from a private site.';
                return;
            }
            const redirect = requestedRedirect || '/';
            window.location.href = redirect;
        } catch (err) {
            error = 'Invalid email or password';
        } finally {
            loading = false;
        }
    }
</script>

<div class="max-w-md mx-auto mt-20">
    <div class="card p-8">
        <div class="text-center mb-6">
            <h1 class="text-2xl font-bold">Log in to MiStatic</h1>
            <p class="text-sm text-gray-500 mt-1">Manage your static sites securely</p>
        </div>

        {#if error}
            <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 px-4 py-3 rounded mb-4 text-sm">
                {error}
            </div>
        {/if}

        <form onsubmit={(e) => { e.preventDefault(); login(); }} class="space-y-4">
            <div>
                <label for="email" class="block text-sm font-medium mb-1">Email</label>
                <input id="email" type="email" bind:value={email} required class="w-full border border-gray-300 dark:border-[#30363d] dark:bg-[#0d1117] rounded px-3 py-2 focus:outline-none focus:border-blue-500" />
            </div>
            
            <div>
                <label for="password" class="block text-sm font-medium mb-1">Password</label>
                <input id="password" type="password" bind:value={password} required class="w-full border border-gray-300 dark:border-[#30363d] dark:bg-[#0d1117] rounded px-3 py-2 focus:outline-none focus:border-blue-500" />
            </div>

            <div class="pt-2">
                <button type="submit" disabled={loading} class="btn-primary w-full py-2">
                    {loading ? 'Logging in...' : 'Sign In'}
                </button>
            </div>
        </form>
    </div>
</div>
