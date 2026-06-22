'use client'

export type AuthMode = 'login' | 'register'

export interface AuthTransitionRect {
  top: number
  left: number
  width: number
  height: number
}

interface AuthTransitionState {
  rect: AuthTransitionRect
  mode: AuthMode
  createdAt: number
}

const AUTH_TRANSITION_KEY = 'keysWift.authTransition'
const AUTH_SUCCESS_KEY = 'keysWift.authSuccess'
const STATE_TTL = 3000

function readState(key: string): AuthTransitionState | null {
  if (typeof window === 'undefined') return null

  try {
    const raw = window.sessionStorage.getItem(key)
    if (!raw) return null

    const state = JSON.parse(raw) as AuthTransitionState
    if (!state?.rect || Date.now() - state.createdAt > STATE_TTL) {
      window.sessionStorage.removeItem(key)
      return null
    }

    return state
  } catch {
    window.sessionStorage.removeItem(key)
    return null
  }
}

function writeState(key: string, mode: AuthMode, rect: AuthTransitionRect) {
  if (typeof window === 'undefined') return

  const state: AuthTransitionState = {
    mode,
    rect,
    createdAt: Date.now(),
  }
  window.sessionStorage.setItem(key, JSON.stringify(state))
}

export function captureAuthTrigger(element: HTMLElement | null, mode: AuthMode) {
  if (!element || typeof window === 'undefined') return

  const rect = element.getBoundingClientRect()
  writeState(AUTH_TRANSITION_KEY, mode, {
    top: rect.top,
    left: rect.left,
    width: rect.width,
    height: rect.height,
  })
}

export function takeAuthTriggerState() {
  const state = readState(AUTH_TRANSITION_KEY)
  if (typeof window !== 'undefined') {
    window.sessionStorage.removeItem(AUTH_TRANSITION_KEY)
  }
  return state
}

export function saveAuthSuccessTarget(mode: AuthMode, rect: AuthTransitionRect) {
  writeState(AUTH_SUCCESS_KEY, mode, rect)
}

export function takeAuthSuccessTarget() {
  const state = readState(AUTH_SUCCESS_KEY)
  if (typeof window !== 'undefined') {
    window.sessionStorage.removeItem(AUTH_SUCCESS_KEY)
  }
  return state
}
