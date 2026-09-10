export function getSiteUrl(subdomain: string, rootDomain: string) {
    const isLocal = rootDomain.includes('.local') || rootDomain.includes('localhost');
    const protocol = isLocal ? 'http' : 'https';
    const port = isLocal ? ':8090' : '';
    return `${protocol}://${subdomain}.${rootDomain}${port}`;
}
