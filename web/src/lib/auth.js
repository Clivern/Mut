import { ref } from 'vue'
import { loadUserFromStorage, saveUserToStorage, removeUserFromStorage } from '@/utils/storage'

export const user = ref(loadUserFromStorage())

export function saveUser(userData) {
  user.value = userData
  saveUserToStorage(userData)
}

export function clearUser() {
  user.value = null
  removeUserFromStorage()
}
