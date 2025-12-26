import { ref } from 'vue'

const message = ref(null)
const duration = ref(2500)
let timer = null

function show(msg, ms = 2500) {
  clearTimeout(timer)
  message.value = msg
  duration.value = ms
  timer = setTimeout(() => {
    message.value = null
  }, ms)
}

function clear() {
  clearTimeout(timer)
  message.value = null
}

export function useFlash() {
  return { message, duration, show, clear }
}

export const flashMessage = message
export const flashDuration = duration
export const showFlash = show
export const clearFlash = clear
