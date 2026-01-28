import { computed } from 'vue'
import { loadUserFromStorage, loadWorkspaceFromStorage } from '@/utils/storage'

export const WORKSPACE_ROLE_READONLY = 'readonly'
export const WORKSPACE_ROLE_ADMIN = 'admin'
export const WORKSPACE_ROLE_OWNER = 'owner'
export const WORKSPACE_ROLE_REGULAR = 'regular'

export function canManageWorkspace(user, workspace) {
  if (!workspace?.id) return false
  return workspace.role === WORKSPACE_ROLE_ADMIN || workspace.role === WORKSPACE_ROLE_OWNER
}

export function useWorkspaceContext() {
  const currentUser = loadUserFromStorage()
  const currentWorkspace = loadWorkspaceFromStorage()
  const canManage = computed(() => canManageWorkspace(currentUser, currentWorkspace))

  return { currentUser, currentWorkspace, canManage }
}
