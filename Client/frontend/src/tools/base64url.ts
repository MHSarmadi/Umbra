export const encodeBase64Url = (bytes: Uint8Array<ArrayBuffer>) =>
	btoa(String.fromCharCode(...bytes))
		.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

export const base64urlToBase64 = (base64url: string) => {
	const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
	const padded = base64.padEnd(base64.length + (4 - base64.length % 4) % 4, '=');
	return padded
}
export const decodeBase64Url = (base64url: string) => {
	const base64 = base64urlToBase64(base64url)
	return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
};