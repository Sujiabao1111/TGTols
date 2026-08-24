const INVITE_CODE_STORAGE_KEY = "invite:code"
const INVITE_ACTIVITY_POPUP_PENDING_KEY = "invite:activity-popup-pending"

const normalizeInviteCode = (inviteCode?: string | null) => {
  const normalized = inviteCode?.trim()
  return normalized || undefined
}

export const saveInviteCode = (inviteCode?: string | null) => {
  const normalized = normalizeInviteCode(inviteCode)

  if (!normalized || typeof window === "undefined") {
    return
  }

  try {
    window.localStorage.setItem(INVITE_CODE_STORAGE_KEY, normalized)
  } catch (error) {
    console.error("Failed to save invite code:", error)
  }

  try {
    window.sessionStorage.setItem(INVITE_ACTIVITY_POPUP_PENDING_KEY, "1")
  } catch (error) {
    console.error("Failed to save invite popup flag:", error)
  }
}

export const getStoredInviteCode = () => {
  if (typeof window === "undefined") {
    return undefined
  }

  try {
    return normalizeInviteCode(window.localStorage.getItem(INVITE_CODE_STORAGE_KEY))
  } catch (error) {
    console.error("Failed to read invite code:", error)
    return undefined
  }
}

export const resolveInviteCode = (inviteCode?: string | null) => {
  return normalizeInviteCode(inviteCode) || getStoredInviteCode()
}

export const clearStoredInviteCode = () => {
  if (typeof window === "undefined") {
    return
  }

  try {
    window.localStorage.removeItem(INVITE_CODE_STORAGE_KEY)
  } catch (error) {
    console.error("Failed to clear invite code:", error)
  }
}

export const consumeInviteActivityPopupPending = () => {
  if (typeof window === "undefined") {
    return false
  }

  try {
    const isPending = window.sessionStorage.getItem(INVITE_ACTIVITY_POPUP_PENDING_KEY) === "1"
    window.sessionStorage.removeItem(INVITE_ACTIVITY_POPUP_PENDING_KEY)
    return isPending
  } catch (error) {
    console.error("Failed to consume invite popup flag:", error)
    return false
  }
}
