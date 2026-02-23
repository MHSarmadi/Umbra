export const encodeBase64 = (bytes: Uint8Array<ArrayBuffer>) => btoa(String.fromCharCode(...bytes));

export const decodeBase64 = (base64: string) => {
	return Uint8Array.from(atob(base64), c => c.charCodeAt(0));
};

export const decodeBufferIntoDate = (buffer: Uint8Array<ArrayBuffer>) => {
	const unixMillis = buffer.reduce((acc, byte) => (acc * 256) + byte, 0);
	return new Date(unixMillis);
}

export const decodeBase64IntoDate = (millisec_base64: string) => {
	const bytes = decodeBase64(millisec_base64);
	return decodeBufferIntoDate(bytes);
}