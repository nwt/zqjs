import './wasm_exec.js'

if (globalThis.global === undefined) {
    // for bridge.js
    globalThis.global = globalThis
}
const bridge = await import('./bridge.js')

const wasm = await fetch('./main.wasm')
const proxy = bridge.default(wasm.arrayBuffer())

export default proxy.zq
