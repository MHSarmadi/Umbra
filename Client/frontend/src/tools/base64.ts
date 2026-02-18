export const encodeBase64 = (bytes: Uint8Array<ArrayBuffer>) => btoa(String.fromCharCode(...bytes));

export const decodeBase64 = (base64: string) => {
	return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
};