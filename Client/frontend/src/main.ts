import { createApp } from 'vue'
import './style.css'
import App from './App.vue'

const app = createApp(App)

async function loadWasm() {
	if (!('instantiateStreaming' in WebAssembly)) {
		console.error('Browser does not support WebAssembly.instantiateStreaming')
		return
	}
	const go = new Go();
	try {
		const result = await WebAssembly.instantiateStreaming(fetch("/umbra.wasm"), go.importObject)
		go.run(result.instance)
		console.log("Go-Wasm loaded successfully")
		return true
	} catch (err) {
		console.error('Error loading wasm:', err)
		return false
	}
}

loadWasm().then((success) => {
	if (success) {
		app.mount('#app')
	} else {
		alert("Failed to load WASM module")
	}
})