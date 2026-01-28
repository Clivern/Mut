<template>
  <div class="min-h-screen bg-theme-bg">
    <AppNav />

    <main class="w-full py-8 px-4 sm:px-6 lg:px-8 space-y-8">
      <div>
        <h1 class="text-3xl font-semibold text-theme-text">{{ $t('members_page.title') }}</h1>
        <p class="text-theme-textLight mt-2">{{ $t('members_page.subtitle') }}</p>
      </div>

      <div
        v-if="errorMessage"
        class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
      >
        {{ errorMessage }}
      </div>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-theme-border">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('members_page.invite_member') }}</h2>
          <p class="mt-1 text-sm text-theme-textLight">{{ $t('members_page.invite_desc') }}</p>
        </div>

        <form class="grid gap-4 p-6 md:grid-cols-[1fr_180px_auto]" @submit.prevent="handleInvite">
          <div>
            <label for="invite-email" class="form-label">{{ $t('members_page.email') }}</label>
            <input
              id="invite-email"
              v-model="inviteEmail"
              type="email"
              required
              maxlength="255"
              class="input-field"
              placeholder="user@example.com"
              :disabled="sendingInvite"
            >
          </div>
          <div>
            <label for="invite-role" class="form-label">{{ $t('members_page.role') }}</label>
            <select id="invite-role" v-model="inviteRole" class="input-field" :disabled="sendingInvite">
              <option v-for="role in assignableRoles" :key="role" :value="role">{{ roleLabel(role) }}</option>
            </select>
          </div>
          <div class="flex items-end">
            <button type="submit" class="btn-primary" :disabled="sendingInvite">
              {{ sendingInvite ? $t('members_page.sending') : $t('members_page.send_invite') }}
            </button>
          </div>
        </form>

        <div class="px-6 pb-6 space-y-6">
          <div
            v-if="inviteError"
            class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
          >
            {{ inviteError }}
          </div>

          <div v-if="memberError" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
            {{ memberError }}
          </div>

          <div v-if="members.length === 0" class="text-sm text-theme-textLight py-4">
            {{ $t('members_page.no_members') }}
          </div>
          <div v-else class="border border-theme-border rounded-lg overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-bg">
                <tr>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.name') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.email') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.role') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="member in members" :key="member.userId" class="hover:bg-theme-hover/50">
                  <td class="px-4 py-3 text-sm text-theme-text">{{ member.name }}</td>
                  <td class="px-4 py-3 text-sm text-theme-textLight">{{ member.email }}</td>
                  <td class="px-4 py-3 text-sm text-theme-text">
                    <select
                      v-if="member.role !== ROLE_OWNER"
                      :value="member.role"
                      class="input-field py-1"
                      :aria-label="$t('members_page.role')"
                      @change="handleRoleChange(member, $event.target.value)"
                    >
                      <option v-for="role in assignableRoles" :key="role" :value="role">{{ roleLabel(role) }}</option>
                    </select>
                    <span v-else>{{ roleLabel(member.role) }}</span>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <button
                      v-if="member.role !== ROLE_OWNER"
                      type="button"
                      class="text-sm text-red-600 hover:text-red-700 font-medium"
                      @click="removeModalMember = member"
                    >
                      {{ $t('members_page.remove') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <section class="rounded-xl border border-theme-border bg-white shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-theme-border">
          <h2 class="text-lg font-semibold text-theme-text">{{ $t('members_page.pending_invites') }}</h2>
        </div>

        <div class="p-6 space-y-6">
          <div
            v-if="cancelInviteError"
            class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
          >
            {{ cancelInviteError }}
          </div>
          <div v-if="invites.length === 0" class="text-sm text-theme-textLight text-center py-4">
            {{ $t('members_page.no_invites') }}
          </div>
          <div v-else class="border border-theme-border rounded-lg overflow-x-auto">
            <table class="min-w-full divide-y divide-theme-border">
              <thead class="bg-theme-bg">
                <tr>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.email') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.role') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-left text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.expires') }}</th>
                  <th scope="col" class="px-4 py-2.5 text-right text-xs font-medium text-theme-textLight uppercase tracking-wide">{{ $t('members_page.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-theme-border">
                <tr v-for="invite in invites" :key="invite.id" class="hover:bg-theme-hover/50">
                  <td class="px-4 py-3 text-sm text-theme-text">{{ invite.email }}</td>
                  <td class="px-4 py-3 text-sm text-theme-text">{{ roleLabel(invite.role) }}</td>
                  <td class="px-4 py-3 text-sm text-theme-textLight">{{ formatCreatedAt(invite.expiresAt, locale) }}</td>
                  <td class="px-4 py-3 text-right">
                    <button type="button" class="text-sm text-red-600 hover:text-red-700 font-medium" @click="cancelModalInvite = invite">
                      {{ $t('members_page.cancel_invite') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>
    </main>

    <ConfirmModal
      :open="!!cancelModalInvite"
      :title="$t('members_page.cancel_invite_title')"
      :message="$t('members_page.cancel_invite_confirm', { email: cancelModalInvite?.email })"
      :busy="cancellingInvite"
      :confirm-label="$t('members_page.cancel_invite')"
      :busy-label="$t('members_page.cancelling')"
      @cancel="cancelModalInvite = null"
      @confirm="confirmCancelInvite"
    />
    <ConfirmModal
      :open="!!removeModalMember"
      :title="$t('members_page.remove_title')"
      :message="$t('members_page.remove_confirm', { name: removeModalMember?.name })"
      :busy="removingMember"
      :confirm-label="$t('members_page.remove')"
      :busy-label="$t('members_page.removing')"
      @cancel="removeModalMember = null"
      @confirm="confirmRemoveMember"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppNav from '@/components/AppNav.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import { showFlash } from '@/lib/flash'
import { errorText } from '@/lib/error'
import { workspace_members_api, workspace_invite_api } from '@/api'
import { formatCreatedAt } from '@/utils/date'
import { loadWorkspaceFromStorage } from '@/utils/storage'

const ROLE_OWNER = 'owner'
const ROLES = ['admin', 'regular', 'readonly']

const { t, locale } = useI18n()

const workspaceId = computed(() => loadWorkspaceFromStorage()?.id)
const members = ref([])
const errorMessage = ref(null)

const memberError = ref(null)
const inviteEmail = ref('')
const inviteRole = ref('regular')
const inviteError = ref(null)
const sendingInvite = ref(false)
const invites = ref([])
const cancelModalInvite = ref(null)
const cancellingInvite = ref(false)
const cancelInviteError = ref(null)
const removeModalMember = ref(null)
const removingMember = ref(false)

const assignableRoles = ROLES

function roleLabel(role) {
  return t(`members_page.role_${role}`)
}

async function loadMembers() {
  const res = await workspace_members_api.list(workspaceId.value, { limit: 100, offset: 0 })
  members.value = res.data.members ?? []
}

async function loadInvites() {
  const res = await workspace_invite_api.list(workspaceId.value, { limit: 100, offset: 0 })
  invites.value = (res.data.invites ?? []).filter((invite) => invite.status === 'pending')
}

async function loadAll() {
  errorMessage.value = null

  try {
    await loadMembers()
    await loadInvites()
  } catch (err) {
    errorMessage.value = errorText(err, t('members_page.failed_load'))
  }
}

async function handleInvite() {
  sendingInvite.value = true
  inviteError.value = null
  try {
    await workspace_invite_api.create(workspaceId.value, {
      email: inviteEmail.value.trim(),
      role: inviteRole.value,
    })
    inviteEmail.value = ''
    inviteRole.value = 'regular'
    await loadInvites()
    showFlash(t('members_page.invite_sent'))
  } catch (err) {
    inviteError.value = errorText(err, t('members_page.failed_create_invite'))
  } finally {
    sendingInvite.value = false
  }
}

async function confirmCancelInvite() {
  const invite = cancelModalInvite.value
  if (!invite) return
  cancellingInvite.value = true
  cancelInviteError.value = null
  try {
    await workspace_invite_api.delete(workspaceId.value, invite.id)
    cancelModalInvite.value = null
    await loadInvites()
    showFlash(t('members_page.invite_cancelled'))
  } catch (err) {
    cancelInviteError.value = errorText(err, t('members_page.failed_cancel_invite'))
    cancelModalInvite.value = null
  } finally {
    cancellingInvite.value = false
  }
}

async function handleRoleChange(member, role) {
  memberError.value = null

  try {
    await workspace_members_api.updateRole(workspaceId.value, member.userId, { role })
    await loadMembers()
    showFlash(t('common.saved'))
  } catch (err) {
    memberError.value = errorText(err, t('members_page.failed_update_member'))
    await loadMembers()
  }
}

async function confirmRemoveMember() {
  const member = removeModalMember.value
  if (!member) return

  removingMember.value = true
  memberError.value = null

  try {
    await workspace_members_api.remove(workspaceId.value, member.userId)
    removeModalMember.value = null
    await loadMembers()
    showFlash(t('common.deleted'))
  } catch (err) {
    memberError.value = errorText(err, t('members_page.failed_remove_member'))
    removeModalMember.value = null
  } finally {
    removingMember.value = false
  }
}

onMounted(loadAll)
</script>
