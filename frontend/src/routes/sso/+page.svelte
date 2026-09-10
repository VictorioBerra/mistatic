<script>
    import { pb } from '$lib/pb';
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { goto } from '$app/navigation';

    let redirectUrl = $state('');
    let token = $state('');
    let formRef = $state();

    onMount(() => {
        redirectUrl = $page.url.searchParams.get('redirect') || '';
        
        if (!redirectUrl) {
            goto('/');
            return;
        }

        if (!pb.authStore.isValid) {
            // Send them to login but preserve the SSO redirect
            goto(`/login?redirect=/sso?redirect=${encodeURIComponent(redirectUrl)}`);
            return;
        }

        token = pb.authStore.token;
        
        // Auto submit the form in the next tick
        setTimeout(() => {
            if (formRef) formRef.submit();
        }, 50);
    });
</script>

<div class="flex items-center justify-center min-h-[50vh]">
    <div class="text-center">
        <svg class="animate-spin h-8 w-8 text-blue-500 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <h2 class="text-xl font-semibold">Authenticating...</h2>
        <p class="text-gray-500 mt-2">Securely logging you into {new URL(redirectUrl || 'http://localhost').hostname}</p>
    </div>
</div>

{#if token && redirectUrl}
    <form bind:this={formRef} method="POST" action={redirectUrl} class="hidden">
        <input type="hidden" name="token" value={token} />
    </form>
{/if}
