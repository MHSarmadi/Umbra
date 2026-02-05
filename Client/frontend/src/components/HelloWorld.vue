<script setup lang="ts">
import { ref } from 'vue'
import WasmWorker from '../workers/wasm-worker?worker'

defineProps<{ msg: string }>()

const worker = new WasmWorker()

const cipher = ref<Uint8Array | null>(null)
const salt = ref<Uint8Array | null>(null)

worker.onmessage = (event) => {
  if (event.data.type === 'encrypt') {
    if (event.data.success) {
      console.log('Encryption successful:', event.data.cipher, event.data.salt)
      cipher.value = new Uint8Array(event.data.cipher)
      salt.value = new Uint8Array(event.data.salt)
    } else {
      console.error('Encryption failed:', event.data.error)
    }
  }
}

function encryptData(key: Uint8Array, data: Uint8Array, context: string, difficulty: number) {
  worker.postMessage({
    type: 'encrypt',
    payload: {
      key,
      data,
      context,
      difficulty
    }
  })
}

const data = ref<string>("")
const key = ref<string>("")
const context = ref<string>("test-context")
const difficulty = ref<number>(5)

const base64url = (bytes: Uint8Array) => {
  const str = btoa(String.fromCharCode(...bytes))
  return str.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function encrypt() {
  const keyUint8 = new TextEncoder().encode(key.value)
  const dataUint8 = new TextEncoder().encode(data.value)
  encryptData(keyUint8, dataUint8, context.value, difficulty.value)
}

const count = ref(0)
</script>

<template>
  <h1>{{ msg }}</h1>

  <div class="card">
    <button type="button" @click="count++">count is {{ count }}</button>
    <p>
      Edit
      <code>components/HelloWorld.vue</code> to test HMR
    </p>
  </div>

  <div class="card">
    <h2>WASM Encryption Example</h2>
    <p><input v-model="key" placeholder="Enter key" /></p>
    <p><input v-model="data" placeholder="Enter data" /></p>
    <button type="button" @click="encrypt()">
      Encrypt Data
    </button>
    <p v-if="!cipher">Click the button to encrypt some data</p>
    <template v-else>
      <p>Ciphertext (base64url):</p>
      <pre>{{ base64url(cipher!) }}</pre>
      <p>Salt (base64url):</p>
      <pre>{{ base64url(salt!) }}</pre>
    </template>
  </div>

  <p>
    Check out
    <a href="https://vuejs.org/guide/quick-start.html#local" target="_blank"
      >create-vue</a
    >, the official Vue + Vite starter
  </p>
  <p>
    Learn more about IDE Support for Vue in the
    <a
      href="https://vuejs.org/guide/scaling-up/tooling.html#ide-support"
      target="_blank"
      >Vue Docs Scaling up Guide</a
    >.
  </p>
  <p class="read-the-docs">Click on the Vite and Vue logos to learn more</p>
  <h2>{{  }}</h2>
</template>

<style scoped>
.read-the-docs {
  color: #888;
}
</style>
