const USER_STORAGE_KEY = 'mut_user'

export const loadUserFromStorage = () => {
  try {
    const stored = localStorage.getItem(USER_STORAGE_KEY)
    return stored ? JSON.parse(stored) : null
  } catch (err) {
    console.error('Failed to load user from localStorage:', err)
    return null
  }
}

export const saveUserToStorage = (userData) => {
  try {
    if (userData) {
      localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(userData))
    }
  } catch (err) {
    console.error('Failed to save user to localStorage:', err)
  }
}

export const removeUserFromStorage = () => {
  try {
    localStorage.removeItem(USER_STORAGE_KEY)
  } catch (err) {
    console.error('Failed to clear user from localStorage:', err)
  }
}
