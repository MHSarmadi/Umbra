/// <reference types="vite/client" />

interface Go {
  new(): Go;
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): void;
}

declare const Go: {
  new(): Go;
}